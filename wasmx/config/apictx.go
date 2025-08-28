package config

import (
	cmtcfg "github.com/cometbft/cometbft/config"

	srvconfig "github.com/loredanacirstea/wasmx/server/config"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	mctx "github.com/loredanacirstea/wasmx/context"
	menc "github.com/loredanacirstea/wasmx/encoding"
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
	SetMultiapp(*MultiChainApp)
}
