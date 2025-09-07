package runtime

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/utils"
)

var _ memc.IVm = (*WazeroVm)(nil)
var _ memc.IWasmVmMeta = (*WazeroVmMeta)(nil)

type WazeroExport struct {
	name string
	fn   interface{}
}

func (f WazeroExport) Name() string {
	return f.name
}

func (f WazeroExport) Fn() interface{} {
	return f.fn
}

type WazeroImport struct {
	moduleName string
	name       string
	fn         interface{}
}

type WazeroMeta struct {
	imports []WazeroImport
	exports []WazeroExport
}

func (f WazeroMeta) ListImports() []memc.WasmImport {
	result := make([]memc.WasmImport, len(f.imports))
	for i, v := range f.imports {
		result[i] = v
	}
	return result
}

func (f WazeroMeta) ListExports() []memc.WasmExport {
	result := make([]memc.WasmExport, len(f.exports))
	for i, v := range f.exports {
		result[i] = v
	}
	return result
}

func (f WazeroImport) ModuleName() string {
	return f.moduleName
}

func (f WazeroImport) Name() string {
	return f.name
}

func (f WazeroImport) Fn() interface{} {
	return f.fn
}

type WazeroFn struct {
	name        string
	fn          func(context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error)
	inputTypes  []api.ValueType
	outputTypes []api.ValueType
	cost        int32
}

func (f WazeroFn) Name() string {
	return f.name
}

func (f WazeroFn) Cost() int32 {
	return f.cost
}

func (f WazeroFn) InputTypes() []interface{} {
	return FromValTypeSlice(f.inputTypes)
}

func (f WazeroFn) OutputTypes() []interface{} {
	return FromValTypeSlice(f.outputTypes)
}

func (f WazeroFn) Fn(context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	return f.fn(context, rnh, params)
}

func (f WazeroFn) WrappedFn(rnh memc.RuntimeHandler, _context interface{}, inputTypes []api.ValueType, outputTypes []api.ValueType) func(ctx context.Context, m api.Module, stack []uint64) {
	return func(ctx context.Context, m api.Module, stack []uint64) {
		vm := rnh.GetVm().(*WazeroVm)
		vm.vm = m
		params := make([]interface{}, len(inputTypes))
		for i, val := range stack[0:len(inputTypes)] {
			params[i] = ValueFromUint64(val, inputTypes[i])
		}
		results, err := f.fn(_context, rnh, params)
		if err != nil {
			if err.Error() == memc.VM_TERMINATE_ERROR {
				panic(err)
			}
			panic(err)
		}
		for i, res := range results {
			stack[i] = ValueToUint64(res, outputTypes[i])
		}
	}
}

type WazeroVm struct {
	ctx         context.Context
	cache       *WasmEngineCache
	vm          api.Module
	r           wazero.Runtime
	cleanups    []func()
	moduleNames []string
	aot         bool
	wasi        bool
	cfg         wazero.ModuleConfig
	args        []string
	envs        []string
	preopens    []string
	fileMap     map[string][]byte
}

type WasmEngineCache struct {
	ctx               context.Context
	InterpreterConfig wazero.RuntimeConfig
	CompiledConfig    wazero.RuntimeConfig

	// wazero has a compilation cache that we share per chain executions
	CompilationCache wazero.CompilationCache
	// we also cache compiled modules - this is what takes a lot of time
	// the deserializing compiled modules process is expensive
	// caching this takes from 4-5s block production to 0.6-0.8s
	// with tendermintp2p with timeouts 0
	CompiledModules     map[string]wazero.CompiledModule
	compiledModulesLock sync.Mutex

	deserializeLock sync.Mutex

	cleanups []func()
}

func (v *WasmEngineCache) GetInterpreterConfig() interface{} {
	return v.InterpreterConfig
}

func (v *WasmEngineCache) GetCompiledConfig() interface{} {
	return v.CompiledConfig
}

func (v *WasmEngineCache) GetCompilationCache() interface{} {
	return v.CompilationCache
}
func (v *WasmEngineCache) Close() {
	v.CompiledModules = map[string]wazero.CompiledModule{}
	v.CompilationCache.Close(v.ctx)
	// run in inverse order
	for i := len(v.cleanups) - 1; i >= 0; i-- {
		v.cleanups[i]()
	}
	fmt.Println("closed Wazero cache and compiled modules")
}

