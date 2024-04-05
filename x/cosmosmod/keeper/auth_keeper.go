package keeper

import (
	"fmt"

	"cosmossdk.io/core/address"
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	networkkeeper "mythos/v1/x/network/keeper"

	"mythos/v1/x/cosmosmod/types"
)

type (
	KeeperAuth struct {
		jsoncdc           codec.JSONCodec
		cdc               codec.Codec
		storeKey          storetypes.StoreKey
		paramstore        paramtypes.Subspace
		InterfaceRegistry cdctypes.InterfaceRegistry
		WasmxKeeper       types.WasmxKeeper
		NetworkKeeper     networkkeeper.Keeper
		actionExecutor    *networkkeeper.ActionExecutor

		// the address capable of executing messages through governance. Typically, this
		// should be the x/gov module account.
		authority string

		validatorAddressCodec addresscodec.Codec
		consensusAddressCodec addresscodec.Codec
		addressCodec          address.Codec
		permAddrs             map[string]authtypes.PermissionsForAddress
	}
)

func NewKeeperAuth(
	jsoncdc codec.JSONCodec,
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	wasmxKeeper types.WasmxKeeper,
	networkKeeper networkkeeper.Keeper,
	actionExecutor *networkkeeper.ActionExecutor,
	authority string,
	interfaceRegistry cdctypes.InterfaceRegistry,
	validatorAddressCodec addresscodec.Codec,
	consensusAddressCodec addresscodec.Codec,
	addressCodec address.Codec,
	permAddrs map[string]authtypes.PermissionsForAddress,
) *KeeperAuth {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	keeper := &KeeperAuth{
		jsoncdc:               jsoncdc,
		cdc:                   cdc,
		storeKey:              storeKey,
		paramstore:            ps,
		WasmxKeeper:           wasmxKeeper,
		NetworkKeeper:         networkKeeper,
		actionExecutor:        actionExecutor,
		authority:             authority,
		InterfaceRegistry:     interfaceRegistry,
		validatorAddressCodec: validatorAddressCodec,
		consensusAddressCodec: consensusAddressCodec,
		addressCodec:          addressCodec,
		permAddrs:             permAddrs,
	}
	return keeper
}

func (k *KeeperAuth) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.AuthModuleName()))
}

// GetAuthority returns the module's authority.
func (k *KeeperAuth) GetAuthority() string {
	return k.authority
}

func (k *KeeperAuth) JSONCodec() codec.JSONCodec {
	return k.jsoncdc
}
