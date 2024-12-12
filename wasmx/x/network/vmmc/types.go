package vmmc

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"

	mctx "github.com/loredanacirstea/wasmx/v1/context"
	menc "github.com/loredanacirstea/wasmx/v1/encoding"
	vmtypes "github.com/loredanacirstea/wasmx/v1/x/wasmx/vm"
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
	CurrentNodeId    int32                 `json:"current_node_id"`
	InitialPorts     mctx.NodePorts        `json:"initial_ports"`
}

type StartSubChainMsg struct {
	ChainId     string           `json:"chain_id"`
	ChainConfig menc.ChainConfig `json:"chain_config"`
	NodePorts   mctx.NodePorts   `json:"node_ports"`
}

type StartSubChainResponse struct {
	Error string `json:"error"`
}

type StateSyncConfig struct {
	Enable              bool     `json:"enable"`
	TempDir             string   `json:"temp_dir"`
	DiscoveryTime       int64    `json:"discovery_time"`
	ChunkRequestTimeout int64    `json:"chunk_request_timeout"`
	ChunkFetchers       int32    `json:"chunk_fetchers"`
	RpcServers          []string `json:"rpc_servers"`
	TrustPeriod         int64    `json:"trust_period"`
	TrustHeight         int64    `json:"trust_height"`
	TrustHash           string   `json:"trust_hash"`
}

type StateSyncRequestMsg struct {
	ProtocolId                  string           `json:"protocol_id"`
	PeerAddress                 string           `json:"peer_address"`
	ChainId                     string           `json:"chain_id"`
	ChainConfig                 menc.ChainConfig `json:"chain_config"`
	NodePorts                   mctx.NodePorts   `json:"node_ports"`
	InitialPorts                mctx.NodePorts   `json:"initial_node_ports"`
	StatesyncConfig             StateSyncConfig  `json:"statesync_config"`
	Peers                       []string         `json:"peers"`
	CurrentNodeId               int32            `json:"current_node_id"`
	VerificationChainId         string           `json:"verification_chain_id"`
	VerificationContractAddress string           `json:"verification_contract_address"`
}

type StateSyncResponse struct {
	Error string `json:"error"`
}
