package types

import (
	"errors"
	fmt "fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	codec "github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	wasmxtypes "mythos/v1/x/wasmx/types"
)

// NewGenesisState creates a new genesis state.
func NewGenesisState(staking StakingGenesisState, bank BankGenesisState, gov GovGenesisState, auth AuthGenesisState, slashing slashingtypes.GenesisState, distrib DistributionGenesisState) *GenesisState {
	return &GenesisState{
		Staking:      staking,
		Bank:         bank,
		Gov:          gov,
		Auth:         auth,
		Slashing:     slashing,
		Distribution: distrib,
	}
}

// NewBankGenesisState returns a default bank module genesis state.
func NewBankGenesisState(params banktypes.Params, balances []banktypes.Balance, supply sdk.Coins, denomInfo []DenomDeploymentInfo, sendEnabled []banktypes.SendEnabled) *BankGenesisState {
	return &BankGenesisState{
		Params:      params,
		Balances:    balances,
		Supply:      supply,
		DenomInfo:   denomInfo,
		SendEnabled: sendEnabled,
	}
}

// NewStakingGenesisState returns a default staking module genesis state.
func NewStakingGenesisState(params stakingtypes.Params, validators []stakingtypes.Validator, delegations []Delegation) *StakingGenesisState {
	return &StakingGenesisState{
		Params:      params,
		Validators:  validators,
		Delegations: delegations,
	}
}

// NewAuthGenesisState returns a default staking module genesis state.
func NewAuthGenesisState(params authtypes.Params, accounts []*AnyAccount) *AuthGenesisState {
	return &AuthGenesisState{
		Params:   params,
		Accounts: accounts,
	}
}

// NewSlashingGenesisState returns a default staking module genesis state.
func NewSlashingGenesisState(params slashingtypes.Params, signingInfos []slashingtypes.SigningInfo, missedBlocks []slashingtypes.ValidatorMissedBlocks) *slashingtypes.GenesisState {
	return &slashingtypes.GenesisState{
		Params:       params,
		SigningInfos: signingInfos,
		MissedBlocks: missedBlocks,
	}
}

// NewDistributingGenesisState returns a default staking module genesis state.
func NewDistributingGenesisState(
	params distributiontypes.Params,
	feePool distributiontypes.FeePool,
	delegatorWithdrawInfos []distributiontypes.DelegatorWithdrawInfo,
	previousProposer string,
	outstandingRewards []distributiontypes.ValidatorOutstandingRewardsRecord,
	validatorAccumulatedCommissions []distributiontypes.ValidatorAccumulatedCommissionRecord,
	validatorCurrentRewards []distributiontypes.ValidatorCurrentRewardsRecord,
	delegatorStartingInfos []distributiontypes.DelegatorStartingInfoRecord,
	validatorSlashEvents []distributiontypes.ValidatorSlashEventRecord,
	baseDenom string,
) *DistributionGenesisState {
	return &DistributionGenesisState{
		Params:                          params,
		FeePool:                         feePool,
		DelegatorWithdrawInfos:          delegatorWithdrawInfos,
		PreviousProposer:                previousProposer,
		OutstandingRewards:              outstandingRewards,
		ValidatorAccumulatedCommissions: validatorAccumulatedCommissions,
		ValidatorCurrentRewards:         validatorCurrentRewards,
		DelegatorStartingInfos:          delegatorStartingInfos,
		ValidatorSlashEvents:            validatorSlashEvents,
		BaseDenom:                       baseDenom,
	}
}

// NewAuthGenesisState returns a default staking module genesis state.
func NewAuthGenesisStateFromCosmos(cdc codec.Codec, params authtypes.Params, accounts []authtypes.GenesisAccount) (*AuthGenesisState, error) {
	accs := make([]*AnyAccount, len(accounts))
	for i, account := range accounts {
		acc, err := AccountIToAnyAccount(account, cdc)
		if err != nil {
			return nil, err
		}
		accs[i] = acc
	}
	return &AuthGenesisState{
		Params:   params,
		Accounts: accs,
	}, nil
}

// DefaultStakingGenesisState returns a default bank module genesis state.
func DefaultStakingGenesisState() *StakingGenesisState {
	return &StakingGenesisState{
		Params: stakingtypes.DefaultParams(),
	}
}

