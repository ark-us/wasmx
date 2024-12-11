package config

var NodePortStep = int32(100)

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

func GetChainNodePorts(i int32, portOffset int32) NodePorts {
	return NodePorts{
		CosmosRestApi:    1317 + i + portOffset,
		CosmosGrpc:       9090 + i + portOffset,
		WasmxNetworkGrpc: 8090 + i + portOffset,
		WebsrvWebServer:  9900 + i + portOffset,
		EvmJsonRpc:       8545 + i*2 + portOffset*2,
		EvmJsonRpcWs:     8546 + i*2 + portOffset*2,
		Pprof:            6060 + i + portOffset,
		WasmxNetworkP2P:  5001 + i + portOffset,
		TendermintRpc:    26657 + i + portOffset,
	}
}

func GetChainNodePortsInitial(i int32, portOffset int32) NodePorts {
	portstep := NodePortStep
	subchainPortOffset := portOffset * int32(portstep)
	return getChainNodePortsInitial(i, portstep, subchainPortOffset)
}

func getChainNodePortsInitial(i int32, portstep int32, subchainPortOffset int32) NodePorts {
	return NodePorts{
		CosmosRestApi:    1330 + i*portstep + subchainPortOffset,
		CosmosGrpc:       9100 + i*portstep + subchainPortOffset,
		WasmxNetworkGrpc: 8100 + i*portstep + subchainPortOffset,
		WebsrvWebServer:  9910 + i*portstep + subchainPortOffset,
		EvmJsonRpc:       8555 + i*portstep + subchainPortOffset,
		EvmJsonRpcWs:     8656 + i*portstep + subchainPortOffset,
		Pprof:            6070 + i*portstep + subchainPortOffset,
		WasmxNetworkP2P:  5010 + i*portstep + subchainPortOffset,
		TendermintRpc:    26670 + i*portstep + subchainPortOffset,
	}
}
