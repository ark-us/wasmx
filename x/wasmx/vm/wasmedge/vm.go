package wasmedge

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/second-state/WasmEdge-go/wasmedge"

	memc "mythos/v1/x/wasmx/vm/memory/common"
)

var _ memc.IVm = (*WasmEdgeVm)(nil)

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
		if err != nil {
			return res, wasmedge.Result_Fail
		}
		return res, wasmedge.Result_Success
	}
}

type WasmEdgeVm struct {
	vm        *wasmedge.VM
	callframe *wasmedge.CallingFrame
	cleanups  []func()
}

func NewWasmEdgeVm(_ sdk.Context) memc.IVm {
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

func (wm *WasmEdgeVm) New(ctx sdk.Context) memc.IVm {
	return NewWasmEdgeVm(ctx)
}

func (wm *WasmEdgeVm) Cleanup() {
	for _, fn := range wm.cleanups {
		fn()
	}
}

func (wm *WasmEdgeVm) Call(funcname string, args []interface{}) ([]int32, error) {
	result, err := wm.vm.Execute(funcname, args...)
	if err != nil {
		return nil, err
	}
	return memc.ToInt32Slice(result)
}

func (wm *WasmEdgeVm) GetMemory() (memc.IMemory, error) {
	mem := wm.callframe.GetMemoryByIndex(0)
	if mem == nil {
		return nil, fmt.Errorf("could not find memory")
	}
	return WasmEdgeMemory{mem}, nil
}

func (wm *WasmEdgeVm) InstantiateWasm(filePath string, wasmbuffer []byte) error {
	var err error
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
	for _, fndef := range fndefs {
		envmod.AddFunction(fndef.name, wasmedge.NewFunction(wasmedge.NewFunctionType(fndef.inputTypes, fndef.outputTypes), fndef.WrappedFn(rnh), context, uint(fndef.cost)))
	}
	return envmod
}

func (wm *WasmEdgeVm) VerifyEnv(version string, imports []*wasmedge.ImportType) error {
	// TODO check that all imports are supported by the given version

	// for _, mimport := range imports {
	// 	fmt.Println("Import:", mimport.GetModuleName(), mimport.GetExternalName())
	// }
	return nil
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
