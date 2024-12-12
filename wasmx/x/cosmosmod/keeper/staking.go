package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	math "cosmossdk.io/math"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	mcfg "github.com/loredanacirstea/wasmx/config"
	"github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// BondedRatio the fraction of the staking tokens which are currently bonded
func (k KeeperStaking) BondedRatio(goCtx context.Context) (math.LegacyDec, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO
	k.Logger(ctx).Debug("KeeperStaking.BondedRatio not implemented")
	return math.LegacyOneDec(), nil
}

// StakingTokenSupply staking tokens from the total supply
func (k KeeperStaking) StakingTokenSupply(goCtx context.Context) (math.Int, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	config := k.actionExecutor.GetApp().GetChainCfg()
	derc20Address, err := k.GetDERC20Address(ctx, config.BondBaseDenom)
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

// Delegation gets the delegation interface for a particular set of delegator and validator addresses
func (k KeeperStaking) Delegation(goCtx context.Context, addrDel sdk.AccAddress, addrVal sdk.ValAddress) (stakingtypes.DelegationI, error) {
	validatorsResp, err := k.DelegationInternal(goCtx, k.accBech32Codec.BytesToAccAddressPrefixed(addrDel), k.valBech32Codec.BytesToValAddressPrefixed(addrVal))
	if err != nil {
		return nil, err
	}
	return validatorsResp.DelegationResponse.Delegation, nil
}

func (k KeeperStaking) DelegationInternal(goCtx context.Context, addrDel mcodec.AccAddressPrefixed, addrVal mcodec.ValAddressPrefixed) (*stakingtypes.QueryDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleQuery(ctx, "GetDelegation", &stakingtypes.QueryDelegationRequest{DelegatorAddr: addrDel.String(), ValidatorAddr: addrVal.String()})
	if err != nil {
		return nil, err
	}
	var validatorsResp stakingtypes.QueryDelegationResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &validatorsResp)
	if err != nil {
		return nil, err
	}
	return &validatorsResp, nil
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
	validator, err := k.ValidatorByConsAddr(goCtx, addr)
	if err != nil {
		return false, err
	}
	return validator.IsJailed(), nil
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
	k.Logger(ctx).Error("KeeperStaking.Jail not implemented")
	return nil
}

// MaxValidators - Maximum number of validators
func (k KeeperStaking) MaxValidators(goCtx context.Context) (uint32, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.MaxValidators not implemented")
	return 100, nil
}

// Infraction was committed at the current height or at a past height,
// but not at a height in the future
func (k KeeperStaking) Slash(goCtx context.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor math.LegacyDec) (math.Int, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.Slash not implemented")
	return math.ZeroInt(), nil
}

// SlashWithInfractionReason implementation doesn't require the infraction (types.Infraction) to work but is required by Interchain Security.
func (k KeeperStaking) SlashWithInfractionReason(goCtx context.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor math.LegacyDec, _ stakingtypes.Infraction) (math.Int, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.SlashWithInfractionReason not implemented")
	return k.Slash(goCtx, consAddr, infractionHeight, power, slashFactor)
}

// unjail a validator
func (k KeeperStaking) Unjail(goCtx context.Context, consAddr sdk.ConsAddress) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.Unjail not implemented")
	return nil
}

// Validator gets the Validator interface for a particular address
func (k KeeperStaking) Validator(goCtx context.Context, addrVal sdk.ValAddress) (stakingtypes.ValidatorI, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addrValStr, err := k.ValidatorAddressCodec().BytesToString(addrVal)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "validator: %s", mcfg.ERRORMSG_ACC_TOSTRING)
	}

	msgbz := []byte(fmt.Sprintf(`{"GetValidator":{"validator_addr":"%s"}}`, addrValStr))
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
func (k KeeperStaking) ValidatorByConsAddr(goCtx context.Context, addrCons sdk.ConsAddress) (stakingtypes.ValidatorI, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addrConsStr, err := k.ConsensusAddressCodec().BytesToString(addrCons)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "cons: %s", mcfg.ERRORMSG_ACC_TOSTRING)
	}

	msgbz := []byte(fmt.Sprintf(`{"ValidatorByConsAddr":{"validator_addr":"%s"}}`, addrConsStr))
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

// GetAllDelegatorDelegations returns all delegations of a delegator
func (k KeeperStaking) GetAllDelegatorDelegations(goCtx context.Context, delegator sdk.AccAddress) ([]stakingtypes.Delegation, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.GetAllDelegatorDelegations not implemented")
	return make([]stakingtypes.Delegation, 0), nil
}

