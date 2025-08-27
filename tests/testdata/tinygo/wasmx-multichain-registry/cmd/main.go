package main

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx-multichain-registry/lib"
)

//go:wasm-module consensus
//export wasmx_consensus_json_1
func wasmx_consensus_json_1() {}

//go:wasm-module multichain
//export wasmx_multichain_1
func wasmx_multichain_1() {}

//go:wasm-module consensus
//export memory_ptrlen_i64_1
func memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func wasmx_env_i64_2() {}

//go:wasm-module wasmx-multichain-registry
//export instantiate
func Instantiate() {
	calldata, err := lib.GetCallDataInitialize()
	if err != nil {
		lib.Revert("invalid Instantiate calldata: " + err.Error())
		return
	}
	lib.SetParams(calldata.Params)
}

func main() {
	calldata, err := lib.GetCallDataWrap()
	if err != nil {
		lib.Revert("invalid call data: " + err.Error())
		return
	}

	var result []byte

	switch {
	case calldata.InitSubChain != nil:
		result = lib.InitSubChain(*calldata.InitSubChain)
	case calldata.RegisterSubChain != nil:
		result = lib.RegisterSubChain(*calldata.RegisterSubChain)
	case calldata.RegisterDefaultSubChain != nil:
		result = lib.RegisterDefaultSubChain(*calldata.RegisterDefaultSubChain)
	case calldata.RegisterSubChainValidator != nil:
		result = lib.RegisterSubChainValidator(*calldata.RegisterSubChainValidator)
	case calldata.RemoveSubChain != nil:
		result = lib.RemoveSubChain(*calldata.RemoveSubChain)
	case calldata.GetSubChainById != nil:
		result = lib.GetSubChainById(*calldata.GetSubChainById)
	case calldata.GetSubChainConfigById != nil:
		result = lib.GetSubChainConfigById(*calldata.GetSubChainConfigById)
	case calldata.GetSubChainConfigByIds != nil:
		result = lib.GetSubChainConfigByIds(*calldata.GetSubChainConfigByIds)
	case calldata.GetSubChainsByIds != nil:
		result = lib.GetSubChainsByIds(*calldata.GetSubChainsByIds)
	case calldata.GetSubChains != nil:
		result = lib.GetSubChains(*calldata.GetSubChains)
	case calldata.GetSubChainIds != nil:
		result = lib.GetSubChainIds(*calldata.GetSubChainIds)
	case calldata.GetSubChainIdsByLevel != nil:
		result = lib.GetSubChainIdsByLevel(*calldata.GetSubChainIdsByLevel)
	case calldata.GetSubChainIdsByValidator != nil:
		result = lib.GetSubChainIdsByValidator(*calldata.GetSubChainIdsByValidator)
	case calldata.GetValidatorsByChainId != nil:
		result = lib.GetValidatorsByChainId(*calldata.GetValidatorsByChainId)
	case calldata.GetValidatorAddressesByChainId != nil:
		result = lib.GetValidatorAddressesByChainId(*calldata.GetValidatorAddressesByChainId)
	case calldata.ConvertAddressByChainId != nil:
		result = lib.ConvertAddressByChainId(*calldata.ConvertAddressByChainId)
	case calldata.GetCurrentLevel != nil:
		result = lib.QueryGetCurrentLevelAction(*calldata.GetCurrentLevel)
	case calldata.CrossChainTx != nil:
		result = lib.CrossChainTx(*calldata.CrossChainTx)
	case calldata.CrossChainQuery != nil:
		result = lib.CrossChainQuery(*calldata.CrossChainQuery)
	case calldata.CrossChainQueryNonDeterministic != nil:
		result = lib.CrossChainQueryNonDeterministic(*calldata.CrossChainQueryNonDeterministic)
	default:
		databz := wasmx.GetCallData()
		lib.Revert("invalid function call data: " + string(databz))
		return
	}

	wasmx.SetFinishData(result)
}
