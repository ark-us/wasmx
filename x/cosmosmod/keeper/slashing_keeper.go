package keeper

import (
	"fmt"

	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	networkkeeper "mythos/v1/x/network/keeper"

	"mythos/v1/x/cosmosmod/types"
)

type (
	KeeperSlashing struct {
		jsoncdc           codec.JSONCodec
		cdc               codec.Codec
		storeKey          storetypes.StoreKey
		paramstore        paramtypes.Subspace
		InterfaceRegistry cdctypes.InterfaceRegistry
		sk                *KeeperStaking
		WasmxKeeper       types.WasmxKeeper
		NetworkKeeper     networkkeeper.Keeper
		actionExecutor    *networkkeeper.ActionExecutor

		// the address capable of executing messages through governance. Typically, this
		// should be the x/gov module account.
		authority string

		validatorAddressCodec addresscodec.Codec
		consensusAddressCodec addresscodec.Codec
	}
)

func NewKeeperSlashing(
	jsoncdc codec.JSONCodec,
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	sk *KeeperStaking,
	wasmxKeeper types.WasmxKeeper,
	networkKeeper networkkeeper.Keeper,
	actionExecutor *networkkeeper.ActionExecutor,
	authority string,
) *KeeperSlashing {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	keeper := &KeeperSlashing{
		jsoncdc:        jsoncdc,
		cdc:            cdc,
		storeKey:       storeKey,
		paramstore:     ps,
		sk:             sk,
		WasmxKeeper:    wasmxKeeper,
		NetworkKeeper:  networkKeeper,
		actionExecutor: actionExecutor,
		authority:      authority,
	}
	return keeper
}

func (k *KeeperSlashing) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.SlashingModuleName()))
}

// GetAuthority returns the module's authority.
func (k *KeeperSlashing) GetAuthority() string {
	return k.authority
}

func (k *KeeperSlashing) JSONCodec() codec.JSONCodec {
	return k.jsoncdc
}
