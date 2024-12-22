package runtime

import (
	"fmt"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/second-state/WasmEdge-go/wasmedge"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

var _ memc.IVm = (*WasmEdgeVm)(nil)
var _ memc.IWasmVmMeta = (*WasmEdgeVmMeta)(nil)

type WasmEdgeExport struct {
	name string
	// inputTypes  []wasmedge.ValType
	// outputTypes []wasmedge.ValType
	fn interface{}
}

func (f WasmEdgeExport) Name() string {
	return f.name
}

// func (f WasmEdgeExport) InputTypes() []interface{} {
// 	return FromValTypeSlice(f.inputTypes)
// }

// func (f WasmEdgeExport) OutputTypes() []interface{} {
// 	return FromValTypeSlice(f.outputTypes)
// }

func (f WasmEdgeExport) Fn() interface{} {
	return f.fn
}

type WasmEdgeImport struct {
	moduleName string
	name       string
	// inputTypes  []wasmedge.ValType
	// outputTypes []wasmedge.ValType
	fn interface{}
}

func (f WasmEdgeImport) ModuleName() string {
	return f.moduleName
}

func (f WasmEdgeImport) Name() string {
	return f.name
}

// func (f WasmEdgeImport) InputTypes() []interface{} {
// 	return FromValTypeSlice(f.inputTypes)
// }

// func (f WasmEdgeImport) OutputTypes() []interface{} {
// 	return FromValTypeSlice(f.outputTypes)
// }

func (f WasmEdgeImport) Fn() interface{} {
	return f.fn
}

type WasmEdgeMeta struct {
	imports []WasmEdgeImport
	exports []WasmEdgeExport
}

func (f WasmEdgeMeta) ListImports() []memc.WasmImport {
	result := make([]memc.WasmImport, len(f.imports))
	for i, v := range f.imports {
		result[i] = v
	}
	return result
}

func (f WasmEdgeMeta) ListExports() []memc.WasmExport {
	result := make([]memc.WasmExport, len(f.exports))
	for i, v := range f.exports {
		result[i] = v
	}
	return result
}

type WasmEdgeFn struct {
	name        string
	fn          func(context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error)
	inputTypes  []wasmedge.ValType
	outputTypes []wasmedge.ValType
	cost        int32
}

func (f WasmEdgeFn) Name() string {
	return f.name
}

func (f WasmEdgeFn) Cost() int32 {
	return f.cost
}

func (f WasmEdgeFn) InputTypes() []interface{} {
	return FromValTypeSlice(f.inputTypes)
}

func (f WasmEdgeFn) OutputTypes() []interface{} {
	return FromValTypeSlice(f.outputTypes)
}

func (f WasmEdgeFn) Fn(context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	return f.fn(context, rnh, params)
}

func (f WasmEdgeFn) WrappedFn(rnh memc.RuntimeHandler) func(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	return func(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
		vm := rnh.GetVm().(*WasmEdgeVm)
		vm.callframe = callframe
		res, err := f.fn(context, rnh, params)
		// 0, 1, 2 are used by wasmedge for success, terminate, fail
		if err != nil {
			if err.Error() == memc.VM_TERMINATE_ERROR {
				return res, wasmedge.Result_Terminate
			}
			return res, wasmedge.Result_Fail
		}
		// TODO terminate
		return res, wasmedge.Result_Success
	}
}

type WasmEdgeVm struct {
	vm        *wasmedge.VM
	callframe *wasmedge.CallingFrame
	cleanups  []func()
}

func (wm *WasmEdgeVm) New(ctx sdk.Context, aot bool) memc.IVm {
	return NewWasmEdgeVm(ctx, aot)
}

func (wm *WasmEdgeVm) Cleanup() {
	// run in inverse order
	for i := len(wm.cleanups) - 1; i >= 0; i-- {
		wm.cleanups[i]()
	}
}

func (wm *WasmEdgeVm) Call(funcname string, args []interface{}, gasMeter memc.GasMeter) ([]int32, error) {
	result, err := wm.vm.Execute(funcname, args...)
	if err != nil {
		return nil, err
	}
	return memc.ToInt32Slice(result)
}

func (wm *WasmEdgeVm) GetMemory() (memc.IMemory, error) {
	var mem *wasmedge.Memory
	if wm.callframe != nil {
		mem = wm.callframe.GetMemoryByIndex(0)
		if mem == nil {
			return nil, fmt.Errorf("could not find memory")
		}
	}
	if wm.vm == nil {
		return nil, fmt.Errorf("could not find wasm vm")
	}
	mod := wm.vm.GetActiveModule()
	if mod == nil {
		return nil, fmt.Errorf("could not find vm active module")
	}
	mem = mod.FindMemory("memory")
	return WasmEdgeMemory{mem}, nil
}

func (wm *WasmEdgeVm) GetFunctionList() []string {
	fnlist, _ := wm.vm.GetFunctionList()
	return fnlist
}

// func (wm *WasmEdgeVm) ListGlobals() []string {
// 	return wm.vm.GetActiveModule().ListGlobal()
// }

func (wm *WasmEdgeVm) FindGlobal(name string) interface{} {
	glob := wm.vm.GetActiveModule().FindGlobal(name)
	return glob.GetValue()
}

func (wm *WasmEdgeVm) ListRegisteredModule() []string {
	return wm.vm.ListRegisteredModule()
}

func (wm *WasmEdgeVm) InitWasi(args []string, envs []string, preopens []string) error {
	mod := wm.vm.GetImportModule(wasmedge.WASI)
	if mod == nil {
		return fmt.Errorf("WASI module not found")
	}
	mod.InitWasi(args, envs, preopens)
	return nil
}

func (wm *WasmEdgeVm) InstantiateWasm(wasmFilePath string, aotFilePath string, wasmbuffer []byte) error {
	var err error
	filePath := wasmFilePath
	if wasmFilePath == "" {
		filePath = aotFilePath
	}
	if wasmbuffer == nil {
		err = wm.vm.LoadWasmFile(filePath)
		if err != nil {
			return sdkerrors.Wrapf(err, "load wasm file failed %s", filePath)
		}
	} else {
		err = wm.vm.LoadWasmBuffer(wasmbuffer)
		if err != nil {
			return sdkerrors.Wrapf(err, "load wasm file failed from buffer")
		}
	}
	err = wm.vm.Validate()
	if err != nil {
		return sdkerrors.Wrapf(err, "wasm module VM validate failed")
	}
	err = wm.vm.Instantiate()
	if err != nil {
		return sdkerrors.Wrapf(err, "wasm module VM instantiate failed")
	}
	return nil
}

func (wm *WasmEdgeVm) RegisterModule(mod interface{}) error {
	_mod, ok := mod.(*wasmedge.Module)
	if !ok {
		return fmt.Errorf("wasm module registration failed, invalid module interface")
	}
	return wm.RegisterModuleInner(_mod)
}

func (wm *WasmEdgeVm) RegisterModuleInner(mod *wasmedge.Module) error {
	return wm.vm.RegisterModule(mod)
}

func (wm *WasmEdgeVm) ValType_I32() interface{} {
	return wasmedge.ValType_I32
}

func (wm *WasmEdgeVm) ValType_I64() interface{} {
	return wasmedge.ValType_I64
}

func (wm *WasmEdgeVm) ValType_F64() interface{} {
	return wasmedge.ValType_F64
}

func (wm *WasmEdgeVm) BuildFn(fnname string, fnval memc.IFnVal, inputTypes []interface{}, outputTypes []interface{}, cost int32) memc.IFn {
	newfn, err := wm.BuildFnInner(fnname, fnval, inputTypes, outputTypes, cost)
	if err != nil {
		panic(err)
	}
	return newfn
}

func (wm *WasmEdgeVm) BuildFnInner(fnname string, fnval memc.IFnVal, inputTypes []interface{}, outputTypes []interface{}, cost int32) (memc.IFn, error) {
	input, err := toValTypeSlice(inputTypes)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "invalid wasm host function input")
	}
	output, err := toValTypeSlice(outputTypes)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "invalid wasm host function output")
	}
	return WasmEdgeFn{
		name:        fnname,
		fn:          fnval,
		inputTypes:  input,
		outputTypes: output,
		cost:        cost,
	}, nil
}

