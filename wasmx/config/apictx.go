package config

import (
	cmtcfg "github.com/cometbft/cometbft/config"

	srvconfig "wasmx/v1/server/config"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	mctx "wasmx/v1/context"
	menc "wasmx/v1/encoding"
)

type APICtxI interface {
	BuildConfigs(
		chainId string,
		chainCfg *menc.ChainConfig,
		ports mctx.NodePorts,
	) (MythosApp, *server.Context, client.Context, *srvconfig.Config, *cmtcfg.Config, client.CometRPC, error)
	StartChainApis(
		chainId string,
		chainCfg *menc.ChainConfig,
		ports mctx.NodePorts,
	) (MythosApp, *server.Context, client.Context, *srvconfig.Config, *cmtcfg.Config, client.CometRPC, error)
}