func (v *WasmEngineCache) GetCompiledMod(wasmFilePath string) wazero.CompiledModule {
	v.compiledModulesLock.Lock()
	defer v.compiledModulesLock.Unlock()
	mod, ok := v.CompiledModules[wasmFilePath]
	if ok && mod != nil {
		return mod
	}
	return nil
}
func (v *WasmEngineCache) SetCompiledMod(wasmFilePath string, mod wazero.CompiledModule) {
	v.compiledModulesLock.Lock()
	defer v.compiledModulesLock.Unlock()
	v.CompiledModules[wasmFilePath] = mod
}

func NewWazeroRuntime(parentCtx context.Context) *WasmEngineCache {
	cache := &WasmEngineCache{ctx: parentCtx}
	cache.CompilationCache = wazero.NewCompilationCache()
	cache.CompiledModules = map[string]wazero.CompiledModule{}

	cache.InterpreterConfig = wazero.NewRuntimeConfigInterpreter().
		WithCloseOnContextDone(true).                // for now, we let the execution finish in case we need to save block data in our core contracts
		WithCompilationCache(cache.CompilationCache) // .WithDebugInfoEnabled(true)
	cache.CompiledConfig = wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).                // for now, we let the execution finish in case we need to save block data in our core contracts
		WithCompilationCache(cache.CompilationCache) // .WithDebugInfoEnabled(true)

	return cache
}

func NewWazeroVm(cache *WasmEngineCache, ctx sdk.Context, aot bool) memc.IVm {
	var cleanups []func()
	var config wazero.RuntimeConfig

	if aot {
		config = cache.CompiledConfig
	} else {
		config = cache.InterpreterConfig
	}

	r := wazero.NewRuntimeWithConfig(ctx, config)
	cleanups = append(cleanups, func() {
		r.Close(ctx)
	})

	return &WazeroVm{
		ctx:      ctx,
		cache:    cache,
		r:        r,
		cleanups: cleanups,
		aot:      aot,
	}
}

func NewWazeroVmRaw(ctx sdk.Context, cache *WasmEngineCache, r wazero.Runtime, cleanups []func(), aot bool) *WazeroVm {
	return &WazeroVm{
		ctx:      ctx,
		cache:    cache,
		r:        r,
		cleanups: cleanups,
		aot:      aot,
	}
}

func (wm *WazeroVm) New(ctx sdk.Context, aot bool) memc.IVm {
	return NewWazeroVm(wm.cache, ctx, aot)
}

func (wm *WazeroVm) Cleanup() {
	// run in inverse order
	for i := len(wm.cleanups) - 1; i >= 0; i-- {
		wm.cleanups[i]()
	}
}

func (wm *WazeroVm) Call(funcname string, args []interface{}, gasMeter memc.GasMeter) ([]int32, error) {
	if wm.vm == nil {
		panic(fmt.Sprintf("Call %s: WazeroVm not instantiated", funcname))
	}
	_args := make([]uint64, len(args))
	for i, arg := range args {
		_args[i] = uint64(arg.(int32))
	}
	fn := wm.vm.ExportedFunction(funcname)
	if fn == nil {
		return nil, fmt.Errorf("WazeroVm: exported function not found: %s", funcname)
	}
	var wrappedMeter *GasMeter
	if gasMeter != nil {
		wrappedMeter = NewGasMeter(gasMeter.GasRemaining(), uint64(0), gasMeter)
		fn = fn.WithGasMeter(wrappedMeter)
	}
	result, err := fn.Call(wm.ctx, _args...)
	if gasMeter != nil {
		consumed := wrappedMeter.GasConsumed()
		if consumed > 0 {
			gasMeter.ConsumeGas(consumed, "wasm execution: "+funcname)
		}
	}
	if err != nil {
		expected := memc.VM_TERMINATE_ERROR + " (recovered by wazero)"
		if !strings.Contains(err.Error(), expected) {
			return nil, err
		}
	}
	_result := make([]int32, len(result))
	for i, res := range result {
		_result[i] = int32(res)
	}
	return _result, nil
}

func (wm *WazeroVm) GetMemory() (memc.IMemory, error) {
	if wm.vm == nil {
		panic("GetMemory: WazeroVm not instantiated")
	}
	mem := wm.vm.Memory()
	if mem == nil {
		return nil, fmt.Errorf("could not find memory")
	}
	return WazeroMemory{mem}, nil
}

