package wazero

import (
	"context"
	"fmt"
	"os"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	memc "mythos/v1/x/wasmx/vm/memory/common"
)

var CONTEXT_CACHE_KEY = "wazero_cache"

var _ memc.IVm = (*WazeroVm)(nil)

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

func (f WazeroFn) WrappedFn(rnh memc.RuntimeHandler) func(ctx context.Context, m api.Module, stack []uint64) {
	return func(ctx context.Context, m api.Module, stack []uint64) {
		vm := rnh.GetVm().(*WazeroVm)
		vm.vm = m
		params := make([]interface{}, len(stack))
		for i, val := range stack {
			params[i] = val
		}
		results, err := f.fn(ctx, rnh, params)
		if err != nil {
			panic(err)
		}
		for i, res := range results {
			stack[i] = res.(uint64)
		}
	}
}

type WazeroVm struct {
	ctx      context.Context
	cache    wazero.CompilationCache
	vm       api.Module
	r        wazero.Runtime
	cleanups []func()
}

// TODO clean cache!
func NewWazeroVm(ctx sdk.Context) memc.IVm {
	cache, ok := ctx.Value(CONTEXT_CACHE_KEY).(wazero.CompilationCache)
	if !ok {
		cache = wazero.NewCompilationCache()
		ctx = ctx.WithValue(CONTEXT_CACHE_KEY, cache)
	}
	var cleanups []func()
	config := wazero.NewRuntimeConfigCompiler().WithCompilationCache(cache)
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
	for _, fn := range wm.cleanups {
		fn()
	}
}

func (wm *WazeroVm) Call(funcname string, args []interface{}) ([]int32, error) {
	_args := make([]uint64, len(args))
	for i, arg := range args {
		_args[i] = uint64(arg.(int32))
	}
	result, err := wm.vm.ExportedFunction(funcname).Call(wm.ctx, _args...)
	if err != nil {
		return nil, err
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

func (wm *WazeroVm) InstantiateWasm(filePath string, wasmbuffer []byte) error {
	var err error
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
	_, err := mod.Instantiate(wm.ctx)
	return err
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

func (wm *WazeroVm) BuildModuleInner(rnh memc.RuntimeHandler, modname string, context interface{}, fndefs []WazeroFn) wazero.HostModuleBuilder {
	envmod := wm.r.NewHostModuleBuilder(modname)
	// TODO cost for each function
	for _, fndef := range fndefs {
		envmod = envmod.NewFunctionBuilder().WithGoModuleFunction(
			api.GoModuleFunc(fndef.WrappedFn(rnh)),
			fndef.inputTypes,
			fndef.outputTypes,
		).Export(fndef.name)
	}
	return envmod
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