func DefaultBankDenoms(denomUnit string, baseDenomUnit uint32, denomName string) []DenomDeploymentInfo {
	erc20jsonCodeId := -1
	derc20jsonCodeId := -1
	for i, sysc := range wasmxtypes.DefaultSystemContracts("", "") {
		if sysc.Label == wasmxtypes.ERC20_v001 {
			erc20jsonCodeId = i + 1
		} else if sysc.Label == wasmxtypes.DERC20_v001 {
			derc20jsonCodeId = i + 1
		}
	}
	if erc20jsonCodeId == -1 {
		panic(fmt.Sprintf("%s missing", wasmxtypes.ERC20_v001))
	}
	if derc20jsonCodeId == -1 {
		panic(fmt.Sprintf("%s missing", wasmxtypes.DERC20_v001))
	}
	return []DenomDeploymentInfo{
		{
			Metadata: banktypes.Metadata{
				Description: "main gas token",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    denomUnit,
						Exponent: baseDenomUnit,
						Aliases:  []string{},
					},
					{
						Denom:    fmt.Sprintf("a%s", denomUnit),
						Exponent: 1,
						Aliases:  []string{},
					},
				},
				Base:    fmt.Sprintf("a%s", denomUnit),
				Display: strings.ToUpper(denomUnit),
				Name:    strings.ToUpper(denomUnit),
				Symbol:  denomUnit,
				URI:     "",
				URIHash: "",
			},
			CodeId:  uint64(erc20jsonCodeId),
			Admins:  []string{wasmxtypes.ROLE_BANK, wasmxtypes.ROLE_GOVERNANCE},
			Minters: []string{wasmxtypes.ROLE_BANK, wasmxtypes.ROLE_GOVERNANCE, wasmxtypes.ROLE_DISTRIBUTION},
		},
		{
			Metadata: banktypes.Metadata{
				Description: "staking token",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    fmt.Sprintf("s%s", denomUnit),
						Exponent: baseDenomUnit,
						Aliases:  []string{},
					},
					{
						Denom:    fmt.Sprintf("as%s", denomUnit),
						Exponent: 1,
						Aliases:  []string{},
					},
				},
				Base:    fmt.Sprintf("as%s", denomUnit),
				Display: strings.ToUpper(fmt.Sprintf("s%s", denomUnit)),
				Name:    strings.ToUpper(fmt.Sprintf("s%s", denomUnit)),
				Symbol:  fmt.Sprintf("s%s", denomUnit),
				URI:     "",
				URIHash: "",
			},
			CodeId:    uint64(derc20jsonCodeId),
			Admins:    []string{wasmxtypes.ROLE_STAKING, wasmxtypes.ROLE_BANK},
			Minters:   []string{wasmxtypes.ROLE_STAKING, wasmxtypes.ROLE_BANK},
			BaseDenom: fmt.Sprintf("a%s", denomUnit),
		},
		{
			Metadata: banktypes.Metadata{
				Description: "rewards token",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    fmt.Sprintf("r%s", denomUnit),
						Exponent: baseDenomUnit,
						Aliases:  []string{},
					},
					{
						Denom:    fmt.Sprintf("ar%s", denomUnit),
						Exponent: 1,
						Aliases:  []string{},
					},
				},
				Base:    fmt.Sprintf("ar%s", denomUnit),
				Display: strings.ToUpper(fmt.Sprintf("r%s", denomUnit)),
				Name:    strings.ToUpper(fmt.Sprintf("r%s", denomUnit)),
				Symbol:  fmt.Sprintf("r%s", denomUnit),
				URI:     "",
				URIHash: "",
			},
			CodeId:    uint64(erc20jsonCodeId),
			Admins:    []string{wasmxtypes.ROLE_BANK, wasmxtypes.ROLE_DISTRIBUTION},
			Minters:   []string{wasmxtypes.ROLE_BANK, wasmxtypes.ROLE_DISTRIBUTION},
			BaseDenom: fmt.Sprintf("a%s", denomUnit),
		},
		{
			Metadata: banktypes.Metadata{
				Description: "arbitration token",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    "arb",
						Exponent: baseDenomUnit,
						Aliases:  []string{},
					},
					{
						Denom:    "aarb",
						Exponent: 1,
						Aliases:  []string{},
					},
				},
				Base:    "aarb",
				Display: "ARBITRATION",
				Name:    "ARBITRATION",
				Symbol:  "arb",
				URI:     "",
				URIHash: "",
			},
			CodeId:  uint64(erc20jsonCodeId),
			Admins:  []string{wasmxtypes.ROLE_GOVERNANCE},
			Minters: []string{wasmxtypes.ROLE_GOVERNANCE},
		},
	}
}