func (wm *WasmEdgeVm) BuildModule(rnh memc.RuntimeHandler, modname string, context interface{}, fndefs []memc.IFn) (interface{}, error) {
	_fndefs := make([]WasmEdgeFn, len(fndefs))
	for i, fndef := range fndefs {
		_fndef, ok := fndef.(WasmEdgeFn)
		if !ok {
			return nil, fmt.Errorf("wasm module build failed, invalid function interface")
		}
		_fndefs[i] = _fndef
	}
	mod := wm.BuildModuleInner(rnh, modname, context, _fndefs)
	return mod, nil
}

func (wm *WasmEdgeVm) BuildModuleInner(rnh memc.RuntimeHandler, modname string, context interface{}, fndefs []WasmEdgeFn) *wasmedge.Module {
	envmod := wasmedge.NewModule(modname)
	wm.cleanups = append(wm.cleanups, envmod.Release)
	for _, fndef := range fndefs {
		envmod.AddFunction(fndef.name, wasmedge.NewFunction(wasmedge.NewFunctionType(fndef.inputTypes, fndef.outputTypes), fndef.WrappedFn(rnh), context, uint(fndef.cost)))
	}
	return envmod
}

type WasmEdgeVmMeta struct{}

// When cgo is disabled at build time, this returns an error at runtime.
func (WasmEdgeVmMeta) LibVersion() string {
	return wasmedge.GetVersion()
}

