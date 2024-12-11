package wazero

import (
	"context"
	"fmt"
	"os"
	"strings"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"mythos/v1/x/wasmx/types"
	memc "mythos/v1/x/wasmx/vm/memory/common"
	"mythos/v1/x/wasmx/vm/utils"
)

var CONTEXT_CACHE_KEY = "wazero_cache"

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
	cache       wazero.CompilationCache
	vm          api.Module
	r           wazero.Runtime
	cleanups    []func()
	moduleNames []string
}

// TODO clean cache!
func NewWazeroVm(ctx sdk.Context) memc.IVm {
	cache, ok := ctx.Value(CONTEXT_CACHE_KEY).(wazero.CompilationCache)
	if !ok {
		cache = wazero.NewCompilationCache()
		ctx = ctx.WithValue(CONTEXT_CACHE_KEY, cache)
	}
	var cleanups []func()
	// TODO WASI
	// NewRuntimeConfigCompiler
	// NewRuntimeConfig
	// .WithDebugInfoEnabled(true)
	// TODO check if compiler is suppported
	config := wazero.NewRuntimeConfigCompiler().
		WithCloseOnContextDone(false). // for now, we let the execution finish in case we need to save block data in our core contracts
		WithCompilationCache(cache)    // .WithDebugInfoEnabled(true)

	r := wazero.NewRuntimeWithConfig(ctx, config)
	cleanups = append(cleanups, func() {
		r.Close(ctx)
	})

	return &WazeroVm{
		ctx:      ctx,
		cache:    cache,
		r:        r,
		cleanups: cleanups,
	}
}

func (wm *WazeroVm) New(ctx sdk.Context) memc.IVm {
	return NewWazeroVm(ctx)
}

func (wm *WazeroVm) Cleanup() {
	// run in inverse order
	for i := len(wm.cleanups) - 1; i >= 0; i-- {
		wm.cleanups[i]()
	}
}

func (wm *WazeroVm) Call(funcname string, args []interface{}) ([]int32, error) {
	_args := make([]uint64, len(args))
	for i, arg := range args {
		_args[i] = uint64(arg.(int32))
	}
	result, err := wm.vm.ExportedFunction(funcname).Call(wm.ctx, _args...)
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
	mem := wm.vm.Memory()
	if mem == nil {
		return nil, fmt.Errorf("could not find memory")
	}
	return WazeroMemory{mem}, nil
}

func (wm *WazeroVm) GetFunctionList() []string {
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
	glob := wm.vm.ExportedGlobal(name)
	val := glob.Get()
	return ValueFromUint64(val, glob.Type())
}

func (wm *WazeroVm) InitWasi(args []string, envs []string, preopens []string) error {
	// modcfg := wazero.NewModuleConfig().WithArgs(args).WithEnv(envs).WithFSConfig()
	// wm.r.InstantiateWithConfig(ctx, wasmbuffer, modcfg)

	// WithWorkDirFS

	// mod := wm.vm.GetImportModule(WASI)
	// if mod == nil {
	// 	return fmt.Errorf("WASI module not found")
	// }
	// mod.InitWasi(args, envs, preopens)
	return nil
}

func (wm *WazeroVm) InstantiateWasm(filePath string, wasmbuffer []byte) error {
	var err error
	if strings.Contains(filePath, types.PINNED_FOLDER) {
		content, err := os.Open(filePath)
		if err != nil {
			return sdkerrors.Wrapf(err, "load original wasm file failed %s", filePath)
		}

		// TODO better - just provide the original wasm & the pinned file
		originalWasmPath := strings.Replace(filePath, fmt.Sprintf("/%s/", types.PINNED_FOLDER), "/", 1)
		origwasmbuffer, err := os.ReadFile(originalWasmPath)
		if err != nil {
			return sdkerrors.Wrapf(err, "load wasm file failed %s", originalWasmPath)
		}
		compiledmod, err := wm.r.DeserializeCompiledModule(wm.ctx, origwasmbuffer, content)
		if err != nil {
			return sdkerrors.Wrapf(err, "module deserialization failed from buffer")
		}
		vm, err := wm.r.InstantiateModule(wm.ctx, compiledmod, wazero.NewModuleConfig())
		if err != nil {
			return sdkerrors.Wrapf(err, "load wasm file failed from buffer")
		}
		wm.vm = vm
	} else {
		if wasmbuffer == nil {
			wasmbuffer, err = os.ReadFile(filePath)
			if err != nil {
				return sdkerrors.Wrapf(err, "load wasm file failed %s", filePath)
			}
		}
		vm, err := wm.r.Instantiate(wm.ctx, wasmbuffer)
		if err != nil {
			return sdkerrors.Wrapf(err, "load wasm file failed from buffer")
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

func (wm *WazeroVm) InitWASI(args []string, envs []string, preopens []string) error {
	wasiConfig := wazero.NewModuleConfig()
	// Add arguments
	for _, arg := range args {
		wasiConfig = wasiConfig.WithArgs(arg)
	}

	// Add environment variables
	for _, env := range envs {
		keyValue := splitEnv(env)
		if keyValue != nil {
			wasiConfig = wasiConfig.WithEnv(keyValue[0], keyValue[1])
		}
	}

	// Add preopened directories
	for _, dir := range preopens {
		wasiConfig = wasiConfig.WithFSConfig(wazero.NewFSConfig().WithDirMount(dir, "/"))
	}

	// Instantiate the WASI module
	_, err := wasi_snapshot_preview1.Instantiate(wm.ctx, wm.r)
	if err != nil {
		return err
	}
	wm.moduleNames = append(wm.moduleNames, wasi_snapshot_preview1.ModuleName)

	// TODO WASI
	// // InstantiateModule runs the "_start" function, WASI's "main".
	// // * Set the program name (arg[0]) to "wasi"; arg[1] should be "/test.txt".
	// if _, err = r.InstantiateWithConfig(ctx, catWasm, config.WithArgs("wasi", os.Args[1])); err != nil {
	// 	// Note: Most compilers do not exit the module after running "_start",
	// 	// unless there was an error. This allows you to call exported functions.
	// 	if exitErr, ok := err.(*sys.ExitError); ok && exitErr.ExitCode() != 0 {
	// 		fmt.Fprintf(os.Stderr, "exit_code: %d\n", exitErr.ExitCode())
	// 	} else if !ok {
	// 		log.Panicln(err)
	// 	}
	// }

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

type WazeroVmMeta struct{}

func (WazeroVmMeta) NewWasmVm(ctx sdk.Context) memc.IVm {
	return NewWazeroVm(ctx)
}

func (WazeroVmMeta) AnalyzeWasm(ctx sdk.Context, wasmbuffer []byte) (memc.WasmMeta, error) {
	config := wazero.NewRuntimeConfigInterpreter()
	r := wazero.NewRuntimeWithConfig(ctx, config)
	cmod, err := r.CompileModule(ctx, wasmbuffer)
	if err != nil {
		return nil, err
	}

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

func (WazeroVmMeta) AotCompile(ctx sdk.Context, inPath string, outPath string) error {
	config := wazero.NewRuntimeConfigCompiler()
	r := wazero.NewRuntimeWithConfig(ctx, config)

	wasmbuffer, err := os.ReadFile(inPath)
	if err != nil {
		return err
	}

	_, reader, err := r.CompileModuleAndSerialize(ctx, wasmbuffer)
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
