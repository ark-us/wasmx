package keeper

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection

	"github.com/cosmos/ibc-go/v8/testing/mock"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/bytes"
	cmtquery "github.com/cometbft/cometbft/libs/pubsub/query"
	"github.com/cometbft/cometbft/libs/pubsub/query/syntax"
	cometp2p "github.com/cometbft/cometbft/p2p"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"mythos/v1/server/config"
	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type ABCIClient struct {
	bapp           types.BaseApp
	nk             types.WasmxWrapper
	logger         log.Logger
	actionExecutor *ActionExecutor
	serverConfig   *cmtconfig.Config
	config         *config.Config
}

func NewABCIClient(
	bapp types.BaseApp,
	logger log.Logger,
	networkKeeper types.WasmxWrapper,
	serverConfig *cmtconfig.Config,
	config *config.Config,
	actionExecutor *ActionExecutor,
) *ABCIClient {
	return &ABCIClient{
		bapp:           bapp,
		nk:             networkKeeper,
		logger:         logger,
		serverConfig:   serverConfig,
		config:         config,
		actionExecutor: actionExecutor,
	}
}

func (c *ABCIClient) ABCIInfo(context.Context) (*rpctypes.ResultABCIInfo, error) {
	resInfo, err := c.bapp.Info(RequestInfo)
	if err != nil {
		return nil, err
	}
	return &rpctypes.ResultABCIInfo{Response: *resInfo}, nil
}

func (c *ABCIClient) ABCIQuery(ctx context.Context, path string, data bytes.HexBytes) (*rpctypes.ResultABCIQuery, error) {
	return c.ABCIQueryWithOptions(ctx, path, data, rpcclient.DefaultABCIQueryOptions)
}

func (c *ABCIClient) ABCIQueryWithOptions(ctx context.Context, path string, data bytes.HexBytes, opts rpcclient.ABCIQueryOptions) (*rpctypes.ResultABCIQuery, error) {
	c.logger.Debug("ABCIClient.ABCIQueryWithOptions", "path", path, "height", opts.Height)
	// TODO prove
	req := &abci.RequestQuery{
		Data:   data,
		Height: opts.Height,
		Path:   path,
		Prove:  opts.Prove,
	}
	res, err := c.bapp.Query(context.TODO(), req)
	if err != nil {
		return nil, err
	}
	resp := &rpctypes.ResultABCIQuery{Response: *res}
	return resp, nil
}

// func (c *ABCIClient) Simulate(_ context.Context, tx cmttypes.SimulateRequest) (*rpctypes.SimulateResponse, error) {
// 	fmt.Println("--BroadcastTxCommit--")
// 	return nil, nil
// }

func (c *ABCIClient) BroadcastTxCommit(_ context.Context, tx cmttypes.Tx) (*rpctypes.ResultBroadcastTxCommit, error) {
	return nil, fmt.Errorf("ABCIClient.BroadcastTxCommit not implemented")
}

