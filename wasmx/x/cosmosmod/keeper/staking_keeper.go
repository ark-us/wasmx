package keeper

import (
	"fmt"

	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/v1/codec"
	networkkeeper "github.com/loredanacirstea/wasmx/v1/x/network/keeper"

	"github.com/loredanacirstea/wasmx/v1/x/cosmosmod/types"
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

		addressCodec          addresscodec.Codec
		validatorAddressCodec addresscodec.Codec
		consensusAddressCodec addresscodec.Codec
		accBech32Codec        mcodec.AccBech32Codec
		valBech32Codec        mcodec.ValBech32Codec
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
	addressCodec addresscodec.Codec,
) *KeeperStaking {
	accBech32Codec := mcodec.MustUnwrapAccBech32Codec(addressCodec)
	valBech32Codec := mcodec.MustUnwrapValBech32Codec(validatorAddressCodec)
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
		accBech32Codec:        accBech32Codec,
		valBech32Codec:        valBech32Codec,
	}
	return keeper
}

func (k *KeeperStaking) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With(log.ModuleKey, fmt.Sprintf("x/%s", types.StakingModuleName()), "chain_id", ctx.ChainID())
}

// GetAuthority returns the module's authority.
func (k *KeeperStaking) GetAuthority() string {
	return k.authority
}

func (k *KeeperStaking) JSONCodec() codec.JSONCodec {
	return k.jsoncdc
}

func (k *KeeperStaking) AddressCodec() addresscodec.Codec {
	return k.addressCodec
}

func (k *KeeperStaking) ValidatorAddressCodec() addresscodec.Codec {
	return k.validatorAddressCodec
}

func (k *KeeperStaking) ConsensusAddressCodec() addresscodec.Codec {
	return k.consensusAddressCodec
}

func (k *KeeperStaking) AccBech32Codec() mcodec.AccBech32Codec {
	return k.accBech32Codec
}
