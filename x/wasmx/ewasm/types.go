package ewasm

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"wasmx/x/wasmx/types"
)

var (
	// 0, 1, 2 are used by wasmedge for success, terminate, fail
	Result_OutOfGas = wasmedge.NewResult(wasmedge.ErrCategory_UserLevel, 10)
)

var (
	EventTypeEwasmFunction = "ewasm_function"
	EventTypeEwasmLog      = "ewasm_log"
	AttributeKeyIndex      = "index"
	AttributeKeyData       = "data"
	AttributeKeyTopic      = "topic_"
)

type ContractContext struct {
	FilePath         string
	Vm               *wasmedge.VM
	VmAst            *wasmedge.AST
	VmExecutor       *wasmedge.Executor
	ContractStoreKey []byte
	Context          *Context
}

func (c ContractContext) Execute_() ([]byte, error) {
	store := wasmedge.NewStore()
	mod, err := c.VmExecutor.Instantiate(store, c.VmAst)
	if err != nil {
		return nil, err
	}

	funcinst := mod.FindFunction("main")
	if funcinst == nil {
		return nil, err
	}
	_, err = c.VmExecutor.Invoke(store, funcinst)
	if err != nil {
		return nil, err
	}
	store.Release()
	return c.Context.ReturnData, nil
}

func (c ContractContext) Execute(newctx *Context) ([]byte, error) {
	contractVm, ewasmEnv, err := InitiateWasm(newctx, c.FilePath, nil)
	if err != nil {
		return nil, err
	}
	setCodeSize(newctx, contractVm, "main")

	_, err = contractVm.Execute("main")
	if err != nil {
		return nil, err
	}
	ewasmEnv.Release()
	contractVm.Release()
	return newctx.ReturnData, nil
}

// key is a bech32 string
type ContractRouter = map[string]ContractContext

type Context struct {
	Ctx                sdk.Context
	GasMeter           types.GasMeter
	Env                *types.Env
	ContractRouter     ContractRouter
	ContractStore      types.KVStore
	CallContext        types.MessageInfo
	CosmosHandler      types.WasmxCosmosHandler
	Calldata           []byte
	Callvalue          *big.Int
	ReturnData         []byte
	CurrentCallId      uint32
	Logs               []EwasmLog
	DeploymentCodeSize int32
	CodeSize           int32
}

type EwasmFunctionWrapper struct {
	Name string
	Vm   *wasmedge.VM
}

type EwasmLog struct {
	ContractAddress sdk.AccAddress
	Data            []byte
	Topics          [][]byte
}
