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

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	"github.com/cosmos/ibc-go/v8/testing/mock"

	abcicli "github.com/cometbft/cometbft/abci/client"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtconfig "github.com/cometbft/cometbft/config"
	cometbftenc "github.com/cometbft/cometbft/crypto/encoding"
	"github.com/cometbft/cometbft/libs/bytes"
	cmtquery "github.com/cometbft/cometbft/libs/pubsub/query"
	"github.com/cometbft/cometbft/libs/pubsub/query/syntax"
	cometp2p "github.com/cometbft/cometbft/p2p"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"

	cfg "mythos/v1/config"
	"mythos/v1/server/config"
	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type ABCIClient struct {
	mapp           cfg.MythosApp
	bapp           cfg.BaseApp
	nk             types.WasmxWrapper
	logger         log.Logger
	actionExecutor cfg.ActionExecutor
	serverConfig   *cmtconfig.Config
	config         *config.Config
}

func NewABCIClient(
	mapp cfg.MythosApp,
	bapp cfg.BaseApp,
	logger log.Logger,
	networkKeeper types.WasmxWrapper,
	serverConfig *cmtconfig.Config,
	config *config.Config,
	actionExecutor cfg.ActionExecutor,
) sdkclient.CometRPC {
	logger = logger.With("module", "network", "client", "abci")
	return &ABCIClient{
		mapp:           mapp,
		bapp:           bapp,
		nk:             networkKeeper,
		logger:         logger,
		serverConfig:   serverConfig,
		config:         config,
		actionExecutor: actionExecutor,
	}
}

func (c *ABCIClient) ABCIInfo(context.Context) (*rpctypes.ResultABCIInfo, error) {
	c.logger.Debug("ABCIClient.ABCIInfo")
	resInfo, err := c.bapp.Info(types.RequestInfo)
	if err != nil {
		return nil, err
	}
	return &rpctypes.ResultABCIInfo{Response: *resInfo}, nil
}

func (c *ABCIClient) ABCIQuery(goctx context.Context, path string, data bytes.HexBytes) (*rpctypes.ResultABCIQuery, error) {
	c.logger.Debug("ABCIClient.ABCIQuery", "path", path, "data", data.String())
	return c.ABCIQueryWithOptions(goctx, path, data, rpcclient.DefaultABCIQueryOptions)
}

