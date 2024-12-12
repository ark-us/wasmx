package config

import (
	networkconfig "github.com/loredanacirstea/wasmx/v1/x/network/server/config"
	jsonrpcconfig "github.com/loredanacirstea/wasmx/v1/x/wasmx/server/config"
	websrvconfig "github.com/loredanacirstea/wasmx/v1/x/websrv/server/config"
)

const DefaultConfigTemplate = websrvconfig.DefaultConfigTemplate + jsonrpcconfig.DefaultConfigTemplate + networkconfig.DefaultConfigTemplate
