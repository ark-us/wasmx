package rpc

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/ethereum/go-ethereum/rpc"

	rpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"

	menc "github.com/loredanacirstea/wasmx/encoding"

	"github.com/loredanacirstea/wasmx/x/wasmx/rpc/backend"
	"github.com/loredanacirstea/wasmx/x/wasmx/rpc/namespaces/eth"
	"github.com/loredanacirstea/wasmx/x/wasmx/rpc/namespaces/net"
)

// RPC namespaces and API version
const (
	// Cosmos namespaces

	CosmosNamespace = "cosmos"

	// Ethereum namespaces

	Web3Namespace     = "web3"
	EthNamespace      = "eth"
	PersonalNamespace = "personal"
	NetNamespace      = "net"
	TxPoolNamespace   = "txpool"
	DebugNamespace    = "debug"
	MinerNamespace    = "miner"

	apiVersion = "1.0"
)

// APICreator creates the JSON-RPC API implementations.
type APICreator = func(
	svrCtx *server.Context,
	clientCtx client.Context,
	ctx context.Context,
	tendermintWebsocketClient *rpcclient.WSClient,
	allowUnprotectedTxs bool,
	chainId string,
	chainConfig menc.ChainConfig,
) []rpc.API

// apiCreators defines the JSON-RPC API namespaces.
var apiCreators map[string]APICreator

func init() {
	apiCreators = map[string]APICreator{
		EthNamespace: func(svrCtx *server.Context,
			clientCtx client.Context,
			ctx context.Context,
			tmWSClient *rpcclient.WSClient,
			allowUnprotectedTxs bool,
			chainId string,
			chainConfig menc.ChainConfig,
		) []rpc.API {
			evmBackend := backend.NewBackend(svrCtx, svrCtx.Logger, clientCtx, ctx, allowUnprotectedTxs, chainId, chainConfig)
			return []rpc.API{
				{
					Namespace: EthNamespace,
					Version:   apiVersion,
					Service:   eth.NewPublicAPI(svrCtx.Logger, evmBackend),
					Public:    true,
				},
				// {
				// 	Namespace: EthNamespace,
				// 	Version:   apiVersion,
				// 	Service:   filters.NewPublicAPI(ctx.Logger, clientCtx, tmWSClient, evmBackend),
				// 	Public:    true,
				// },
			}
		},
		// Web3Namespace: func(*server.Context, client.Context, *rpcclient.WSClient, bool) []rpc.API {
		// 	return []rpc.API{
		// 		{
		// 			Namespace: Web3Namespace,
		// 			Version:   apiVersion,
		// 			Service:   web3.NewPublicAPI(),
		// 			Public:    true,
		// 		},
		// 	}
		// },
		NetNamespace: func(_ *server.Context, clientCtx client.Context, _ context.Context, _ *rpcclient.WSClient, _ bool, _ string, _ menc.ChainConfig) []rpc.API {
			return []rpc.API{
				{
					Namespace: NetNamespace,
					Version:   apiVersion,
					Service:   net.NewPublicAPI(clientCtx),
					Public:    true,
				},
			}
		},
		// PersonalNamespace: func(ctx *server.Context,
		// 	clientCtx client.Context,
		// 	_ *rpcclient.WSClient,
		// 	allowUnprotectedTxs bool,
		// 	indexer types.EVMTxIndexer,
		// ) []rpc.API {
		// 	evmBackend := backend.NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs, indexer)
		// 	return []rpc.API{
		// 		{
		// 			Namespace: PersonalNamespace,
		// 			Version:   apiVersion,
		// 			Service:   personal.NewAPI(ctx.Logger, evmBackend),
		// 			Public:    false,
		// 		},
		// 	}
		// },
		// TxPoolNamespace: func(ctx *server.Context, _ client.Context, _ *rpcclient.WSClient, _ bool, _ types.EVMTxIndexer) []rpc.API {
		// 	return []rpc.API{
		// 		{
		// 			Namespace: TxPoolNamespace,
		// 			Version:   apiVersion,
		// 			Service:   txpool.NewPublicAPI(ctx.Logger),
		// 			Public:    true,
		// 		},
		// 	}
		// },
		// DebugNamespace: func(ctx *server.Context,
		// 	clientCtx client.Context,
		// 	_ *rpcclient.WSClient,
		// 	allowUnprotectedTxs bool,
		// 	indexer types.EVMTxIndexer,
		// ) []rpc.API {
		// 	evmBackend := backend.NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs, indexer)
		// 	return []rpc.API{
		// 		{
		// 			Namespace: DebugNamespace,
		// 			Version:   apiVersion,
		// 			Service:   debug.NewAPI(ctx, evmBackend),
		// 			Public:    true,
		// 		},
		// 	}
		// },
		// MinerNamespace: func(ctx *server.Context,
		// 	clientCtx client.Context,
		// 	_ *rpcclient.WSClient,
		// 	allowUnprotectedTxs bool,
		// 	indexer types.EVMTxIndexer,
		// ) []rpc.API {
		// 	evmBackend := backend.NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs, indexer)
		// 	return []rpc.API{
		// 		{
		// 			Namespace: MinerNamespace,
		// 			Version:   apiVersion,
		// 			Service:   miner.NewPrivateAPI(ctx, evmBackend),
		// 			Public:    false,
		// 		},
		// 	}
		// },
	}
}

// GetRPCAPIs returns the list of all APIs
func GetRPCAPIs(svrCtx *server.Context,
	clientCtx client.Context,
	ctx context.Context,
	tmWSClient *rpcclient.WSClient,
	allowUnprotectedTxs bool,
	selectedAPIs []string,
	chainId string,
	chainConfig menc.ChainConfig,
) []rpc.API {
	var apis []rpc.API

	for _, ns := range selectedAPIs {
		if creator, ok := apiCreators[ns]; ok {
			apis = append(apis, creator(svrCtx, clientCtx, ctx, tmWSClient, allowUnprotectedTxs, chainId, chainConfig)...)
		} else {
			svrCtx.Logger.Error("invalid namespace value", "namespace", ns)
		}
	}

	return apis
}

// RegisterAPINamespace registers a new API namespace with the API creator.
// This function fails if the namespace is already registered.
func RegisterAPINamespace(ns string, creator APICreator) error {
	if _, ok := apiCreators[ns]; ok {
		return fmt.Errorf("duplicated api namespace %s", ns)
	}
	apiCreators[ns] = creator
	return nil
}
