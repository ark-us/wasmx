package keeper

import (
	"context"
	"encoding/json"
	"time"

	addresscodec "cosmossdk.io/core/address"
	math "cosmossdk.io/math"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"

	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// BondedRatio the fraction of the staking tokens which are currently bonded
func (k Keeper) BondedRatio(ctx context.Context) (math.LegacyDec, error) {
	return math.LegacyZeroDec(), nil
}

// StakingTokenSupply staking tokens from the total supply
func (k Keeper) StakingTokenSupply(ctx context.Context) (math.Int, error) {
	return math.ZeroInt(), nil
}

// ValidatorAddressCodec returns the app validator address codec.
func (k Keeper) ValidatorAddressCodec() addresscodec.Codec {
	return k.validatorAddressCodec
}

// ConsensusAddressCodec returns the app consensus address codec.
func (k Keeper) ConsensusAddressCodec() addresscodec.Codec {
	return k.consensusAddressCodec
}

// Delegation gets the delegation interface for a particular set of delegator and validator addresses
func (k Keeper) Delegation(ctx context.Context, addrDel sdk.AccAddress, addrVal sdk.ValAddress) (stakingtypes.DelegationI, error) {
	return nil, nil
}

// GetAllValidators gets the set of all validators with no limits, used during genesis dump
func (k Keeper) GetAllValidators(goCtx context.Context) (validators []stakingtypes.Validator, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgbz := []byte(`{"GetAllValidators":{}}`)
	res, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_STAKING,
		Contract: wasmxtypes.ROLE_STAKING,
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
	var validatorsResp stakingtypes.QueryValidatorsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &validatorsResp)
	if err != nil {
		return nil, err
	}
	return validatorsResp.Validators, nil
}

// IsValidatorJailed checks and returns boolean of a validator status jailed or not.
func (k Keeper) IsValidatorJailed(ctx context.Context, addr sdk.ConsAddress) (bool, error) {
	return false, nil
}

// IterateValidators iterates through the validator set and perform the provided function
func (k Keeper) IterateValidators(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error {
	return nil
}

// jail a validator
func (k Keeper) Jail(ctx context.Context, consAddr sdk.ConsAddress) error {
	return nil
}

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx context.Context) (uint32, error) {
	return 100, nil
}

// Infraction was committed at the current height or at a past height,
// but not at a height in the future
func (k Keeper) Slash(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor math.LegacyDec) (math.Int, error) {
	return math.ZeroInt(), nil
}

// SlashWithInfractionReason implementation doesn't require the infraction (types.Infraction) to work but is required by Interchain Security.
func (k Keeper) SlashWithInfractionReason(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor math.LegacyDec, _ types.Infraction) (math.Int, error) {
	return k.Slash(ctx, consAddr, infractionHeight, power, slashFactor)
}

// unjail a validator
func (k Keeper) Unjail(ctx context.Context, consAddr sdk.ConsAddress) error {
	return nil
}

// Validator gets the Validator interface for a particular address
func (k Keeper) Validator(ctx context.Context, address sdk.ValAddress) (types.ValidatorI, error) {
	return nil, nil
}

// ValidatorByConsAddr gets the validator interface for a particular pubkey
func (k Keeper) ValidatorByConsAddr(ctx context.Context, addr sdk.ConsAddress) (types.ValidatorI, error) {
	return nil, nil
}

// GetAllDelegatorDelegations returns all delegations of a delegator
func (k Keeper) GetAllDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress) ([]types.Delegation, error) {
	return make([]types.Delegation, 0), nil
}

// GetAllSDKDelegations returns all delegations used during genesis dump
// TODO: remove this func, change all usage for iterate functionality
func (k Keeper) GetAllSDKDelegations(ctx context.Context) (delegations []types.Delegation, err error) {
	return make([]types.Delegation, 0), nil
}

// IterateDelegations iterates through all of the delegations from a delegator
func (k Keeper) IterateDelegations(ctx context.Context, delAddr sdk.AccAddress,
	fn func(index int64, del types.DelegationI) (stop bool),
) error {
	return nil
}

// GetHistoricalInfo gets the historical info at a given height
func (k Keeper) GetHistoricalInfo(ctx context.Context, height int64) (types.HistoricalInfo, error) {
	return types.HistoricalInfo{}, nil
}

// UnbondingTime - The time duration for unbonding
func (k Keeper) UnbondingTime(ctx context.Context) (time.Duration, error) {
	return time.Minute, nil
}

