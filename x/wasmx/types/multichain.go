package types

import (
	"math/big"

	abci "github.com/cometbft/cometbft/abci/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	ChainId string `json:"chainId"`
}
type QueryGetValidatorsByChainIdRequest struct {
	ChainId string `json:"chainId"`
}
type QueryValidatorAddressesByChainIdRequest struct {
	ChainId string `json:"chainId"`
}
type QueryGetCurrentLevelResponse struct {
	Level int32 `json:"level"`
}

type MultiChainRegistryCallData struct {
	RegisterSubChain          *RegisterSubChainRequest          `json:"RegisterSubChain,omitempty"`
	RegisterDefaultSubChain   *RegisterDefaultSubChainRequest   `json:"RegisterDefaultSubChain,omitempty"`
	RemoveSubChain            *RemoveSubChainRequest            `json:"RemoveSubChain,omitempty"`
	RegisterSubChainValidator *RegisterSubChainValidatorRequest `json:"RegisterSubChainValidator,omitempty"`
	InitSubChain              *InitSubChainRequest              `json:"InitSubChain,omitempty"`

	// query
	GetSubChains                   *QueryGetSubChainsRequest                `json:"GetSubChains,omitempty"`
	GetSubChainsByIds              *QueryGetSubChainsByIdsRequest           `json:"GetSubChainsByIds,omitempty"`
	GetSubChainIds                 *QueryGetSubChainIdsRequest              `json:"GetSubChainIds,omitempty"`
	GetSubChainById                *QueryGetSubChainRequest                 `json:"GetSubChainById,omitempty"`
	GetSubChainConfigById          *QueryGetSubChainRequest                 `json:"GetSubChainConfigById,omitempty"`
	GetValidatorsByChainId         *QueryGetValidatorsByChainIdRequest      `json:"GetValidatorsByChainId,omitempty"`
	GetValidatorAddressesByChainId *QueryValidatorAddressesByChainIdRequest `json:"GetValidatorAddressesByChainId,omitempty"`
}

type QueryBuildGenTxRequest struct {
	ChainId string                          `json:"chainId"`
	Msg     stakingtypes.MsgCreateValidator `json:"msg"`
}

type ActionParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Erc20TokenInstantiate struct {
	Admins    []string `json:"admins"`
	Minters   []string `json:"minters"`
	Name      string   `json:"name"`
	Symbol    string   `json:"symbol"`
	Decimals  int32    `json:"decimals"`
	BaseDenom string   `json:"base_denom"`
}

type Erc20RollupTokenInstantiate struct {
	Admins      []string `json:"admins"`
	Minters     []string `json:"minters"`
	Name        string   `json:"name"`
	Symbol      string   `json:"symbol"`
	Decimals    int32    `json:"decimals"`
	BaseDenom   string   `json:"base_denom"`
	SubChainIds []string `json:"sub_chain_ids"`
}

type MsgTotalSupplyResponse struct {
	Supply sdk.Coin `json:"supply"`
}

type MsgBalanceOfResponse struct {
	Balance sdk.Coin `json:"balance"`
}

type CoinPerChain struct {
	ChainId string      `json:"chain_id"`
	Value   sdkmath.Int `json:"value"`
}

type MsgTotalSupplyCrossChainResponse struct {
	Supply sdk.Coin       `json:"supply"`
	Chains []CoinPerChain `json:"chains"`
}

type MsgBalanceOfCrossChainResponse struct {
	Balance sdk.Coin       `json:"balance"`
	Chains  []CoinPerChain `json:"chains"`
}
