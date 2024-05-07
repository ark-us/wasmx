package keeper

import (
	"context"
	"fmt"

	sdkerr "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "mythos/v1/codec"
	cfg "mythos/v1/config"
	"mythos/v1/x/network/types"
)

var _ types.QueryServer = &Keeper{}

func (k *Keeper) QueryMultiChain(goCtx context.Context, req *types.QueryMultiChainRequest) (*types.QueryMultiChainResponse, error) {
	abciReq, err := mcodec.RequestQueryFromBz(req.QueryData)

	multichainapp, err := cfg.GetMultiChainApp(k.goContextParent)
	if err != nil {
		return nil, err
	}
	iapp, err := multichainapp.GetApp(req.MultiChainId)
	if err != nil {
		return nil, err
	}
	app, ok := iapp.(cfg.MythosApp)
	if !ok {
		return nil, fmt.Errorf("error App interface from multichainapp")
	}

	resp, err := app.Query(context.TODO(), &abciReq)
	if err != nil {
		return nil, err
	}
	if resp != nil && resp.Code != 0 {
		return nil, sdkerr.ABCIError(resp.Codespace, resp.Code, resp.Log)
	}
	res := resp.Value

	return &types.QueryMultiChainResponse{
		Data: res,
	}, nil
}

// TODO remove this, because we should use QueryMultiChain
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
