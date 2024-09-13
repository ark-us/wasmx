package interfaces

import (
	"github.com/second-state/WasmEdge-go/wasmedge"
)

type MemoryHandler interface {
	ReadMemFromPtr(callframe *wasmedge.CallingFrame, pointer interface{}) ([]byte, error)
	AllocateWriteMem(vm *wasmedge.VM, callframe *wasmedge.CallingFrame, data []byte) (int32, error)
	ReadStringFromPtr(callframe *wasmedge.CallingFrame, pointer interface{}) (string, error)
	ReadJsString(arr []byte) string
}
