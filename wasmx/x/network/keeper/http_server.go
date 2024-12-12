package keeper

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	cometcore "github.com/cometbft/cometbft/rpc/core"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	cometrpc "github.com/cometbft/cometbft/rpc/jsonrpc/server"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	comettypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/loredanacirstea/wasmx/v1/server/config"
)

type Environment struct {
	app          servertypes.Application
	networkWrap  client.CometRPC
	serverConfig *cmtconfig.Config
	config       *config.Config
}

func NewEnvironment(
	app servertypes.Application,
	networkWrap client.CometRPC,
	serverConfig *cmtconfig.Config,
	config *config.Config,
) Environment {
	return Environment{
		app,
		networkWrap,
		serverConfig,
		config,
	}
}

// Routes is a map of available routes.
func (env *Environment) GetRoutes() cometcore.RoutesMap {
	return cometcore.RoutesMap{
		// subscribe/unsubscribe are reserved for websocket events.
		"subscribe":       cometrpc.NewWSRPCFunc(env.Subscribe, "query"),
		"unsubscribe":     cometrpc.NewWSRPCFunc(env.Unsubscribe, "query"),
		"unsubscribe_all": cometrpc.NewWSRPCFunc(env.UnsubscribeAll, ""),

		// info AP
		"health":               cometrpc.NewRPCFunc(env.Health, ""),
		"status":               cometrpc.NewRPCFunc(env.Status, ""),
		"net_info":             cometrpc.NewRPCFunc(env.NetInfo, ""),
		"blockchain":           cometrpc.NewRPCFunc(env.BlockchainInfo, "minHeight,maxHeight", cometrpc.Cacheable()),
		"genesis":              cometrpc.NewRPCFunc(env.Genesis, "", cometrpc.Cacheable()),
		"genesis_chunked":      cometrpc.NewRPCFunc(env.GenesisChunked, "chunk", cometrpc.Cacheable()),
		"block":                cometrpc.NewRPCFunc(env.Block, "height", cometrpc.Cacheable("height")),
		"block_by_hash":        cometrpc.NewRPCFunc(env.BlockByHash, "hash", cometrpc.Cacheable()),
		"block_results":        cometrpc.NewRPCFunc(env.BlockResults, "height", cometrpc.Cacheable("height")),
		"commit":               cometrpc.NewRPCFunc(env.Commit, "height", cometrpc.Cacheable("height")),
		"header":               cometrpc.NewRPCFunc(env.Header, "height", cometrpc.Cacheable("height")),
		"header_by_hash":       cometrpc.NewRPCFunc(env.HeaderByHash, "hash", cometrpc.Cacheable()),
		"check_tx":             cometrpc.NewRPCFunc(env.CheckTx, "tx"),
		"tx":                   cometrpc.NewRPCFunc(env.Tx, "hash,prove", cometrpc.Cacheable()),
		"tx_search":            cometrpc.NewRPCFunc(env.TxSearch, "query,prove,page,per_page,order_by"),
		"block_search":         cometrpc.NewRPCFunc(env.BlockSearch, "query,page,per_page,order_by"),
		"validators":           cometrpc.NewRPCFunc(env.Validators, "height,page,per_page", cometrpc.Cacheable("height")),
		"dump_consensus_state": cometrpc.NewRPCFunc(env.DumpConsensusState, ""),
		"consensus_state":      cometrpc.NewRPCFunc(env.GetConsensusState, ""),
		"consensus_params":     cometrpc.NewRPCFunc(env.ConsensusParams, "height", cometrpc.Cacheable("height")),
		"unconfirmed_txs":      cometrpc.NewRPCFunc(env.UnconfirmedTxs, "limit"),
		"num_unconfirmed_txs":  cometrpc.NewRPCFunc(env.NumUnconfirmedTxs, ""),

		// tx broadcast API
		"broadcast_tx_commit": cometrpc.NewRPCFunc(env.BroadcastTxCommit, "tx"),
		"broadcast_tx_sync":   cometrpc.NewRPCFunc(env.BroadcastTxSync, "tx"),
		"broadcast_tx_async":  cometrpc.NewRPCFunc(env.BroadcastTxAsync, "tx"),

		// abci API
		"abci_query": cometrpc.NewRPCFunc(env.ABCIQuery, "path,data,height,prove"),
		"abci_info":  cometrpc.NewRPCFunc(env.ABCIInfo, "", cometrpc.Cacheable()),

		// evidence API
		"broadcast_evidence": cometrpc.NewRPCFunc(env.BroadcastEvidence, "evidence"),
	}
}