func (wm *WazeroVm) GetFunctionList() []string {
	if wm.vm == nil {
		panic("GetFunctionList: WazeroVm not instantiated")
	}
	fnlist := []string{}
	for fnname := range wm.vm.ExportedFunctionDefinitions() {
		fnlist = append(fnlist, fnname)
	}
	return fnlist
}

func (wm *WazeroVm) ListRegisteredModule() []string {
	return wm.moduleNames
}

func (wm *WazeroVm) FindGlobal(name string) interface{} {
	if wm.vm == nil {
		panic(fmt.Sprintf("FindGlobal %s: WazeroVm not instantiated", name))
	}
	glob := wm.vm.ExportedGlobal(name)
	val := glob.Get()
	return ValueFromUint64(val, glob.Type())
}

func (wm *WazeroVm) ConfigWASI(config wazero.ModuleConfig, args []string, envs []string, preopens []string) wazero.ModuleConfig {
	// Add arguments
	for _, arg := range args {
		config = config.WithArgs(arg)
	}

	// Add environment variables
	for _, env := range envs {
		keyValue := splitEnv(env)
		if keyValue != nil {
			config = config.WithEnv(keyValue[0], keyValue[1])
		}
	}

	// Add preopened directories
	for _, dir := range preopens {
		config = config.WithFSConfig(wazero.NewFSConfig().WithDirMount(dir, "/"))
	}
	return config
}

func (wm *WazeroVm) InstantiateWasi(args []string, envs []string, preopens []string, fileMap map[string][]byte) {
	wm.args = args
	wm.envs = envs
	wm.preopens = preopens
	wm.fileMap = fileMap
}

func (wm *WazeroVm) WasiArgs() []string {
	return wm.args
}
func (wm *WazeroVm) WasiEnvs() []string {
	return wm.envs
}
func (wm *WazeroVm) WasiPreopens() []string {
	return wm.preopens
}
func (wm *WazeroVm) WasiFileMap() map[string][]byte {
	return wm.fileMap
}

func (wm *WazeroVm) InstantiateWasm(wasmFilePath string, aotFilePath string, wasmbuffer []byte) error {
	var err error
	cfg := wm.cfg
	if cfg == nil {
		cfg = wazero.NewModuleConfig()
	}
	// make sure wazero does not call any functions automatically
	cfg = cfg.WithStartFunctions([]string{}...)

	compilerSupported := wazero.CompilerSupported()
	if compilerSupported && aotFilePath != "" {
		compiledmod := wm.cache.GetCompiledMod(wasmFilePath)
		if compiledmod == nil {
			compiled, err := os.Open(aotFilePath)
			if err != nil {
				return sdkerrors.Wrapf(err, "load original wasm file failed %s", aotFilePath)
			}
			origwasmbuffer, err := os.ReadFile(wasmFilePath)
			if err != nil {
				return sdkerrors.Wrapf(err, "load wasm file failed %s", wasmFilePath)
			}
			// lock this, otherwise we get race conditions on the deserialization process
			// due to this function being used by both deterministic and non-deterministic executions on the same chain
			wm.cache.deserializeLock.Lock()
			// give the top-most context here (parentGoCtx)
			compiledmod, err = wm.r.DeserializeCompiledModule(wm.cache.ctx, origwasmbuffer, compiled)
			wm.cache.deserializeLock.Unlock()
			if err != nil {
				return sdkerrors.Wrapf(err, "module deserialization failed from buffer")
			}
			wm.cache.SetCompiledMod(wasmFilePath, compiledmod)
		}

		vm, err := wm.r.InstantiateModule(wm.ctx, compiledmod, cfg)
		if err != nil {
			wrappedErr := sdkerrors.Wrapf(err, "instantiate wasm failed for compiled module: %s", wasmFilePath)
			// we panic now, to make sure we know what happens here
			// ideally, it should never panic
			// if it panics, it is probably an issue with the cached compiled module
			// we may want to consider recompiling the module, but still, an error should not exist and it must be fixed first
			panic(wrappedErr)
			// return wrappedErr
		}
		wm.vm = vm
	} else {
		if wasmbuffer == nil {
			wasmbuffer, err = os.ReadFile(wasmFilePath)
			if err != nil {
				return sdkerrors.Wrapf(err, "load wasm file failed %s", wasmFilePath)
			}
		}
		vm, err := wm.r.InstantiateWithConfig(wm.ctx, wasmbuffer, cfg)
		if err != nil {
			return sdkerrors.Wrapf(err, "instantiate wasm failed for interpreted module")
		}
		wm.vm = vm
	}
	return nil
}

