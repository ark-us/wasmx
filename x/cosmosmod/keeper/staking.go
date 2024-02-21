package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	addresscodec "cosmossdk.io/core/address"
	errors "cosmossdk.io/errors"
	math "cosmossdk.io/math"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"mythos/v1/x/cosmosmod/types"
	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// BondedRatio the fraction of the staking tokens which are currently bonded
func (k KeeperStaking) BondedRatio(goCtx context.Context) (math.LegacyDec, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("BondedRatio not implemented")
	return math.LegacyOneDec(), nil
}

// StakingTokenSupply staking tokens from the total supply
func (k KeeperStaking) StakingTokenSupply(goCtx context.Context) (math.Int, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	derc20Address, err := k.GetDERC20Address(ctx)
	if err != nil {
		return math.NewInt(0), err
	}
	msgbz := []byte(`{"totalSupply":{}}`)
	res, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_STAKING,
		Contract: derc20Address.String(),
		Msg:      msgbz,
	})
	if err != nil {
		return math.NewInt(0), err
	}
	var resp wasmxtypes.ContractResponse
	err = json.Unmarshal(res.Data, &resp)
	if err != nil {
		return math.NewInt(0), err
	}
	var response types.QueryTotalSupplyResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &response)
	if err != nil {
		return math.NewInt(0), err
	}
	return response.Supply.Amount, nil
}

// ValidatorAddressCodec returns the app validator address codec.
func (k KeeperStaking) ValidatorAddressCodec() addresscodec.Codec {
	return k.validatorAddressCodec
}

// ConsensusAddressCodec returns the app consensus address codec.
func (k KeeperStaking) ConsensusAddressCodec() addresscodec.Codec {
	return k.consensusAddressCodec
}

// Delegation gets the delegation interface for a particular set of delegator and validator addresses
func (k KeeperStaking) Delegation(goCtx context.Context, addrDel sdk.AccAddress, addrVal sdk.ValAddress) (stakingtypes.DelegationI, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgbz := []byte(fmt.Sprintf(`{"GetDelegation":{"delegator_addr":"%s","validator_addr":"%s"}}`, addrDel.String(), addrVal.String()))
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
	var validatorsResp stakingtypes.QueryDelegationResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &validatorsResp)
	if err != nil {
		return nil, err
	}
	return validatorsResp.DelegationResponse.Delegation, nil
}

// GetAllValidators gets the set of all validators with no limits, used during genesis dump
func (k KeeperStaking) GetAllValidators(goCtx context.Context) (validators []stakingtypes.Validator, err error) {
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
func (k KeeperStaking) IsValidatorJailed(goCtx context.Context, addr sdk.ConsAddress) (bool, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("IsValidatorJailed not implemented")
	return false, nil
}

// IterateValidators iterates through the validator set and perform the provided function
func (k KeeperStaking) IterateValidators(goCtx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error {
	// TODO more efficient - can we iterate on the contract storage directly?
	validators, err := k.GetAllValidators(goCtx)
	if err != nil {
		return err
	}
	for i, validator := range validators {
		// we need to unpack, for usage with GetCachedValue() by other modules
		var pkI cryptotypes.PubKey
		err = k.InterfaceRegistry.UnpackAny(validator.ConsensusPubkey, &pkI)
		if err != nil {
			panic(err)
		}

		if fn(int64(i), validator) {
			break
		}
	}
	return nil
}

// jail a validator
func (k KeeperStaking) Jail(goCtx context.Context, consAddr sdk.ConsAddress) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Jail not implemented")
	return nil
}

// MaxValidators - Maximum number of validators
func (k KeeperStaking) MaxValidators(goCtx context.Context) (uint32, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("MaxValidators not implemented")
	return 100, nil
}

// Infraction was committed at the current height or at a past height,
// but not at a height in the future
func (k KeeperStaking) Slash(goCtx context.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor math.LegacyDec) (math.Int, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Slash not implemented")
	return math.ZeroInt(), nil
}

// SlashWithInfractionReason implementation doesn't require the infraction (types.Infraction) to work but is required by Interchain Security.
func (k KeeperStaking) SlashWithInfractionReason(goCtx context.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor math.LegacyDec, _ stakingtypes.Infraction) (math.Int, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("SlashWithInfractionReason not implemented")
	return k.Slash(goCtx, consAddr, infractionHeight, power, slashFactor)
}

// unjail a validator
func (k KeeperStaking) Unjail(goCtx context.Context, consAddr sdk.ConsAddress) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Unjail not implemented")
	return nil
}

