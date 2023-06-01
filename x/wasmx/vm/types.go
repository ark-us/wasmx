package vm

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/native"
)

var (
	// 0, 1, 2 are used by wasmedge for success, terminate, fail
	Result_OutOfGas = wasmedge.NewResult(wasmedge.ErrCategory_UserLevel, 10)

	LOG_TYPE_WASMX = "wasmx"
)

var (
	EventTypeWasmxLog = "log_"
	AttributeKeyIndex = "index"
	AttributeKeyData  = "data"
	AttributeKeyTopic = "topic_"
)

type ContractContext struct {
	FilePath         string
	Vm               *wasmedge.VM
	VmAst            *wasmedge.AST
	VmExecutor       *wasmedge.Executor
	ContractStoreKey []byte
	Context          *Context
	SystemDeps       []types.SystemDep
}

func (c ContractContext) Execute(newctx *Context) ([]byte, error) {
	hexaddr := EvmAddressFromAcc(newctx.Env.Contract.Address).Hex()
	nativePrecompile, found := native.NativeMap[hexaddr]
	if found {
		data := nativePrecompile(newctx.Calldata)
		newctx.ReturnData = data
		return data, nil
	}

	contractVm, cleanups, err := InitiateWasm(newctx, c.FilePath, nil, c.SystemDeps)
	if err != nil {
		runCleanups(cleanups)
		return nil, err
	}
	setExecutionBytecode(newctx, contractVm, "main")
	newctx.ContractRouter[newctx.Env.Contract.Address.String()].Vm = contractVm

	_, err = contractVm.Execute("main")
	if err != nil {
		return nil, err
	}
	runCleanups(cleanups)
	contractVm.Release()
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
	CallContext    types.MessageInfo
	CosmosHandler  types.WasmxCosmosHandler
	Calldata       []byte
	Callvalue      *big.Int
	ReturnData     []byte
	CurrentCallId  uint32
	Logs           []WasmxLog
	// instantiate -> this is the constructor + runtime + constructor args
	// execute -> this is the runtime bytecode
	ExecutionBytecode []byte
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
