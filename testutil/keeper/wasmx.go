package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	app "mythos/v1/app"
	"mythos/v1/x/wasmx/keeper"
	"mythos/v1/x/wasmx/types"

	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

func WasmxKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	encodingConfig := app.MakeEncodingConfig()
	_, legacyAmino := encodingConfig.Marshaler, encodingConfig.Amino

	paramsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"WasmxParams",
	)
	paramsKeeper := paramskeeper.NewKeeper(
		cdc,
		legacyAmino,
		sdk.NewKVStoreKey(paramstypes.StoreKey),
		sdk.NewTransientStoreKey(paramstypes.TStoreKey),
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
		sdk.NewKVStoreKey(authtypes.StoreKey), // target store
		subspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, // prototype
		maccPerms,
		app.Bech32PrefixAccAddr,
	)
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		sdk.NewKVStoreKey(banktypes.StoreKey),
		accountKeeper,
		subspace(banktypes.ModuleName),
		make(map[string]bool),
	)
	transferKeeper := ibctransferkeeper.NewKeeper(
		cdc,
		sdk.NewKVStoreKey(ibctransfertypes.StoreKey),
		subspace(ibctransfertypes.ModuleName),
		// app.IBCKeeper.ChannelKeeper,
		// app.IBCKeeper.ChannelKeeper,
		// &app.IBCKeeper.PortKeeper,
		nil, nil, nil,
		accountKeeper,
		bankKeeper,
		nil, //scopedTransferKeeper,
	)
	stakingKeeper := stakingkeeper.NewKeeper(
		cdc,
		sdk.NewKVStoreKey(stakingtypes.StoreKey),
		accountKeeper,
		bankKeeper,
		subspace(stakingtypes.ModuleName),
	)
	distrKeeper := distrkeeper.NewKeeper(
		cdc,
		sdk.NewKVStoreKey(distrtypes.StoreKey),
		subspace(distrtypes.ModuleName),
		accountKeeper,
		bankKeeper,
		&stakingKeeper,
		authtypes.FeeCollectorName,
	)
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		accountKeeper,
		bankKeeper,
		transferKeeper,
		stakingKeeper,
		distrKeeper,
		nil,
		types.DefaultWasmConfig(),
		app.DefaultNodeHome,
		app.BaseDenom,
		app.MakeEncodingConfig().InterfaceRegistry,
		nil,
		nil,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
