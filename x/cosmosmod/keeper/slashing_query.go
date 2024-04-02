package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// QuerierSlashing is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type QuerierSlashing struct {
	Keeper *KeeperSlashing
}

var _ slashingtypes.QueryServer = QuerierSlashing{}

func NewQuerierSlashing(keeper *KeeperSlashing) QuerierSlashing {
	return QuerierSlashing{Keeper: keeper}
}

func (k QuerierSlashing) Params(goCtx context.Context, req *slashingtypes.QueryParamsRequest) (*slashingtypes.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params, err := k.Keeper.Params(ctx)
	if err != nil {
		return nil, err
	}
	return &slashingtypes.QueryParamsResponse{Params: *params}, nil
}

func (k QuerierSlashing) SigningInfo(goCtx context.Context, req *slashingtypes.QuerySigningInfoRequest) (*slashingtypes.QuerySigningInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	info, err := k.Keeper.SigningInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return &slashingtypes.QuerySigningInfoResponse{ValSigningInfo: *info}, nil
}

func (k QuerierSlashing) SigningInfos(goCtx context.Context, req *slashingtypes.QuerySigningInfosRequest) (*slashingtypes.QuerySigningInfosResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.Keeper.SigningInfos(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
