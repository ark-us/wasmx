package keeper

import (
	"fmt"

	address "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	networkkeeper "mythos/v1/x/network/keeper"

	"mythos/v1/x/cosmosmod/types"
)

type (
	KeeperSlashing struct {
		jsoncdc           codec.JSONCodec
		cdc               codec.Codec
		InterfaceRegistry cdctypes.InterfaceRegistry
		sk                *KeeperStaking
		WasmxKeeper       types.WasmxKeeper
		NetworkKeeper     networkkeeper.Keeper
		actionExecutor    *networkkeeper.ActionExecutor

		// the address capable of executing messages through governance. Typically, this
		// should be the x/gov module account.
		authority string

		addressCodec address.Codec
	}
)

func NewKeeperSlashing(
	jsoncdc codec.JSONCodec,
	cdc codec.Codec,
	sk *KeeperStaking,
	wasmxKeeper types.WasmxKeeper,
	networkKeeper networkkeeper.Keeper,
	actionExecutor *networkkeeper.ActionExecutor,
	authority string,
	addressCodec address.Codec,
) *KeeperSlashing {
	keeper := &KeeperSlashing{
		jsoncdc:        jsoncdc,
		cdc:            cdc,
		sk:             sk,
		WasmxKeeper:    wasmxKeeper,
		NetworkKeeper:  networkKeeper,
		actionExecutor: actionExecutor,
		authority:      authority,
		addressCodec:   addressCodec,
	}
	return keeper
}

func (k *KeeperSlashing) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.SlashingModuleName()), "chain_id", ctx.ChainID())
}

// GetAuthority returns the module's authority.
func (k *KeeperSlashing) GetAuthority() string {
	return k.authority
}

func (k *KeeperSlashing) JSONCodec() codec.JSONCodec {
	return k.jsoncdc
}

func (k *KeeperSlashing) AddressCodec() address.Codec {
	return k.addressCodec
}
