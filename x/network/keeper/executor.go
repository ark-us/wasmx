package keeper

import (
	"context"
	"fmt"
	"sync"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino" // Import amino.proto file for reflection

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	cfg "mythos/v1/config"
	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

func checkNegativeHeight(height int64) error {
	if height < 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot query with height < 0; please provide a valid height")
	}

	return nil
}

// createQueryContext creates a new sdk.Context for a query, taking as args
// the block height and whether the query needs a proof or not.
func CreateQueryContext(app types.BaseApp, logger log.Logger, height int64, prove bool) (sdk.Context, func(), storetypes.CacheMultiStore, error) {
	if err := checkNegativeHeight(height); err != nil {
		return sdk.Context{}, nil, nil, err
	}

	cms := app.CommitMultiStore()
	qms := cms.(storetypes.MultiStore)

	lastBlockHeight := qms.LatestVersion()
	if lastBlockHeight == 0 {
		return sdk.Context{}, nil, nil, errorsmod.Wrapf(sdkerrors.ErrInvalidHeight, "%s is not ready; please wait for first block", app.Name())
	}

	if height > lastBlockHeight {
		return sdk.Context{}, nil, nil,
			errorsmod.Wrap(
				sdkerrors.ErrInvalidHeight,
				"cannot query with height in the future; please provide a valid height",
			)
	}

	// when a client did not provide a query height, manually inject the latest
	if height == 0 {
		height = lastBlockHeight
	}

	if height <= 1 && prove {
		return sdk.Context{}, nil, nil,
			errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"cannot query with proof when height <= 1; please provide a valid height",
			)
	}

	// cacheMS, err := qms.CacheMultiStoreWithVersion(height)
	// if err != nil {
	// 	return sdk.Context{}, nil,CacheMultiStoreWithVersion
	// 		errorsmod.Wrapf(
	// 			sdkerrors.ErrInvalidRequest,
	// 			"failed to load state at height %d; %s (latest height: %d)", height, err, lastBlockHeight,
	// 		)
	// }
	cacheMS := qms.CacheMultiStore()

	// tmpctx, err := app.CreateQueryContext(height, false)
	// if err != nil {
	// 	return sdk.Context{}, nil, err
	// }
	// tmpctx := app.GetContextForFinalizeBlock(make([]byte, 0))
	// tmpctx := app.GetContextForCheckTx(make([]byte, 0))

	// TODO fixme!!!
	header := cmtproto.Header{
		ChainID:            app.ChainID(),
		Height:             height,
		Time:               time.Now().UTC(),
		ProposerAddress:    []byte("proposer"),
		NextValidatorsHash: []byte("proposer"),
		AppHash:            app.LastCommitID().Hash,
		// Version: tmversion.Consensus{
		// 	Block: version.BlockProtocol,
		// },
		// LastBlockId: tmproto.BlockID{
		// 	Hash: tmhash.Sum([]byte("block_id")),
		// 	PartSetHeader: tmproto.PartSetHeader{
		// 		Total: 11,
		// 		Hash:  tmhash.Sum([]byte("partset_header")),
		// 	},
		// },
		// AppHash:            tmhash.Sum([]byte("app")),
		// DataHash:           tmhash.Sum([]byte("data")),
		// EvidenceHash:       tmhash.Sum([]byte("evidence")),
		// ValidatorsHash:     tmhash.Sum([]byte("validators")),
		// NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		// ConsensusHash:      tmhash.Sum([]byte("consensus")),
		// LastResultsHash:    tmhash.Sum([]byte("last_result")),
	}
	tmpctx := app.NewUncachedContext(false, header)

	// branch the commit multi-store for safety
	ctx := sdk.NewContext(cacheMS, tmpctx.BlockHeader(), true, logger).
		WithMinGasPrices(nil).
		WithBlockHeight(height).
		WithGasMeter(storetypes.NewGasMeter(NETWORK_GAS_LIMIT))

	if height != lastBlockHeight {
		rms, ok := app.CommitMultiStore().(*rootmulti.Store)
		if ok {
			cInfo, err := rms.GetCommitInfo(height)
			if cInfo != nil && err == nil {
				ctx = ctx.WithBlockTime(cInfo.Timestamp)
			}
		}
	}

	sdkCtx, commitCacheCtx := ctx.CacheContext()
	return sdkCtx, commitCacheCtx, cacheMS, nil
}