func (c *ABCIClient) ABCIQueryWithOptions(goctx context.Context, path string, data bytes.HexBytes, opts rpcclient.ABCIQueryOptions) (*rpctypes.ResultABCIQuery, error) {
	c.logger.Debug("ABCIClient.ABCIQueryWithOptions", "path", path, "height", opts.Height, "data", data.String())

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

func (c *ABCIClient) BroadcastTxAsync(goctx context.Context, tx cmttypes.Tx) (*rpctypes.ResultBroadcastTx, error) {
	c.logger.Debug("ABCIClient.BroadcastTxAsync", "txhash", hex.EncodeToString(tx.Hash()))

	bapp := c.bapp
	mapp := c.mapp
	ok := false

	sdktx, err := c.bapp.TxDecode(tx)
	if err != nil {
		return nil, err
	}
	multiChainIds := []string{}
	txWithExtensions, ok := sdktx.(authante.HasExtensionOptionsTx)
	if ok {
		opts := txWithExtensions.GetExtensionOptions()
		if len(opts) > 0 && opts[0].GetTypeUrl() == types.TypeURL_ExtensionOptionAtomicMultiChainTx {
			ext := opts[0].GetCachedValue().(*types.ExtensionOptionAtomicMultiChainTx)
			multiChainIds = ext.ChainIds
		}
	}
	if len(sdktx.GetMsgs()) > 0 {
		msg, ok := sdktx.GetMsgs()[0].(*types.MsgMultiChainWrap)
		if ok {
			multiChainIds = []string{msg.MultiChainId}
		}
	}
	if len(multiChainIds) == 0 {
		multiChainIds = []string{bapp.ChainID()}
	}

	// if atomic transaction, we send it to all respective chains
	for _, multiChainId := range multiChainIds {
		multichainapp, err := cfg.GetMultiChainApp(c.mapp.GetGoContextParent())
		if err != nil {
			return nil, err
		}
		iapp, err := multichainapp.GetApp(multiChainId)
		if err != nil {
			return nil, err
		}
		mapp, ok = iapp.(cfg.MythosApp)
		if !ok {
			return nil, fmt.Errorf("error App interface from multichainapp")
		}
		bapp, ok = iapp.(cfg.BaseApp)
		if !ok {
			return nil, fmt.Errorf("error BaseApp interface from multichainapp")
		}

		// TODO use ctx from params?
		cb := func(goctx context.Context) (any, error) {
			ctx := sdk.UnwrapSDKContext(goctx)
			sdktx, err := bapp.TxDecode(tx)

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
					contractAddress, err = mapp.AddressCodec().StringToBytes(msgExec.Contract)
					if err != nil {
						return nil, err
					}
				}
				// }
				// if consensusless or consensusmeta contract -> just execute it
				// whitelist of contracts exposed like this - just chat
				if len(contractAddress.Bytes()) > 0 {
					contractInfo := c.nk.GetContractInfo(ctx, contractAddress)

					// whitelist is in hex
					addrhex := wasmxtypes.EvmAddressFromAcc(contractAddress).Hex()

					if contractInfo != nil && contractInfo.StorageType != wasmxtypes.ContractStorageType_CoreConsensus && slices.Contains(types.CONSENSUSLESS_EXTERNAL_WHITELIST, addrhex) {
						contractAddressStr, err := mapp.AddressCodec().BytesToString(contractAddress)
						if err != nil {
							return nil, err
						}
						c.logger.Info("ABCIClient.BroadcastTxAsync executing consensusless or consensusmeta contract", "address", contractAddressStr)

						// we sent directly to the contract
						rresp, err := c.nk.ExecuteContract(ctx, &types.MsgExecuteContract{
							Sender:   msgExec.Sender, // sender will be taken from decoded tx
							Contract: contractAddressStr,
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
			rresp, err := mapp.GetNetworkKeeper().ExecuteContract(ctx, &types.MsgExecuteContract{
				Sender:   wasmxtypes.ROLE_CONSENSUS,
				Contract: wasmxtypes.ROLE_CONSENSUS,
				Msg:      msg,
			})
			if err != nil {
				return nil, err
			}
			return rresp, nil
		}
		_, err = mapp.GetActionExecutor().ExecuteWithMockHeader(context.Background(), cb)
		// TODO handle resp, err ?
		if err != nil {
			c.logger.Error("ABCIClient.BroadcastTxAsync", "txhash", hex.EncodeToString(tx.Hash()), "error", err.Error())
		}
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

// Validators gets the validator set at the given block height.
//
// If no height is provided, it will fetch the latest validator set. Note the
// validators are sorted by their voting power - this is the canonical order
// for the validators in the set as used in computing their Merkle root.
//
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/validators
func (c *ABCIClient) Validators(ctx context.Context, height *int64, page, perPage *int) (*rpctypes.ResultValidators, error) {
	entry, blockHeight, err := c.GetBlockEntry(ctx, height)
	if err != nil {
		return nil, err
	}

	var bmeta types.TendermintValidators
	err = c.nk.Codec().UnmarshalJSON(entry.ValidatorInfo, &bmeta)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Validators failed to decode cmttypes.Validator")
	}

	count := len(bmeta.Validators)
	cmtvals := make([]*cmttypes.Validator, count)
	for i, val := range bmeta.Validators {
		var pubkey cryptotypes.PubKey
		err = c.mapp.InterfaceRegistry().UnpackAny(val.PubKey, &pubkey)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "ABCIClient.Validators failed to convert unpack cryptotypes.PubKey")
		}
		tmPk, err := cryptocodec.ToCmtProtoPublicKey(pubkey)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "ABCIClient.Validators failed to convert cryptotypes.PubKey to proto")
		}
		tmPk2, err := cometbftenc.PubKeyFromProto(tmPk)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "ABCIClient.Validators failed to convert cryptotypes.PubKey from proto")
		}
		valaddr, err := hex.DecodeString(val.HexAddress)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "ABCIClient.Validators failed to decode hex address")
		}
		v := &cmttypes.Validator{
			Address:          valaddr,
			PubKey:           tmPk2,
			VotingPower:      val.VotingPower,
			ProposerPriority: val.ProposerPriority,
		}
		cmtvals[i] = v
	}

	result := &rpctypes.ResultValidators{
		BlockHeight: blockHeight,
		Validators:  cmtvals,
		Count:       len(cmtvals),
		Total:       len(cmtvals), // TODO fixme
	}

	return result, nil
}

