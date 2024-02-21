package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"mythos/v1/x/cosmosmod/types"
	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type msgGovServer struct {
	Keeper *KeeperGov
}

// NewMsgGovServerImpl returns an implementation of the MsgServer interface
func NewMsgGovServerImpl(keeper *KeeperGov) govtypes1.MsgServer {
	return &msgGovServer{
		Keeper: keeper,
	}
}

var _ govtypes1.MsgServer = msgGovServer{}

func (m msgGovServer) SubmitProposal(goCtx context.Context, msg *govtypes1.MsgSubmitProposal) (*govtypes1.MsgSubmitProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	msgs := make([][]byte, len(msg.Messages))
	for i, submsg := range msg.Messages {
		msgjson, err := m.Keeper.JSONCodec().MarshalJSON(submsg)
		if err != nil {
			return nil, err
		}
		msgs[i] = msgjson
	}
	internalMsg := &types.MsgSubmitProposal{
		Messages:       msgs,
		InitialDeposit: msg.InitialDeposit,
		Proposer:       msg.Proposer,
		Metadata:       msg.Metadata,
		Title:          msg.Title,
		Summary:        msg.Summary,
		Expedited:      msg.Expedited,
	}

	msgjson, err := m.Keeper.JSONCodec().MarshalJSON(internalMsg)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"SubmitProposal":%s}`, string(msgjson)))
	resp, err := m.Keeper.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   msg.Proposer,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	var response govtypes1.MsgSubmitProposalResponse
	err = m.Keeper.cdc.UnmarshalJSON(resp.Data, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (m msgGovServer) Vote(goCtx context.Context, msg *govtypes1.MsgVote) (*govtypes1.MsgVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgjson, err := m.Keeper.JSONCodec().MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"Vote":%s}`, string(msgjson)))
	_, err = m.Keeper.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   msg.Voter,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	return &govtypes1.MsgVoteResponse{}, nil
}

func (m msgGovServer) VoteWeighted(goCtx context.Context, msg *govtypes1.MsgVoteWeighted) (*govtypes1.MsgVoteWeightedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgjson, err := m.Keeper.JSONCodec().MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"VoteWeighted":%s}`, string(msgjson)))
	fmt.Println("-VoteWeighted-", string(msgbz))
	_, err = m.Keeper.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   msg.Voter,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	return &govtypes1.MsgVoteWeightedResponse{}, nil
}

func (m msgGovServer) Deposit(goCtx context.Context, msg *govtypes1.MsgDeposit) (*govtypes1.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("Deposit not implemented")
	return &govtypes1.MsgDepositResponse{}, nil
}

func (m msgGovServer) CancelProposal(goCtx context.Context, msg *govtypes1.MsgCancelProposal) (*govtypes1.MsgCancelProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m.Keeper.Logger(ctx).Error("CancelProposal not implemented")
	return &govtypes1.MsgCancelProposalResponse{}, nil
}

func (m msgGovServer) ExecLegacyContent(context.Context, *govtypes1.MsgExecLegacyContent) (*govtypes1.MsgExecLegacyContentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecLegacyContent not implemented")
}

func (m msgGovServer) UpdateParams(context.Context, *govtypes1.MsgUpdateParams) (*govtypes1.MsgUpdateParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateParams not implemented")
}
