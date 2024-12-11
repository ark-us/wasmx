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
	KeeperDistribution struct {
		jsoncdc           codec.JSONCodec
		cdc               codec.Codec
		InterfaceRegistry cdctypes.InterfaceRegistry
		ak                types.AccountKeeper
		bk                *KeeperBank
		sk                *KeeperStaking
		WasmxKeeper       types.WasmxKeeper
		NetworkKeeper     networkkeeper.Keeper
		actionExecutor    *networkkeeper.ActionExecutor

		// the address capable of executing messages through governance. Typically, this
		// should be the x/gov module account.
		authority        string
		feeCollectorName string

		validatorAddressCodec address.Codec
		consensusAddressCodec address.Codec
		addressCodec          address.Codec
		accBech32Codec        mcodec.AccBech32Codec
	}
)

func NewKeeperDistribution(
	jsoncdc codec.JSONCodec,
	cdc codec.Codec,
	accountKeeper types.AccountKeeper,
	bk *KeeperBank,
	sk *KeeperStaking,
	wasmxKeeper types.WasmxKeeper,
	networkKeeper networkkeeper.Keeper,
	actionExecutor *networkkeeper.ActionExecutor,
	authority string,
	feeCollectorName string,
	interfaceRegistry cdctypes.InterfaceRegistry,
	validatorAddressCodec address.Codec,
	consensusAddressCodec address.Codec,
	addressCodec address.Codec,
) *KeeperDistribution {
	accBech32Codec := mcodec.MustUnwrapAccBech32Codec(addressCodec)
	keeper := &KeeperDistribution{
		jsoncdc:               jsoncdc,
		cdc:                   cdc,
		ak:                    accountKeeper,
		bk:                    bk,
		sk:                    sk,
		WasmxKeeper:           wasmxKeeper,
		NetworkKeeper:         networkKeeper,
		actionExecutor:        actionExecutor,
		authority:             authority,
		feeCollectorName:      feeCollectorName,
		InterfaceRegistry:     interfaceRegistry,
		validatorAddressCodec: validatorAddressCodec,
		consensusAddressCodec: consensusAddressCodec,
		addressCodec:          addressCodec,
		accBech32Codec:        accBech32Codec,
	}
	return keeper
}

func (k *KeeperDistribution) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With(log.ModuleKey, fmt.Sprintf("x/%s", types.DistributionModuleName()), "chain_id", ctx.ChainID())
}

// GetAuthority returns the module's authority.
func (k *KeeperDistribution) GetAuthority() string {
	return k.authority
}

func (k *KeeperDistribution) JSONCodec() codec.JSONCodec {
	return k.jsoncdc
}

func (k *KeeperDistribution) AddressCodec() address.Codec {
	return k.addressCodec
}

func (k *KeeperDistribution) ValidatorAddressCodec() address.Codec {
	return k.validatorAddressCodec
}

func (k *KeeperDistribution) ConsensusAddressCodec() address.Codec {
	return k.consensusAddressCodec
}

func (k *KeeperDistribution) AccBech32Codec() mcodec.AccBech32Codec {
	return k.accBech32Codec
}
