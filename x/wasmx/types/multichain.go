package types

import (
	"math/big"

	abci "github.com/cometbft/cometbft/abci/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	menc "mythos/v1/encoding"
)

type InitSubChainDeterministicRequest struct {
	InitChainRequest abci.RequestInitChain `json:"init_chain_request"`
	ChainConfig      menc.ChainConfig      `json:"chain_config"`
	Peers            []string              `json:"peers"`
}

type RegisterSubChainRequest struct {
	Data           InitSubChainDeterministicRequest `json:"data"`
	GenTxs         [][]byte                         `json:"genTxs"`
	InitialBalance *big.Int                         `json:"initial_balance"`
}

type RegisterDefaultSubChainRequest struct {
	ChainBaseName  string   `json:"chain_base_name"`
	DenomUnit      string   `json:"denom_unit"`
	Decimals       uint32   `json:"base_denom_unit"`
	LevelIndex     uint32   `json:"level_index"`
	InitialBalance *big.Int `json:"initial_balance"`
}

type RegisterSubChainValidatorRequest struct {
	ChainId string `json:"chainId"`
	GenTx   []byte `json:"genTx"`
}

type InitSubChainRequest struct {
	ChainId string `json:"chainId"`
}

type QueryGetSubChainIdsRequest struct{}

type QueryGetSubChainRequest struct {
	ChainId string `json:"chainId"`
}

type RemoveSubChainRequest struct {
	ChainId string `json:"chainId"`
}

type QueryGetSubChainsRequest struct{}
type QueryGetSubChainsByIdsRequest struct {
	Ids []string `json:"ids"`
}

type MultiChainRegistryCallData struct {
	RegisterSubChain          *RegisterSubChainRequest          `json:"RegisterSubChain,omitempty"`
	RegisterDefaultSubChain   *RegisterDefaultSubChainRequest   `json:"RegisterDefaultSubChain,omitempty"`
	RemoveSubChain            *RemoveSubChainRequest            `json:"RemoveSubChain,omitempty"`
	RegisterSubChainValidator *RegisterSubChainValidatorRequest `json:"RegisterSubChainValidator,omitempty"`
	InitSubChain              *InitSubChainRequest              `json:"InitSubChain,omitempty"`

	// query
	GetSubChains          *QueryGetSubChainsRequest      `json:"GetSubChains,omitempty"`
	GetSubChainsByIds     *QueryGetSubChainsByIdsRequest `json:"GetSubChainsByIds,omitempty"`
	GetSubChainIds        *QueryGetSubChainIdsRequest    `json:"GetSubChainIds,omitempty"`
	GetSubChainById       *QueryGetSubChainRequest       `json:"GetSubChainById,omitempty"`
	GetSubChainConfigById *QueryGetSubChainRequest       `json:"GetSubChainConfigById,omitempty"`
}

type QueryBuildGenTxRequest struct {
	ChainId string                          `json:"chainId"`
	Msg     stakingtypes.MsgCreateValidator `json:"msg"`
}

type ActionParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
