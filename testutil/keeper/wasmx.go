package keeper

import (
	"testing"

	"wasmx/app"
	"wasmx/x/wasmx/keeper"
	"wasmx/x/wasmx/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
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
	subspace := func(m string) paramstypes.Subspace {
		r, ok := paramsKeeper.GetSubspace(m)
		require.True(t, ok)
		return r
	}
	maccPerms := map[string][]string{}
	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		sdk.NewKVStoreKey(authtypes.StoreKey), // target store
		subspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, // prototype
		maccPerms,
		sdk.Bech32PrefixAccAddr,
	)
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		sdk.NewKVStoreKey(banktypes.StoreKey),
		accountKeeper,
		subspace(banktypes.ModuleName),
		make(map[string]bool),
	)
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		accountKeeper,
		bankKeeper,
		types.DefaultWasmConfig(),
		app.DefaultNodeHome,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
