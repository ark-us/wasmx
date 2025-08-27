package lib

import (
	sdkmath "cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env"
)

const MODULE_NAME = "distribution"

const FEE_COLLECTOR_ROLE = "fee_collector"

// Params represents distribution module parameters
type Params struct {
	CommunityTax        string `json:"community_tax"`         // f64 as string
	BaseProposerReward  string `json:"base_proposer_reward"`  // f64 as string
	BonusProposerReward string `json:"bonus_proposer_reward"` // f64 as string
	WithdrawAddrEnabled bool   `json:"withdraw_addr_enabled"`
}

// DelegatorWithdrawInfo represents delegator withdrawal information
type DelegatorWithdrawInfo struct {
	DelegatorAddress wasmx.Bech32String `json:"delegator_address"`
	WithdrawAddress  wasmx.Bech32String `json:"withdraw_address"`
}

// ValidatorOutstandingRewardsRecord represents outstanding rewards record
type ValidatorOutstandingRewardsRecord struct {
	ValidatorAddress   wasmx.ValidatorAddressString `json:"validator_address"`
	OutstandingRewards []wasmx.DecCoin              `json:"outstanding_rewards"`
}

// ValidatorAccumulatedCommission represents accumulated commission
type ValidatorAccumulatedCommission struct {
	Commission []wasmx.DecCoin `json:"commission"`
}

// ValidatorAccumulatedCommissionRecord represents accumulated commission record
type ValidatorAccumulatedCommissionRecord struct {
	ValidatorAddress wasmx.ValidatorAddressString    `json:"validator_address"`
	Accumulated      ValidatorAccumulatedCommission `json:"accumulated"`
}

// ValidatorHistoricalRewards represents historical rewards
type ValidatorHistoricalRewards struct {
	CumulativeRewardRatio []wasmx.DecCoin `json:"cumulative_reward_ratio"`
	ReferenceCount        uint32          `json:"reference_count"`
}

// ValidatorCurrentRewards represents current rewards
type ValidatorCurrentRewards struct {
	Rewards []wasmx.DecCoin `json:"rewards"`
	Period  uint64          `json:"period"`
}

// ValidatorHistoricalRewardsRecord represents historical rewards record
type ValidatorHistoricalRewardsRecord struct {
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
	Period           uint64                       `json:"period"`
	Rewards          ValidatorHistoricalRewards   `json:"rewards"`
}

// ValidatorCurrentRewardsRecord represents current rewards record
type ValidatorCurrentRewardsRecord struct {
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
	Rewards          ValidatorCurrentRewards      `json:"rewards"`
}

// DelegatorStartingInfo represents delegator starting information
type DelegatorStartingInfo struct {
	PreviousPeriod uint64  `json:"previous_period"`
	Stake          float64 `json:"stake"`
	Height         uint64  `json:"height"` // LegacyDec
}

// DelegatorStartingInfoRecord represents delegator starting info record
type DelegatorStartingInfoRecord struct {
	DelegatorAddress wasmx.Bech32String           `json:"delegator_address"`
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
	StartingInfo     DelegatorStartingInfo        `json:"starting_info"`
}

// ValidatorSlashEvent represents a validator slash event
type ValidatorSlashEvent struct {
	ValidatorPeriod uint64  `json:"validator_period"`
	Fraction        float64 `json:"fraction"` // LegacyDec
}

// ValidatorSlashEventRecord represents a slash event record
type ValidatorSlashEventRecord struct {
	ValidatorAddress     wasmx.ValidatorAddressString `json:"validator_address"`
	Height               uint64                       `json:"height"`
	Period               uint64                       `json:"period"`
	ValidatorSlashEvent  ValidatorSlashEvent          `json:"validator_slash_event"`
}

// FeePool represents the fee pool
type FeePool struct {
	CommunityPool []wasmx.DecCoin `json:"community_pool"`
}

// ValidatorOutstandingRewards represents outstanding rewards
type ValidatorOutstandingRewards struct {
	Rewards []wasmx.DecCoin `json:"rewards"`
}

// ValidatorSlashEvents represents multiple slash events
type ValidatorSlashEvents struct {
	ValidatorSlashEvents []ValidatorSlashEvent `json:"validator_slash_events"`
}

// CommunityPoolSpendProposal represents a community pool spend proposal
type CommunityPoolSpendProposal struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Recipient   string         `json:"recipient"`
	Amount      []wasmx.Coin   `json:"amount"`
}

// CommunityPoolSpendProposalWithDeposit represents a proposal with deposit
type CommunityPoolSpendProposalWithDeposit struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Recipient   string `json:"recipient"`
	Amount      string `json:"amount"`
	Deposit     string `json:"deposit"`
}

// DelegationDelegatorReward represents delegation rewards
type DelegationDelegatorReward struct {
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
	Reward           []wasmx.DecCoin              `json:"reward"`
}

