package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	app "mythos/v1/app"
	"mythos/v1/x/wasmx/keeper"
	"mythos/v1/x/wasmx/types"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

func WasmxKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	tStoreKey := storetypes.NewTransientStoreKey(types.TStoreKey)
	clessStoreKey := storetypes.NewConsensuslessStoreKey(types.CLessStoreKey)

	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	stateStore := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	stateStore.MountStoreWithDB(tStoreKey, storetypes.StoreTypeTransient, db)
	stateStore.MountStoreWithDB(clessStoreKey, storetypes.StoreTypeConsensusless, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	encodingConfig := app.MakeEncodingConfig()
	_, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino

	appOpts := app.DefaultAppOptions{}
	g, _, _ := app.GetTestCtx(logger, true)
	appOpts.Set("goroutineGroup", g)

	paramsSubspace := typesparams.NewSubspace(cdc,
		codec.NewLegacyAmino(),
		storeKey,
		memStoreKey,
		"WasmxParams",
	)
	paramsKeeper := paramskeeper.NewKeeper(
		cdc,
		legacyAmino,
		storetypes.NewKVStoreKey(paramstypes.StoreKey),
		storetypes.NewTransientStoreKey(paramstypes.TStoreKey),
	)
	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	subspace := func(m string) paramstypes.Subspace {
		r, ok := paramsKeeper.GetSubspace(m)
		require.True(t, ok)
		return r
	}
	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
	}
	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(storetypes.NewKVStoreKey(authtypes.StoreKey)), // target store
		// subspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, // prototype
		maccPerms,
		authcodec.NewBech32Codec(app.Bech32PrefixAccAddr),
		app.Bech32PrefixAccAddr,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(storetypes.NewKVStoreKey(banktypes.StoreKey)),
		accountKeeper,
		// subspace(banktypes.ModuleName),
		make(map[string]bool),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)
	transferKeeper := ibctransferkeeper.NewKeeper(
		cdc,
		storetypes.NewKVStoreKey(ibctransfertypes.StoreKey),
		subspace(ibctransfertypes.ModuleName),
		// app.IBCKeeper.ChannelKeeper,
		// app.IBCKeeper.ChannelKeeper,
		// &app.IBCKeeper.PortKeeper,
		nil, nil, nil,
		accountKeeper,
		bankKeeper,
		nil, //scopedTransferKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	stakingKeeper := stakingkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storetypes.NewKVStoreKey(stakingtypes.StoreKey)),
		accountKeeper,
		bankKeeper,
		// subspace(stakingtypes.ModuleName),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(app.Bech32PrefixValAddr),
		authcodec.NewBech32Codec(app.Bech32PrefixConsAddr),
	)
	distrKeeper := distrkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storetypes.NewKVStoreKey(distrtypes.StoreKey)),
		// subspace(distrtypes.ModuleName),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	mapp := app.New(logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, 0, encodingConfig, appOpts)
	k := keeper.NewKeeper(
		nil,
		cdc,
		storeKey,
		memStoreKey,
		tStoreKey,
		clessStoreKey,
		paramsSubspace,
		accountKeeper,
		bankKeeper,
		transferKeeper,
		stakingKeeper,
		distrkeeper.NewQuerier(distrKeeper),
		nil,
		types.DefaultWasmConfig(),
		app.DefaultNodeHome,
		app.BaseDenom,
		app.MakeEncodingConfig().InterfaceRegistry,
		nil,
		nil,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		mapp,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
