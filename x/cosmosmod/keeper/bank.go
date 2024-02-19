package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// SendCoins transfers amt coins from a sending account to a receiving account.
// An error is returned upon failure.
func (k KeeperBank) SendCoins(goCtx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msg := banktypes.NewMsgSend(fromAddr, toAddr, amt)
	bankmsgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return err
	}
	msgbz := []byte(fmt.Sprintf(`{"SendCoins":%s}`, string(bankmsgbz)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return err
	}
	// TODO replace account keeper
	// Create account if recipient does not exist.
	//
	// NOTE: This should ultimately be removed in favor a more flexible approach
	// such as delegated fee messages.
	accExists := k.ak.HasAccount(ctx, toAddr)
	if !accExists {
		defer telemetry.IncrCounter(1, "new", "account")
		k.ak.SetAccount(ctx, k.ak.NewAccountWithAddress(ctx, toAddr))
	}

	// bech32 encoding is expensive! Only do it once for fromAddr
	fromAddrString := fromAddr.String()
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTransfer,
			sdk.NewAttribute(types.AttributeKeyRecipient, toAddr.String()),
			sdk.NewAttribute(types.AttributeKeySender, fromAddrString),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amt.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(types.AttributeKeySender, fromAddr.String()),
		),
	})
	return nil
}

// MintCoins creates new coins from thin air and adds it to the module account.
// It will panic if the module account does not exist or is unauthorized.
func (k KeeperBank) MintCoins(goCtx context.Context, moduleName string, amounts sdk.Coins) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Debug("MintCoins", moduleName, amounts)
	msg := &banktypes.Balance{
		Address: authtypes.NewModuleAddress(moduleName).String(),
		Coins:   amounts,
	}
	bankmsgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return err
	}
	msgbz := []byte(fmt.Sprintf(`{"MintCoins":%s}`, string(bankmsgbz)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return err
	}
	return nil
}

// SendCoinsFromModuleToAccount transfers coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if
// the recipient address is black-listed or if sending the tokens fails.
func (k KeeperBank) SendCoinsFromModuleToAccount(
	goCtx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msg := banktypes.NewMsgSend(authtypes.NewModuleAddress(senderModule), recipientAddr, amt)
	bankmsgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return err
	}
	msgbz := []byte(fmt.Sprintf(`{"SendCoins":%s}`, string(bankmsgbz)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return err
	}
	return nil
}

// SendCoinsFromModuleToModule transfers coins from a ModuleAccount to another.
// It will panic if either module account does not exist.
func (k KeeperBank) SendCoinsFromModuleToModule(
	goCtx context.Context, senderModule, recipientModule string, amt sdk.Coins,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msg := banktypes.NewMsgSend(authtypes.NewModuleAddress(senderModule), authtypes.NewModuleAddress(recipientModule), amt)
	bankmsgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return err
	}
	msgbz := []byte(fmt.Sprintf(`{"SendCoins":%s}`, string(bankmsgbz)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return err
	}
	return nil
}

// SendCoinsFromAccountToModule transfers coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k KeeperBank) SendCoinsFromAccountToModule(
	goCtx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msg := banktypes.NewMsgSend(senderAddr, authtypes.NewModuleAddress(recipientModule), amt)
	bankmsgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return err
	}
	msgbz := []byte(fmt.Sprintf(`{"SendCoins":%s}`, string(bankmsgbz)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return err
	}
	return nil
}

// BurnCoins burns coins deletes coins from the balance of the module account.
// It will panic if the module account does not exist or is unauthorized.
func (k KeeperBank) BurnCoins(goCtx context.Context, moduleName string, amounts sdk.Coins) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// authtypes.NewModuleAddress(moduleName).String()
	k.Logger(ctx).Error("BurnCoins not implemented")
	return nil
}

// SetDenomMetaData sets the denominations metadata
func (k KeeperBank) SetDenomMetaData(goCtx context.Context, denomMetaData banktypes.Metadata) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("SetDenomMetaData not implemented")
}

// HasDenomMetaData checks if the denomination metadata exists in store.
func (k KeeperBank) HasDenomMetaData(goCtx context.Context, denom string) bool {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("HasDenomMetaData not implemented")
	return true
}

// IsSendEnabledCoin returns the current SendEnabled status of the provided coin's denom
func (k KeeperBank) IsSendEnabledCoin(goCtx context.Context, coin sdk.Coin) bool {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("IsSendEnabledCoin not implemented")
	return true
}

// IsSendEnabledCoins checks the coins provided and returns an ErrSendDisabled
// if any of the coins are not configured for sending. Returns nil if sending is
// enabled for all provided coins.
func (k KeeperBank) IsSendEnabledCoins(goCtx context.Context, coins ...sdk.Coin) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("IsSendEnabledCoins not implemented")
	return nil
}

// LockedCoins returns all the coins that are not spendable (i.e. locked) for an
// account by address. For standard accounts, the result will always be no coins.
// For vesting accounts, LockedCoins is delegated to the concrete vesting account
// type.
func (k KeeperBank) LockedCoins(goCtx context.Context, addr sdk.AccAddress) sdk.Coins {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("LockedCoins not implemented")
	return nil
}

// BlockedAddr checks if a given address is restricted from
// receiving funds.
func (k KeeperBank) BlockedAddr(addr sdk.AccAddress) bool {
	fmt.Println("BlockedAddr not implemented")
	// return k.blockedAddrs[addr.String()]
	// TODO
	return false
}

// SpendableCoins returns the total balances of spendable coins for an account
// by address. If the account has no spendable coins, an empty Coins slice is
// returned.
func (k KeeperBank) SpendableCoins(goCtx context.Context, addr sdk.AccAddress) sdk.Coins {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("SpendableCoins not implemented")
	return nil
}

// GetAllBalances returns all the account balances for the given account address.
func (k KeeperBank) GetAllBalances(goCtx context.Context, addr sdk.AccAddress) sdk.Coins {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msg := banktypes.NewQueryAllBalancesRequest(addr, nil, false)
	bankmsgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return nil
	}
	msgbz := []byte(fmt.Sprintf(`{"GetAllBalances":%s}`, string(bankmsgbz)))
	resp, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return nil
	}
	var contractResp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp.Data, &contractResp)
	if err != nil {
		return nil
	}

	var response banktypes.QueryAllBalancesResponse
	err = k.cdc.UnmarshalJSON(contractResp.Data, &response)
	if err != nil {
		return nil
	}

	return response.Balances
}

// GetBalance returns the balance of a specific denomination for a given account
// by address.
func (k KeeperBank) GetBalance(goCtx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	ctx := sdk.UnwrapSDKContext(goCtx)
	msg := banktypes.NewQueryBalanceRequest(addr, denom)
	bankmsgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return sdk.Coin{}
	}
	msgbz := []byte(fmt.Sprintf(`{"GetBalance":%s}`, string(bankmsgbz)))
	resp, err := k.NetworkKeeper.QueryContract(ctx, &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		return sdk.Coin{}
	}
	var contractResp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp.Data, &contractResp)
	if err != nil {
		return sdk.Coin{}
	}

	var response banktypes.QueryBalanceResponse
	err = k.cdc.UnmarshalJSON(contractResp.Data, &response)
	if err != nil {
		return sdk.Coin{}
	}

	return *response.Balance
}

// // GetSupply retrieves the Supply from store
// func (k KeeperBank) GetSupply(goCtx context.Context, denom string) sdk.Coin {
// 	return sdk.NewCoin(denom, math.NewInt(0))
// }
