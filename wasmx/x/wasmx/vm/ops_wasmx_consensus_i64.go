package vm

import (
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func InstantiateWasmxConsensusJson_i64(context *Context, rnh memc.RuntimeHandler, dep *types.SystemDep) error {
	var err error
	wasmx, err := BuildWasmxConsensus_i64(context, rnh)
	if err != nil {
		return err
	}
	err = rnh.GetVm().RegisterModule(wasmx)
	if err != nil {
		return err
	}
	return nil
}

func BuildWasmxConsensus_i64(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("CheckTx", CheckTx, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("PrepareProposal", PrepareProposal, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ProcessProposal", ProcessProposal, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("OptimisticExecution", OptimisticExecution, []interface{}{vm.ValType_I64(), vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("FinalizeBlock", FinalizeBlock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("BeginBlock", BeginBlock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("EndBlock", EndBlock, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("Commit", Commit, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("RollbackToVersion", RollbackToVersion, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("HeaderHash", wasmxHeaderHash, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ValidatorsHash", wasmxValidatorsHash, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ConsensusParamsHash", wasmxConsensusParamsHash, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("BlockCommitVoteBytes", wasmxBlockCommitVoteBytes, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
	}

	// TODO
	// // ApplySnapshotChunk(req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error)
	// env.AddFunction("ApplySnapshotChunk", NewFunction(functype_i32_i32, ApplySnapshotChunk, context, 0))

	// // LoadSnapshotChunk(req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error)
	// env.AddFunction("LoadSnapshotChunk", NewFunction(functype_i32_i32, LoadSnapshotChunk, context, 0))

	// // OfferSnapshot(req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error)
	// env.AddFunction("OfferSnapshot", NewFunction(functype_i32_i32, OfferSnapshot, context, 0))

	// // ListSnapshots(req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error)
	// env.AddFunction("ListSnapshots", NewFunction(functype_i32_i32, ListSnapshots, context, 0))

	return vm.BuildModule(rnh, "consensus", context, fndefs)
}
