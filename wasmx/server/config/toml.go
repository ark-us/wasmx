package config

import (
	networkconfig "github.com/loredanacirstea/wasmx/x/network/server/config"
	jsonrpcconfig "github.com/loredanacirstea/wasmx/x/wasmx/server/config"
	websrvconfig "github.com/loredanacirstea/wasmx/x/websrv/server/config"
)

const DefaultConfigTemplate = websrvconfig.DefaultConfigTemplate + jsonrpcconfig.DefaultConfigTemplate + networkconfig.DefaultConfigTemplate
