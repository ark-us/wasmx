package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/network/types"
)

var _ types.QueryServer = &Keeper{}

func (k *Keeper) ContractCall(goCtx context.Context, req *types.QueryContractCallRequest) (*types.QueryContractCallResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info("Not Implemented")

	return &types.QueryContractCallResponse{Data: make(types.RawContractMessage, 0)}, nil
}
