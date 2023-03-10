package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"wasmx/x/websrv/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) HttpGet(c context.Context, req *types.QueryHttpGetRequest) (*types.QueryHttpGetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var request types.HttpRequestGet
	k.cdc.Unmarshal(req.HttpRequest, &request)

	rsp, err := k.HttpGetInternal(sdk.UnwrapSDKContext(c), request)
	if err != nil {
		return nil, err
	}
	return &types.QueryHttpGetResponse{Data: rsp}, nil
}

func (k Keeper) ContractByRoute(c context.Context, req *types.QueryContractByRouteRequest) (*types.QueryContractByRouteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Path == "" {
		return nil, types.ErrEmptyRoute
	}
	contractAddress := k.GetRouteToContract(sdk.UnwrapSDKContext(c), req.Path)
	return &types.QueryContractByRouteResponse{
		ContractAddress: contractAddress.String(),
	}, nil
}

func (k Keeper) RouteByContract(c context.Context, req *types.QueryRouteByContractRequest) (*types.QueryRouteByContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	contractAddress, err := sdk.AccAddressFromBech32(req.ContractAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}
	path := k.GetContractToRoute(sdk.UnwrapSDKContext(c), contractAddress)
	return &types.QueryRouteByContractResponse{
		Path: path,
	}, nil
}
