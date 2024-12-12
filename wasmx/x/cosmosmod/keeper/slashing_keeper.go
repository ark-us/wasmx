package keeper

import (
	"fmt"

	address "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/v1/codec"
	networkkeeper "github.com/loredanacirstea/wasmx/v1/x/network/keeper"

	"github.com/loredanacirstea/wasmx/v1/x/cosmosmod/types"
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

		addressCodec   address.Codec
		accBech32Codec mcodec.AccBech32Codec
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
	accBech32Codec := mcodec.MustUnwrapAccBech32Codec(addressCodec)
	keeper := &KeeperSlashing{
		jsoncdc:        jsoncdc,
		cdc:            cdc,
		sk:             sk,
		WasmxKeeper:    wasmxKeeper,
		NetworkKeeper:  networkKeeper,
		actionExecutor: actionExecutor,
		authority:      authority,
		addressCodec:   addressCodec,
		accBech32Codec: accBech32Codec,
	}
	return keeper
}

func (k *KeeperSlashing) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With(log.ModuleKey, fmt.Sprintf("x/%s", types.SlashingModuleName()), "chain_id", ctx.ChainID())
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

func (k *KeeperSlashing) AccBech32Codec() mcodec.AccBech32Codec {
	return k.accBech32Codec
}
