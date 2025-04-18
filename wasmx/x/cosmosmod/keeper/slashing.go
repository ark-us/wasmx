package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// AddPubkey sets a address-pubkey relation
func (k KeeperSlashing) AddPubkey(goCtx context.Context, pubkey cryptotypes.PubKey) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperSlashing.AddPubkey not implemented")
	return nil
}

// GetPubkey returns the pubkey from the adddress-pubkey relation
func (k KeeperSlashing) GetPubkey(goCtx context.Context, a cryptotypes.Address) (cryptotypes.PubKey, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperSlashing.GetPubkey not implemented")
	return nil, nil
}

func (k KeeperSlashing) Unjail(goCtx sdk.Context, msg *slashingtypes.MsgUnjail) (*slashingtypes.MsgUnjailResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_, err := k.ContractModuleExecution(ctx, "Unjail", msg)
	if err != nil {
		return nil, err
	}
	return &slashingtypes.MsgUnjailResponse{}, nil
}

func (k KeeperSlashing) UpdateParams(ctx sdk.Context, params slashingtypes.Params) error {
	// TODO
	k.Logger(ctx).Debug("KeeperSlashing.UpdateParams not implemented")
	return nil
}

func (k KeeperSlashing) Params(ctx sdk.Context) (*slashingtypes.Params, error) {
	resp, err := k.ContractModuleQuery(ctx, "Params", &slashingtypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	var cresp slashingtypes.QueryParamsResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp.Params, nil
}

func (k KeeperSlashing) SigningInfo(ctx sdk.Context, req *slashingtypes.QuerySigningInfoRequest) (*slashingtypes.ValidatorSigningInfo, error) {
	resp, err := k.ContractModuleQuery(ctx, "SigningInfo", req)
	if err != nil {
		return nil, err
	}
	var cresp slashingtypes.QuerySigningInfoResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp.ValSigningInfo, nil
}

func (k KeeperSlashing) SigningInfos(ctx sdk.Context, req *slashingtypes.QuerySigningInfosRequest) (*slashingtypes.QuerySigningInfosResponse, error) {
	resp, err := k.ContractModuleQuery(ctx, "SigningInfos", req)
	if err != nil {
		return nil, err
	}
	var cresp slashingtypes.QuerySigningInfosResponse
	err = k.JSONCodec().UnmarshalJSON(resp.Data, &cresp)
	if err != nil {
		return nil, err
	}
	return &cresp, nil
}

// HasValidatorSigningInfo returns if a given validator has signing information
// persisted.
func (k KeeperSlashing) HasValidatorSigningInfo(goCtx context.Context, consAddr sdk.ConsAddress) bool {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperSlashing.HasValidatorSigningInfo not implemented")
	return true
}

// IsTombstoned returns if a given validator by consensus address is tombstoned.
func (k KeeperSlashing) IsTombstoned(goCtx context.Context, consAddr sdk.ConsAddress) bool {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperSlashing.IsTombstoned not implemented")
	return false
}

// Jail attempts to jail a validator. The slash is delegated to the staking module
// to make the necessary validator changes.
func (k KeeperSlashing) Jail(goCtx context.Context, consAddr sdk.ConsAddress) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.sk.Jail(ctx, consAddr)
}

// JailUntil attempts to set a validator's JailedUntil attribute in its signing
// info. It will panic if the signing info does not exist for the validator.
func (k KeeperSlashing) JailUntil(goCtx context.Context, consAddr sdk.ConsAddress, jailTime time.Time) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperSlashing.JailUntil not implemented")
	return nil
}

// Slash attempts to slash a validator. The slash is delegated to the staking
// module to make the necessary validator changes. It specifies no intraction reason.
func (k KeeperSlashing) Slash(goCtx context.Context, consAddr sdk.ConsAddress, fraction sdkmath.LegacyDec, power, distributionHeight int64) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.SlashWithInfractionReason(ctx, consAddr, fraction, power, distributionHeight, stakingtypes.Infraction_INFRACTION_UNSPECIFIED)
}

// SlashFractionDoubleSign - fraction of power slashed in case of double sign
func (k KeeperSlashing) SlashFractionDoubleSign(goCtx context.Context) (sdkmath.LegacyDec, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperSlashing.SlashFractionDoubleSign not implemented")
	return sdkmath.LegacyDec{}, nil
}

// SlashWithInfractionReason attempts to slash a validator. The slash is delegated to the staking
// module to make the necessary validator changes. It specifies an intraction reason.
func (k KeeperSlashing) SlashWithInfractionReason(goCtx context.Context, consAddr sdk.ConsAddress, fraction sdkmath.LegacyDec, power, distributionHeight int64, infraction stakingtypes.Infraction) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_, err := k.sk.SlashWithInfractionReason(ctx, consAddr, distributionHeight, power, fraction, infraction)
	if err != nil {
		return err
	}
	return nil
}

// Tombstone attempts to tombstone a validator. It will panic if signing info for
// the given validator does not exist.
func (k KeeperSlashing) Tombstone(goCtx context.Context, consAddr sdk.ConsAddress) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperSlashing.Tombstone not implemented")
	return nil
}

// IterateValidatorSigningInfos iterates over the stored ValidatorSigningInfo
func (k KeeperSlashing) IterateValidatorSigningInfos(goCtx context.Context,
	handler func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool),
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperSlashing.IterateValidatorSigningInfos not implemented")
	return nil
}

// SetValidatorSigningInfo sets the validator signing info to a consensus address key
func (k KeeperSlashing) SetValidatorSigningInfo(goCtx context.Context, address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("KeeperSlashing.SetValidatorSigningInfo not implemented")
	return nil
}

func (k KeeperSlashing) ContractModuleQuery(ctx sdk.Context, fname string, req interface{}) (*wasmxtypes.ContractResponse, error) {
	msgbz, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	msgbz2 := []byte(fmt.Sprintf(`{"%s":%s}`, fname, string(msgbz)))
	res1, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_SLASHING,
		Contract: wasmxtypes.ROLE_SLASHING,
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

func (k KeeperSlashing) ContractModuleExecution(ctx sdk.Context, fname string, req interface{}) (*wasmxtypes.ContractResponse, error) {
	msgbz, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	msgbz2 := []byte(fmt.Sprintf(`{"%s":%s}`, fname, string(msgbz)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_SLASHING,
		Contract: wasmxtypes.ROLE_SLASHING,
		Msg:      msgbz2,
	})
	if err != nil {
		return nil, err
	}
	// var resp wasmxtypes.ContractResponse
	// err = json.Unmarshal(res1.Data, &resp)
	// if err != nil {
	// 	return nil, err
	// }
	// return &resp, nil
	return &wasmxtypes.ContractResponse{}, nil
}
