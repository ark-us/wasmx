package config

type NodePorts struct {
	CosmosRestApi    int32 `json:"cosmos_rest_api"`
	CosmosGrpc       int32 `json:"cosmos_grpc"`
	TendermintRpc    int32 `json:"tendermint_rpc"`
	WasmxNetworkGrpc int32 `json:"wasmx_network_grpc"`
	EvmJsonRpc       int32 `json:"evm_jsonrpc"`
	EvmJsonRpcWs     int32 `json:"evm_jsonrpc_ws"`
	WebsrvWebServer  int32 `json:"websrv_web_server"`
	Pprof            int32 `json:"pprof"`
	WasmxNetworkP2P  int32 `json:"wasmx_network_p2p"`
}

func GetInitialChainNodePorts(i int32, portOffset int32) NodePorts {
	return NodePorts{
		CosmosRestApi:    1317 + i + portOffset,
		CosmosGrpc:       9090 + i + portOffset,
		WasmxNetworkGrpc: 8090 + i + portOffset,
		WebsrvWebServer:  9900 + i + portOffset,
		EvmJsonRpc:       8545 + i*2 + portOffset,
		EvmJsonRpcWs:     8546 + i + portOffset,
		Pprof:            6060 + i + portOffset,
		WasmxNetworkP2P:  5001 + i + portOffset,
		TendermintRpc:    26657 + i + portOffset,
	}
}
