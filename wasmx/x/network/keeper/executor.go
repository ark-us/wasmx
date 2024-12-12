package keeper

import (
	"context"
	"encoding/json"
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
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	cmttypes "github.com/cometbft/cometbft/types"

	mcfg "github.com/loredanacirstea/wasmx/v1/config"
	"github.com/loredanacirstea/wasmx/v1/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/v1/x/wasmx/types"
)

func checkNegativeHeight(height int64) error {
	if height < 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "cannot query with height < 0; please provide a valid height")
	}

	return nil
}

func (k *Keeper) GetHeaderByHeight(app mcfg.MythosApp, logger log.Logger, height int64, prove bool) (*cmtproto.Header, error) {
	return GetHeaderByHeight(app, logger, height, prove)
}

func GetHeaderByHeight(app mcfg.MythosApp, logger log.Logger, height int64, prove bool) (*cmtproto.Header, error) {
	bapp := app.GetBaseApp()
	if err := checkNegativeHeight(height); err != nil {
		return nil, err
	}
	cms := bapp.CommitMultiStore()
	qms := cms.(storetypes.MultiStore)

	lastBlockHeight := qms.LatestVersion()
	if lastBlockHeight == 0 {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidHeight, "%s is not ready; please wait for first block", bapp.ChainID())
	}

	if height > lastBlockHeight {
		return nil,
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
		return nil,
			errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"cannot query with proof when height <= 1; please provide a valid height",
			)
	}

	if height == lastBlockHeight {
		checkCtx := bapp.GetCheckStateContext()
		header := checkCtx.BlockHeader()
		var emptyTime time.Time
		if header.Height > 0 && header.Time != emptyTime {
			return &header, nil
		}
	}
	client := NewABCIClient(app, bapp, app.GetBaseApp().Logger(), app.GetNetworkKeeper(), nil, nil, app.GetActionExecutor())

	// Important! GetBlockEntryByHeight must not create a cycle, so it must only use ExecuteWithHeader
	entry, _, err := client.(*ABCIClient).GetBlockEntryByHeight(context.TODO(), height)
	if err != nil {
		return nil, err
	}
	var bheader cmttypes.Header
	err = json.Unmarshal(entry.Header, &bheader)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "CreateQueryContext failed to decode Header")
	}
	return bheader.ToProto(), nil
}

func GetMockHeader(bapp mcfg.BaseApp, height int64) cmtproto.Header {
	return cmtproto.Header{
		ChainID:            bapp.ChainID(),
		Height:             height,
		Time:               time.Now().UTC(),
		ProposerAddress:    []byte("proposer"),
		NextValidatorsHash: []byte("proposer"),
		AppHash:            bapp.LastCommitID().Hash,
		Version: tmversion.Consensus{
			Block: types.RequestInfo.BlockVersion,
		},
	}
}

func CreateQueryContextWithHeader(app mcfg.BaseApp, logger log.Logger, header cmtproto.Header, prove bool) (sdk.Context, func(), storetypes.CacheMultiStore, error) {
	cms := app.CommitMultiStore()
	qms := cms.(storetypes.MultiStore)

	// cacheMS, err := qms.CacheMultiStoreWithVersion(height)
	// if err != nil {
	// 	return sdk.Context{}, nil,CacheMultiStoreWithVersion
	// 		errorsmod.Wrapf(
	// 			sdkerrors.ErrInvalidRequest,
	// 			"failed to load state at height %d; %s (latest height: %d)", height, err, lastBlockHeight,
	// 		)
	// }
	cacheMS := qms.CacheMultiStore()
	lastBlockHeight := qms.LatestVersion()

	tmpctx := app.NewUncachedContext(false, header)

	// branch the commit multi-store for safety
	ctx := sdk.NewContext(cacheMS, tmpctx.BlockHeader(), true, logger).
		WithMinGasPrices(nil).
		WithBlockHeight(header.Height).
		WithGasMeter(storetypes.NewGasMeter(NETWORK_GAS_LIMIT))

	var emptyTime time.Time
	if header.Height != lastBlockHeight || header.Time == emptyTime {
		rms, ok := app.CommitMultiStore().(*rootmulti.Store)
		if ok {
			cInfo, err := rms.GetCommitInfo(header.Height)
			if cInfo != nil && err == nil {
				ctx = ctx.WithBlockTime(cInfo.Timestamp)
			}
		}
	}

	sdkCtx, commitCacheCtx := ctx.CacheContext()
	return sdkCtx, commitCacheCtx, cacheMS, nil
}

func commitCtx(mythosapp mcfg.MythosApp, sdkCtx sdk.Context, commitCacheCtx func(), ctxcachems storetypes.CacheMultiStore) error {
	commitCacheCtx()

	origtstore := ctxcachems.GetStore(mythosapp.GetCMetaKey(wasmxtypes.MetaConsensusStoreKey))
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
	mtx    sync.Mutex
	app    mcfg.MythosApp
	logger log.Logger
}

func NewActionExecutor(app mcfg.MythosApp, logger log.Logger) *ActionExecutor {
	return &ActionExecutor{
		app:    app,
		logger: logger,
	}
}

func (r *ActionExecutor) GetLogger() log.Logger {
	return r.logger
}

func (r *ActionExecutor) GetApp() mcfg.MythosApp {
	return r.app
}

func (r *ActionExecutor) GetBaseApp() mcfg.BaseApp {
	return r.app.GetBaseApp()
}

func (r *ActionExecutor) Execute(goCtx context.Context, height int64, cb func(goctx context.Context) (any, error)) (any, error) {
	header, err := GetHeaderByHeight(r.app, r.logger, height, false)
	if err != nil {
		return nil, err
	}
	return r.ExecuteWithHeader(goCtx, *header, cb)
}

func (r *ActionExecutor) ExecuteWithMockHeader(goCtx context.Context, cb func(goctx context.Context) (any, error)) (any, error) {
	return r.ExecuteWithHeader(context.Background(), GetMockHeader(r.GetBaseApp(), r.GetBaseApp().LastBlockHeight()), cb)
}

func (r *ActionExecutor) ExecuteWithHeader(goCtx context.Context, header cmtproto.Header, cb func(goctx context.Context) (any, error)) (any, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	sdkCtx, commitCacheCtx, ctxcachems, err := CreateQueryContextWithHeader(r.app.GetBaseApp(), r.logger, header, false)
	if err != nil {
		return nil, err
	}
	return r.ExecuteInternal(goCtx, sdkCtx, commitCacheCtx, ctxcachems, cb)
}

func (r *ActionExecutor) ExecuteInternal(
	goCtx context.Context,
	sdkCtx sdk.Context,
	commitCacheCtx func(),
	ctxcachems storetypes.CacheMultiStore,
	cb func(goctx context.Context) (any, error),
) (any, error) {
	if goCtx == nil {
		goCtx = context.Background()
	}
	goCtx = context.WithValue(goCtx, sdk.SdkContextKey, sdkCtx)
	res, err := cb(goCtx)
	if err != nil {
		return nil, err
	}

	// we only commit if callback was successful
	err = commitCtx(r.app, sdkCtx, commitCacheCtx, ctxcachems)
	if err != nil {
		return nil, err
	}
	return res, nil
}
