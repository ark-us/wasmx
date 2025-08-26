package types

import (
	fmt "fmt"
	"math/big"
	"strconv"
	"time"

	codec "github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func CosmosProposalFromInternal(cdc codec.JSONCodec, proposal GovProposal) (*govtypes.Proposal, error) {
	anymsgs := make([]*cdctypes.Any, len(proposal.Messages))
	for i, submsgbz := range proposal.Messages {
		var anymsg cdctypes.Any
		err := cdc.UnmarshalJSON(submsgbz, &anymsg)
		if err != nil {
			return nil, err
		}
		anymsgs[i] = &anymsg
	}
	tally, err := CosmosTallyFromInternal(proposal.FinalTallyResult)
	if err != nil {
		return nil, err
	}
	return &govtypes.Proposal{
		Messages:         anymsgs,
		Id:               proposal.Id,
		Status:           proposal.Status,
		FinalTallyResult: tally,
		SubmitTime:       proposal.SubmitTime,
		DepositEndTime:   proposal.DepositEndTime,
		TotalDeposit:     proposal.TotalDeposit,
		VotingStartTime:  proposal.VotingStartTime,
		VotingEndTime:    proposal.VotingEndTime,
		Metadata:         proposal.Metadata,
		Title:            proposal.Title,
		Summary:          proposal.Summary,
		Proposer:         proposal.Proposer,
		Expedited:        proposal.Expedited,
		FailedReason:     proposal.FailedReason,
	}, nil
}

func CosmosTallyFromInternal(tally *govtypes.TallyResult) (*govtypes.TallyResult, error) {
	yes, ok := new(big.Int).SetString(tally.YesCount, 0)
	if !ok {
		return nil, fmt.Errorf("could not parse tally yes_count")
	}
	no, ok := new(big.Int).SetString(tally.NoCount, 0)
	if !ok {
		return nil, fmt.Errorf("could not parse tally no_count")
	}
	abstain, ok := new(big.Int).SetString(tally.AbstainCount, 0)
	if !ok {
		return nil, fmt.Errorf("could not parse tally abstain_count")
	}
	veto, ok := new(big.Int).SetString(tally.NoWithVetoCount, 0)
	if !ok {
		return nil, fmt.Errorf("could not parse tally no_with_veto_count")
	}
	tally.YesCount = yes.String()
	tally.NoCount = no.String()
	tally.AbstainCount = abstain.String()
	tally.NoWithVetoCount = veto.String()
	return tally, nil
}

func CosmosProposalToInternal(cdc codec.JSONCodec, proposal govtypes.Proposal) (*GovProposal, error) {
	encodedMsgs := make([][]byte, len(proposal.Messages))
	for i, msg := range proposal.Messages {
		msgbz, err := cdc.MarshalJSON(msg)
		if err != nil {
			return nil, err
		}
		encodedMsgs[i] = msgbz
	}
	return &GovProposal{
		Messages:         encodedMsgs,
		Id:               proposal.Id,
		Status:           proposal.Status,
		FinalTallyResult: proposal.FinalTallyResult,
		SubmitTime:       proposal.SubmitTime,
		DepositEndTime:   proposal.DepositEndTime,
		TotalDeposit:     proposal.TotalDeposit,
		VotingStartTime:  proposal.VotingStartTime,
		VotingEndTime:    proposal.VotingEndTime,
		Metadata:         proposal.Metadata,
		Title:            proposal.Title,
		Summary:          proposal.Summary,
		Proposer:         proposal.Proposer,
		Expedited:        proposal.Expedited,
		FailedReason:     proposal.FailedReason,
	}, nil
}

func CosmosProposalsFromInternal(cdc codec.JSONCodec, proposals []GovProposal) ([]govtypes.Proposal, error) {
	cproposals := make([]govtypes.Proposal, len(proposals))
	for i, prop := range proposals {
		cprop, err := CosmosProposalFromInternal(cdc, prop)
		if err != nil {
			return nil, err
		}
		cproposals[i] = *cprop
	}
	return cproposals, nil
}

