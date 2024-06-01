package vmmc

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"

	menc "mythos/v1/encoding"
	vmtypes "mythos/v1/x/wasmx/vm"
)

const HOST_WASMX_ENV_MULTICHAIN_VER1 = "wasmx_multichain_1"

const HOST_WASMX_ENV_EXPORT = "wasmx_multichain_"

const HOST_WASMX_ENV_MULTICHAIN = "multichain"

type Context struct {
	*vmtypes.Context
}

type InitSubChainMsg struct {
	InitChainRequest abci.RequestInitChain `json:"init_chain_request"`
	ChainConfig      menc.ChainConfig      `json:"chain_config"`
	ValidatorAddress bytes.HexBytes        `json:"validator_address"`
	ValidatorPrivKey []byte                `json:"validator_privkey"`
	ValidatorPubKey  []byte                `json:"validator_pubkey"`
	Peers            []string              `json:"peers"`
}

type StartSubChainMsg struct {
	ChainId     string           `json:"chain_id"`
	ChainConfig menc.ChainConfig `json:"chain_config"`
}

type StartSubChainResponse struct {
	Error string `json:"error"`
}