// DefaultBankGenesisState returns a default bank module genesis state.
func DefaultBankGenesisState(denomUnit string, baseDenomUnit uint32, denomName string) *BankGenesisState {
	return NewBankGenesisState(banktypes.DefaultParams(), []banktypes.Balance{}, sdk.Coins{}, DefaultBankDenoms(denomUnit, baseDenomUnit, denomName), []banktypes.SendEnabled{})
}

// DefaultGovGenesisState returns a default bank module genesis state.
func DefaultGovGenesisState() *GovGenesisState {
	govstate := govtypes1.DefaultGenesisState()
	return &GovGenesisState{
		StartingProposalId: govstate.StartingProposalId,
		Deposits:           govstate.Deposits,
		Votes:              govstate.Votes,
		Proposals:          make([]*GovProposal, 0),
		Params:             CosmosParamsToInternal(govstate.Params),
		Constitution:       govstate.Constitution,
	}
}

// DefaultSlashingGenesisState returns a default bank module genesis state.
func DefaultSlashingGenesisState() *slashingtypes.GenesisState {
	return slashingtypes.DefaultGenesisState()
}

// DefaultDistributionGenesisState returns a default bank module genesis state.
func DefaultDistributionGenesisState(baseDenom string) *DistributionGenesisState {
	gen := distributiontypes.DefaultGenesisState()
	return NewDistributingGenesisState(
		gen.Params,
		gen.FeePool,
		gen.DelegatorWithdrawInfos,
		gen.PreviousProposer,
		gen.OutstandingRewards,
		gen.ValidatorAccumulatedCommissions,
		gen.ValidatorCurrentRewards,
		gen.DelegatorStartingInfos,
		gen.ValidatorSlashEvents,
		baseDenom,
	)
}

// DefaultAuthGenesisState returns a default bank module genesis state.
func DefaultAuthGenesisState() *AuthGenesisState {
	state := authtypes.DefaultGenesisState()
	accounts := make([]*AnyAccount, len(state.Accounts))
	for i, acc := range state.Accounts {
		// val := acc.GetCachedValue()
		accounts[i] = AnyToAnyAccount(acc)
	}
	return &AuthGenesisState{
		Params:   state.Params,
		Accounts: accounts,
	}
}

func AnyToAnyAccount(acc *cdctypes.Any) *AnyAccount {
	return &AnyAccount{
		TypeUrl: acc.GetTypeUrl(),
		Value:   acc.GetValue(),
	}
}

func AccountIToAnyAccount(acc sdk.AccountI, cdc codec.Codec) (*AnyAccount, error) {
	bz, err := cdc.MarshalJSON(acc)
	if err != nil {
		return nil, err
	}
	return &AnyAccount{
		TypeUrl: sdk.MsgTypeURL(acc),
		Value:   bz,
	}, nil
}

