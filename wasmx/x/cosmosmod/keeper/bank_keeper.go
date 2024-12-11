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
	KeeperBank struct {
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

func NewKeeperBank(
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
) *KeeperBank {
	accBech32Codec := mcodec.MustUnwrapAccBech32Codec(addressCodec)

	keeper := &KeeperBank{
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
		addressCodec:          addressCodec,
		accBech32Codec:        accBech32Codec,
	}
	return keeper
}

func (k *KeeperBank) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With(log.ModuleKey, fmt.Sprintf("x/%s", types.BankModuleName()), "chain_id", ctx.ChainID())
}

// GetAuthority returns the module's authority.
func (k *KeeperBank) GetAuthority() string {
	return k.authority
}

func (k *KeeperBank) JSONCodec() codec.JSONCodec {
	return k.jsoncdc
}

func (k *KeeperBank) AddressCodec() address.Codec {
	return k.addressCodec
}

func (k *KeeperBank) ValidatorAddressCodec() address.Codec {
	return k.validatorAddressCodec
}

func (k *KeeperBank) ConsensusAddressCodec() address.Codec {
	return k.consensusAddressCodec
}

func (k *KeeperBank) AccBech32Codec() mcodec.AccBech32Codec {
	return k.accBech32Codec
}