func (env *Environment) Subscribe(ctx *rpctypes.Context, query string) (*ctypes.ResultSubscribe, error) {
	fmt.Println("= WS Subscribe")
	return nil, fmt.Errorf("Subscribe not implemented")
}

func (env *Environment) Unsubscribe(ctx *rpctypes.Context, query string) (*ctypes.ResultUnsubscribe, error) {
	fmt.Println("= WS Unsubscribe")
	return nil, fmt.Errorf("Unsubscribe not implemented")
}

func (env *Environment) UnsubscribeAll(ctx *rpctypes.Context) (*ctypes.ResultUnsubscribe, error) {
	fmt.Println("= WS UnsubscribeAll")
	return nil, fmt.Errorf("UnsubscribeAll not implemented")
}

func (env *Environment) Health(*rpctypes.Context) (*ctypes.ResultHealth, error) {
	fmt.Println("= WS Health")
	return nil, fmt.Errorf("Health not implemented")
}

func (env *Environment) Status(ctx *rpctypes.Context) (*ctypes.ResultStatus, error) {
	fmt.Println("= WS Status")
	return env.networkWrap.Status(context.TODO())
}

func (env *Environment) NetInfo(*rpctypes.Context) (*ctypes.ResultNetInfo, error) {
	fmt.Println("= WS NetInfo")
	return nil, fmt.Errorf("NetInfo not implemented")
}

func (env *Environment) BlockchainInfo(
	_ *rpctypes.Context,
	minHeight, maxHeight int64,
) (*ctypes.ResultBlockchainInfo, error) {
	fmt.Println("= WS BlockchainInfo")
	return env.networkWrap.BlockchainInfo(context.TODO(), minHeight, maxHeight)
}

func (env *Environment) Genesis(*rpctypes.Context) (*ctypes.ResultGenesis, error) {
	fmt.Println("= WS Genesis")
	return nil, fmt.Errorf("Genesis not implemented")
}

func (env *Environment) GenesisChunked(_ *rpctypes.Context, chunk uint) (*ctypes.ResultGenesisChunk, error) {
	fmt.Println("= WS GenesisChunked")
	return nil, fmt.Errorf("GenesisChunked not implemented")
}

func (env *Environment) Block(ctx *rpctypes.Context, height *int64) (*ctypes.ResultBlock, error) {
	fmt.Println("= WS Block")
	return env.networkWrap.Block(context.TODO(), height)
}

func (env *Environment) BlockByHash(_ *rpctypes.Context, hash []byte) (*ctypes.ResultBlock, error) {
	fmt.Println("= WS BlockByHash")
	return env.networkWrap.BlockByHash(context.TODO(), hash)
}

func (env *Environment) BlockResults(_ *rpctypes.Context, heightPtr *int64) (*ctypes.ResultBlockResults, error) {
	fmt.Println("= WS BlockResults")
	return env.networkWrap.BlockResults(context.TODO(), heightPtr)
}

// Commit gets block commit at a given height.
// If no height is provided, it will fetch the commit for the latest block.
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/commit
func (env *Environment) Commit(_ *rpctypes.Context, heightPtr *int64) (*ctypes.ResultCommit, error) {
	fmt.Println("= WS Commit")
	return env.networkWrap.Commit(context.TODO(), heightPtr)
}

func (env *Environment) Header(_ *rpctypes.Context, heightPtr *int64) (*ctypes.ResultHeader, error) {
	fmt.Println("= WS Header")
	return nil, fmt.Errorf("Header not implemented")
}

func (env *Environment) HeaderByHash(_ *rpctypes.Context, hash bytes.HexBytes) (*ctypes.ResultHeader, error) {
	fmt.Println("= WS HeaderByHash")
	return nil, fmt.Errorf("HeaderByHash not implemented")
}

func (env *Environment) CheckTx(_ *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultCheckTx, error) {
	fmt.Println("= WS CheckTx")
	req := &abci.RequestCheckTx{
		Tx:   tx,
		Type: abci.CheckTxType_New,
	}
	abciclient := env.networkWrap.(*ABCIClient)
	resp, err := abciclient.CheckTx(context.TODO(), req)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultCheckTx{ResponseCheckTx: *resp}, nil
}

// Tx allows you to query the transaction results. `nil` could mean the
// transaction is in the mempool, invalidated, or was not sent in the first
// place.
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/tx
func (env *Environment) Tx(ctx *rpctypes.Context, hash []byte, prove bool) (*ctypes.ResultTx, error) {
	fmt.Println("= WS Tx")
	return env.networkWrap.Tx(context.TODO(), hash, prove)
}

