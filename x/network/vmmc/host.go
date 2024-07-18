package vmmc

import (
	"encoding/json"

	"github.com/second-state/WasmEdge-go/wasmedge"

	mcfg "mythos/v1/config"
	vmtypes "mythos/v1/x/wasmx/vm"
	asmem "mythos/v1/x/wasmx/vm/memory/assemblyscript"
)

// InitSubChain(*InitSubChainMsg) (*abci.ResponseInitChain, error)
func InitSubChain(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req InitSubChainMsg
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	response, err := InitApp(ctx, &req)
	if err != nil {
		ctx.Logger(ctx.Ctx).Error("could not initiate subchain app", "error", err.Error())
		return nil, wasmedge.Result_Fail
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// StartSubChain(StartSubChainMsg): void
func StartSubChain(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req StartSubChainMsg
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	err = StartApp(ctx, &req)
	response := &StartSubChainResponse{Error: ""}
	if err != nil {
		ctx.Logger(ctx.Ctx).Error("could not start subchain app", "error", err.Error())
		response.Error = err.Error()
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func GetSubChainIds(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	multichainapp, err := mcfg.GetMultiChainApp(ctx.GoContextParent)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	responsebz, err := json.Marshal(multichainapp.ChainIds)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := asmem.AllocateWriteMem(ctx.Context.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

func BuildWasmxMultichainJson1(ctx_ *vmtypes.Context) *wasmedge.Module {
	context := &Context{Context: ctx_}
	env := wasmedge.NewModule("multichain")
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("InitSubChain", wasmedge.NewFunction(functype_i32_i32, InitSubChain, context, 0))
	env.AddFunction("StartSubChain", wasmedge.NewFunction(functype_i32_i32, StartSubChain, context, 0))
	env.AddFunction("GetSubChainIds", wasmedge.NewFunction(functype__i32, GetSubChainIds, context, 0))
	return env
}
