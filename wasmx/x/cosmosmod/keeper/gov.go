package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	types "wasmx/v1/x/cosmosmod/types"
	networktypes "wasmx/v1/x/network/types"
	wasmxtypes "wasmx/v1/x/wasmx/types"
)

func (k KeeperGov) Proposal(ctx sdk.Context, req *govtypes.QueryProposalRequest) (*govtypes.QueryProposalResponse, error) {
	reqbz, err := k.JSONCodec().MarshalJSON(req)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetProposal":%s}`, string(reqbz)))
	res, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	var resp wasmxtypes.ContractResponse
	err = json.Unmarshal(res.Data, &resp)
	if err != nil {
		return nil, err
	}
	var internalResp types.QueryProposalResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &internalResp)
	if err != nil {
		return nil, err
	}
	proposal, err := types.CosmosProposalFromInternal(k.JSONCodec(), *internalResp.Proposal)
	if err != nil {
		return nil, err
	}
	return &govtypes.QueryProposalResponse{Proposal: proposal}, nil
}

func (k KeeperGov) Proposals(ctx sdk.Context, req *govtypes.QueryProposalsRequest) (*govtypes.QueryProposalsResponse, error) {
	reqbz, err := k.JSONCodec().MarshalJSON(req)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetProposals":%s}`, string(reqbz)))
	res, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	var resp wasmxtypes.ContractResponse
	err = json.Unmarshal(res.Data, &resp)
	if err != nil {
		return nil, err
	}
	var internalResp types.QueryProposalsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &internalResp)
	if err != nil {
		return nil, err
	}
	proposals := make(govtypes.Proposals, len(internalResp.Proposals))
	for i, proposal := range internalResp.Proposals {
		prop, err := types.CosmosProposalFromInternal(k.JSONCodec(), *proposal)
		if err != nil {
			return nil, err
		}
		proposals[i] = prop
	}
	return &govtypes.QueryProposalsResponse{Proposals: proposals, Pagination: internalResp.Pagination}, nil
}

func (k KeeperGov) Vote(ctx sdk.Context, req *govtypes.QueryVoteRequest) (*govtypes.QueryVoteResponse, error) {
	k.Logger(ctx).Error("Vote not implemented")
	return &govtypes.QueryVoteResponse{Vote: nil}, nil
}

func (k KeeperGov) Votes(ctx sdk.Context, req *govtypes.QueryVotesRequest) (*govtypes.QueryVotesResponse, error) {
	k.Logger(ctx).Error("Votes not implemented")
	return &govtypes.QueryVotesResponse{Votes: make([]*govtypes.Vote, 0), Pagination: &query.PageResponse{Total: 0}}, nil
}

func (k KeeperGov) Params(ctx sdk.Context, req *govtypes.QueryParamsRequest) (*govtypes.QueryParamsResponse, error) {
	reqbz, err := k.JSONCodec().MarshalJSON(req)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetParams":%s}`, string(reqbz)))
	res, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	var resp wasmxtypes.ContractResponse
	err = json.Unmarshal(res.Data, &resp)
	if err != nil {
		return nil, err
	}

	var internalResp types.QueryParamsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &internalResp)
	if err != nil {
		return nil, err
	}
	params, err := types.CosmosParamsFromInternal(internalResp.Params)
	if err != nil {
		return nil, err
	}
	return &govtypes.QueryParamsResponse{Params: params}, nil
}

func (k KeeperGov) Deposit(ctx sdk.Context, req *govtypes.QueryDepositRequest) (*govtypes.QueryDepositResponse, error) {
	k.Logger(ctx).Error("Deposit not implemented")
	return &govtypes.QueryDepositResponse{}, nil
}

func (k KeeperGov) Deposits(ctx sdk.Context, req *govtypes.QueryDepositsRequest) (*govtypes.QueryDepositsResponse, error) {
	k.Logger(ctx).Error("Deposits not implemented")
	return &govtypes.QueryDepositsResponse{Deposits: make([]*govtypes.Deposit, 0), Pagination: &query.PageResponse{Total: 0}}, nil
}

func (k KeeperGov) TallyResult(ctx sdk.Context, req *govtypes.QueryTallyResultRequest) (*govtypes.QueryTallyResultResponse, error) {
	reqbz, err := k.JSONCodec().MarshalJSON(req)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetTallyResult":%s}`, string(reqbz)))
	res, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	var resp wasmxtypes.ContractResponse
	err = json.Unmarshal(res.Data, &resp)
	if err != nil {
		return nil, err
	}

	var internalResp govtypes.QueryTallyResultResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &internalResp)
	if err != nil {
		return nil, err
	}
	tally, err := types.CosmosTallyFromInternal(internalResp.Tally)
	if err != nil {
		return nil, err
	}
	internalResp.Tally = tally
	return &internalResp, nil
}

func (k KeeperGov) Constitution(ctx sdk.Context, req *govtypes.QueryConstitutionRequest) (*govtypes.QueryConstitutionResponse, error) {
	k.Logger(ctx).Error("Constitution not implemented")
	return &govtypes.QueryConstitutionResponse{}, nil
}