// Validator gets the Validator interface for a particular address
func (k KeeperStaking) Validator(goCtx context.Context, address sdk.ValAddress) (stakingtypes.ValidatorI, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msgbz := []byte(fmt.Sprintf(`{"GetValidator":{"validator_addr":"%s"}}`, address.String()))
	res1, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_STAKING,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	var resp wasmxtypes.ContractResponse
	err = json.Unmarshal(res1.Data, &resp)
	if err != nil {
		return nil, err
	}
	var validatorsResp stakingtypes.QueryValidatorResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &validatorsResp)
	if err != nil {
		return nil, err
	}
	return validatorsResp.Validator, nil
}

// ValidatorByConsAddr gets the validator interface for a particular pubkey
func (k KeeperStaking) ValidatorByConsAddr(goCtx context.Context, addr sdk.ConsAddress) (stakingtypes.ValidatorI, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("ValidatorByConsAddr not implemented")
	return nil, nil
}

// GetAllDelegatorDelegations returns all delegations of a delegator
func (k KeeperStaking) GetAllDelegatorDelegations(goCtx context.Context, delegator sdk.AccAddress) ([]stakingtypes.Delegation, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("GetAllDelegatorDelegations not implemented")
	return make([]stakingtypes.Delegation, 0), nil
}

// GetAllSDKDelegations returns all delegations used during genesis dump
// TODO: remove this func, change all usage for iterate functionality [sdk comment]
func (k KeeperStaking) GetAllSDKDelegations(goCtx context.Context) (delegations []stakingtypes.Delegation, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	derc20Address, err := k.GetDERC20Address(ctx)
	if err != nil {
		return nil, err
	}
	k.WasmxKeeper.IterateContractState(ctx, derc20Address, func(key []byte, value []byte) bool {
		if !strings.HasPrefix(string(key), types.STAKING_DELEGATOR_TO_DELEGATION_KEY) {
			return false
		}
		delegator, validator, amount, err := types.ParseStoredDelegation(key, value)
		if err == nil {
			delegations = append(delegations, stakingtypes.Delegation{
				DelegatorAddress: delegator,
				ValidatorAddress: validator,
				// TODO amount will be big int
				Shares: math.LegacyNewDec(amount.Int64()),
			})
		}
		return false
	})
	return delegations, nil
}

// IterateDelegations iterates through all of the delegations from a delegator
func (k KeeperStaking) IterateDelegations(goCtx context.Context, delAddr sdk.AccAddress,
	fn func(index int64, del stakingtypes.DelegationI) (stop bool),
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("IterateDelegations not implemented")
	return nil
}

// GetHistoricalInfo gets the historical info at a given height
func (k KeeperStaking) GetHistoricalInfo(goCtx context.Context, height int64) (stakingtypes.HistoricalInfo, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("GetHistoricalInfo not implemented")
	return stakingtypes.HistoricalInfo{}, nil
}

// UnbondingTime - The time duration for unbonding
func (k KeeperStaking) UnbondingTime(goCtx context.Context) (time.Duration, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("UnbondingTime not implemented")
	return time.Minute, nil
}

// GetParams gets the x/staking module parameters.
func (k KeeperStaking) GetParams(goCtx context.Context) (params stakingtypes.Params, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("GetParams not implemented")
	return stakingtypes.Params{}, nil
}

// IterateBondedValidatorsByPower iterates through the bonded validator set and perform the provided function
func (k KeeperStaking) IterateBondedValidatorsByPower(goCtx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("IterateBondedValidatorsByPower not implemented")
	return nil
}