func (c *ABCIClient) BroadcastTxAsync(_ context.Context, tx cmttypes.Tx) (*rpctypes.ResultBroadcastTx, error) {
	c.logger.Debug("ABCIClient.BroadcastTxAsync", "txhash", hex.EncodeToString(tx.Hash()))

	// TODO use ctx from params?
	cb := func(goctx context.Context) (any, error) {
		ctx := sdk.UnwrapSDKContext(goctx)
		sdktx, err := c.bapp.TxDecode(tx)

		if len(sdktx.GetMsgs()) > 0 {
			// we just take the first one // TODO fixme
			msg := sdktx.GetMsgs()[0]
			var contractAddress sdk.AccAddress
			// TODO all messages, like MsgExecuteEth
			// msgEthTx, ok := msg.(*wasmxtypes.MsgExecuteEth)
			// if ok {
			// 	ethTx := msgEthTx.AsTransaction()
			// 	to := ethTx.To()
			// 	contractAddress = wasmxtypes.AccAddressFromEvm(*to)
			// } else {
			msgExec, ok := msg.(*wasmxtypes.MsgExecuteContract)
			if ok {
				contractAddress, err = sdk.AccAddressFromBech32(msgExec.Contract)
				if err != nil {
					return nil, err
				}
			}
			// }
			// if consensusless contract -> just execute it
			// whitelist of contracts exposed like this - just chat
			if len(contractAddress.Bytes()) > 0 {
				contractInfo := c.nk.GetContractInfo(ctx, contractAddress)

				// whitelist is in hex
				addrhex := wasmxtypes.EvmAddressFromAcc(contractAddress).Hex()

				if contractInfo != nil && contractInfo.StorageType != wasmxtypes.ContractStorageType_CoreConsensus && slices.Contains(types.CONSENSUSLESS_EXTERNAL_WHITELIST, addrhex) {
					c.logger.Info("ABCIClient.BroadcastTxAsync executing consensusless contract", "address", contractAddress.String())

					// we sent directly to the contract
					rresp, err := c.nk.ExecuteContract(ctx, &types.MsgExecuteContract{
						Sender:   msgExec.Sender, // sender will be taken from decoded tx
						Contract: contractAddress.String(),
						Msg:      []byte(fmt.Sprintf(`{"HandleTx":"%s"}`, base64.StdEncoding.EncodeToString(tx))),
					})

					if err != nil {
						return nil, err
					}
					return rresp, nil
				}
			}
		}

		msg := []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, base64.StdEncoding.EncodeToString(tx)))
		rresp, err := c.nk.ExecuteContract(ctx, &types.MsgExecuteContract{
			Sender:   wasmxtypes.ROLE_CONSENSUS,
			Contract: wasmxtypes.ROLE_CONSENSUS,
			Msg:      msg,
		})
		if err != nil {
			return nil, err
		}
		return rresp, nil
	}
	_, err := c.actionExecutor.Execute(context.Background(), c.bapp.LastBlockHeight(), cb, c.bapp.ChainID())
	// TODO handle resp, err ?
	if err != nil {
		c.logger.Error("ABCIClient.BroadcastTxAsync", "txhash", hex.EncodeToString(tx.Hash()), "error", err.Error())
	}

	return &rpctypes.ResultBroadcastTx{Hash: tx.Hash()}, nil
}

func (c *ABCIClient) BroadcastTxSync(ctx context.Context, tx cmttypes.Tx) (*rpctypes.ResultBroadcastTx, error) {
	c.logger.Debug("ABCIClient.BroadcastTxSync")
	return c.BroadcastTxAsync(ctx, tx)

	// TODO fixme

	// resCh := make(chan *abci.ResponseCheckTx, 1)
	// err := env.Mempool.CheckTx(tx, func(res *abci.ResponseCheckTx) {
	// 	select {
	// 	case <-ctx.Context().Done():
	// 	case resCh <- res:
	// 	}
	// }, mempl.TxInfo{})
	// if err != nil {
	// 	return nil, err
	// }

	// select {
	// case <-ctx.Context().Done():
	// 	return nil, fmt.Errorf("broadcast confirmation not received: %w", ctx.Context().Err())
	// case res := <-resCh:
	// 	return &ctypes.ResultBroadcastTx{
	// 		Code:      res.Code,
	// 		Data:      res.Data,
	// 		Log:       res.Log,
	// 		Codespace: res.Codespace,
	// 		Hash:      tx.Hash(),
	// 	}, nil
	// }
}

func (c *ABCIClient) Validators(ctx context.Context, height *int64, page, perPage *int) (*rpctypes.ResultValidators, error) {
	return nil, fmt.Errorf("ABCIClient.Validators not implemented")
}

