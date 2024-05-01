package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// QuerierBank is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type QuerierAuth struct {
	Keeper *KeeperAuth
}

var _ authtypes.QueryServer = QuerierAuth{}

func NewQuerierAuth(keeper *KeeperAuth) QuerierAuth {
	return QuerierAuth{Keeper: keeper}
}

func (k QuerierAuth) Accounts(goCtx context.Context, req *authtypes.QueryAccountsRequest) (*authtypes.QueryAccountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.Accounts not implemented")
	return &authtypes.QueryAccountsResponse{}, nil
}

func (k QuerierAuth) Account(goCtx context.Context, req *authtypes.QueryAccountRequest) (*authtypes.QueryAccountResponse, error) {
	addr, err := k.Keeper.accBech32Codec.StringToAccAddressPrefixed(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	account, err := k.Keeper.GetAccountPrefixed(goCtx, addr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	any, err := codectypes.NewAnyWithValue(account)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &authtypes.QueryAccountResponse{Account: any}, nil
}

func (k QuerierAuth) AccountAddressByID(goCtx context.Context, req *authtypes.QueryAccountAddressByIDRequest) (*authtypes.QueryAccountAddressByIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.Logger(ctx).Error("Auth.AccountAddressByID not implemented")
	return &authtypes.QueryAccountAddressByIDResponse{}, nil
}

func (k QuerierAuth) Params(goCtx context.Context, req *authtypes.QueryParamsRequest) (*authtypes.QueryParamsResponse, error) {
	params := k.Keeper.GetParams(goCtx)
	return &authtypes.QueryParamsResponse{Params: params}, nil
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