// GetAllSDKDelegations returns all delegations used during genesis dump
// TODO: remove this func, change all usage for iterate functionality [sdk comment]
func (k KeeperStaking) GetAllSDKDelegations(goCtx context.Context) (delegations []stakingtypes.Delegation, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	config := k.actionExecutor.GetApp().GetChainCfg()
	derc20Address, err := k.GetDERC20Address(ctx, config.BondBaseDenom)
	if err != nil {
		return nil, err
	}
	k.WasmxKeeper.IterateContractState(ctx, derc20Address.Bytes(), func(key []byte, value []byte) bool {
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
	k.Logger(ctx).Error("KeeperStaking.IterateDelegations not implemented")
	return nil
}

// GetHistoricalInfo gets the historical info at a given height
func (k KeeperStaking) GetHistoricalInfo(goCtx context.Context, height int64) (stakingtypes.HistoricalInfo, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.GetHistoricalInfo not implemented")
	return stakingtypes.HistoricalInfo{}, nil
}

// UnbondingTime - The time duration for unbonding
func (k KeeperStaking) UnbondingTime(goCtx context.Context) (time.Duration, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.UnbondingTime not implemented")
	return time.Minute, nil
}

// GetParams gets the x/staking module parameters.
func (k KeeperStaking) GetParams(goCtx context.Context) (params stakingtypes.Params, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.ContractModuleQuery(ctx, "Params", &stakingtypes.QueryParamsRequest{})
	if err != nil {
		return stakingtypes.Params{}, err
	}
	var cresp stakingtypes.QueryParamsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return stakingtypes.Params{}, err
	}
	return cresp.Params, nil
}

// IterateBondedValidatorsByPower iterates through the bonded validator set and perform the provided function
func (k KeeperStaking) IterateBondedValidatorsByPower(goCtx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.IterateBondedValidatorsByPower not implemented")
	return nil
}

// TotalBondedTokens total staking tokens supply which is bonded
func (k KeeperStaking) TotalBondedTokens(goCtx context.Context) (math.Int, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.TotalBondedTokens not implemented")
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
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "expecting cryptotypes.PubKey, got %T", pk)
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
	k.Logger(ctx).Error("KeeperStaking.GetAllDelegations not implemented")
	return make([]stakingtypes.Delegation, 0), err
}

// IterateRedelegations iterates through all redelegations.
func (k KeeperStaking) IterateRedelegations(goCtx context.Context, fn func(index int64, red stakingtypes.Redelegation) (stop bool)) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.IterateRedelegations not implemented")
	return nil
}

// SetRedelegation sets a redelegation and associated index.
func (k KeeperStaking) SetRedelegation(goCtx context.Context, red stakingtypes.Redelegation) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.SetRedelegation not implemented")
	return nil
}

// IterateUnbondingDelegations iterates through all of the unbonding delegations.
func (k KeeperStaking) IterateUnbondingDelegations(goCtx context.Context, fn func(index int64, ubd stakingtypes.UnbondingDelegation) (stop bool)) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.IterateUnbondingDelegations not implemented")
	return nil
}

// SetUnbondingDelegation sets the unbonding delegation and associated index.
func (k KeeperStaking) SetUnbondingDelegation(goCtx context.Context, ubd stakingtypes.UnbondingDelegation) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.SetUnbondingDelegation not implemented")
	return nil
}

// GetValidator gets a single validator
func (k KeeperStaking) GetValidator(goCtx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.GetValidator not implemented")
	return stakingtypes.Validator{}, nil
}

// SetValidator sets the main record holding validator details
func (k KeeperStaking) SetValidator(goCtx context.Context, validator stakingtypes.Validator) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.SetValidator not implemented")
	return nil
}

// IterateLastValidators iterates through the active validator set and perform the provided function
func (k KeeperStaking) IterateLastValidators(goCtx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.IterateLastValidators not implemented")
	return nil
}

// SetValidatorByConsAddr sets a validator by conesensus address
func (k KeeperStaking) SetValidatorByConsAddr(goCtx context.Context, validator stakingtypes.Validator) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("KeeperStaking.SetValidatorByConsAddr not implemented")
	return nil
}

// PowerReduction - is the amount of staking tokens required for 1 unit of consensus-engine power.
// Currently, this returns a global variable that the app developer can tweak.
// TODO: we might turn this into an on-chain param:
// https://github.com/cosmos/cosmos-sdk/issues/8365
func (k KeeperStaking) PowerReduction(goCtx context.Context) math.Int {
	return sdk.DefaultPowerReduction
}

func (k KeeperStaking) Pool(goCtx context.Context, req *stakingtypes.QueryPoolRequest) (*stakingtypes.QueryPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	reqbz, err := k.JSONCodec().MarshalJSON(req)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetPool":%s}`, string(reqbz)))
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
	var internalResp stakingtypes.QueryPoolResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &internalResp)
	if err != nil {
		return nil, err
	}
	return &internalResp, nil
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
func (k KeeperStaking) GetDERC20Address(ctx sdk.Context, bondDenom string) (derc20Address mcodec.AccAddressPrefixed, err error) {
	// TODO asmyt denom
	msgbz := []byte(fmt.Sprintf(`{"GetAddressByDenom":{"denom":"%s"}}`, bondDenom))
	res1, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return mcodec.AccAddressPrefixed{}, err
	}
	var resp1 wasmxtypes.ContractResponse
	err = json.Unmarshal(res1.Data, &resp1)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, err
	}
	var qaddrResp types.QueryAddressByDenomResponse
	err = json.Unmarshal(resp1.Data, &qaddrResp)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, err
	}
	if len(qaddrResp.Address.Bytes()) == 0 {
		return mcodec.AccAddressPrefixed{}, fmt.Errorf("contract not found for bank denom: %s", bondDenom)
	}
	return qaddrResp.Address, nil
}

func (k KeeperStaking) ContractModuleQuery(ctx sdk.Context, fname string, req interface{}) (*wasmxtypes.ContractResponse, error) {
	msgbz, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	msgbz2 := []byte(fmt.Sprintf(`{"%s":%s}`, fname, string(msgbz)))
	res1, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_STAKING,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz2,
	})
	if err != nil {
		return nil, err
	}
	var resp wasmxtypes.ContractResponse
	err = json.Unmarshal(res1.Data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (k KeeperStaking) ContractModuleExecution(ctx sdk.Context, fname string, req interface{}) (*wasmxtypes.ContractResponse, error) {
	msgbz, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	msgbz2 := []byte(fmt.Sprintf(`{"%s":%s}`, fname, string(msgbz)))
	res1, err := k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_STAKING,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz2,
	})
	if err != nil {
		return nil, err
	}
	var resp wasmxtypes.ContractResponse
	err = json.Unmarshal(res1.Data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