func (wm *WazeroVm) RegisterModule(mod interface{}) error {
	_mod, ok := mod.(wazero.HostModuleBuilder)
	if !ok {
		return fmt.Errorf("wasm module registration failed, invalid module interface")
	}
	return wm.RegisterModuleInner(_mod)
}

func (wm *WazeroVm) RegisterModuleInner(mod wazero.HostModuleBuilder) error {
	_mod, err := mod.Instantiate(wm.ctx)
	if err != nil {
		return err
	}
	wm.moduleNames = append(wm.moduleNames, _mod.Name())
	return nil
}

func (wm *WazeroVm) ValType_I32() interface{} {
	return api.ValueTypeI32
}

func (wm *WazeroVm) ValType_I64() interface{} {
	return api.ValueTypeI64
}

func (wm *WazeroVm) ValType_F64() interface{} {
	return api.ValueTypeF64
}

func (wm *WazeroVm) BuildFn(fnname string, fnval memc.IFnVal, inputTypes []interface{}, outputTypes []interface{}, cost int32) memc.IFn {
	newfn, err := wm.BuildFnInner(fnname, fnval, inputTypes, outputTypes, cost)
	if err != nil {
		panic(err)
	}
	return newfn
}

func (wm *WazeroVm) BuildFnInner(fnname string, fnval memc.IFnVal, inputTypes []interface{}, outputTypes []interface{}, cost int32) (memc.IFn, error) {
	input, err := toValTypeSlice(inputTypes)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "invalid wasm host function input")
	}
	output, err := toValTypeSlice(outputTypes)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "invalid wasm host function output")
	}
	return WazeroFn{
		name:        fnname,
		fn:          fnval,
		inputTypes:  input,
		outputTypes: output,
		cost:        cost,
	}, nil
}

func (wm *WazeroVm) BuildModule(rnh memc.RuntimeHandler, modname string, context interface{}, fndefs []memc.IFn) (interface{}, error) {
	_fndefs := make([]WazeroFn, len(fndefs))
	for i, fndef := range fndefs {
		_fndef, ok := fndef.(WazeroFn)
		if !ok {
			return nil, fmt.Errorf("wasm module build failed, invalid function interface")
		}
		_fndefs[i] = _fndef
	}
	mod := wm.BuildModuleInner(rnh, modname, context, _fndefs)
	return mod, nil
}

func (wm *WazeroVm) BuildModuleInner(rnh memc.RuntimeHandler, modname string, _context interface{}, fndefs []WazeroFn) wazero.HostModuleBuilder {
	envmod := wm.r.NewHostModuleBuilder(modname)
	// TODO cost for each function
	for _, fndef := range fndefs {
		envmod = envmod.NewFunctionBuilder().WithGoModuleFunction(
			api.GoModuleFunc(fndef.WrappedFn(rnh, _context, fndef.inputTypes, fndef.outputTypes)),
			fndef.inputTypes,
			fndef.outputTypes,
		).Export(fndef.name)
	}
	return envmod
}

// should we keep one global instance for caching compilations across all subchains?
// we now keep one per chain, so they don't bother one another with locks on deserializing the aot compiled module
type WazeroVmMeta struct {
	runtime     map[string]*WasmEngineCache
	runtimeLock sync.Mutex
}

func NewWazeroVmMeta() *WazeroVmMeta {
	return &WazeroVmMeta{runtime: map[string]*WasmEngineCache{}}
}

// When cgo is disabled at build time, this returns an error at runtime.
func (*WazeroVmMeta) LibVersion() string {
	return wazero.Version()
}

func (m *WazeroVmMeta) GetRuntime(chainId string) *WasmEngineCache {
	m.runtimeLock.Lock()
	defer m.runtimeLock.Unlock()
	r, ok := m.runtime[chainId]
	if ok {
		return r
	}
	return nil
}

func (m *WazeroVmMeta) MustGetRuntime(chainId string) *WasmEngineCache {
	r := m.GetRuntime(chainId)
	if r == nil {
		panic("wasm runtime cache not initialized: " + chainId)
	}
	return r
}

func (m *WazeroVmMeta) SetRuntime(chainId string, cache *WasmEngineCache) {
	m.runtimeLock.Lock()
	defer m.runtimeLock.Unlock()
	m.runtime[chainId] = cache
}

