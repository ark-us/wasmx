package vmmc

import (
	abci "github.com/cometbft/cometbft/abci/types"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// InitSubChain(*InitSubChainMsg) (*abci.ResponseInitChain, error)
func InitSubChainMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &abci.ResponseInitChain{}
	return prepareResponse(rnh, response)
}

// StartSubChain(StartSubChainMsg): void
func StartSubChainMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &StartSubChainResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func GetSubChainIdsMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	return prepareResponse(rnh, []string{})
}

// this is what we use to statesync subchains
// StartStateSyncRequest(StateSyncRequestMsg): void
func StartStateSyncRequestMock(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &StartSubChainResponse{Error: ""}
	return prepareResponse(rnh, response)
}

func BuildWasmxMultichainJson1Mock(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("InitSubChain", InitSubChainMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("StartSubChain", StartSubChainMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("GetSubChainIds", GetSubChainIdsMock, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("StartStateSync", StartStateSyncRequestMock, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "multichain", context, fndefs)
}
