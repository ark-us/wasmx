package ewasm

import (
	"github.com/second-state/WasmEdge-go/wasmedge"
)

type WasmEthMessage struct {
	Readonly bool   `json:"readonly"`
	Data     []byte `json:"data"`
}

type Context struct {
	Calldata   []byte
	ReturnData []byte
}

type EwasmFunctionWrapper struct {
	Name string
	Vm   *wasmedge.VM
}
