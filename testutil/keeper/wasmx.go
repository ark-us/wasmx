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
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	mcodec "mythos/v1/codec"
	config "mythos/v1/config"
	networkkeeper "mythos/v1/x/network/keeper"
	networkvm "mythos/v1/x/network/vm"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

func WasmxKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	tStoreKey := storetypes.NewTransientStoreKey(types.TStoreKey)
	mStoreKey := storetypes.NewConsensuslessStoreKey(types.MetaConsensusStoreKey)
	sStoreKey := storetypes.NewConsensuslessStoreKey(types.SingleConsensusStoreKey)

	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	stateStore := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	stateStore.MountStoreWithDB(tStoreKey, storetypes.StoreTypeTransient, db)
	stateStore.MountStoreWithDB(mStoreKey, storetypes.StoreTypeConsensusless, db)
	stateStore.MountStoreWithDB(sStoreKey, storetypes.StoreTypeConsensusless, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	chainCfg, err := config.GetChainConfig(config.MYTHOS_CHAIN_ID_TEST)
	if err != nil {
		panic(err)
	}
	encodingConfig := app.MakeEncodingConfig(chainCfg)
	_, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino

	appOpts := app.DefaultAppOptions{}
	g, goctx, _ := app.GetTestCtx(logger, true)
	goctx = wasmxtypes.ContextWithBackgroundProcesses(goctx)
	goctx = networkvm.WithP2PEmptyContext(goctx)
	goctx, bapps := config.WithMultiChainAppEmpty(goctx)
	appOpts.Set("goroutineGroup", g)
	appOpts.Set("goContextParent", goctx)

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
	// subspace := func(m string) paramstypes.Subspace {
	// 	r, ok := paramsKeeper.GetSubspace(m)
	// 	require.True(t, ok)
	// 	return r
	// }
	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
	}
	permAddrs := make(map[string]authtypes.PermissionsForAddress)
	for name, perms := range maccPerms {
		permAddrs[name] = authtypes.NewPermissionsForAddress(name, perms)
	}
	valCodec := mcodec.NewValBech32Codec(chainCfg.Bech32PrefixValAddr, mcodec.NewAddressPrefixedFromVal)
	consCodec := mcodec.NewConsBech32Codec(chainCfg.Bech32PrefixConsAddr, mcodec.NewAddressPrefixedFromCons)
	addrCodec := mcodec.NewAccBech32Codec(chainCfg.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)

	govAddr, err := addrCodec.BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	// transferKeeper := ibctransferkeeper.NewKeeper(
	// 	cdc,
	// 	storetypes.NewKVStoreKey(ibctransfertypes.StoreKey),
	// 	subspace(ibctransfertypes.ModuleName),
	// 	// app.IBCKeeper.ChannelKeeper,
	// 	// app.IBCKeeper.ChannelKeeper,
	// 	// &app.IBCKeeper.PortKeeper,
	// 	nil, nil, nil,
	// 	accountKeeper,
	// 	bankKeeper,
	// 	nil, //scopedTransferKeeper,
	// 	govAddr,
	// )
	actionExecutor := networkkeeper.NewActionExecutor(bapps, logger)
	mapp := app.NewApp(actionExecutor, logger, db, nil, true, map[int64]bool{}, app.DefaultNodeHome, 0, encodingConfig, appOpts)
	k := keeper.NewKeeper(
		g,
		goctx,
		cdc,
		encodingConfig.TxConfig,
		storeKey,
		memStoreKey,
		tStoreKey,
		mStoreKey,
		sStoreKey,
		paramsSubspace,
		// transferKeeper,
		// stakingKeeper,
		// distrkeeper.NewQuerier(distrKeeper),
		// nil,
		types.DefaultWasmConfig(),
		app.DefaultNodeHome,
		config.BaseDenom,
		permAddrs,
		app.MakeEncodingConfig(chainCfg).InterfaceRegistry,
		nil,
		nil,
		govAddr,
		valCodec,
		consCodec,
		addrCodec,
		mapp,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
