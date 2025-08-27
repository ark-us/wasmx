package vmmc

import (
	abci "github.com/cometbft/cometbft/abci/types"

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