func (c *ABCIClient) Status(context.Context) (*rpctypes.ResultStatus, error) {
	c.logger.Debug("ABCIClient.Status")
	// TODO finalize

	res, err := c.bapp.Info(types.RequestInfo)
	if err != nil {
		return nil, fmt.Errorf("error calling Info: %v", err)
	}
	// TODO fixme
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		return nil, err
	}
	if c.serverConfig == nil {
		return nil, fmt.Errorf("error calling Info: serverConfig is nil")
	}
	result := &rpctypes.ResultStatus{
		NodeInfo: cometp2p.DefaultNodeInfo{
			ProtocolVersion: cometp2p.ProtocolVersion{
				P2P:   types.RequestInfo.P2PVersion,
				Block: types.RequestInfo.BlockVersion,
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
	entry, blockHeight, err := c.GetBlockEntry(ctx, height)
	if err != nil {
		return nil, err
	}

	var bmeta types.RequestProcessProposalWithMetaInfo
	err = c.nk.Codec().UnmarshalJSON(entry.Data, &bmeta)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed to decode RequestProcessProposalWithMetaInfo")
	}
	b := bmeta.Request

	var header cmttypes.Header
	err = json.Unmarshal(entry.Header, &header)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed to decode Header")
	}

	var lastCommit cmttypes.Commit
	err = json.Unmarshal(entry.LastCommit, &lastCommit)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed to decode Commit")
	}

	var evidence cmttypes.EvidenceData
	err = json.Unmarshal(entry.Evidence, &evidence)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed to decode EvidenceData")
	}

	if len(b.Hash) == 0 {
		return nil, fmt.Errorf("block (%d) not found", blockHeight)
	}

	// TODO fixme
	blockId := cmttypes.BlockID{
		Hash:          b.Hash,
		PartSetHeader: cmttypes.PartSetHeader{Total: 1, Hash: b.Hash},
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
	c.logger.Debug("ABCIClient.Commit", "height", height)
	block, err := c.Block(ctx, height)
	if err != nil {
		return nil, err
	}
	var lastCommit *cmttypes.Commit
	lastCommit = nil

	latestHeight, err := c.LatestBlockHeight(ctx)
	if err != nil {
		return nil, err
	}

	// If the next block has not been committed yet,
	// use a non-canonical commit
	if block.Block.Height == latestHeight {
		resp, err := c.consensusQuery("getLastBlockCommit", "[]")
		if err != nil {
			return nil, err
		}
		var commit cmttypes.Commit
		err = json.Unmarshal(resp.Data, &commit)
		if err != nil {
			return nil, err
		}
		if commit.Height != block.Block.Height {
			return nil, fmt.Errorf("ABCIClient.Commit commit height mismatch block height: expected %d, got %d", block.Block.Height, commit.Height)
		}
		return rpctypes.NewResultCommit(&block.Block.Header, &commit, true), nil
	}

	// Return the canonical commit (comes from the block at height+1)
	nextHeight := block.Block.Height + 1
	blockNext, err := c.Block(ctx, &nextHeight)

	// LoadBlockCommit returns the Commit for the given height.
	// This commit consists of the +2/3 and other Precommit-votes for block at `height`,
	// and it comes from the block.LastCommit for `height+1`.
	// If no commit is found for the given height, it returns nil.
	if err == nil && blockNext != nil {
		lastCommit = blockNext.Block.LastCommit
	}

	return rpctypes.NewResultCommit(&block.Block.Header, lastCommit, true), nil
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
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed to decode BlockEntry")
	}

	var bmeta types.RequestProcessProposalWithMetaInfo
	err = c.nk.Codec().UnmarshalJSON(entry.Data, &bmeta)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.Block failed to decode RequestProcessProposalWithMetaInfo")
	}
	blockData := bmeta.Request

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

// Important! fsmQuery must not create a cycle, so it must only use ExecuteWithHeader
func (c *ABCIClient) fsmQuery(key string) (*wasmxtypes.ContractResponse, error) {
	msg := fmt.Sprintf(`{"getContextValue":{"key":"%s"}}`, key)
	return c.storageQuery(msg)
}

// Important! storageQuery must not create a cycle, so it must only use ExecuteWithHeader
func (c *ABCIClient) storageQuery(msg string) (*wasmxtypes.ContractResponse, error) {
	return StorageQuery(c.nk, c.actionExecutor, c.bapp, msg)
}