// GetParams gets the x/staking module parameters.
func (k Keeper) GetParams(ctx context.Context) (params types.Params, err error) {
	return types.Params{}, nil
}

// IterateBondedValidatorsByPower iterates through the bonded validator set and perform the provided function
func (k Keeper) IterateBondedValidatorsByPower(ctx context.Context, fn func(index int64, validator types.ValidatorI) (stop bool)) error {
	return nil
}

// TotalBondedTokens total staking tokens supply which is bonded
func (k Keeper) TotalBondedTokens(ctx context.Context) (math.Int, error) {
	return math.ZeroInt(), nil
}

// ApplyAndReturnValidatorSetUpdates applies and return accumulated updates to the bonded validator set. Also,
// * Updates the active valset as keyed by LastValidatorPowerKey.
// * Updates the total power as keyed by LastTotalPowerKey.
// * Updates validator status' according to updated powers.
// * Updates the fee pool bonded vs not-bonded tokens.
// * Updates relevant indices.
// It gets called once after genesis, another time maybe after genesis transactions,
// then once at every EndBlock.
//
// CONTRACT: Only validators with non-zero power or zero-power that were bonded
// at the previous block height or were removed from the validator set entirely
// are returned to CometBFT.
func (k Keeper) ApplyAndReturnValidatorSetUpdates(ctx context.Context) (updates []abci.ValidatorUpdate, err error) {
	return make([]abci.ValidatorUpdate, 0), nil
}

// GetAllDelegations returns all delegations used during genesis dump.
func (k Keeper) GetAllDelegations(ctx context.Context) (delegations []types.Delegation, err error) {
	return make([]types.Delegation, 0), err
}

// IterateRedelegations iterates through all redelegations.
func (k Keeper) IterateRedelegations(ctx context.Context, fn func(index int64, red types.Redelegation) (stop bool)) error {
	return nil
}

// SetRedelegation sets a redelegation and associated index.
func (k Keeper) SetRedelegation(ctx context.Context, red types.Redelegation) error {
	return nil
}

// IterateUnbondingDelegations iterates through all of the unbonding delegations.
func (k Keeper) IterateUnbondingDelegations(ctx context.Context, fn func(index int64, ubd types.UnbondingDelegation) (stop bool)) error {
	return nil
}

// SetUnbondingDelegation sets the unbonding delegation and associated index.
func (k Keeper) SetUnbondingDelegation(ctx context.Context, ubd types.UnbondingDelegation) error {
	return nil
}

// GetValidator gets a single validator
func (k Keeper) GetValidator(ctx context.Context, addr sdk.ValAddress) (validator types.Validator, err error) {
	return types.Validator{}, nil
}

// SetValidator sets the main record holding validator details
func (k Keeper) SetValidator(ctx context.Context, validator types.Validator) error {
	return nil
}

// IterateLastValidators iterates through the active validator set and perform the provided function
func (k Keeper) IterateLastValidators(ctx context.Context, fn func(index int64, validator types.ValidatorI) (stop bool)) error {
	return nil
}

// SetValidatorByConsAddr sets a validator by conesensus address
func (k Keeper) SetValidatorByConsAddr(ctx context.Context, validator types.Validator) error {
	return nil
}

// PowerReduction - is the amount of staking tokens required for 1 unit of consensus-engine power.
// Currently, this returns a global variable that the app developer can tweak.
// TODO: we might turn this into an on-chain param:
// https://github.com/cosmos/cosmos-sdk/issues/8365
func (k Keeper) PowerReduction(ctx context.Context) math.Int {
	return sdk.DefaultPowerReduction
}

// WriteValidators returns a slice of bonded genesis validators.
func WriteValidators(ctx sdk.Context, keeper *Keeper) (vals []cmttypes.GenesisValidator, returnErr error) {
	err := keeper.IterateLastValidators(ctx, func(_ int64, validator types.ValidatorI) (stop bool) {
		pk, err := validator.ConsPubKey()
		if err != nil {
			returnErr = err
			return true
		}
		cmtPk, err := cryptocodec.ToCmtPubKeyInterface(pk)
		if err != nil {
			returnErr = err
			return true
		}

		vals = append(vals, cmttypes.GenesisValidator{
			Address: sdk.ConsAddress(cmtPk.Address()).Bytes(),
			PubKey:  cmtPk,
			Power:   validator.GetConsensusPower(keeper.PowerReduction(ctx)),
			Name:    validator.GetMoniker(),
		})

		return false
	})
	if err != nil {
		return nil, err
	}

	return
}
