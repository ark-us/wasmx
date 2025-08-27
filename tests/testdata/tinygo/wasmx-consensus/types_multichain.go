package consensus

import (
	"fmt"
	"strconv"
	"strings"
)

const START_EVM_ID = 1000

type ChainConfig struct {
	Bech32PrefixAccAddr  string `json:"Bech32PrefixAccAddr"`
	Bech32PrefixAccPub   string `json:"Bech32PrefixAccPub"`
	Bech32PrefixValAddr  string `json:"Bech32PrefixValAddr"`
	Bech32PrefixValPub   string `json:"Bech32PrefixValPub"`
	Bech32PrefixConsAddr string `json:"Bech32PrefixConsAddr"`
	Bech32PrefixConsPub  string `json:"Bech32PrefixConsPub"`
	Name                 string `json:"Name"`
	HumanCoinUnit        string `json:"HumanCoinUnit"`
	BaseDenom            string `json:"BaseDenom"`
	DenomUnit            string `json:"DenomUnit"`
	BaseDenomUnit        uint32 `json:"BaseDenomUnit"`
	BondBaseDenom        string `json:"BondBaseDenom"`
	BondDenom            string `json:"BondDenom"`
}

type InitSubChainDeterministicRequest struct {
	InitChainRequest RequestInitChain `json:"init_chain_request"`
	ChainConfig      ChainConfig      `json:"chain_config"`
	Peers            []string         `json:"peers"`
}

type GenesisState map[string][]byte

type GenutilGenesis struct {
	GenTxs [][]byte `json:"gen_txs"`
}

type InitSubChainMsg struct {
	InitChainRequest RequestInitChain `json:"init_chain_request"`
	ChainConfig      ChainConfig      `json:"chain_config"`
	ValidatorAddress string           `json:"validator_address"`
	ValidatorPrivkey []byte           `json:"validator_privkey"`
	ValidatorPubkey  []byte           `json:"validator_pubkey"`
	Peers            []string         `json:"peers"`
	CurrentNodeID    int32            `json:"current_node_id"`
	InitialPorts     NodePorts        `json:"initial_ports"`
}

type NewSubChainDeterministicData struct {
	InitChainRequest RequestInitChain `json:"init_chain_request"`
	ChainConfig      ChainConfig      `json:"chain_config"`
}

type StartSubChainMsg struct {
	ChainID     string      `json:"chain_id"`
	ChainConfig ChainConfig `json:"chain_config"`
	NodePorts   NodePorts   `json:"node_ports"`
}

type StartSubChainResponse struct {
	Error string `json:"error"`
}

type ChainIDParts struct {
	Full      string
	BaseName  string
	Level     uint32
	EvmID     uint64
	ForkIndex uint32
}

type ChainId struct {
	Full      string `json:"full"`
	BaseName  string `json:"base_name"`
	Level     uint32 `json:"level"`
	EvmID     uint64 `json:"evmid"`
	ForkIndex uint32 `json:"fork_index"`
}

// FromString parses chain IDs like: mythos_1_7000-1 or mythos_7000-1
func ChainIdFromString(chainId string) (ChainId, error) {
	parts := strings.Split(chainId, "_")
	if len(parts) < 2 {
		return ChainId{}, fmt.Errorf("invalid chain id: %s", chainId)
	}
	baseName := parts[0]
	var level uint32
	var lastpart string
	if len(parts) == 2 {
		lastpart = parts[1]
	} else {
		lv, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return ChainId{}, fmt.Errorf("invalid level: %w", err)
		}
		level = uint32(lv)
		lastpart = parts[2]
	}
	parts2 := strings.Split(lastpart, "-")
	if len(parts2) != 2 {
		return ChainId{}, fmt.Errorf("invalid evmid-fork part: %s", lastpart)
	}
	evmid, err := strconv.ParseUint(parts2[0], 10, 64)
	if err != nil {
		return ChainId{}, fmt.Errorf("invalid evmid: %w", err)
	}
	forkIndexU64, err := strconv.ParseUint(parts2[1], 10, 32)
	if err != nil {
		return ChainId{}, fmt.Errorf("invalid fork index: %w", err)
	}
	return ChainId{
		Full:      chainId,
		BaseName:  baseName,
		Level:     level,
		EvmID:     evmid,
		ForkIndex: uint32(forkIndexU64),
	}, nil
}

