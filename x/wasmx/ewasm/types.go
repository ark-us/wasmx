package ewasm

import (
	"math/big"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"wasmx/x/wasmx/types"
)

type ContractContext struct {
	Vm      *wasmedge.VM
	Store   types.KVStore // TODO remove
	Context *Context
}

// key is a bech32 string
type ContractRouter = map[string]ContractContext

type Context struct {
	Env                *types.Env
	ContractRouter     ContractRouter
	ContractStore      types.KVStore
	CallContext        types.MessageInfo
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