// GenesisState represents the distribution module's genesis state
type GenesisState struct {
	Params                           Params                                  `json:"params"`
	FeePool                          FeePool                                 `json:"fee_pool"`
	DelegatorWithdrawInfos           []DelegatorWithdrawInfo                 `json:"delegator_withdraw_infos"`
	PreviousProposer                 wasmx.Bech32String                      `json:"previous_proposer"`
	OutstandingRewards               []ValidatorOutstandingRewardsRecord     `json:"outstanding_rewards"`
	ValidatorAccumulatedCommissions  []ValidatorAccumulatedCommissionRecord  `json:"validator_accumulated_commissions"`
	ValidatorHistoricalRewards       []ValidatorHistoricalRewardsRecord      `json:"validator_historical_rewards"`
	ValidatorCurrentRewards          []ValidatorCurrentRewardsRecord         `json:"validator_current_rewards"`
	DelegatorStartingInfos           []DelegatorStartingInfoRecord           `json:"delegator_starting_infos"`
	ValidatorSlashEvents             []ValidatorSlashEventRecord             `json:"validator_slash_events"`
	BaseDenom                        string                                  `json:"base_denom"`
	RewardsDenom                     string                                  `json:"rewards_denom"`
}

// Message types
type MsgSetWithdrawAddress struct {
	DelegatorAddress wasmx.Bech32String `json:"delegator_address"`
	WithdrawAddress  wasmx.Bech32String `json:"withdraw_address"`
}

type MsgSetWithdrawAddressResponse struct{}

type MsgWithdrawDelegatorReward struct {
	DelegatorAddress wasmx.Bech32String           `json:"delegator_address"`
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
}

type MsgWithdrawDelegatorRewardResponse struct {
	Amount []wasmx.Coin `json:"amount"`
}

type MsgWithdrawValidatorCommission struct {
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
}

type MsgWithdrawValidatorCommissionResponse struct {
	Amount []wasmx.Coin `json:"amount"`
}

type MsgFundCommunityPool struct {
	Amount    []wasmx.Coin       `json:"amount"`
	Depositor wasmx.Bech32String `json:"depositor"`
}

type MsgFundCommunityPoolResponse struct{}

type MsgUpdateParams struct {
	Authority wasmx.Bech32String `json:"authority"`
	Params    Params             `json:"params"`
}

type MsgUpdateParamsResponse struct{}

type MsgCommunityPoolSpend struct {
	Authority wasmx.Bech32String `json:"authority"`
	Recipient string             `json:"recipient"`
	Amount    []wasmx.Coin       `json:"amount"`
}

type MsgCommunityPoolSpendResponse struct{}

type MsgDepositValidatorRewardsPool struct {
	Depositor        wasmx.Bech32String           `json:"depositor"`
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
	Amount           []wasmx.Coin                 `json:"amount"`
}

type MsgDepositValidatorRewardsPoolResponse struct{}

// Query types
type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QueryValidatorDistributionInfoRequest struct {
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
}

type QueryValidatorDistributionInfoResponse struct {
	OperatorAddress  wasmx.ValidatorAddressString `json:"operator_address"`
	SelfBondRewards  []wasmx.DecCoin              `json:"self_bond_rewards"`
	Commission       []wasmx.DecCoin              `json:"commission"`
}

type QueryValidatorOutstandingRewardsRequest struct {
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
}

type QueryValidatorOutstandingRewardsResponse struct {
	Rewards ValidatorOutstandingRewards `json:"rewards"`
}

type QueryValidatorCommissionRequest struct {
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
}

type QueryValidatorCommissionResponse struct {
	Commission ValidatorAccumulatedCommission `json:"commission"`
}

type QueryValidatorSlashesRequest struct {
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
	StartingHeight   uint64                       `json:"starting_height"`
	EndingHeight     uint64                       `json:"ending_height"`
	Pagination       wasmx.PageRequest            `json:"pagination"`
}

type QueryValidatorSlashesResponse struct {
	Slashes    []ValidatorSlashEvent `json:"slashes"`
	Pagination wasmx.PageResponse    `json:"pagination"`
}

type QueryDelegationRewardsRequest struct {
	DelegatorAddress wasmx.Bech32String           `json:"delegator_address"`
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
}

type QueryDelegationRewardsResponse struct {
	Rewards []wasmx.DecCoin `json:"rewards"`
}

type QueryDelegationTotalRewardsRequest struct {
	DelegatorAddress wasmx.Bech32String `json:"delegator_address"`
}

type QueryDelegationTotalRewardsResponse struct {
	Rewards []DelegationDelegatorReward `json:"rewards"`
	Total   []wasmx.DecCoin             `json:"total"`
}

type QueryDelegatorValidatorsRequest struct {
	DelegatorAddress wasmx.Bech32String `json:"delegator_address"`
}

type QueryDelegatorValidatorsResponse struct {
	Validators []wasmx.Bech32String `json:"validators"`
}

type QueryDelegatorWithdrawAddressRequest struct {
	DelegatorAddress wasmx.Bech32String `json:"delegator_address"`
}

type QueryDelegatorWithdrawAddressResponse struct {
	WithdrawAddress wasmx.Bech32String `json:"withdraw_address"`
}

type QueryCommunityPoolRequest struct{}

type QueryCommunityPoolResponse struct {
	Pool []wasmx.DecCoin `json:"pool"`
}

// Calldata structure
type CallData struct {
	GetParams *MsgGetParams `json:"GetParams"`

	InitGenesis *MsgInitGenesis `json:"InitGenesis"`
}

type MsgInitGenesis struct{}

type MsgGetParams struct{}
