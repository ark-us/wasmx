package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"mythos/v1/x/cosmosmod/types"
)

type msgGovServer struct {
	*Keeper
}

// NewMsgGovServerImpl returns an implementation of the MsgServer interface
func NewMsgGovServerImpl(keeper *Keeper) types.MsgGovServer {
	return &msgGovServer{
		Keeper: keeper,
	}
}

var _ types.MsgGovServer = msgGovServer{}

func (m msgGovServer) SubmitProposal(goCtx context.Context, msg *govtypes.MsgSubmitProposal) (*govtypes.MsgSubmitProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("SubmitProposal not implemented")
	return &govtypes.MsgSubmitProposalResponse{}, nil
}

func (m msgGovServer) Vote(goCtx context.Context, msg *govtypes.MsgVote) (*govtypes.MsgVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("Vote not implemented")
	return &govtypes.MsgVoteResponse{}, nil
}

func (m msgGovServer) VoteWeighted(goCtx context.Context, msg *govtypes.MsgVoteWeighted) (*govtypes.MsgVoteWeightedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("VoteWeighted not implemented")
	return &govtypes.MsgVoteWeightedResponse{}, nil
}

func (m msgGovServer) Deposit(goCtx context.Context, msg *govtypes.MsgDeposit) (*govtypes.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("Deposit not implemented")
	return &govtypes.MsgDepositResponse{}, nil
}
