package keeper

import (
	"fmt"

	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	networkkeeper "mythos/v1/x/network/keeper"

	"mythos/v1/x/cosmosmod/types"
)

type (
	KeeperDistribution struct {
		jsoncdc           codec.JSONCodec
		cdc               codec.Codec
		storeKey          storetypes.StoreKey
		paramstore        paramtypes.Subspace
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

		validatorAddressCodec addresscodec.Codec
		consensusAddressCodec addresscodec.Codec
	}
)

func NewKeeperDistribution(
	jsoncdc codec.JSONCodec,
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bk *KeeperBank,
	sk *KeeperStaking,
	wasmxKeeper types.WasmxKeeper,
	networkKeeper networkkeeper.Keeper,
	actionExecutor *networkkeeper.ActionExecutor,
	authority string,
	feeCollectorName string,
	interfaceRegistry cdctypes.InterfaceRegistry,
	validatorAddressCodec addresscodec.Codec,
	consensusAddressCodec addresscodec.Codec,
) *KeeperDistribution {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	keeper := &KeeperDistribution{
		jsoncdc:               jsoncdc,
		cdc:                   cdc,
		storeKey:              storeKey,
		paramstore:            ps,
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
	}
	return keeper
}

func (k *KeeperDistribution) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.DistributionModuleName()))
}

// GetAuthority returns the module's authority.
func (k *KeeperDistribution) GetAuthority() string {
	return k.authority
}

func (k *KeeperDistribution) JSONCodec() codec.JSONCodec {
	return k.jsoncdc
}
