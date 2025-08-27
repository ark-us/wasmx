package vm

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func BuildWasmxConsensusJson1Mocki32(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
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