func commitCtx(bapp types.BaseApp, sdkCtx sdk.Context, commitCacheCtx func(), ctxcachems storetypes.CacheMultiStore) error {
	mythosapp, ok := bapp.(MythosApp)
	if !ok {
		return fmt.Errorf("commitCtx: failed to get MythosApp from server Application")
	}
	commitCacheCtx()

	origtstore := ctxcachems.GetStore(mythosapp.GetCLessKey(wasmxtypes.MetaConsensusStoreKey))
	origtstore.(storetypes.CacheWrap).Write()

	origtstore = ctxcachems.GetStore(mythosapp.GetCLessKey(wasmxtypes.SingleConsensusStoreKey))
	origtstore.(storetypes.CacheWrap).Write()

	// origtstore2 := ctxcachems.GetKVStore(mythosapp.GetCLessKey(wasmxtypes.CLessStoreKey))
	// origtstore2.CacheWrap().Write()

	// cms := app.CommitMultiStore()
	// origtstore3 := cms.GetCommitKVStore(mythosapp.GetCLessKey(wasmxtypes.CLessStoreKey))
	// origtstore3.Commit()

	return nil
}

type ActionExecutor struct {
	mtx       sync.Mutex
	multiapps *cfg.MultiChainApp
	logger    log.Logger
}

func NewActionExecutor(multiapps *cfg.MultiChainApp, logger log.Logger) *ActionExecutor {
	return &ActionExecutor{
		multiapps: multiapps,
		logger:    logger,
	}
}

func (r *ActionExecutor) GetLogger() log.Logger {
	return r.logger
}

func (r *ActionExecutor) GetMultiApp() *cfg.MultiChainApp {
	return r.multiapps
}

func (r *ActionExecutor) GetMythosApp(chainId string) (MythosApp, error) {
	iapp, err := r.multiapps.GetApp(chainId)
	if err != nil {
		return nil, err
	}
	app, ok := iapp.(MythosApp)
	if !ok {
		return nil, fmt.Errorf("cannot get MythosApp")
	}
	return app, nil
}

func (r *ActionExecutor) GetApp(chainId string) (types.BaseApp, error) {
	app, err := r.GetMythosApp(chainId)
	if err != nil {
		return nil, err
	}
	bapp, ok := app.(types.BaseApp)
	if !ok {
		return nil, fmt.Errorf("cannot get BaseApp")
	}
	return bapp, nil
}

func (r *ActionExecutor) Execute(goCtx context.Context, height int64, cb func(goctx context.Context) (any, error), chainId string) (any, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()


	cfg.SetGlobalChainConfig(chainId)

	bapp, err := r.GetApp(chainId)
	if err != nil {
		return nil, err
	}
	if bapp.ChainID() != chainId {
		return nil, fmt.Errorf("BaseApp ChainID %s is different than expected %s", bapp.ChainID(), chainId)
	}

	sdkCtx, commitCacheCtx, ctxcachems, err := CreateQueryContext(bapp, r.logger, height, false)
	if err != nil {
		return nil, err
	}
	if goCtx == nil {
		goCtx = context.Background()
	}
	// goCtx, cancelFn := context.WithCancel(goCtx)
	// defer cancelFn()
	goCtx = context.WithValue(goCtx, sdk.SdkContextKey, sdkCtx)
	res, err := cb(goCtx)
	if err != nil {
		return nil, err
	}

	// we only commit if callback was successful
	err = commitCtx(bapp, sdkCtx, commitCacheCtx, ctxcachems)
	if err != nil {
		return nil, err
	}
	return res, nil
}
