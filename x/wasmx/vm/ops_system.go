package vm

import (
	"math"
	"math/big"
	"time"

	"github.com/second-state/WasmEdge-go/wasmedge"

	asmem "mythos/v1/x/wasmx/vm/memory/assemblyscript"
)

type SysContext struct {
	ctx *Context
}

func timeNow(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*SysContext)
	magnitude := params[0].(int32)
	if magnitude > 9 {
		magnitude = 9
	}
	if magnitude < 0 {
		magnitude = 0
	}
	timed := time.Now().UnixNano() / int64(math.Pow10(int(9-magnitude)))
	ptr, err := asmem.AllocateWriteMem(ctx.ctx.MustGetVmFromContext(), callframe, big.NewInt(timed).FillBytes(make([]byte, 32)))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// timeNow is non-deterministic
func BuildSysEnv(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("sys")
	ctx := &SysContext{ctx: context}
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	env.AddFunction("timeNow", wasmedge.NewFunction(functype_i32_i32, timeNow, ctx, 0))
	return env
}
