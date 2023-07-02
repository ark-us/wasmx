package config

import (
	jsonrpcconfig "mythos/v1/x/wasmx/server/config"
	websrvconfig "mythos/v1/x/websrv/server/config"
)

const DefaultConfigTemplate = websrvconfig.DefaultConfigTemplate + jsonrpcconfig.DefaultConfigTemplate
