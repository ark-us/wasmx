package lib

import (
	"time"

	sdkmath "cosmossdk.io/math"
	consensus "github.com/loredanacirstea/wasmx-env-consensus/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "staking"

// Bond status constants
const (
	Unspecified int32 = 0
	Unbonded    int32 = 1
	Unbonding   int32 = 2
	Bonded      int32 = 3
)

// Bond status strings
const (
	UnspecifiedS = "BOND_STATUS_UNSPECIFIED"
	UnbondedS    = "BOND_STATUS_UNBONDED"
	UnbondingS   = "BOND_STATUS_UNBONDING"
	BondedS      = "BOND_STATUS_BONDED"
)

// Hook constants
const (
	AfterValidatorCreated = "AfterValidatorCreated"
	AfterValidatorBonded  = "AfterValidatorBonded"
)

const TypeUrl_MsgCreateValidator = "/cosmos.staking.v1beta1.MsgCreateValidator"

// Type aliases
type BondStatusNumber = int32
type BondStatusString = string

// Description represents validator description
type Description struct {
	Moniker         string `json:"moniker"`
	Identity        string `json:"identity"`
	Website         string `json:"website"`
	SecurityContact string `json:"security_contact"`
	Details         string `json:"details"`
}

// CommissionRates represents validator commission rates
type CommissionRates struct {
	Rate          string `json:"rate"`            // f64 as string
	MaxRate       string `json:"max_rate"`        // f64 as string
	MaxChangeRate string `json:"max_change_rate"` // f64 as string
}

// MsgCreateValidator represents a create validator message
type MsgCreateValidator struct {
	Description       Description        `json:"description"`
	Commission        CommissionRates    `json:"commission"`
	MinSelfDelegation sdkmath.Int        `json:"min_self_delegation"`
	DelegatorAddress  string             `json:"delegator_address"`
	ValidatorAddress  wasmx.Bech32String `json:"validator_address"`
	Pubkey            *wasmx.PublicKey   `json:"pubkey"`
	Value             wasmx.Coin         `json:"value"`
}

func (MsgCreateValidator) TypeUrl() string {
	return TypeUrl_MsgCreateValidator
}

type MsgCreateValidatorResponse struct{}

// ValidatorInfo represents validator info from tendermint
type ValidatorInfo struct {
	Address          wasmx.HexString `json:"address"`
	PubKey           string          `json:"pub_key"` // base64
	VotingPower      int64           `json:"voting_power"`
	ProposerPriority int64           `json:"proposer_priority"`
}

// Delegation represents a delegation
type Delegation struct {
	DelegatorAddress wasmx.Bech32String           `json:"delegator_address"`
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
	Amount           sdkmath.Int                  `json:"amount"`
}

// LastValidatorPower tracks validator power
type LastValidatorPower struct {
	Address string `json:"address"`
	Power   int64  `json:"power"`
}

// UnbondingDelegationEntry represents an unbonding delegation entry
type UnbondingDelegationEntry struct {
	CreationHeight          int64       `json:"creation_height"`
	CompletionTime          int64       `json:"completion_time"`
	InitialBalance          sdkmath.Int `json:"initial_balance"`
	Balance                 sdkmath.Int `json:"balance"`
	UnbondingId             uint64      `json:"unbonding_id"`
	UnbondingOnHoldRefCount int64       `json:"unbonding_on_hold_ref_count"`
}

// UnbondingDelegation represents an unbonding delegation
type UnbondingDelegation struct {
	DelegatorAddress string                       `json:"delegator_address"`
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
	Entries          []UnbondingDelegationEntry   `json:"entries"`
}

// RedelegationEntry represents a redelegation entry
type RedelegationEntry struct{}

// Redelegation represents a redelegation
type Redelegation struct {
	DelegatorAddress    string                       `json:"delegator_address"`
	ValidatorSrcAddress wasmx.ValidatorAddressString `json:"validator_src_address"`
	ValidatorDstAddress wasmx.ValidatorAddressString `json:"validator_dst_address"`
	Entries             []RedelegationEntry          `json:"entries"`
}

// GenesisState represents the staking module's genesis state
type GenesisState struct {
	Params               Params                `json:"params"`
	LastTotalPower       string                `json:"last_total_power"`
	LastValidatorPowers  []LastValidatorPower  `json:"last_validator_powers"`
	Validators           []Validator           `json:"validators"`
	Delegations          []Delegation          `json:"delegations"`
	UnbondingDelegations []UnbondingDelegation `json:"unbonding_delegations"`
	Redelegations        []Redelegation        `json:"redelegations"`
	BaseDenom            string                `json:"base_denom"`
}

// InitGenesisResponse represents the response from init genesis
type InitGenesisResponse struct {
	Updates []consensus.ValidatorUpdate `json:"updates"`
}

// Params represents staking module parameters
type Params struct {
	UnbondingTime     string `json:"unbonding_time"`
	MaxValidators     uint32 `json:"max_validators"`
	MaxEntries        uint32 `json:"max_entries"`
	HistoricalEntries uint32 `json:"historical_entries"`
	BondDenom         string `json:"bond_denom"`
	MinCommissionRate string `json:"min_commission_rate"`
}

// Commission represents validator commission
type Commission struct {
	CommissionRates CommissionRates `json:"commission_rates"`
	UpdateTime      time.Time       `json:"update_time"`
}

// Validator represents a validator
type Validator struct {
	OperatorAddress         wasmx.Bech32String `json:"operator_address"`
	ConsensusPubkey         *wasmx.PublicKey   `json:"consensus_pubkey"`
	Jailed                  bool               `json:"jailed"`
	Status                  BondStatusString   `json:"status"`
	Tokens                  sdkmath.Int        `json:"tokens"`
	DelegatorShares         string             `json:"delegator_shares"` // f64 as string
	Description             Description        `json:"description"`
	UnbondingHeight         int64              `json:"unbonding_height"`
	UnbondingTime           time.Time          `json:"unbonding_time"`
	Commission              Commission         `json:"commission"`
	MinSelfDelegation       sdkmath.Int        `json:"min_self_delegation"`
	UnbondingOnHoldRefCount int64              `json:"unbonding_on_hold_ref_count"`
	UnbondingIds            []uint64           `json:"unbonding_ids"`
}

// ValidatorSimple represents a simplified validator (without tokens/shares)
type ValidatorSimple struct {
	OperatorAddress         wasmx.Bech32String `json:"operator_address"`
	ConsensusPubkey         *wasmx.PublicKey   `json:"consensus_pubkey"`
	Jailed                  bool               `json:"jailed"`
	Status                  BondStatusString   `json:"status"`
	Description             Description        `json:"description"`
	UnbondingHeight         int64              `json:"unbonding_height"`
	UnbondingTime           time.Time          `json:"unbonding_time"`
	Commission              Commission         `json:"commission"`
	MinSelfDelegation       sdkmath.Int        `json:"min_self_delegation"`
	UnbondingOnHoldRefCount int64              `json:"unbonding_on_hold_ref_count"`
	UnbondingIds            []uint64           `json:"unbonding_ids"`
}

// FromValidator creates a ValidatorSimple from a Validator
func (vs *ValidatorSimple) FromValidator(v Validator) {
	vs.OperatorAddress = v.OperatorAddress
	vs.ConsensusPubkey = v.ConsensusPubkey
	vs.Jailed = v.Jailed
	vs.Status = v.Status
	vs.Description = v.Description
	vs.UnbondingHeight = v.UnbondingHeight
	vs.UnbondingTime = v.UnbondingTime
	vs.Commission = v.Commission
	vs.MinSelfDelegation = v.MinSelfDelegation
	vs.UnbondingOnHoldRefCount = v.UnbondingOnHoldRefCount
	vs.UnbondingIds = v.UnbondingIds
}

// ToValidator creates a Validator from ValidatorSimple
func (vs ValidatorSimple) ToValidator(tokens sdkmath.Int, shares string) Validator {
	return Validator{
		OperatorAddress:         vs.OperatorAddress,
		ConsensusPubkey:         vs.ConsensusPubkey,
		Jailed:                  vs.Jailed,
		Status:                  vs.Status,
		Tokens:                  tokens,
		DelegatorShares:         shares,
		Description:             vs.Description,
		UnbondingHeight:         vs.UnbondingHeight,
		UnbondingTime:           vs.UnbondingTime,
		Commission:              vs.Commission,
		MinSelfDelegation:       vs.MinSelfDelegation,
		UnbondingOnHoldRefCount: vs.UnbondingOnHoldRefCount,
		UnbondingIds:            vs.UnbondingIds,
	}
}

// RedelegationEntryResponse represents a redelegation entry response
type RedelegationEntryResponse struct {
	RedelegationEntry RedelegationEntry `json:"redelegation_entry"`
	Balance           sdkmath.Int       `json:"balance"`
}

// RedelegationResponse represents a redelegation response
type RedelegationResponse struct {
	Redelegation Redelegation                `json:"redelegation"`
	Entries      []RedelegationEntryResponse `json:"entries"`
}

// HistoricalInfo represents historical validator info
type HistoricalInfo struct {
	Header consensus.Header `json:"header"`
	Valset []Validator      `json:"valset"`
}

// MsgDelegate represents a delegation message
type MsgDelegate struct {
	DelegatorAddress string                       `json:"delegator_address"`
	ValidatorAddress wasmx.ValidatorAddressString `json:"validator_address"`
	Amount           wasmx.Coin                   `json:"amount"`
}

// MsgUpdateValidators represents a validator update message
type MsgUpdateValidators struct {
	Updates []consensus.ValidatorUpdate `json:"updates"`
}

// Query types
type QueryGetAllValidators struct{}

type QueryGetAllValidatorInfos struct{}

type QueryValidatorsRequest struct {
	Pagination wasmx.PageRequest `json:"pagination"`
}

type QueryValidatorsResponse struct {
	Validators []Validator        `json:"validators"`
	Pagination wasmx.PageResponse `json:"pagination"`
}

type QueryValidatorInfosResponse struct {
	Validators []ValidatorSimple  `json:"validators"`
	Pagination wasmx.PageResponse `json:"pagination"`
}

type QueryValidatorRequest struct {
	ValidatorAddr wasmx.ValidatorAddressString `json:"validator_addr"`
}

type QueryValidatorResponse struct {
	Validator Validator `json:"validator"`
}

type QueryValidatorDelegationsRequest struct {
	ValidatorAddr wasmx.ValidatorAddressString `json:"validator_addr"`
	Pagination    wasmx.PageRequest            `json:"pagination"`
}

type QueryValidatorDelegationsResponse struct {
	DelegationResponses []DelegationResponse `json:"delegation_responses"`
	Pagination          wasmx.PageResponse   `json:"pagination"`
}

type QueryValidatorUnbondingDelegationsRequest struct {
	ValidatorAddr wasmx.ValidatorAddressString `json:"validator_addr"`
	Pagination    wasmx.PageRequest            `json:"pagination"`
}

type QueryValidatorUnbondingDelegationsResponse struct {
	UnbondingResponses []UnbondingDelegation `json:"unbonding_responses"`
	Pagination         wasmx.PageResponse    `json:"pagination"`
}

type QueryDelegationRequest struct {
	DelegatorAddr wasmx.Bech32String           `json:"delegator_addr"`
	ValidatorAddr wasmx.ValidatorAddressString `json:"validator_addr"`
}

// DelegationCosmos represents a cosmos delegation
type DelegationCosmos struct {
	DelegatorAddress wasmx.Bech32String `json:"delegator_address"`
	ValidatorAddress wasmx.Bech32String `json:"validator_address"`
	Shares           sdkmath.Int        `json:"shares"`
}

// DelegationResponse represents a delegation response
type DelegationResponse struct {
	Delegation DelegationCosmos `json:"delegation"`
	Balance    wasmx.Coin       `json:"balance"`
}

type QueryDelegationResponse struct {
	DelegationResponse DelegationResponse `json:"delegation_response"`
}

type QueryUnbondingDelegationRequest struct {
	DelegatorAddr wasmx.Bech32String           `json:"delegator_addr"`
	ValidatorAddr wasmx.ValidatorAddressString `json:"validator_addr"`
}

type QueryUnbondingDelegationResponse struct {
	Unbond UnbondingDelegation `json:"unbond"`
}

type QueryDelegatorDelegationsRequest struct {
	DelegatorAddr wasmx.Bech32String `json:"delegator_addr"`
	Pagination    wasmx.PageRequest  `json:"pagination"`
}

type QueryDelegatorDelegationsResponse struct {
	DelegationResponses []DelegationResponse `json:"delegation_responses"`
	Pagination          wasmx.PageResponse   `json:"pagination"`
}

type QueryDelegatorUnbondingDelegationsRequest struct {
	DelegatorAddr wasmx.Bech32String `json:"delegator_addr"`
	Pagination    wasmx.PageRequest  `json:"pagination"`
}

type QueryDelegatorUnbondingDelegationsResponse struct {
	UnbondingResponses []UnbondingDelegation `json:"unbonding_responses"`
	Pagination         wasmx.PageResponse    `json:"pagination"`
}

type QueryRedelegationsRequest struct {
	DelegatorAddr    wasmx.Bech32String `json:"delegator_addr"`
	SrcValidatorAddr wasmx.Bech32String `json:"src_validator_addr"`
	DstValidatorAddr wasmx.Bech32String `json:"dst_validator_addr"`
	Pagination       wasmx.PageRequest  `json:"pagination"`
}

type QueryRedelegationsResponse struct {
	RedelegationResponses []RedelegationResponse `json:"redelegation_responses"`
	Pagination            wasmx.PageResponse     `json:"pagination"`
}

type QueryDelegatorValidatorsRequest struct {
	DelegatorAddr wasmx.Bech32String `json:"delegator_addr"`
	Pagination    wasmx.PageRequest  `json:"pagination"`
}

type QueryDelegatorValidatorsResponse struct {
	Validators []Validator        `json:"validators"`
	Pagination wasmx.PageResponse `json:"pagination"`
}

type QueryDelegatorValidatorRequest struct {
	DelegatorAddr wasmx.Bech32String           `json:"delegator_addr"`
	ValidatorAddr wasmx.ValidatorAddressString `json:"validator_addr"`
}

type QueryDelegatorValidatorResponse struct {
	Validator Validator `json:"validator"`
}

type QueryHistoricalInfoRequest struct {
	Height int64 `json:"height"`
}

type QueryHistoricalInfoResponse struct {
	Hist HistoricalInfo `json:"hist"`
}

type QueryPoolRequest struct{}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

// Pool represents the staking pool
type Pool struct {
	NotBondedTokens sdkmath.Int `json:"not_bonded_tokens"`
	BondedTokens    sdkmath.Int `json:"bonded_tokens"`
}

type QueryPoolResponse struct {
	Pool Pool `json:"pool"`
}

type QueryContractInfoResponse struct {
	ContractInfo *wasmx.ContractInfo `json:"contract_info"`
}

type QueryIsValidatorJailed struct {
	Consaddr wasmx.ConsensusAddressString `json:"consaddr"`
}

type QueryIsValidatorJailedResponse struct {
	Jailed bool `json:"jailed"`
}

// Slashing messages
type MsgSlash struct {
	Consaddr         wasmx.ConsensusAddressString `json:"consaddr"`
	InfractionHeight int64                        `json:"infractionHeight"`
	Power            int64                        `json:"power"`
	SlashFactor      string                       `json:"slashFactor"`
}

type MsgSlashWithInfractionReason struct {
	Consaddr         wasmx.ConsensusAddressString `json:"consaddr"`
	InfractionHeight int64                        `json:"infractionHeight"`
	Power            int64                        `json:"power"`
	SlashFactor      string                       `json:"slashFactor"`
	InfractionReason string                       `json:"infractionReason"`
}

type MsgSlashWithInfractionReasonResponse struct {
	AmountBurned sdkmath.Int `json:"amount_burned"`
}

type MsgJail struct {
	Consaddr wasmx.ConsensusAddressString `json:"consaddr"`
}

type MsgUnjail struct {
	Address wasmx.Bech32String `json:"address"`
}

type MsgUnjailResponse struct{}

type QueryConsensusAddressByOperatorAddress struct {
	ValidatorAddr wasmx.Bech32String `json:"validator_addr"`
}

type QueryConsensusAddressByOperatorAddressResponse struct {
	Consaddr wasmx.ConsensusAddressString `json:"consaddr"`
}

// Utility function to get validator from create message
func GetValidatorFromMsgCreate(req MsgCreateValidator) Validator {
	return Validator{
		OperatorAddress:         req.ValidatorAddress,
		ConsensusPubkey:         req.Pubkey,
		Jailed:                  false,
		Status:                  BondedS,
		Tokens:                  req.Value.Amount,
		DelegatorShares:         "0.0",
		Description:             req.Description,
		UnbondingHeight:         0,
		UnbondingTime:           time.Unix(0, 0),
		Commission:              Commission{CommissionRates: req.Commission, UpdateTime: time.Unix(0, 0)},
		MinSelfDelegation:       req.MinSelfDelegation,
		UnbondingOnHoldRefCount: 0,
		UnbondingIds:            []uint64{},
	}
}

// Calldata structure
type CallData struct {
	GetParams *MsgGetParams `json:"GetParams"`

	InitGenesis *MsgInitGenesis `json:"InitGenesis"`
}

type MsgInitGenesis struct{}

type MsgGetParams struct{}
