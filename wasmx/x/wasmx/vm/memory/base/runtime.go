package base

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

type RuntimeHandlerBase struct {
	vm memc.IVm
}

var _ memc.RuntimeHandler = (*RuntimeHandlerBase)(nil)

func NewRuntimeHandlerBase(vm memc.IVm) memc.RuntimeHandler {
	return RuntimeHandlerBase{vm}
}

func (h RuntimeHandlerBase) GetVm() memc.IVm {
	return h.vm
}

func (h RuntimeHandlerBase) GetMemory() (memc.IMemory, error) {
	mem, err := h.vm.GetMemory()
	if err != nil {
		return nil, err
	}
	return mem, nil
}

func (h RuntimeHandlerBase) ReadMemFromPtr(pointer interface{}) ([]byte, error) {
	panic("RuntimeHandlerBase.ReadMemFromPtr not implemented")
}

func (h RuntimeHandlerBase) AllocateWriteMem(data []byte) (int32, error) {
	panic("RuntimeHandlerBase.AllocateWriteMem not implemented")
}

func (RuntimeHandlerBase) ReadJsString(arr []byte) string {
	panic("RuntimeHandlerBase.ReadJsString not implemented")
}

func (h RuntimeHandlerBase) ReadStringFromPtr(pointer interface{}) (string, error) {
	panic("RuntimeHandlerBase.ReadStringFromPtr not implemented")
}
