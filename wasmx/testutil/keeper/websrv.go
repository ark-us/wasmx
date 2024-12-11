package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	mcodec "wasmx/v1/codec"
	"wasmx/v1/x/websrv/keeper"
	"wasmx/v1/x/websrv/types"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"

	config "wasmx/v1/config"
)

func WebsrvKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	chainCfg, err := config.GetChainConfig(config.MYTHOS_CHAIN_ID_TEST)
	if err != nil {
		panic(err)
	}
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	addrCodec := mcodec.NewAccBech32Codec(chainCfg.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)
	govAddr, err := addrCodec.BytesToString(authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)

	paramsSubspace := typesparams.NewSubspace(cdc,
		codec.NewLegacyAmino(),
		storeKey,
		memStoreKey,
		"WebsrvParams",
	)
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		nil,
		nil,
		govAddr,
		addrCodec,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
