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
