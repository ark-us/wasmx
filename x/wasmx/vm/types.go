package vm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dbm "github.com/tendermint/tm-db"

	"github.com/second-state/WasmEdge-go/wasmedge"

	cw8types "mythos/v1/x/wasmx/cw8/types"
	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/native"
)

var (
	// 0, 1, 2 are used by wasmedge for success, terminate, fail
	Result_OutOfGas = wasmedge.NewResult(wasmedge.ErrCategory_UserLevel, 10)

	LOG_TYPE_WASMX = "wasmx"
)

type ContractContext struct {
	FilePath         string
	Vm               *wasmedge.VM
	VmAst            *wasmedge.AST
	VmExecutor       *wasmedge.Executor
	ContractStoreKey []byte
	Context          *Context
	SystemDeps       []types.SystemDep
	Bytecode         []byte // runtime bytecode
	CodeHash         []byte
}

func (c ContractContext) Execute(newctx *Context) ([]byte, error) {
	hexaddr := types.EvmAddressFromAcc(newctx.Env.Contract.Address).Hex()
	nativePrecompile, found := native.NativeMap[hexaddr]
	if found {
		data := nativePrecompile(newctx.Env.CurrentCall.CallData)
		newctx.ReturnData = data
		return data, nil
	}

	contractVm, cleanups, err := InitiateWasm(newctx, c.FilePath, nil, c.SystemDeps)
	if err != nil {
		runCleanups(cleanups)
		return nil, err
	}
	setExecutionBytecode(newctx, contractVm, types.ENTRY_POINT_EXECUTE)
	newctx.ContractRouter[newctx.Env.Contract.Address.String()].Vm = contractVm

	executeHandler := GetExecuteFunctionHandler(c.SystemDeps)
	_, err = executeHandler(newctx, contractVm, types.ENTRY_POINT_EXECUTE)
	if err != nil {
		runCleanups(cleanups)
		return nil, err
	}
	runCleanups(cleanups)
	return newctx.ReturnData, nil
}

// key is a bech32 string
type ContractRouter = map[string]*ContractContext

type Context struct {
	Ctx            sdk.Context
	GasMeter       types.GasMeter
	Env            *types.Env
	ContractRouter ContractRouter
	ContractStore  types.KVStore
	CosmosHandler  types.WasmxCosmosHandler
	ReturnData     []byte
	CurrentCallId  uint32
	Logs           []WasmxLog
	Messages       []cw8types.SubMsg `json:"messages"`
	dbIterators    map[int32]dbm.Iterator
}

type EwasmFunctionWrapper struct {
	Name string
	Vm   *wasmedge.VM
}

type WasmxLog struct {
	ContractAddress sdk.AccAddress
	Data            []byte
	Topics          [][32]byte
	Type            string
}
