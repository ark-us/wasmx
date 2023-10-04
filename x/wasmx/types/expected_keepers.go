package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
)

// AccountKeeper defines a subset of methods implemented by the cosmos-sdk account keeper
type AccountKeeper interface {
	// Return a new account with the next account number and the specified address. Does not save the new account to the store.
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	// Retrieve an account from the store.
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	// Set an account in the store.
	SetAccount(ctx context.Context, acc sdk.AccountI)
}

// BankKeeper defines a subset of methods implemented by the cosmos-sdk bank keeper
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetSupply(ctx context.Context, denom string) sdk.Coin
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	BlockedAddr(addr sdk.AccAddress) bool
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

// DistributionKeeper defines a subset of methods implemented by the cosmos-sdk distribution keeper
type DistributionKeeper interface {
	DelegationRewards(c context.Context, req *distrtypes.QueryDelegationRewardsRequest) (*distrtypes.QueryDelegationRewardsResponse, error)
}

// StakingKeeper defines a subset of methods implemented by the cosmos-sdk staking keeper
type StakingKeeper interface {
	// BondDenom - Bondable coin denomination
	BondDenom(ctx context.Context) (string, error)
	// GetValidator get a single validator
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
	// GetBondedValidatorsByPower get the current group of bonded validators sorted by power-rank
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
	// GetAllDelegatorDelegations return all delegations for a delegator
	GetAllDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.Delegation, error)
	// GetDelegation return a specific delegation
	GetDelegation(ctx context.Context,
		delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, err error)
	// HasReceivingRedelegation check if validator is receiving a redelegation
	HasReceivingRedelegation(ctx context.Context,
		delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) (bool, error)
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	// GetChannel(ctx context.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	ChanCloseInit(ctx sdk.Context, portID, channelID string, chanCap *capabilitytypes.Capability) error
	// GetAllChannels(ctx sdk.Context) (channels []channeltypes.IdentifiedChannel)
	// IterateChannels(ctx sdk.Context, cb func(channeltypes.IdentifiedChannel) bool)
	// SetChannel(ctx sdk.Context, portID, channelID string, channel channeltypes.Channel)
}
