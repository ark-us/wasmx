package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (k KeeperAuth) SetParams(ctx sdk.Context, params authtypes.Params) error {
	k.Logger(ctx).Error("Auth.Params not implemented")
	return nil
}

func (k KeeperAuth) GetParams(goCtx context.Context) (params authtypes.Params) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.GetParams not implemented")
	return params
}

func (k KeeperAuth) AddressCodec() address.Codec {
	return k.addressCodec
}

func (k KeeperAuth) HasAccount(goCtx context.Context, addr sdk.AccAddress) bool {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.HasAccount not implemented")
	return true
}

func (k KeeperAuth) GetAccount(goCtx context.Context, addr sdk.AccAddress) sdk.AccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.GetAccount not implemented")
	return nil
}

func (k KeeperAuth) SetAccount(goCtx context.Context, acc sdk.AccountI) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.SetAccount not implemented")
}

func (k KeeperAuth) NewAccountWithAddress(goCtx context.Context, addr sdk.AccAddress) sdk.AccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.NewAccountWithAddress not implemented")
	return nil
}

func (k KeeperAuth) GetModuleAddress(moduleName string) sdk.AccAddress {
	fmt.Println("Auth.GetModuleAddress not implemented")
	// permAddr, ok := k.permAddrs[moduleName]
	// if !ok {
	// 	return nil
	// }
	// return permAddr.GetAddress()
	return nil
}

func (k KeeperAuth) GetModuleAccount(goCtx context.Context, moduleName string) sdk.ModuleAccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.GetModuleAccount not implemented")
	return nil
}

func (k KeeperAuth) SetModuleAccount(goCtx context.Context, macc sdk.ModuleAccountI) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.SetModuleAccount not implemented")
}

func (k KeeperAuth) NewAccount(goCtx context.Context, acc sdk.AccountI) sdk.AccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.NewAccount not implemented")
	return nil
}

func (k KeeperAuth) RemoveAccount(goCtx context.Context, acc sdk.AccountI) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.RemoveAccount not implemented")
}

func (k KeeperAuth) IterateAccounts(goCtx context.Context, cb func(account sdk.AccountI) (stop bool)) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.IterateAccounts not implemented")
}

func (k KeeperAuth) GetSequence(goCtx context.Context, addr sdk.AccAddress) (uint64, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.GetSequence not implemented")
	return 0, nil
}

// func (authkeeper.AccountKeeper).AddressCodec() address.Codec
// func (authkeeper.AccountKeeper).ExportGenesis(ctx sdk.Context) *authtypes.GenesisState
// func (authkeeper.AccountKeeper).GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
// func (authkeeper.AccountKeeper).GetAllAccounts(ctx context.Context) (accounts []sdk.AccountI)
// func (authkeeper.AccountKeeper).GetAuthority() string
// func (authkeeper.AccountKeeper).GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
// func (authkeeper.AccountKeeper).GetModuleAccountAndPermissions(ctx context.Context, moduleName string) (sdk.ModuleAccountI, []string)
// func (authkeeper.AccountKeeper).GetModuleAddress(moduleName string) sdk.AccAddress
// func (authkeeper.AccountKeeper).GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string)
// func (authkeeper.AccountKeeper).GetModulePermissions() map[string]authtypes.PermissionsForAddress
// func (authkeeper.AccountKeeper).GetParams(ctx context.Context) (params authtypes.Params)
// func (authkeeper.AccountKeeper).GetPubKey(ctx context.Context, addr sdk.AccAddress) (types.PubKey, error)
// func (authkeeper.AccountKeeper).GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error)
// func (authkeeper.AccountKeeper).HasAccount(ctx context.Context, addr sdk.AccAddress) bool
// func (authkeeper.AccountKeeper).InitGenesis(ctx sdk.Context, data authtypes.GenesisState)
// func (authkeeper.AccountKeeper).IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) (stop bool))
// func (authkeeper.AccountKeeper).Logger(ctx context.Context) log.Logger
// func (authkeeper.AccountKeeper).NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
// func (authkeeper.AccountKeeper).NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
// func (authkeeper.AccountKeeper).NextAccountNumber(ctx context.Context) uint64
// func (authkeeper.AccountKeeper).RemoveAccount(ctx context.Context, acc sdk.AccountI)
// func (authkeeper.AccountKeeper).SetAccount(ctx context.Context, acc sdk.AccountI)
// func (authkeeper.AccountKeeper).SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI)
// func (authkeeper.AccountKeeper).ValidatePermissions(macc sdk.ModuleAccountI) error
