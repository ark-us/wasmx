package ewasm

import (
	"math/big"

	"github.com/second-state/WasmEdge-go/wasmedge"

	"wasmx/x/wasmx/types"
)

type Context struct {
	Env                types.Env
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
