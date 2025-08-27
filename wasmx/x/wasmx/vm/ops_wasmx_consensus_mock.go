package vm

import (
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func MockCheckTx(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	return rnh.AllocateWriteMem(make([]byte, 0))
}
func MockPrepareProposal(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	return rnh.AllocateWriteMem(make([]byte, 0))
}
func MockProcessProposal(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	return rnh.AllocateWriteMem(make([]byte, 0))
}
func MockFinalizeBlock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	return rnh.AllocateWriteMem(make([]byte, 0))
}
func MockCommit(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	return rnh.AllocateWriteMem(make([]byte, 0))
}
func MockRollbackToVersion(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	return rnh.AllocateWriteMem(make([]byte, 0))
}
