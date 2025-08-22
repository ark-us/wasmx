package gov

import (
	wasmx "github.com/loredanacirstea/wasmx-env"
	utils "github.com/loredanacirstea/wasmx-utils"
)

// Defaults for gov module (mirrors AssemblyScript defaults.ts)

const (
	DefaultStartingProposalID             = uint64(1)
	DefaultMinDepositTokens               = "10000000"
	DefaultMinExpeditedDepositTokensRatio = uint64(5)
	DefaultDepositPeriod                  = uint64(172800000) // 48h in ms
	DefaultVotingPeriod                   = uint64(172800000) // 48h in ms
	DefaultVotingExpedited                = uint64(86400000)  // 24h in ms
	DefaultQuorum                         = "0.334000000000000000"
	DefaultThreshold                      = "0.500000000000000000"
	DefaultVetoThreshold                  = "0.334000000000000000"
	MinInitialDepositRatio                = "0.000000000000000000"
	ProposalCancelRatio                   = "0.500000000000000000"
	ProposalCancelDest                    = ""
	ExpeditedThreshold                    = "0.667000000000000000"
	BurnVoteQuorum                        = false
	BurnProposalDepositPrevote            = false
	BurnVoteVeto                          = true
	MinDepositRatio                       = "0.010000000000000000"
)

func GetDefaultParams(defaultBondDenom string) Params {
	min := Coin{Denom: defaultBondDenom, Amount: NewBigFromString(DefaultMinDepositTokens)}
	expeditedMin := Coin{Denom: defaultBondDenom, Amount: NewBigFromString(DefaultMinDepositTokens).Mul(NewBigFromString("5"))}
	return Params{
		MinDeposit:                 []Coin{min},
		MaxDepositPeriod:           utils.StringUint64(DefaultDepositPeriod),
		VotingPeriod:               utils.StringUint64(DefaultVotingPeriod),
		Quorum:                     DefaultQuorum,
		Threshold:                  DefaultThreshold,
		VetoThreshold:              DefaultVetoThreshold,
		MinInitialDepositRatio:     MinInitialDepositRatio,
		ProposalCancelRatio:        ProposalCancelRatio,
		ProposalCancelDest:         wasmx.Bech32String(ProposalCancelDest),
		ExpeditedVotingPeriod:      utils.StringUint64(DefaultVotingExpedited),
		ExpeditedThreshold:         ExpeditedThreshold,
		ExpeditedMinDeposit:        []Coin{expeditedMin},
		BurnVoteQuorum:             BurnVoteQuorum,
		BurnProposalDepositPrevote: BurnProposalDepositPrevote,
		BurnVoteVeto:               BurnVoteVeto,
		MinDepositRatio:            MinDepositRatio,
	}
}

func GetDefaultGenesis(baseDenom string, defaultBondDenom string, rewardsBaseDenom string) GenesisState {
	params := GetDefaultParams(defaultBondDenom)
	return GenesisState{
		StartingProposalID: utils.StringUint64(DefaultStartingProposalID),
		Deposits:           []Deposit{},
		Votes:              []Vote{},
		Proposals:          []Proposal{},
		Params:             params,
		Constitution:       "",
	}
}
