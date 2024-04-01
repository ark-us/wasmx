package keeper

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AddPubkey sets a address-pubkey relation
func (k KeeperSlashing) AddPubkey(ctx context.Context, pubkey cryptotypes.PubKey) error {
	return nil
}

// GetPubkey returns the pubkey from the adddress-pubkey relation
func (k KeeperSlashing) GetPubkey(ctx context.Context, a cryptotypes.Address) (cryptotypes.PubKey, error) {
	return nil, nil
}

func (k KeeperSlashing) Unjail(ctx sdk.Context) error {
	// TODO
	k.Logger(ctx).Debug("KeeperSlashing.Unjail not implemented")
	return nil
}

func (k KeeperSlashing) UpdateParams(ctx sdk.Context, params slashingtypes.Params) error {
	// TODO
	k.Logger(ctx).Debug("KeeperSlashing.UpdateParams not implemented")
	return nil
}

func (k KeeperSlashing) Params(ctx sdk.Context) (*slashingtypes.Params, error) {
	// TODO
	k.Logger(ctx).Debug("KeeperSlashing.Params not implemented")
	return nil, nil
}

func (k KeeperSlashing) SigningInfo(ctx sdk.Context, consAddress sdk.ConsAddress) (*slashingtypes.ValidatorSigningInfo, error) {
	// TODO
	k.Logger(ctx).Debug("KeeperSlashing.SigningInfo not implemented")
	return nil, nil
}

func (k KeeperSlashing) SigningInfos(ctx sdk.Context, req *slashingtypes.QuerySigningInfosRequest) (*slashingtypes.QuerySigningInfosResponse, error) {
	// TODO
	k.Logger(ctx).Debug("KeeperSlashing.SigningInfo not implemented")
	return nil, nil
}

// HasValidatorSigningInfo returns if a given validator has signing information
// persisted.
func (k KeeperSlashing) HasValidatorSigningInfo(ctx context.Context, consAddr sdk.ConsAddress) bool {
	return true
}

// IsTombstoned returns if a given validator by consensus address is tombstoned.
func (k KeeperSlashing) IsTombstoned(ctx context.Context, consAddr sdk.ConsAddress) bool {
	return false
}

// Jail attempts to jail a validator. The slash is delegated to the staking module
// to make the necessary validator changes.
func (k KeeperSlashing) Jail(ctx context.Context, consAddr sdk.ConsAddress) error {
	return nil
}

// JailUntil attempts to set a validator's JailedUntil attribute in its signing
// info. It will panic if the signing info does not exist for the validator.
func (k KeeperSlashing) JailUntil(ctx context.Context, consAddr sdk.ConsAddress, jailTime time.Time) error {
	return nil
}

// Slash attempts to slash a validator. The slash is delegated to the staking
// module to make the necessary validator changes. It specifies no intraction reason.
func (k KeeperSlashing) Slash(ctx context.Context, consAddr sdk.ConsAddress, fraction sdkmath.LegacyDec, power, distributionHeight int64) error {
	return nil
}

// SlashFractionDoubleSign - fraction of power slashed in case of double sign
func (k KeeperSlashing) SlashFractionDoubleSign(ctx context.Context) (sdkmath.LegacyDec, error) {
	return sdkmath.LegacyDec{}, nil
}

// SlashWithInfractionReason attempts to slash a validator. The slash is delegated to the staking
// module to make the necessary validator changes. It specifies an intraction reason.
func (k KeeperSlashing) SlashWithInfractionReason(ctx context.Context, consAddr sdk.ConsAddress, fraction sdkmath.LegacyDec, power, distributionHeight int64, infraction stakingtypes.Infraction) error {
	return nil
}

// Tombstone attempts to tombstone a validator. It will panic if signing info for
// the given validator does not exist.
func (k KeeperSlashing) Tombstone(ctx context.Context, consAddr sdk.ConsAddress) error {
	return nil
}

// IterateValidatorSigningInfos iterates over the stored ValidatorSigningInfo
func (k KeeperSlashing) IterateValidatorSigningInfos(ctx context.Context,
	handler func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool),
) error {
	return nil
}

// SetValidatorSigningInfo sets the validator signing info to a consensus address key
func (k KeeperSlashing) SetValidatorSigningInfo(ctx context.Context, address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) error {
	return nil
}
