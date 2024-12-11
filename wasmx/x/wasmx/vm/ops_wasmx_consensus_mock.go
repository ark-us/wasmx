package vm

import (
	memc "wasmx/v1/x/wasmx/vm/memory/common"
)

func BuildWasmxConsensusJson1Mock(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("CheckTx", MockCheckTx, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("PrepareProposal", MockPrepareProposal, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ProcessProposal", MockProcessProposal, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("FinalizeBlock", MockFinalizeBlock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("BeginBlock", MockFinalizeBlock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("EndBlock", MockFinalizeBlock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Commit", MockCommit, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("RollbackToVersion", MockRollbackToVersion, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "consensus", context, fndefs)
}

func MockCheckTx(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	ptr, err := rnh.AllocateWriteMem(make([]byte, 0))
	if err != nil {
		return nil, err
	}
	returns[0] = ptr
	return returns, nil
}
func MockPrepareProposal(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	ptr, err := rnh.AllocateWriteMem(make([]byte, 0))
	if err != nil {
		return nil, err
	}
	returns[0] = ptr
	return returns, nil
}
func MockProcessProposal(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	ptr, err := rnh.AllocateWriteMem(make([]byte, 0))
	if err != nil {
		return nil, err
	}
	returns[0] = ptr
	return returns, nil
}
func MockFinalizeBlock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	ptr, err := rnh.AllocateWriteMem(make([]byte, 0))
	if err != nil {
		return nil, err
	}
	returns[0] = ptr
	return returns, nil
}
func MockCommit(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	ptr, err := rnh.AllocateWriteMem(make([]byte, 0))
	if err != nil {
		return nil, err
	}
	returns[0] = ptr
	return returns, nil
}
func MockRollbackToVersion(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	ptr, err := rnh.AllocateWriteMem(make([]byte, 0))
	if err != nil {
		return nil, err
	}
	returns[0] = ptr
	return returns, nil
}
