package keeper

import (
	"fmt"

	address "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "wasmx/v1/codec"
	networkkeeper "wasmx/v1/x/network/keeper"

	"wasmx/v1/x/cosmosmod/types"
)

type (
	KeeperGov struct {
		jsoncdc           codec.JSONCodec
		cdc               codec.Codec
		InterfaceRegistry cdctypes.InterfaceRegistry
		ak                types.AccountKeeper
		WasmxKeeper       types.WasmxKeeper
		NetworkKeeper     networkkeeper.Keeper
		actionExecutor    *networkkeeper.ActionExecutor

		// the address capable of executing messages through governance. Typically, this
		// should be the x/gov module account.
		authority string

		validatorAddressCodec address.Codec
		consensusAddressCodec address.Codec
		addressCodec          address.Codec
		accBech32Codec        mcodec.AccBech32Codec
	}
)

func NewKeeperGov(
	jsoncdc codec.JSONCodec,
	cdc codec.Codec,
	accountKeeper types.AccountKeeper,
	wasmxKeeper types.WasmxKeeper,
	networkKeeper networkkeeper.Keeper,
	actionExecutor *networkkeeper.ActionExecutor,
	authority string,
	interfaceRegistry cdctypes.InterfaceRegistry,
	validatorAddressCodec address.Codec,
	consensusAddressCodec address.Codec,
	addressCodec address.Codec,
) *KeeperGov {
	accBech32Codec := mcodec.MustUnwrapAccBech32Codec(addressCodec)
	keeper := &KeeperGov{
		jsoncdc:               jsoncdc,
		cdc:                   cdc,
		ak:                    accountKeeper,
		WasmxKeeper:           wasmxKeeper,
		NetworkKeeper:         networkKeeper,
		actionExecutor:        actionExecutor,
		authority:             authority,
		InterfaceRegistry:     interfaceRegistry,
		validatorAddressCodec: validatorAddressCodec,
		consensusAddressCodec: consensusAddressCodec,
		accBech32Codec:        accBech32Codec,
	}
	return keeper
}

func (k *KeeperGov) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With(log.ModuleKey, fmt.Sprintf("x/%s", types.GovModuleName()), "chain_id", ctx.ChainID())
}

// GetAuthority returns the module's authority.
func (k *KeeperGov) GetAuthority() string {
	return k.authority
}

func (k *KeeperGov) JSONCodec() codec.JSONCodec {
	return k.jsoncdc
}

func (k *KeeperGov) AddressCodec() address.Codec {
	return k.addressCodec
}

func (k *KeeperGov) ValidatorAddressCodec() address.Codec {
	return k.validatorAddressCodec
}

func (k *KeeperGov) ConsensusAddressCodec() address.Codec {
	return k.consensusAddressCodec
}

func (k *KeeperGov) AccBech32Codec() mcodec.AccBech32Codec {
	return k.accBech32Codec
}
