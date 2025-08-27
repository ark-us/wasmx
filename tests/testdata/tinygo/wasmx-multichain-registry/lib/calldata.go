package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
	xchain "github.com/loredanacirstea/wasmx-env-crosschain"
)

type CallData struct {
	RegisterSubChain          *RegisterSubChainRequest          `json:"RegisterSubChain,omitempty"`
	RegisterDefaultSubChain   *RegisterDefaultSubChainRequest   `json:"RegisterDefaultSubChain,omitempty"`
	RemoveSubChain            *RemoveSubChainRequest            `json:"RemoveSubChain,omitempty"`
	RegisterSubChainValidator *RegisterSubChainValidatorRequest `json:"RegisterSubChainValidator,omitempty"`
	InitSubChain              *InitSubChainRequest              `json:"InitSubChain,omitempty"`

	// query
	GetCurrentLevel                *QueryGetCurrentLevelRequest             `json:"GetCurrentLevel,omitempty"`
	GetSubChains                   *QueryGetSubChainsRequest                `json:"GetSubChains,omitempty"`
	GetSubChainsByIds              *QueryGetSubChainsByIdsRequest           `json:"GetSubChainsByIds,omitempty"`
	GetSubChainIds                 *QueryGetSubChainIdsRequest              `json:"GetSubChainIds,omitempty"`
	GetSubChainById                *QueryGetSubChainRequest                 `json:"GetSubChainById,omitempty"`
	GetSubChainConfigById          *QueryGetSubChainRequest                 `json:"GetSubChainConfigById,omitempty"`
	GetSubChainConfigByIds         *QuerySubChainConfigByIdsRequest         `json:"GetSubChainConfigByIds,omitempty"`
	GetSubChainIdsByLevel          *QueryGetSubChainIdsByLevelRequest       `json:"GetSubChainIdsByLevel,omitempty"`
	GetSubChainIdsByValidator      *QueryGetSubChainIdsByValidatorRequest   `json:"GetSubChainIdsByValidator,omitempty"`
	GetValidatorsByChainId         *QueryGetValidatorsByChainIdRequest      `json:"GetValidatorsByChainId,omitempty"`
	GetValidatorAddressesByChainId *QueryValidatorAddressesByChainIdRequest `json:"GetValidatorAddressesByChainId,omitempty"`
	ConvertAddressByChainId        *QueryConvertAddressByChainIdRequest     `json:"ConvertAddressByChainId,omitempty"`

	// crosschain
	CrossChainTx                    *xchain.MsgCrossChainCallRequest `json:"CrossChainTx,omitempty"`
	CrossChainQuery                 *xchain.MsgCrossChainCallRequest `json:"CrossChainQuery,omitempty"`
	CrossChainQueryNonDeterministic *xchain.MsgCrossChainCallRequest `json:"CrossChainQueryNonDeterministic,omitempty"`
}

func GetCallDataWrap() (CallData, error) {
	raw := wasmx.GetCallData()
	var cd CallData
	err := json.Unmarshal(raw, &cd)
	return cd, err
}

func GetCallDataInitialize() (MsgInitialize, error) {
	raw := wasmx.GetCallData()
	var v MsgInitialize
	err := json.Unmarshal(raw, &v)
	return v, err
}
