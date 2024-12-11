package config

import (
	networkconfig "wasmx/v1/x/network/server/config"
	jsonrpcconfig "wasmx/v1/x/wasmx/server/config"
	websrvconfig "wasmx/v1/x/websrv/server/config"
)

const DefaultConfigTemplate = websrvconfig.DefaultConfigTemplate + jsonrpcconfig.DefaultConfigTemplate + networkconfig.DefaultConfigTemplate