func (c *ABCIClient) Status(context.Context) (*rpctypes.ResultStatus, error) {
	c.logger.Debug("ABCIClient.Status")
	// TODO finalize

	res, err := c.bapp.Info(RequestInfo)
	if err != nil {
		return nil, fmt.Errorf("error calling Info: %v", err)
	}
	// TODO fixme
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		return nil, err
	}
	result := &rpctypes.ResultStatus{
		NodeInfo: cometp2p.DefaultNodeInfo{
			ProtocolVersion: cometp2p.ProtocolVersion{
				P2P:   RequestInfo.P2PVersion,
				Block: RequestInfo.BlockVersion,
				App:   res.AppVersion,
			},
			// TODO client methods per chainId
			Network: c.bapp.ChainID(),
			// TODO
			DefaultNodeID: "9111ccf0de42038bfc305123ee92a6b7eadda2cc",
			ListenAddr:    c.config.Network.Address,
			Version:       res.Version,
			Channels:      []byte{1, 2},
			Moniker:       "fffr",
			Other:         cometp2p.DefaultNodeInfoOther{TxIndex: "on", RPCAddress: c.serverConfig.RPC.ListenAddress},
		},
		SyncInfo: rpctypes.SyncInfo{
			LatestBlockHash:     res.LastBlockAppHash, // TODO fixme
			LatestAppHash:       res.LastBlockAppHash,
			LatestBlockHeight:   res.LastBlockHeight,
			LatestBlockTime:     time.Now(),
			EarliestBlockHash:   res.LastBlockAppHash,
			EarliestAppHash:     res.LastBlockAppHash,
			EarliestBlockHeight: res.LastBlockHeight,
			EarliestBlockTime:   time.Now(),
			CatchingUp:          false,
		},
		ValidatorInfo: rpctypes.ValidatorInfo{
			Address:     bytes.HexBytes(pubKey.Address()),
			PubKey:      pubKey,
			VotingPower: 0,
		},
		// NodeInfo: env.P2PTransport.NodeInfo().(p2p.DefaultNodeInfo),
		// SyncInfo: ctypes.SyncInfo{
		// 	LatestBlockHash:     latestBlockHash,
		// 	LatestAppHash:       latestAppHash,
		// 	LatestBlockHeight:   latestHeight,
		// 	LatestBlockTime:     time.Unix(0, latestBlockTimeNano),
		// 	EarliestBlockHash:   earliestBlockHash,
		// 	EarliestAppHash:     earliestAppHash,
		// 	EarliestBlockHeight: earliestBlockHeight,
		// 	EarliestBlockTime:   time.Unix(0, earliestBlockTimeNano),
		// 	CatchingUp:          env.ConsensusReactor.WaitSync(),
		// },
		// ValidatorInfo: ctypes.ValidatorInfo{
		// 	Address:     env.PubKey.Address(),
		// 	PubKey:      env.PubKey,
		// 	VotingPower: votingPower,
		// },
	}

	return result, nil
}

// height is nil for latest block
func (c *ABCIClient) Block(ctx context.Context, height *int64) (*rpctypes.ResultBlock, error) {
	c.logger.Debug("ABCIClient.Block", "height", height)
	blockHeight := int64(0)
	var err error
	if height == nil {
		blockHeight, err = c.LatestBlockHeight(ctx)
		c.logger.Debug("ABCIClient.Block", "latest height", blockHeight)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed")
		}
	} else {
		blockHeight = *height
	}
	c.logger.Debug("ABCIClient.Block", "height", blockHeight)

	// get indexed tx
	key := types.GetBlockKey(blockHeight)
	resp, err := c.fsmQuery(key)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed")
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("block (%d) not found", blockHeight)
	}

	var entry types.BlockEntry
	err = json.Unmarshal(resp.Data, &entry)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed")
	}

	var b abci.RequestProcessProposal
	err = json.Unmarshal(entry.Data, &b)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed")
	}

	var header cmttypes.Header
	err = json.Unmarshal(entry.Header, &header)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed")
	}

	var lastCommit cmttypes.Commit
	err = json.Unmarshal(entry.LastCommit, &lastCommit)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed")
	}
	var evidence cmttypes.EvidenceData
	err = json.Unmarshal(entry.Evidence, &evidence)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed")
	}

	// TODO fixme
	blockId := cmttypes.BlockID{
		Hash:          b.Hash,
		PartSetHeader: cmttypes.PartSetHeader{Total: 0, Hash: b.Hash[0:3]},
	}

	txs := make([]cmttypes.Tx, len(b.Txs))
	for i, tx := range b.Txs {
		txs[i] = cmttypes.Tx(tx)
	}
	block := cmttypes.MakeBlock(b.Height, txs, &lastCommit, evidence.Evidence)
	block.ChainID = c.bapp.ChainID()
	block.AppHash = header.AppHash
	block.ConsensusHash = header.ConsensusHash
	block.Header = header
	block.LastBlockID = header.LastBlockID
	block.LastResultsHash = header.LastResultsHash
	block.Time = header.Time
	block.ProposerAddress = header.ProposerAddress

	result := rpctypes.ResultBlock{
		BlockID: blockId,
		Block:   block,
	}

	return &result, nil
}