// TotalBondedTokens total staking tokens supply which is bonded
func (k KeeperStaking) TotalBondedTokens(goCtx context.Context) (math.Int, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("TotalBondedTokens not implemented")
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
func (k KeeperStaking) ApplyAndReturnValidatorSetUpdates(goCtx context.Context) (updates []abci.ValidatorUpdate, err error) {
	// TODO replace this mechanism
	validators, err := k.GetAllValidators(goCtx)
	if err != nil {
		return nil, err
	}
	_updates := make([]abci.ValidatorUpdate, len(validators))
	for i, valid := range validators {
		var pkI cryptotypes.PubKey
		err = k.InterfaceRegistry.UnpackAny(valid.ConsensusPubkey, &pkI)
		if err != nil {
			return nil, err
		}
		pk, ok := valid.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
		if !ok {
			return nil, errors.Wrapf(sdkerrors.ErrInvalidType, "expecting cryptotypes.PubKey, got %T", pk)
		}
		tmPk, err := cryptocodec.ToCmtProtoPublicKey(pk)
		if err != nil {
			return nil, err
		}
		upd := abci.ValidatorUpdate{
			PubKey: tmPk,
			Power:  valid.Tokens.BigInt().Int64(),
		}
		_updates[i] = upd
	}
	return _updates, nil
}

// GetAllDelegations returns all delegations used during genesis dump.
func (k KeeperStaking) GetAllDelegations(goCtx context.Context) (delegations []stakingtypes.Delegation, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("GetAllDelegations not implemented")
	return make([]stakingtypes.Delegation, 0), err
}

// IterateRedelegations iterates through all redelegations.
func (k KeeperStaking) IterateRedelegations(goCtx context.Context, fn func(index int64, red stakingtypes.Redelegation) (stop bool)) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("IterateRedelegations not implemented")
	return nil
}

// SetRedelegation sets a redelegation and associated index.
func (k KeeperStaking) SetRedelegation(goCtx context.Context, red stakingtypes.Redelegation) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("SetRedelegation not implemented")
	return nil
}

// IterateUnbondingDelegations iterates through all of the unbonding delegations.
func (k KeeperStaking) IterateUnbondingDelegations(goCtx context.Context, fn func(index int64, ubd stakingtypes.UnbondingDelegation) (stop bool)) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("IterateUnbondingDelegations not implemented")
	return nil
}

// SetUnbondingDelegation sets the unbonding delegation and associated index.
func (k KeeperStaking) SetUnbondingDelegation(goCtx context.Context, ubd stakingtypes.UnbondingDelegation) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("SetUnbondingDelegation not implemented")
	return nil
}

// GetValidator gets a single validator
func (k KeeperStaking) GetValidator(goCtx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("GetValidator not implemented")
	return stakingtypes.Validator{}, nil
}

// SetValidator sets the main record holding validator details
func (k KeeperStaking) SetValidator(goCtx context.Context, validator stakingtypes.Validator) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("SetValidator not implemented")
	return nil
}

// IterateLastValidators iterates through the active validator set and perform the provided function
func (k KeeperStaking) IterateLastValidators(goCtx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("IterateLastValidators not implemented")
	return nil
}

// SetValidatorByConsAddr sets a validator by conesensus address
func (k KeeperStaking) SetValidatorByConsAddr(goCtx context.Context, validator stakingtypes.Validator) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("SetValidatorByConsAddr not implemented")
	return nil
}

// PowerReduction - is the amount of staking tokens required for 1 unit of consensus-engine power.
// Currently, this returns a global variable that the app developer can tweak.
// TODO: we might turn this into an on-chain param:
// https://github.com/cosmos/cosmos-sdk/issues/8365
func (k KeeperStaking) PowerReduction(goCtx context.Context) math.Int {
	return sdk.DefaultPowerReduction
}

// WriteValidators returns a slice of bonded genesis validators.
func WriteValidators(ctx sdk.Context, keeper *KeeperStaking) (vals []cmttypes.GenesisValidator, returnErr error) {
	err := keeper.IterateLastValidators(ctx, func(_ int64, validator stakingtypes.ValidatorI) (stop bool) {
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

// GetAllSDKDelegations returns all delegations used during genesis dump
// TODO: remove this func, change all usage for iterate functionality [sdk comment]
func (k KeeperStaking) GetDERC20Address(ctx sdk.Context) (derc20Address sdk.AccAddress, err error) {
	// TODO asmyt denom
	msgbz := []byte(`{"GetAddressByDenom":{"denom":"asmyt"}}`)
	res1, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return nil, err
	}
	var resp1 wasmxtypes.ContractResponse
	err = json.Unmarshal(res1.Data, &resp1)
	if err != nil {
		return nil, err
	}
	var qaddrResp types.QueryAddressByDenomResponse
	err = json.Unmarshal(resp1.Data, &qaddrResp)
	if err != nil {
		return nil, err
	}
	return qaddrResp.Address, nil
}