// this is a global context, across subchains
func (m *WazeroVmMeta) InitWasmRuntime(parentCtx context.Context, chainId string) {
	r := m.GetRuntime(chainId)
	if r == nil {
		r = NewWazeroRuntime(parentCtx)
		m.SetRuntime(chainId, r)
	}
}

func (m *WazeroVmMeta) Close(chainId string) {
	r := m.GetRuntime(chainId)
	if r != nil {
		r.Close()
	}
}

// entrypoint used by wasmX getRuntimeHandler
func (m *WazeroVmMeta) NewWasmVm(ctx sdk.Context, aot bool) memc.IVm {
	return NewWazeroVm(m.MustGetRuntime(ctx.ChainID()), ctx, aot)
}

func (*WazeroVmMeta) AnalyzeWasm(ctx sdk.Context, wasmbuffer []byte) (memc.WasmMeta, error) {
	config := wazero.NewRuntimeConfigInterpreter()
	r := wazero.NewRuntimeWithConfig(ctx, config)
	defer func() {
		r.Close(ctx)
	}()
	cmod, err := r.CompileModule(ctx, wasmbuffer)
	if err != nil {
		return nil, err
	}
	defer func() {
		cmod.Close(ctx)
	}()

	imports := cmod.ImportedFunctions()
	exports := cmod.ExportedFunctions()
	meta := &WazeroMeta{}
	meta.imports = make([]WazeroImport, len(imports))
	meta.exports = make([]WazeroExport, len(exports))

	for i, mimport := range imports {
		moduleName, name, _ := mimport.Import()
		if moduleName == "" {
			moduleName = mimport.ModuleName()
		}
		if name == "" {
			name = mimport.Name()
		}
		meta.imports[i] = WazeroImport{
			moduleName: moduleName,
			name:       name,
			fn:         mimport.GoFunction(),
		}
	}
	i := 0
	for name, mexport := range exports {
		if name == "" {
			name = mexport.Name()
		}
		if name == "" {
			names := mexport.ExportNames()
			if len(names) > 0 {
				name = names[0]
			}
		}
		meta.exports[i] = WazeroExport{
			name: name,
			fn:   mexport.GoFunction(),
		}
		i++
	}
	return meta, nil
}

func (*WazeroVmMeta) AotCompile(ctx sdk.Context, inPath string, outPath string, meteringOff bool) error {
	if !wazero.CompilerSupported() {
		return nil
	}
	config := wazero.NewRuntimeConfigCompiler()
	r := wazero.NewRuntimeWithConfig(ctx, config)

	wasmbuffer, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}
	_, reader, err := r.CompileModuleAndSerialize(ctx, wasmbuffer, !meteringOff)
	if err != nil {
		return err
	}
	err = utils.SafeWriteReader(outPath, reader)
	if err != nil {
		return err
	}
	return nil
}

func toValTypeSlice(input []interface{}) ([]api.ValueType, error) {
	result := make([]api.ValueType, len(input))
	for i, v := range input {
		val, ok := v.(api.ValueType)
		if !ok {
			return nil, fmt.Errorf("value at index %d is not an api.ValueType", i)
		}
		result[i] = val
	}
	return result, nil
}

func FromValTypeSlice(input []api.ValueType) []interface{} {
	result := make([]interface{}, len(input))
	for i, v := range input {
		result[i] = v
	}
	return result
}

// Helper function to split "KEY=VALUE" into [KEY, VALUE]
func splitEnv(env string) []string {
	for i, c := range env {
		if c == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return nil
}

func ValueFromUint64(val uint64, t api.ValueType) interface{} {
	switch t {
	case api.ValueTypeI32:
		return api.DecodeI32(val)
	case api.ValueTypeI64:
		return int64(val)
	case api.ValueTypeF32:
		return api.DecodeF32(val)
	case api.ValueTypeF64:
		return api.DecodeF64(val)
	default:
		return val
	}
}

func ValueToUint64(val interface{}, t api.ValueType) uint64 {
	switch t {
	case api.ValueTypeI32:
		return api.EncodeI32(val.(int32))
	case api.ValueTypeI64:
		return api.EncodeI64(val.(int64))
	case api.ValueTypeF32:
		return api.EncodeF32(val.(float32))
	case api.ValueTypeF64:
		return api.EncodeF64(val.(float64))
	default:
		return val.(uint64)
	}
}