func (c *ABCIClient) BlockByHash(ctx context.Context, hash []byte) (*rpctypes.ResultBlock, error) {
	return nil, fmt.Errorf("ABCIClient.BlockByHash not implemented")
}

func (c *ABCIClient) BlockResults(ctx context.Context, height *int64) (*rpctypes.ResultBlockResults, error) {
	return nil, fmt.Errorf("ABCIClient.BlockResults not implemented")
}

func (c *ABCIClient) BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*rpctypes.ResultBlockchainInfo, error) {
	return nil, fmt.Errorf("ABCIClient.BlockchainInfo not implemented")
}

func (c *ABCIClient) Commit(ctx context.Context, height *int64) (*rpctypes.ResultCommit, error) {
	return nil, fmt.Errorf("ABCIClient.Commit not implemented")
}

func (c *ABCIClient) Tx(ctx context.Context, hash []byte, prove bool) (*rpctypes.ResultTx, error) {
	c.logger.Debug("ABCIClient.Tx", "hash", hex.EncodeToString(hash))

	// get indexed tx
	key := types.GetTxKey(hash)
	resp, err := c.fsmQuery(key)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("tx (%X) not found", hash)
	}

	var indexedTx types.IndexedTransaction
	err = json.Unmarshal(resp.Data, &indexedTx)
	if err != nil {
		return nil, err
	}

	// get block data
	key = types.GetBlockKey(indexedTx.Height)
	resp, err = c.fsmQuery(key)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("tx block (%d) not found", indexedTx.Height)
	}

	var entry types.BlockEntry
	err = json.Unmarshal(resp.Data, &entry)
	if err != nil {
		return nil, err
	}

	var blockData abci.RequestProcessProposal
	err = json.Unmarshal(entry.Data, &blockData)
	if err != nil {
		return nil, err
	}

	var blockResultData abci.ResponseFinalizeBlock
	err = json.Unmarshal(entry.Result, &blockResultData)
	if err != nil {
		return nil, err
	}

	if len(blockResultData.TxResults) < int(indexedTx.Index) {
		return nil, fmt.Errorf("tx index (%d) not found in block %d", indexedTx.Index, indexedTx.Height)
	}

	var proof cmttypes.TxProof
	// TODO proof

	// this is a query, we do not commit anything
	return &rpctypes.ResultTx{
		Hash:     hash,
		Height:   indexedTx.Height,
		Index:    indexedTx.Index,
		TxResult: *blockResultData.TxResults[indexedTx.Index],
		Tx:       blockData.Txs[indexedTx.Index],
		Proof:    proof,
	}, nil
}

