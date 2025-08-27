package main

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	lib "github.com/loredanacirstea/wasmx-multichain-registry/lib"
)

//go:wasm-module consensus
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module consensus
//export wasmx_consensus_json_i64_1
func Wasmx_consensus_json_i64_1() {}

//go:wasm-module multichain
//export wasmx_multichain_json_i64_1
func Wasmx_multichain_json_i64_1() {}

//go:wasm-module crosschain
//export wasmx_crosschain_json_i64_1
func Wasmx_crosschain_json_i64_1() {}

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
		wasmx.Finish(result)
		return
	case calldata.RegisterSubChain != nil:
		result = lib.RegisterSubChain(*calldata.RegisterSubChain)
		wasmx.Finish(result)
		return
	case calldata.RegisterDefaultSubChain != nil:
		result = lib.RegisterDefaultSubChain(*calldata.RegisterDefaultSubChain)
		wasmx.Finish(result)
		return
	case calldata.RegisterSubChainValidator != nil:
		result = lib.RegisterSubChainValidator(*calldata.RegisterSubChainValidator)
		wasmx.Finish(result)
		return
	case calldata.RemoveSubChain != nil:
		result = lib.RemoveSubChain(*calldata.RemoveSubChain)
		wasmx.Finish(result)
		return
	case calldata.GetSubChainById != nil:
		result = lib.GetSubChainById(*calldata.GetSubChainById)
		wasmx.Finish(result)
		return
	case calldata.GetSubChainConfigById != nil:
		result = lib.GetSubChainConfigById(*calldata.GetSubChainConfigById)
		wasmx.Finish(result)
		return
	case calldata.GetSubChainConfigByIds != nil:
		result = lib.GetSubChainConfigByIds(*calldata.GetSubChainConfigByIds)
		wasmx.Finish(result)
		return
	case calldata.GetSubChainsByIds != nil:
		result = lib.GetSubChainsByIds(*calldata.GetSubChainsByIds)
		wasmx.Finish(result)
		return
	case calldata.GetSubChains != nil:
		result = lib.GetSubChains(*calldata.GetSubChains)
		wasmx.Finish(result)
		return
	case calldata.GetSubChainIds != nil:
		result = lib.GetSubChainIds(*calldata.GetSubChainIds)
		wasmx.Finish(result)
		return
	case calldata.GetSubChainIdsByLevel != nil:
		result = lib.GetSubChainIdsByLevel(*calldata.GetSubChainIdsByLevel)
		wasmx.Finish(result)
		return
	case calldata.GetSubChainIdsByValidator != nil:
		result = lib.GetSubChainIdsByValidator(*calldata.GetSubChainIdsByValidator)
		wasmx.Finish(result)
		return
	case calldata.GetValidatorsByChainId != nil:
		result = lib.GetValidatorsByChainId(*calldata.GetValidatorsByChainId)
		wasmx.Finish(result)
		return
	case calldata.GetValidatorAddressesByChainId != nil:
		result = lib.GetValidatorAddressesByChainId(*calldata.GetValidatorAddressesByChainId)
		wasmx.Finish(result)
		return
	case calldata.ConvertAddressByChainId != nil:
		result = lib.ConvertAddressByChainId(*calldata.ConvertAddressByChainId)
		wasmx.Finish(result)
		return
	case calldata.GetCurrentLevel != nil:
		result = lib.QueryGetCurrentLevelAction(*calldata.GetCurrentLevel)
		wasmx.Finish(result)
		return
	case calldata.CrossChainTx != nil:
		result = lib.CrossChainTx(*calldata.CrossChainTx)
		wasmx.Finish(result)
		return
	case calldata.CrossChainQuery != nil:
		result = lib.CrossChainQuery(*calldata.CrossChainQuery)
		wasmx.Finish(result)
		return
	case calldata.CrossChainQueryNonDeterministic != nil:
		result = lib.CrossChainQueryNonDeterministic(*calldata.CrossChainQueryNonDeterministic)
		wasmx.Finish(result)
		return
	}
	wasmx.Revert(append([]byte("invalid function call data: "), wasmx.GetCallData()...))
}
