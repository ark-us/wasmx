package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"mythos/v1/x/cosmosmod/types"
)

// QuerierBank is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type QuerierAuth struct {
	*Keeper
}

var _ types.QueryAuthServer = QuerierAuth{}

func NewQuerierAuth(keeper *Keeper) QuerierAuth {
	return QuerierAuth{Keeper: keeper}
}

func (k QuerierAuth) Accounts(goCtx context.Context, req *authtypes.QueryAccountsRequest) (*authtypes.QueryAccountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.Accounts not implemented")
	return &authtypes.QueryAccountsResponse{}, nil
}

func (k QuerierAuth) Account(goCtx context.Context, req *authtypes.QueryAccountRequest) (*authtypes.QueryAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.Account not implemented")
	return &authtypes.QueryAccountResponse{}, nil
}

func (k QuerierAuth) AccountAddressByID(goCtx context.Context, req *authtypes.QueryAccountAddressByIDRequest) (*authtypes.QueryAccountAddressByIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.AccountAddressByID not implemented")
	return &authtypes.QueryAccountAddressByIDResponse{}, nil
}

func (k QuerierAuth) Params(goCtx context.Context, req *authtypes.QueryParamsRequest) (*authtypes.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.Params not implemented")
	return &authtypes.QueryParamsResponse{}, nil
}

func (k QuerierAuth) ModuleAccounts(goCtx context.Context, req *authtypes.QueryModuleAccountsRequest) (*authtypes.QueryModuleAccountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.ModuleAccounts not implemented")
	return &authtypes.QueryModuleAccountsResponse{}, nil
}

func (k QuerierAuth) ModuleAccountByName(goCtx context.Context, req *authtypes.QueryModuleAccountByNameRequest) (*authtypes.QueryModuleAccountByNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.ModuleAccountByName not implemented")
	return &authtypes.QueryModuleAccountByNameResponse{}, nil
}

func (k QuerierAuth) Bech32Prefix(goCtx context.Context, req *authtypes.Bech32PrefixRequest) (*authtypes.Bech32PrefixResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.Bech32Prefix not implemented")
	return &authtypes.Bech32PrefixResponse{}, nil
}

func (k QuerierAuth) AddressBytesToString(goCtx context.Context, req *authtypes.AddressBytesToStringRequest) (*authtypes.AddressBytesToStringResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.AddressBytesToString not implemented")
	return &authtypes.AddressBytesToStringResponse{}, nil
}

func (k QuerierAuth) AddressStringToBytes(goCtx context.Context, req *authtypes.AddressStringToBytesRequest) (*authtypes.AddressStringToBytesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.AddressStringToBytes not implemented")
	return &authtypes.AddressStringToBytesResponse{}, nil
}

func (k QuerierAuth) AccountInfo(goCtx context.Context, req *authtypes.QueryAccountInfoRequest) (*authtypes.QueryAccountInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.AccountInfo not implemented")
	return &authtypes.QueryAccountInfoResponse{}, nil
}
