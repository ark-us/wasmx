package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"mythos/v1/x/cosmosmod/types"
)

// QuerierGov is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type QuerierGov struct {
	Keeper *KeeperGov
}

var _ types.QueryGovServer = QuerierGov{}

func NewQuerierGov(keeper *KeeperGov) QuerierGov {
	return QuerierGov{Keeper: keeper}
}

func (k QuerierGov) Proposal(goCtx context.Context, req *govtypes.QueryProposalRequest) (*govtypes.QueryProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.Proposal(ctx, req)
}

func (k QuerierGov) Proposals(goCtx context.Context, req *govtypes.QueryProposalsRequest) (*govtypes.QueryProposalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.Proposals(ctx, req)
}

func (k QuerierGov) Vote(goCtx context.Context, req *govtypes.QueryVoteRequest) (*govtypes.QueryVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.Vote(ctx, req)
}

func (k QuerierGov) Votes(goCtx context.Context, req *govtypes.QueryVotesRequest) (*govtypes.QueryVotesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.Votes(ctx, req)
}

func (k QuerierGov) Params(goCtx context.Context, req *govtypes.QueryParamsRequest) (*govtypes.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.Params(ctx, req)
}

func (k QuerierGov) Deposit(goCtx context.Context, req *govtypes.QueryDepositRequest) (*govtypes.QueryDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.Deposit(ctx, req)
}

func (k QuerierGov) Deposits(goCtx context.Context, req *govtypes.QueryDepositsRequest) (*govtypes.QueryDepositsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.Deposits(ctx, req)
}

func (k QuerierGov) TallyResult(goCtx context.Context, req *govtypes.QueryTallyResultRequest) (*govtypes.QueryTallyResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.TallyResult(ctx, req)
}
