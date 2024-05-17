package types

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	menc "mythos/v1/encoding"
)

type InitSubChainDeterministicRequest struct {
	InitChainRequest abci.RequestInitChain `json:"init_chain_request"`
	ChainConfig      menc.ChainConfig      `json:"chain_config"`
	Peers            []string              `json:"peers"`
}

type RegisterSubChainRequest struct {
	Data     InitSubChainDeterministicRequest `json:"data"`
	GenTxs   [][]byte                         `json:"genTxs"`
	Balances sdk.Coins                        `json:"balances"`
}