// DefaultGenesisState sets default evm genesis state with empty accounts and
// default params and chain config values.
func DefaultGenesisState(denomUnit string, baseDenomUnit uint32, denomName string) *GenesisState {
	return &GenesisState{
		Staking: *DefaultStakingGenesisState(),
		Bank:    *DefaultBankGenesisState(denomUnit, baseDenomUnit, denomName),
		Gov:     *DefaultGovGenesisState(),
		Auth:    *DefaultAuthGenesisState(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Staking.Validate(); err != nil {
		return err
	}
	if err := gs.Bank.Validate(); err != nil {
		return err
	}
	if err := gs.Gov.Validate(); err != nil {
		return err
	}
	if err := gs.Auth.Validate(); err != nil {
		return err
	}
	return nil
}

// ValidateGenesis validates the provided staking genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds, no duplicate validators)
func (gs StakingGenesisState) Validate() error {
	if err := validateGenesisStateValidators(gs.Validators); err != nil {
		return err
	}

	return gs.Params.Validate()
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs BankGenesisState) Validate() error {
	if len(gs.Params.SendEnabled) > 0 && len(gs.SendEnabled) > 0 {
		return errors.New("send_enabled defined in both the send_enabled field and in params (deprecated)")
	}

	if err := gs.Params.Validate(); err != nil {
		return err
	}

	seenSendEnabled := make(map[string]bool)
	seenBalances := make(map[string]bool)
	seenMetadatas := make(map[string]bool)

	totalSupply := sdk.Coins{}

	for _, p := range gs.GetSendEnabled() {
		if _, exists := seenSendEnabled[p.Denom]; exists {
			return fmt.Errorf("duplicate send enabled found: '%s'", p.Denom)
		}
		if err := p.Validate(); err != nil {
			return err
		}
		seenSendEnabled[p.Denom] = true
	}

	for _, balance := range gs.Balances {
		if seenBalances[balance.Address] {
			return fmt.Errorf("duplicate balance for address %s", balance.Address)
		}

		if err := balance.Validate(); err != nil {
			return err
		}

		seenBalances[balance.Address] = true

		totalSupply = totalSupply.Add(balance.Coins...)
	}

	for _, info := range gs.DenomInfo {
		if seenMetadatas[info.Metadata.Base] {
			return fmt.Errorf("duplicate client metadata for denom %s", info.Metadata.Base)
		}

		if err := info.Metadata.Validate(); err != nil {
			return err
		}

		// TODO info.CodeId
		// TODO info.Admins
		// TODO info.Minters

		seenMetadatas[info.Metadata.Base] = true
	}

	if !gs.Supply.Empty() {
		// NOTE: this errors if supply for any given coin is zero
		err := gs.Supply.Validate()
		if err != nil {
			return err
		}

		if !gs.Supply.Equal(totalSupply) {
			return fmt.Errorf("genesis supply is incorrect, expected %v, got %v", gs.Supply, totalSupply)
		}
	}

	return nil
}

func validateGenesisStateValidators(validators []stakingtypes.Validator) error {
	addrMap := make(map[string]bool, len(validators))

	for i := 0; i < len(validators); i++ {
		val := validators[i]
		consPk, err := val.ConsPubKey()
		if err != nil {
			return err
		}

		strKey := string(consPk.Bytes())

		if _, ok := addrMap[strKey]; ok {
			consAddr, err := val.GetConsAddr()
			if err != nil {
				return err
			}
			return fmt.Errorf("duplicate validator in genesis state: moniker %v, address %v", val.Description.Moniker, consAddr)
		}

		if val.Jailed && val.IsBonded() {
			consAddr, err := val.GetConsAddr()
			if err != nil {
				return err
			}
			return fmt.Errorf("validator is bonded and jailed in genesis state: moniker %v, address %v", val.Description.Moniker, consAddr)
		}

		if val.DelegatorShares.IsZero() && !val.IsUnbonding() {
			return fmt.Errorf("bonded/unbonded genesis validator cannot have zero delegator shares, validator: %v", val)
		}

		addrMap[strKey] = true
	}

	return nil
}

// ValidateGenesis validates the provided gov genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds, no duplicate validators)
func (gs GovGenesisState) Validate() error {
	if gs.StartingProposalId == 0 {
		return errors.New("starting proposal id must be greater than 0")
	}

	return gs.Params.ValidateBasic()
}

// ValidateBasic performs basic validation on governance parameters.
func (p GovParams) ValidateBasic() error {
	minDeposit := sdk.Coins(p.MinDeposit)
	if minDeposit.Empty() || !minDeposit.IsValid() {
		return fmt.Errorf("invalid minimum deposit: %s", minDeposit)
	}

	if minExpeditedDeposit := sdk.Coins(p.ExpeditedMinDeposit); minExpeditedDeposit.Empty() || !minExpeditedDeposit.IsValid() {
		return fmt.Errorf("invalid expedited minimum deposit: %s", minExpeditedDeposit)
	} else if minExpeditedDeposit.IsAllLTE(minDeposit) {
		return fmt.Errorf("expedited minimum deposit must be greater than minimum deposit: %s", minExpeditedDeposit)
	}

	if p.MaxDepositPeriod <= 0 {
		return fmt.Errorf("maximum deposit period must be positive: %d", p.MaxDepositPeriod)
	}

	quorum, err := sdkmath.LegacyNewDecFromStr(p.Quorum)
	if err != nil {
		return fmt.Errorf("invalid quorum string: %w", err)
	}
	if quorum.IsNegative() {
		return fmt.Errorf("quorom cannot be negative: %s", quorum)
	}
	if quorum.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("quorom too large: %s", p.Quorum)
	}

	threshold, err := sdkmath.LegacyNewDecFromStr(p.Threshold)
	if err != nil {
		return fmt.Errorf("invalid threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("vote threshold too large: %s", threshold)
	}

	expeditedThreshold, err := sdkmath.LegacyNewDecFromStr(p.ExpeditedThreshold)
	if err != nil {
		return fmt.Errorf("invalid expedited threshold string: %w", err)
	}
	if !threshold.IsPositive() {
		return fmt.Errorf("expedited vote threshold must be positive: %s", threshold)
	}
	if threshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("expedited vote threshold too large: %s", threshold)
	}
	if expeditedThreshold.LTE(threshold) {
		return fmt.Errorf("expedited vote threshold %s, must be greater than the regular threshold %s", expeditedThreshold, threshold)
	}

	vetoThreshold, err := sdkmath.LegacyNewDecFromStr(p.VetoThreshold)
	if err != nil {
		return fmt.Errorf("invalid vetoThreshold string: %w", err)
	}
	if !vetoThreshold.IsPositive() {
		return fmt.Errorf("veto threshold must be positive: %s", vetoThreshold)
	}
	if vetoThreshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("veto threshold too large: %s", vetoThreshold)
	}

	if p.VotingPeriod <= 0 {
		return fmt.Errorf("voting period must be positive: %s", p.VotingPeriod)
	}

	if p.ExpeditedVotingPeriod <= 0 {
		return fmt.Errorf("expedited voting period must be positive: %s", p.ExpeditedVotingPeriod)
	}
	if p.ExpeditedVotingPeriod >= p.VotingPeriod {
		return fmt.Errorf("expedited voting period %s must be strictly less that the regular voting period %s", p.ExpeditedVotingPeriod, p.VotingPeriod)
	}

	minInitialDepositRatio, err := sdkmath.LegacyNewDecFromStr(p.MinInitialDepositRatio)
	if err != nil {
		return fmt.Errorf("invalid mininum initial deposit ratio of proposal: %w", err)
	}
	if minInitialDepositRatio.IsNegative() {
		return fmt.Errorf("mininum initial deposit ratio of proposal must be positive: %s", minInitialDepositRatio)
	}
	if minInitialDepositRatio.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("mininum initial deposit ratio of proposal is too large: %s", minInitialDepositRatio)
	}

	proposalCancelRate, err := sdkmath.LegacyNewDecFromStr(p.ProposalCancelRatio)
	if err != nil {
		return fmt.Errorf("invalid burn rate of cancel proposal: %w", err)
	}
	if proposalCancelRate.IsNegative() {
		return fmt.Errorf("burn rate of cancel proposal must be positive: %s", proposalCancelRate)
	}
	if proposalCancelRate.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("burn rate of cancel proposal is too large: %s", proposalCancelRate)
	}

	// TODO address validator with AddressCodec
	// if len(p.ProposalCancelDest) != 0 {
	// 	_, err := sdk.AccAddressFromBech32(p.ProposalCancelDest)
	// 	if err != nil {
	// 		return fmt.Errorf("deposits destination address is invalid: %s", p.ProposalCancelDest)
	// 	}
	// }

	return nil
}

// Validate performs basic validation of auth genesis data returning an
// error for any failed validation criteria.
func (gs AuthGenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// TODO validation
	// genAccs, err := authtypes.UnpackAccounts(gs.Accounts)
	// if err != nil {
	// 	return err
	// }

	// return authtypes.ValidateGenAccounts(genAccs)
	return nil
}