func (c *ABCIClient) TxSearch(
	ctx context.Context,
	query string,
	prove bool,
	page, perPage *int,
	orderBy string,
) (*rpctypes.ResultTxSearch, error) {
	q, err := cmtquery.New(query)
	if err != nil {
		return nil, err
	}
	c.logger.Debug("ABCIClient.TxSearch", "query", q.String())

	// TODO events

	// cometbft/state/txindex/kv
	// get a list of conditions (like "tx.height > 5")
	conditions := q.Syntax()
	// if there is a hash condition, return the result immediately
	hash, ok, err := lookForHash(conditions)
	if err != nil {
		c.logger.Error("ABCIClient.TxSearch", "ok", ok, "err", err.Error())
	}
	c.logger.Debug("ABCIClient.TxSearch", "hash", hex.EncodeToString(hash), "ok", ok)
	if err != nil {
		return nil, fmt.Errorf("error during searching for a hash in the query: %w", err)
	} else if ok {
		res, err := c.Tx(ctx, hash, false) // TODO prove?
		switch {
		case err != nil:
			return &rpctypes.ResultTxSearch{}, fmt.Errorf("error while retrieving the result: %w", err)
		case res == nil:
			return &rpctypes.ResultTxSearch{Txs: []*rpctypes.ResultTx{}, TotalCount: 0}, nil
		default:
			return &rpctypes.ResultTxSearch{Txs: []*rpctypes.ResultTx{res}, TotalCount: 1}, nil
		}
	}

	topics := make([]string, 0)
	for _, c := range conditions {
		topic := fmt.Sprintf("%s=%s", c.Tag, c.Arg.String())
		topics = append(topics, topic)
	}
	apiResults, err := c.TxsByEvents(ctx, topics)
	if err != nil {
		return nil, err
	}
	return &rpctypes.ResultTxSearch{Txs: apiResults, TotalCount: len(apiResults)}, nil
}

func (c *ABCIClient) TxsByEvents(ctx context.Context, topics []string) ([]*rpctypes.ResultTx, error) {
	c.logger.Debug("ABCIClient.TxsByEvents", "topics", topics)

	txhashes := make([]string, 0)
	for _, topic := range topics {
		key := types.GetTopicDataKey(topic)
		resp, err := c.fsmQuery(key)
		if err != nil {
			return nil, err
		}
		if len(resp.Data) == 0 {
			continue
		}
		var indexedTxs []string
		err = json.Unmarshal(resp.Data, &indexedTxs)
		if err != nil {
			return nil, err
		}

		// intersection
		if len(txhashes) == 0 {
			txhashes = indexedTxs
		} else {
			txhashesTemp := make([]string, 0)
			for _, hash := range indexedTxs {
				if slices.Contains(txhashes, hash) {
					txhashesTemp = append(txhashesTemp, hash)
				}
			}
			txhashes = txhashesTemp
		}
	}

	results := make([]*rpctypes.ResultTx, 0)
	for _, hash := range txhashes {
		hashbz, err := base64.StdEncoding.DecodeString(hash)
		if err != nil {
			continue
		}
		result, err := c.Tx(ctx, hashbz, false)
		if err != nil {
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

func (c *ABCIClient) BlockSearch(
	ctx context.Context,
	query string,
	page, perPage *int,
	orderBy string,
) (*rpctypes.ResultBlockSearch, error) {
	return nil, fmt.Errorf("ABCIClient.BlockSearch not implemented")
}

func (c *ABCIClient) fsmQuery(key string) (*wasmxtypes.ContractResponse, error) {
	cb := func(goctx context.Context) (any, error) {

		msg := []byte(fmt.Sprintf(`{"getContextValue":{"key":"%s"}}`, key))
		return c.nk.QueryContract(sdk.UnwrapSDKContext(goctx), &types.MsgQueryContract{
			Sender:   wasmxtypes.ROLE_STORAGE,
			Contract: wasmxtypes.ROLE_STORAGE,
			Msg:      msg,
		})
	}
	qresp, err := c.actionExecutor.Execute(context.Background(), c.bapp.LastBlockHeight(), cb, c.bapp.ChainID())
	if err != nil {
		return nil, err
	}
	rresp := qresp.(*types.MsgQueryContractResponse)

	var resp wasmxtypes.ContractResponse
	err = json.Unmarshal(rresp.Data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ABCIClient) LatestBlockHeight(ctx context.Context) (int64, error) {
	// get indexed tx
	key := types.GetLastBlockHeightKey()
	resp, err := c.fsmQuery(key)
	if err != nil {
		return 0, err
	}
	if len(resp.Data) == 0 {
		return 0, nil
	}
	return strconv.ParseInt(string(resp.Data), 10, 64)
}

func lookForHash(conditions []syntax.Condition) (hash []byte, ok bool, err error) {
	for _, c := range conditions {
		if c.Tag == cmttypes.TxHashKey {
			decoded, err := hex.DecodeString(c.Arg.Value())
			return decoded, true, err
		}
	}
	return
}
