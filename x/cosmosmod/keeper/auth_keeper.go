package keeper

import (
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	networkkeeper "mythos/v1/x/network/keeper"

	mcodec "mythos/v1/codec"
	"mythos/v1/x/cosmosmod/types"
)

type (
	KeeperAuth struct {
		jsoncdc           codec.JSONCodec
		cdc               codec.Codec
		InterfaceRegistry cdctypes.InterfaceRegistry
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
		permAddrs             map[string]authtypes.PermissionsForAddress
	}
)

func NewKeeperAuth(
	jsoncdc codec.JSONCodec,
	cdc codec.Codec,
	wasmxKeeper types.WasmxKeeper,
	networkKeeper networkkeeper.Keeper,
	actionExecutor *networkkeeper.ActionExecutor,
	authority string,
	interfaceRegistry cdctypes.InterfaceRegistry,
	validatorAddressCodec address.Codec,
	consensusAddressCodec address.Codec,
	addressCodec address.Codec,
	permAddrs map[string]authtypes.PermissionsForAddress,
) *KeeperAuth {
	accBech32Codec := mcodec.MustUnwrapAccBech32Codec(addressCodec)

	keeper := &KeeperAuth{
		jsoncdc:               jsoncdc,
		cdc:                   cdc,
		WasmxKeeper:           wasmxKeeper,
		NetworkKeeper:         networkKeeper,
		actionExecutor:        actionExecutor,
		authority:             authority,
		InterfaceRegistry:     interfaceRegistry,
		validatorAddressCodec: validatorAddressCodec,
		consensusAddressCodec: consensusAddressCodec,
		addressCodec:          addressCodec,
		accBech32Codec:        accBech32Codec,
		permAddrs:             permAddrs,
	}
	return keeper
}

func (k *KeeperAuth) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.AuthModuleName()), "chain_id", ctx.ChainID())
}

// GetAuthority returns the module's authority.
func (k *KeeperAuth) GetAuthority() string {
	return k.authority
}

func (k *KeeperAuth) JSONCodec() codec.JSONCodec {
	return k.jsoncdc
}

func (k *KeeperAuth) AddressCodec() address.Codec {
	return k.addressCodec
}

func (k *KeeperAuth) ValidatorAddressCodec() address.Codec {
	return k.validatorAddressCodec
}

func (k *KeeperAuth) ConsensusAddressCodec() address.Codec {
	return k.consensusAddressCodec
}
