package ewasm

import (
	"math/big"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"wasmx/x/wasmx/types"
)

type ContractContext struct {
	FilePath      string
	Vm            *wasmedge.VM
	VmAst         *wasmedge.AST
	VmExecutor    *wasmedge.Executor
	ContractStore types.KVStore // TODO remove
	Context       *Context
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
	contractVm := wasmedge.NewVM()
	ewasmEnv := BuildEwasmEnv(newctx)
	err := contractVm.RegisterModule(ewasmEnv)
	if err != nil {
		return nil, err
	}
	err = contractVm.LoadWasmFile(c.FilePath)
	if err != nil {
		return nil, err
	}
	err = contractVm.Validate()
	if err != nil {
		return nil, err
	}
	err = contractVm.Instantiate()
	if err != nil {
		return nil, err
	}

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
	DeploymentCodeSize uint32
	CodeSize           uint32
}

type EwasmFunctionWrapper struct {
	Name string
	Vm   *wasmedge.VM
}

type EwasmLog struct {
	Data   []byte
	Topics [][]byte
}
