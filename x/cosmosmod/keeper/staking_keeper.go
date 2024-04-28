package keeper

import (
	"fmt"

	"cosmossdk.io/core/address"
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	networkkeeper "mythos/v1/x/network/keeper"

	"mythos/v1/x/cosmosmod/types"
)

type (
	KeeperStaking struct {
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

		addressCodec          address.Codec
		validatorAddressCodec addresscodec.Codec
		consensusAddressCodec addresscodec.Codec
	}
)

func NewKeeperStaking(
	jsoncdc codec.JSONCodec,
	cdc codec.Codec,
	accountKeeper types.AccountKeeper,
	wasmxKeeper types.WasmxKeeper,
	networkKeeper networkkeeper.Keeper,
	actionExecutor *networkkeeper.ActionExecutor,
	authority string,
	interfaceRegistry cdctypes.InterfaceRegistry,
	validatorAddressCodec addresscodec.Codec,
	consensusAddressCodec addresscodec.Codec,
	addressCodec address.Codec,
) *KeeperStaking {
	keeper := &KeeperStaking{
		jsoncdc:               jsoncdc,
		cdc:                   cdc,
		ak:                    accountKeeper,
		WasmxKeeper:           wasmxKeeper,
		NetworkKeeper:         networkKeeper,
		actionExecutor:        actionExecutor,
		authority:             authority,
		InterfaceRegistry:     interfaceRegistry,
		addressCodec:          addressCodec,
		validatorAddressCodec: validatorAddressCodec,
		consensusAddressCodec: consensusAddressCodec,
	}
	return keeper
}

func (k *KeeperStaking) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.StakingModuleName()), "chain_id", ctx.ChainID())
}

// GetAuthority returns the module's authority.
func (k *KeeperStaking) GetAuthority() string {
	return k.authority
}

func (k *KeeperStaking) JSONCodec() codec.JSONCodec {
	return k.jsoncdc
}

func (k *KeeperStaking) AddressCodec() address.Codec {
	return k.addressCodec
}

func (k *KeeperStaking) ValidatorAddressCodec() address.Codec {
	return k.validatorAddressCodec
}

func (k *KeeperStaking) ConsensusAddressCodec() address.Codec {
	return k.consensusAddressCodec
}
