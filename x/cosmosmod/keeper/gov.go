package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (k Keeper) Proposal(ctx sdk.Context, req *govtypes.QueryProposalRequest) (*govtypes.QueryProposalResponse, error) {
	k.Logger(ctx).Error("Proposal not implemented")
	return &govtypes.QueryProposalResponse{}, nil
}

func (k Keeper) Proposals(ctx sdk.Context, req *govtypes.QueryProposalsRequest) (*govtypes.QueryProposalsResponse, error) {
	k.Logger(ctx).Error("Proposals not implemented")
	return &govtypes.QueryProposalsResponse{}, nil
}

func (k Keeper) Vote(ctx sdk.Context, req *govtypes.QueryVoteRequest) (*govtypes.QueryVoteResponse, error) {
	k.Logger(ctx).Error("Vote not implemented")
	return &govtypes.QueryVoteResponse{}, nil
}

func (k Keeper) Votes(ctx sdk.Context, req *govtypes.QueryVotesRequest) (*govtypes.QueryVotesResponse, error) {
	k.Logger(ctx).Error("Votes not implemented")
	return &govtypes.QueryVotesResponse{}, nil
}

func (k Keeper) Params(ctx sdk.Context, req *govtypes.QueryParamsRequest) (*govtypes.QueryParamsResponse, error) {
	k.Logger(ctx).Error("Params not implemented")
	return &govtypes.QueryParamsResponse{}, nil
}

func (k Keeper) Deposit(ctx sdk.Context, req *govtypes.QueryDepositRequest) (*govtypes.QueryDepositResponse, error) {
	k.Logger(ctx).Error("Deposit not implemented")
	return &govtypes.QueryDepositResponse{}, nil
}

func (k Keeper) Deposits(ctx sdk.Context, req *govtypes.QueryDepositsRequest) (*govtypes.QueryDepositsResponse, error) {
	k.Logger(ctx).Error("Deposits not implemented")
	return &govtypes.QueryDepositsResponse{}, nil
}

func (k Keeper) TallyResult(ctx sdk.Context, req *govtypes.QueryTallyResultRequest) (*govtypes.QueryTallyResultResponse, error) {
	k.Logger(ctx).Error("TallyResult not implemented")
	return &govtypes.QueryTallyResultResponse{}, nil
}