// ToString builds a chain id string like: base_level_evmid-fork
func ChainIdToString(chainBaseName string, level uint32, evmid int64, forkIndex uint32) string {
	if evmid < START_EVM_ID {
		evmid += START_EVM_ID
	}
	return fmt.Sprintf("%s_%d_%d-%d", chainBaseName, level, evmid, forkIndex)
}

type NodePorts struct {
	CosmosRestAPI    int32 `json:"cosmos_rest_api"`
	CosmosGRPC       int32 `json:"cosmos_grpc"`
	TendermintRPC    int32 `json:"tendermint_rpc"`
	WasmxNetworkGRPC int32 `json:"wasmx_network_grpc"`
	EVMJSONRPC       int32 `json:"evm_jsonrpc"`
	EVMJSONRPCWS     int32 `json:"evm_jsonrpc_ws"`
	WebsrvWebServer  int32 `json:"websrv_web_server"`
	Pprof            int32 `json:"pprof"`
	WasmxNetworkP2P  int32 `json:"wasmx_network_p2p"`
}

func DefaultNodePorts() NodePorts {
	return NodePorts{
		CosmosRestAPI:    1330,
		CosmosGRPC:       9100,
		TendermintRPC:    26670,
		WasmxNetworkGRPC: 8100,
		EVMJSONRPC:       8555,
		EVMJSONRPCWS:     8656,
		WebsrvWebServer:  9910,
		Pprof:            6070,
		WasmxNetworkP2P:  5010,
	}
}

func (n NodePorts) Increment() NodePorts {
	return NodePorts{
		CosmosRestAPI:    n.CosmosRestAPI + 1,
		CosmosGRPC:       n.CosmosGRPC + 1,
		TendermintRPC:    n.TendermintRPC + 1,
		WasmxNetworkGRPC: n.WasmxNetworkGRPC + 1,
		EVMJSONRPC:       n.EVMJSONRPC + 1,
		EVMJSONRPCWS:     n.EVMJSONRPCWS + 1,
		WebsrvWebServer:  n.WebsrvWebServer + 1,
		Pprof:            n.Pprof + 1,
		WasmxNetworkP2P:  n.WasmxNetworkP2P + 1,
	}
}

// Empty returns zeroed ports (for tests/configs that want overrides).
func (n NodePorts) Empty() NodePorts {
	return NodePorts{}
}

// State sync config and requests
type StateSyncConfig struct {
	Enable              bool     `json:"enable"`
	TempDir             string   `json:"temp_dir"`
	RPCServers          []string `json:"rpc_servers"`
	TrustPeriod         int64    `json:"trust_period"`
	TrustHeight         int64    `json:"trust_height"`
	TrustHash           string   `json:"trust_hash"`
	DiscoveryTime       int64    `json:"discovery_time"`
	ChunkRequestTimeout int64    `json:"chunk_request_timeout"`
	ChunkFetchers       int32    `json:"chunk_fetchers"`
}

func NewStateSyncConfig(rpcServers []string, trustHeight int64, trustHash string) StateSyncConfig {
	return StateSyncConfig{
		Enable:              true,
		TempDir:             "",
		RPCServers:          rpcServers,
		TrustPeriod:         604800000, // 168h in milliseconds (as used in AS)
		TrustHeight:         trustHeight,
		TrustHash:           trustHash,
		DiscoveryTime:       15000,
		ChunkRequestTimeout: 10000,
		ChunkFetchers:       4,
	}
}

type StartStateSyncRequest struct {
	ProtocolID                  string          `json:"protocol_id"`
	PeerAddress                 string          `json:"peer_address"`
	ChainID                     string          `json:"chain_id"`
	ChainConfig                 ChainConfig     `json:"chain_config"`
	NodePorts                   NodePorts       `json:"node_ports"`
	InitialNodePorts            NodePorts       `json:"initial_node_ports"`
	StateSyncConfig             StateSyncConfig `json:"statesync_config"`
	Peers                       []string        `json:"peers"`
	CurrentNodeID               int32           `json:"current_node_id"`
	VerificationChainID         string          `json:"verification_chain_id"`
	VerificationContractAddress string          `json:"verification_contract_address"`
}

type StartStateSyncResponse struct {
	Error string `json:"error"`
}
