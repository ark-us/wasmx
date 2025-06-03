package vm

import (
	"math"
	"math/big"
	"time"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

type SysContext struct {
	ctx *Context
}

func timeNow(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	magnitude := params[0].(int32)
	if magnitude > 9 {
		magnitude = 9
	}
	if magnitude < 0 {
		magnitude = 0
	}
	timed := time.Now().UnixNano() / int64(math.Pow10(int(9-magnitude)))
	return rnh.AllocateWriteMem(big.NewInt(timed).FillBytes(make([]byte, 32)))
}

// timeNow is non-deterministic
func BuildSysEnv(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("timeNow", timeNow, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "sys", context, fndefs)
}
