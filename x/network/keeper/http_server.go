package keeper

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"
	cometcore "github.com/cometbft/cometbft/rpc/core"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	cometrpc "github.com/cometbft/cometbft/rpc/jsonrpc/server"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	comettypes "github.com/cometbft/cometbft/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

type Environment struct {
	app         servertypes.Application
	networkWrap *ABCIClient
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

func (env *Environment) Subscribe(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS Subscribe")
	return nil, fmt.Errorf("Subscribe not implemented")
}

func (env *Environment) Unsubscribe(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS Unsubscribe")
	return nil, fmt.Errorf("Unsubscribe not implemented")
}

func (env *Environment) UnsubscribeAll(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS UnsubscribeAll")
	return nil, fmt.Errorf("UnsubscribeAll not implemented")
}

func (env *Environment) Health(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS Health")
	return nil, fmt.Errorf("Health not implemented")
}

func (env *Environment) Status(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS Status")
	return nil, fmt.Errorf("Status not implemented")
}

func (env *Environment) NetInfo(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS NetInfo")
	return nil, fmt.Errorf("NetInfo not implemented")
}

func (env *Environment) BlockchainInfo(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS BlockchainInfo")
	return nil, fmt.Errorf("BlockchainInfo not implemented")
}

func (env *Environment) Genesis(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS Genesis")
	return nil, fmt.Errorf("Genesis not implemented")
}

func (env *Environment) GenesisChunked(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS GenesisChunked")
	return nil, fmt.Errorf("GenesisChunked not implemented")
}

func (env *Environment) Block(ctx *rpctypes.Context, height *int64) (*ctypes.ResultBlock, error) {
	fmt.Println("= WS Block")
	return env.networkWrap.Block(context.TODO(), height)
}

func (env *Environment) BlockByHash(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS BlockByHash")
	return nil, fmt.Errorf("BlockByHash not implemented")
}

func (env *Environment) BlockResults(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS BlockResults")
	return nil, fmt.Errorf("BlockResults not implemented")
}

func (env *Environment) Commit(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS Commit")
	return nil, fmt.Errorf("Commit not implemented")
}

func (env *Environment) Header(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS Header")
	return nil, fmt.Errorf("Header not implemented")
}

func (env *Environment) HeaderByHash(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS HeaderByHash")
	return nil, fmt.Errorf("HeaderByHash not implemented")
}

func (env *Environment) CheckTx(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS CheckTx")
	return nil, fmt.Errorf("CheckTx not implemented")
}

// Tx allows you to query the transaction results. `nil` could mean the
// transaction is in the mempool, invalidated, or was not sent in the first
// place.
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/tx
func (env *Environment) Tx(ctx *rpctypes.Context, hash []byte, prove bool) (*ctypes.ResultTx, error) {
	fmt.Println("= WS Tx")
	return env.networkWrap.Tx(context.TODO(), hash, prove)
}

func (env *Environment) TxSearch(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultTxSearch, error) {
	fmt.Println("= WS TxSearch")
	return nil, fmt.Errorf("TxSearch not implemented")
}

func (env *Environment) BlockSearch(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS BlockSearch")
	return nil, fmt.Errorf("BlockSearch not implemented")
}

func (env *Environment) Validators(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS Validators")
	return nil, fmt.Errorf("Validators not implemented")
}

func (env *Environment) DumpConsensusState(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS DumpConsensusState")
	return nil, fmt.Errorf("DumpConsensusState not implemented")
}

func (env *Environment) GetConsensusState(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS GetConsensusState")
	return nil, fmt.Errorf("GetConsensusState not implemented")
}

func (env *Environment) ConsensusParams(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS ConsensusParams")
	return nil, fmt.Errorf("ConsensusParams not implemented")
}

func (env *Environment) UnconfirmedTxs(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS UnconfirmedTxs")
	return nil, fmt.Errorf("UnconfirmedTxs not implemented")
}

func (env *Environment) NumUnconfirmedTxs(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS NumUnconfirmedTxs")
	return nil, fmt.Errorf("NumUnconfirmedTxs not implemented")
}

// BroadcastTxCommit returns with the responses from CheckTx and ExecTxResult.
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Tx/broadcast_tx_commit
func (env *Environment) BroadcastTxCommit(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS BroadcastTxCommit")
	return nil, fmt.Errorf("BroadcastTxCommit not implemented")
}

func (env *Environment) BroadcastTxSync(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTx, error) {
	fmt.Println("= WS BroadcastTxSync")
	return env.networkWrap.BroadcastTxSync(context.TODO(), tx)
}

func (env *Environment) BroadcastTxAsync(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS BroadcastTxAsync")
	return nil, fmt.Errorf("BroadcastTxAsync not implemented")

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
	fmt.Println("= WS ABCIQuery")
	req := &abci.RequestQuery{
		Data:   data,
		Height: height,
		Path:   path,
		Prove:  prove,
	}
	res, err := env.app.Query(context.TODO(), req)
	if err != nil {
		return nil, err
	}
	resp := &ctypes.ResultABCIQuery{Response: *res}
	return resp, nil
}

func (env *Environment) ABCIInfo(ctx *rpctypes.Context) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS ABCIInfo")
	return nil, fmt.Errorf("ABCIInfo not implemented")
}

func (env *Environment) BroadcastEvidence(ctx *rpctypes.Context, tx comettypes.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	fmt.Println("= WS BroadcastEvidence")
	return nil, fmt.Errorf("BroadcastEvidence not implemented")
}
