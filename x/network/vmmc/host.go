package vmmc

import (
	"encoding/json"

	"github.com/second-state/WasmEdge-go/wasmedge"

	abci "github.com/cometbft/cometbft/abci/types"

	vmtypes "mythos/v1/x/wasmx/vm"
	asmem "mythos/v1/x/wasmx/vm/memory/assemblyscript"
)

// InitChain(*abci.RequestInitChain) (*abci.ResponseInitChain, error)
func InitChain(_context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := _context.(*Context)
	requestbz, err := asmem.ReadMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	var req abci.RequestInitChain
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// TODO fill in this info
	response, err := InitApp(ctx, &req, nil, nil, nil, nil, nil)
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

func BuildWasmxMultichainJson1(ctx_ *vmtypes.Context) *wasmedge.Module {
	context := &Context{Context: ctx_}
	env := wasmedge.NewModule("multichain")
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)

	env.AddFunction("InitChain", wasmedge.NewFunction(functype_i32_i32, InitChain, context, 0))

	return env
}