func CosmosProposalsToInternal(cdc codec.JSONCodec, proposals []govtypes.Proposal) ([]GovProposal, error) {
	cproposals := make([]GovProposal, len(proposals))
	for i, prop := range proposals {
		cprop, err := CosmosProposalToInternal(cdc, prop)
		if err != nil {
			return nil, err
		}
		cproposals[i] = *cprop
	}
	return cproposals, nil
}

func CosmosParamsFromInternal(params *GovParams) (*govtypes.Params, error) {
	maxdp, err := time.ParseDuration(strconv.FormatInt(params.MaxDepositPeriod, 10) + "ms")
	if err != nil {
		return nil, err
	}
	vp, err := time.ParseDuration(strconv.FormatInt(params.VotingPeriod, 10) + "ms")
	if err != nil {
		return nil, err
	}
	evp, err := time.ParseDuration(strconv.FormatInt(params.ExpeditedVotingPeriod, 10) + "ms")
	if err != nil {
		return nil, err
	}

	return &govtypes.Params{
		MinDeposit:                 params.MinDeposit,
		MaxDepositPeriod:           &maxdp,
		VotingPeriod:               &vp,
		Quorum:                     params.Quorum,
		Threshold:                  params.Threshold,
		VetoThreshold:              params.VetoThreshold,
		MinInitialDepositRatio:     params.MinInitialDepositRatio,
		ProposalCancelRatio:        params.ProposalCancelRatio,
		ProposalCancelDest:         params.ProposalCancelDest,
		ExpeditedVotingPeriod:      &evp,
		ExpeditedThreshold:         params.ExpeditedThreshold,
		ExpeditedMinDeposit:        params.ExpeditedMinDeposit,
		BurnVoteQuorum:             params.BurnVoteQuorum,
		BurnProposalDepositPrevote: params.BurnProposalDepositPrevote,
		BurnVoteVeto:               params.BurnVoteVeto,
		MinDepositRatio:            params.MinDepositRatio,
	}, nil
}

func CosmosParamsToInternal(params *govtypes.Params) *GovParams {
	return &GovParams{
		MinDeposit:                 params.MinDeposit,
		MaxDepositPeriod:           params.MaxDepositPeriod.Milliseconds(),
		VotingPeriod:               params.VotingPeriod.Milliseconds(),
		Quorum:                     params.Quorum,
		Threshold:                  params.Threshold,
		VetoThreshold:              params.VetoThreshold,
		MinInitialDepositRatio:     params.MinInitialDepositRatio,
		ProposalCancelRatio:        params.ProposalCancelRatio,
		ProposalCancelDest:         params.ProposalCancelDest,
		ExpeditedVotingPeriod:      params.ExpeditedVotingPeriod.Milliseconds(),
		ExpeditedThreshold:         params.ExpeditedThreshold,
		ExpeditedMinDeposit:        params.ExpeditedMinDeposit,
		BurnVoteQuorum:             params.BurnVoteQuorum,
		BurnProposalDepositPrevote: params.BurnProposalDepositPrevote,
		BurnVoteVeto:               params.BurnVoteVeto,
		MinDepositRatio:            params.MinDepositRatio,
	}
}

// DefaultGovParams returns the default governance params
func DefaultGovParams(depositDenom string) govtypes.Params {
	return govtypes.NewParams(
		sdk.NewCoins(sdk.NewCoin(depositDenom, govtypes.DefaultMinDepositTokens)),
		sdk.NewCoins(sdk.NewCoin(depositDenom, govtypes.DefaultMinExpeditedDepositTokens)),
		govtypes.DefaultPeriod,
		govtypes.DefaultPeriod,
		govtypes.DefaultExpeditedPeriod,
		govtypes.DefaultQuorum.String(),
		govtypes.DefaultThreshold.String(),
		govtypes.DefaultExpeditedThreshold.String(),
		govtypes.DefaultVetoThreshold.String(),
		govtypes.DefaultMinInitialDepositRatio.String(),
		govtypes.DefaultProposalCancelRatio.String(),
		govtypes.DefaultProposalCancelDestAddress,
		govtypes.DefaultBurnProposalPrevote,
		govtypes.DefaultBurnVoteQuorom,
		govtypes.DefaultBurnVoteVeto,
		govtypes.DefaultMinDepositRatio.String(),
	)
}