func (WasmEdgeVmMeta) NewWasmVm(ctx sdk.Context, aot bool) memc.IVm {
	return NewWasmEdgeVm(ctx, aot)
}

func (WasmEdgeVmMeta) AnalyzeWasm(_ sdk.Context, wasmbuffer []byte) (memc.WasmMeta, error) {
	loader := wasmedge.NewLoader()
	defer func() {
		loader.Release()
	}()
	ast, err := loader.LoadBuffer(wasmbuffer)
	if err != nil {
		return nil, err
	}
	defer func() {
		ast.Release()
	}()
	imports := ast.ListImports()
	exports := ast.ListExports()
	meta := &WasmEdgeMeta{}
	meta.imports = make([]WasmEdgeImport, len(imports))
	meta.exports = make([]WasmEdgeExport, len(exports))

	for i, mimport := range imports {
		// mimport.GetExternalType()
		meta.imports[i] = WasmEdgeImport{
			moduleName: mimport.GetModuleName(),
			name:       mimport.GetExternalName(),
			fn:         mimport.GetExternalValue(),
		}
	}
	for i, mexport := range exports {
		meta.exports[i] = WasmEdgeExport{
			name: mexport.GetExternalName(),
			fn:   mexport.GetExternalValue(),
		}
	}
	return meta, nil
}

func (WasmEdgeVmMeta) AotCompile(_ sdk.Context, inPath string, outPath string, meteringOff bool) error {
	// Create Configure
	// conf := wasmedge.NewConfigure(wasmedge.THREADS, wasmedge.EXTENDED_CONST, wasmedge.TAIL_CALL, wasmedge.MULTI_MEMORIES)

	// Create Compiler
	// compiler := wasmedge.NewCompilerWithConfig(conf)
	compiler := wasmedge.NewCompiler()
	defer func() {
		compiler.Release()
		// conf.Release()
	}()

	// Compile WASM AOT
	err := compiler.Compile(inPath, outPath)
	if err != nil {
		fmt.Println("Go: Compile WASM to AOT mode Failed!!")
		return err
	}
	return nil
}

func NewWasmEdgeVm(_ sdk.Context, _ bool) memc.IVm {
	var cleanups []func()

	// wasmedge.SetLogOff()
	wasmedge.SetLogErrorLevel()
	// wasmedge.SetLogDebugLevel()

	conf := wasmedge.NewConfigure()
	cleanups = append(cleanups, conf.Release)

	// conf.SetStatisticsInstructionCounting(true)
	// conf.SetStatisticsTimeMeasuring(true)
	// TODO allow wasi only for core contracts
	conf.AddConfig(wasmedge.WASI)
	contractVm := wasmedge.NewVMWithConfig(conf)
	// contractVm := wasmedge.NewVM()

	// first in, last cleaned up
	cleanups = append(cleanups, conf.Release)
	cleanups = append(cleanups, contractVm.Release)

	return &WasmEdgeVm{
		vm:       contractVm,
		cleanups: cleanups,
	}
}

func toValTypeSlice(input []interface{}) ([]wasmedge.ValType, error) {
	result := make([]wasmedge.ValType, len(input))
	for i, v := range input {
		val, ok := v.(wasmedge.ValType)
		if !ok {
			return nil, fmt.Errorf("value at index %d is not an wasmedge.ValType", i)
		}
		result[i] = val
	}
	return result, nil
}

func FromValTypeSlice(input []wasmedge.ValType) []interface{} {
	result := make([]interface{}, len(input))
	for i, v := range input {
		result[i] = v
	}
	return result
}

// Returns the hex address of the interpreter if exists or the version string
func parseDependencyOrHexAddr(contractVersion string, part string) string {
	dep := contractVersion
	if strings.Contains(contractVersion, part) {
		v := contractVersion[len(part):]
		if len(v) > 2 && v[0:2] == "0x" {
			dep = v
		}
	}
	return dep
}