func (env *Environment) TxSearch(
	ctx *rpctypes.Context,
	query string,
	prove bool,
	pagePtr, perPagePtr *int,
	orderBy string,
) (*ctypes.ResultTxSearch, error) {
	fmt.Println("= WS TxSearch")
	return env.networkWrap.TxSearch(context.TODO(), query, prove, pagePtr, perPagePtr, orderBy)
}

func (env *Environment) BlockSearch(
	ctx *rpctypes.Context,
	query string,
	pagePtr, perPagePtr *int,
	orderBy string,
) (*ctypes.ResultBlockSearch, error) {
	fmt.Println("= WS BlockSearch")
	return env.networkWrap.BlockSearch(context.TODO(), query, pagePtr, perPagePtr, orderBy)
}

func (env *Environment) Validators(
	_ *rpctypes.Context,
	heightPtr *int64,
	pagePtr, perPagePtr *int,
) (*ctypes.ResultValidators, error) {
	fmt.Println("= WS Validators")
	return env.networkWrap.Validators(context.TODO(), heightPtr, pagePtr, perPagePtr)
}

func (env *Environment) DumpConsensusState(*rpctypes.Context) (*ctypes.ResultDumpConsensusState, error) {
	fmt.Println("= WS DumpConsensusState")
	return nil, fmt.Errorf("DumpConsensusState not implemented")
}

func (env *Environment) GetConsensusState(*rpctypes.Context) (*ctypes.ResultConsensusState, error) {
	fmt.Println("= WS GetConsensusState")
	return nil, fmt.Errorf("GetConsensusState not implemented")
}

// ConsensusParams gets the consensus parameters at the given block height.
// If no height is provided, it will fetch the latest consensus params.
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/consensus_params
func (env *Environment) ConsensusParams(
	_ *rpctypes.Context,
	heightPtr *int64,
) (*ctypes.ResultConsensusParams, error) {
	fmt.Println("= WS ConsensusParams")
	client := env.networkWrap.(*ABCIClient)
	return client.ConsensusParams(context.TODO(), heightPtr)
}

func (env *Environment) UnconfirmedTxs(_ *rpctypes.Context, limitPtr *int) (*ctypes.ResultUnconfirmedTxs, error) {
	fmt.Println("= WS UnconfirmedTxs")
	return nil, fmt.Errorf("UnconfirmedTxs not implemented")
}

func (env *Environment) NumUnconfirmedTxs(*rpctypes.Context) (*ctypes.ResultUnconfirmedTxs, error) {
	fmt.Println("= WS NumUnconfirmedTxs")
	return nil, fmt.Errorf("NumUnconfirmedTxs not implemented")
}

// BroadcastTxCommit returns with the responses from CheckTx and ExecTxResult.
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Tx/broadcast_tx_commit
func (env *Environment) BroadcastTxCommit(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS BroadcastTxCommit")
	return env.networkWrap.BroadcastTxCommit(context.TODO(), tx)
}

func (env *Environment) BroadcastTxSync(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTx, error) {
	fmt.Println("= WS BroadcastTxSync")
	return env.networkWrap.BroadcastTxSync(context.TODO(), tx)
}

func (env *Environment) BroadcastTxAsync(_ *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTx, error) {
	fmt.Println("= WS BroadcastTxAsync")
	return env.networkWrap.BroadcastTxAsync(context.TODO(), tx)

	// err := env.Mempool.CheckTx(tx, nil, mempl.TxInfo{})
	// if err != nil {
	// 	return nil, err
	// }
	// return &ctypes.ResultBroadcastTx{Hash: tx.Hash()}, nil
}

func (env *Environment) ABCIQuery(
	ctx *rpctypes.Context,
	path string,
	data bytes.HexBytes,
	height int64,
	prove bool,
) (*ctypes.ResultABCIQuery, error) {
	fmt.Println("= WS ABCIQuery-", height, path, prove)
	return env.networkWrap.ABCIQueryWithOptions(context.TODO(), path, data, rpcclient.ABCIQueryOptions{Height: height, Prove: prove})
}

func (env *Environment) ABCIInfo(_ *rpctypes.Context) (*ctypes.ResultABCIInfo, error) {
	return env.networkWrap.ABCIInfo(context.TODO())
}

func (env *Environment) BroadcastEvidence(
	_ *rpctypes.Context,
	ev comettypes.Evidence,
) (*ctypes.ResultBroadcastEvidence, error) {
	fmt.Println("= WS BroadcastEvidence")
	return nil, fmt.Errorf("BroadcastEvidence not implemented")
}
