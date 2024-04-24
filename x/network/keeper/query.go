package keeper

import (
	"context"

	sdkerr "cosmossdk.io/errors"
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

	senderAddr, err := k.wasmxKeeper.GetAddressOrRole(ctx, req.Sender)
	if err != nil {
		return nil, sdkerr.Wrap(err, "sender")
	}
	contractAddress, err := k.wasmxKeeper.GetAddressOrRole(ctx, req.Address)
	if err != nil {
		return nil, sdkerr.Wrap(err, "contract")
	}

	resp, err := k.wasmxKeeper.Query(ctx, contractAddress, senderAddr, req.QueryData, nil, nil)
	if err != nil {
		return nil, err
	}

	return &types.QueryContractCallResponse{Data: resp}, nil
}
