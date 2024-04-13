package keeper

import (
	"fmt"

	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	networkkeeper "mythos/v1/x/network/keeper"

	"mythos/v1/x/cosmosmod/types"
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

		validatorAddressCodec addresscodec.Codec
		consensusAddressCodec addresscodec.Codec
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
	validatorAddressCodec addresscodec.Codec,
	consensusAddressCodec addresscodec.Codec,
) *KeeperBank {
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
	}
	return keeper
}

func (k *KeeperBank) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.BankModuleName()))
}

// GetAuthority returns the module's authority.
func (k *KeeperBank) GetAuthority() string {
	return k.authority
}

func (k *KeeperBank) JSONCodec() codec.JSONCodec {
	return k.jsoncdc
}