func (c *ABCIClient) consensusQuery(method string, params string) (*wasmxtypes.ContractResponse, error) {
	msg := fmt.Sprintf(`{"execute":{"action":{"type":"%s","params":%s,"event":null}}}`, method, params)
	return ContractQuery(c.nk, c.actionExecutor, c.bapp, msg, wasmxtypes.ROLE_CONSENSUS, wasmxtypes.ROLE_CONSENSUS)
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

// API needed for state sync

func (c *ABCIClient) ApplySnapshotChunk(ctx context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	return c.bapp.ApplySnapshotChunk(req)
}

func (c *ABCIClient) LoadSnapshotChunk(ctx context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	return c.bapp.LoadSnapshotChunk(req)
}

func (c *ABCIClient) OfferSnapshot(ctx context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	return c.bapp.OfferSnapshot(req)
}

func (c *ABCIClient) ListSnapshots(ctx context.Context, req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	return c.bapp.ListSnapshots(req)
}

func (c *ABCIClient) CheckTx(ctx context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return c.bapp.CheckTx(req)
}

func (c *ABCIClient) CheckTxAsync(ctx context.Context, req *abci.RequestCheckTx) (*abcicli.ReqRes, error) {
	resp, err := c.bapp.CheckTx(req)
	if err != nil {
		return nil, err
	}
	// TODO
	// return cli.finishAsyncCall(types.ToRequestCheckTx(req), &types.Response{Value: &types.Response_CheckTx{CheckTx: res}}), nil
	reqres := abcicli.NewReqRes(abci.ToRequestCheckTx(req))
	reqres.Response = &abci.Response{Value: &abci.Response_CheckTx{CheckTx: resp}}
	return reqres, nil
}

// height is nil for latest block
func (c *ABCIClient) GetBlockEntry(ctx context.Context, height *int64) (*types.BlockEntry, int64, error) {
	blockHeight := int64(0)
	var err error
	if height == nil {
		blockHeight, err = c.LatestBlockHeight(ctx)
		c.logger.Debug("ABCIClient.Block", "latest height", blockHeight)
		if err != nil {
			return nil, blockHeight, errorsmod.Wrapf(err, "ABCIClient.Block failed")
		}
	} else {
		blockHeight = *height
	}
	return c.GetBlockEntryByHeight(ctx, blockHeight)
}

// Important! GetBlockEntryByHeight must not create a cycle, so it must only use ExecuteWithHeader
func (c *ABCIClient) GetBlockEntryByHeight(ctx context.Context, height int64) (*types.BlockEntry, int64, error) {

	// get indexed tx
	key := types.GetBlockKey(height)
	resp, err := c.fsmQuery(key)
	if err != nil {
		return nil, height, errorsmod.Wrapf(err, "ABCIClient.Block failed")
	}

	if len(resp.Data) == 0 {
		return nil, height, fmt.Errorf("block (%d) not found", height)
	}

	var entry types.BlockEntry
	err = json.Unmarshal(resp.Data, &entry)
	if err != nil {
		return nil, height, errorsmod.Wrapf(err, "ABCIClient.Block failed to decode BlockEntry")
	}
	return &entry, height, nil
}

// height is nil for latest block
func (c *ABCIClient) ConsensusParams(ctx context.Context, height *int64) (*rpctypes.ResultConsensusParams, error) {
	response := &rpctypes.ResultConsensusParams{}
	blockHeight := int64(0)
	var err error
	if height == nil {
		blockHeight, err = c.LatestBlockHeight(ctx)
		c.logger.Debug("ABCIClient.ConsensusParams", "latest height", blockHeight)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "ABCIClient.ConsensusParams failed")
		}
	} else {
		blockHeight = *height
	}
	response.BlockHeight = blockHeight
	c.logger.Debug("ABCIClient.ConsensusParams", "height", blockHeight)

	msg := fmt.Sprintf(`{"getConsensusParams":{"height":%d}}`, blockHeight)
	resp, err := c.storageQuery(msg)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.ConsensusParams failed")
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("consensus params (%d) not found", blockHeight)
	}

	var params cmttypes.ConsensusParams
	err = json.Unmarshal(resp.Data, &params)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "ABCIClient.ConsensusParams failed to decode ConsensusParams")
	}
	response.ConsensusParams = params
	return response, nil
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

// Important! StorageQuery must not create a cycle, so it must only use ExecuteWithHeader
func StorageQuery(nk types.WasmxWrapper, actionExecutor cfg.ActionExecutor, bapp cfg.BaseApp, msg string) (*wasmxtypes.ContractResponse, error) {
	return ContractQuery(nk, actionExecutor, bapp, msg, wasmxtypes.ROLE_STORAGE, wasmxtypes.ROLE_STORAGE)
}

// Important! ContractQuery must not create a cycle, so it must only use ExecuteWithHeader
func ContractQuery(nk types.WasmxWrapper, actionExecutor cfg.ActionExecutor, bapp cfg.BaseApp, msg string, sender, contract string) (*wasmxtypes.ContractResponse, error) {
	cb := func(goctx context.Context) (any, error) {
		return nk.QueryContract(sdk.UnwrapSDKContext(goctx), &types.MsgQueryContract{
			Sender:   sender,
			Contract: contract,
			Msg:      []byte(msg),
		})
	}
	qresp, err := actionExecutor.ExecuteWithMockHeader(context.Background(), cb)
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
