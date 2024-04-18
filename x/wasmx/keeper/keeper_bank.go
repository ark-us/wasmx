package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"mythos/v1/x/wasmx/types"
)

// TransferCoins transfers coins from source to destination account when coin send was enabled for them and the recipient
// is not in the blocked address list.
func (k *Keeper) TransferCoins(parentCtx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount sdk.Coins) error {
	em := sdk.NewEventManager()
	ctx := parentCtx.WithEventManager(em)
	// TODO do we do blocked addresses here?
	// if err := k.IsSendEnabledCoins(ctx, amount...); err != nil {
	// 	return err
	// }
	// if k.BlockedAddr(toAddr) {
	// 	return sdkerr.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", toAddr.String())
	// }

	err := k.SendCoins(ctx, fromAddr, toAddr, amount)
	if err != nil {
		return err
	}
	for _, e := range em.Events() {
		if e.Type == sdk.EventTypeMessage { // skip messages as we talk to the keeper directly
			continue
		}
		parentCtx.EventManager().EmitEvent(e)
	}
	return nil
}

func (k *Keeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, coins sdk.Coins) error {
	aliasFrom, found := k.GetAlias(ctx, fromAddr)
	if found {
		fromAddr = aliasFrom
	}
	aliasTo, found := k.GetAlias(ctx, toAddr)
	if found {
		toAddr = aliasTo
	}
	bankAddress, err := k.GetAddressOrRole(ctx, types.ROLE_BANK)
	if err != nil {
		return err
	}
	msg := banktypes.NewMsgSend(fromAddr, toAddr, coins)
	bankmsgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return err
	}
	msgbz := []byte(fmt.Sprintf(`{"SendCoins":%s}`, string(bankmsgbz)))
	execmsg, err := json.Marshal(types.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		return err
	}
	_, err = k.execute(ctx, bankAddress, bankAddress, execmsg, nil, nil, false)
	return err
}

func (k *Keeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	aliasAddr, found := k.GetAlias(ctx, addr)
	if found {
		addr = aliasAddr
	}
	bankAddress, err := k.GetAddressOrRole(ctx, types.ROLE_BANK)
	if err != nil {
		return sdk.Coin{}, err
	}
	msg := banktypes.NewQueryBalanceRequest(addr, denom)
	bankmsgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return sdk.Coin{}, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetBalance":%s}`, string(bankmsgbz)))
	execmsg, err := json.Marshal(types.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		return sdk.Coin{}, err
	}
	resp, err := k.query(ctx, bankAddress, sdk.AccAddress([]byte(types.ModuleName)), execmsg, nil, nil, false)
	if err != nil {
		return sdk.Coin{}, err
	}
	var response banktypes.QueryBalanceResponse
	err = k.cdc.UnmarshalJSON(resp.Data, &response)
	if err != nil {
		return sdk.Coin{}, err
	}
	return *response.Balance, nil
}

func (k *Keeper) AllBalances(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	aliasAddr, found := k.GetAlias(ctx, addr)
	if found {
		addr = aliasAddr
	}
	bankAddress, err := k.GetAddressOrRole(ctx, types.ROLE_BANK)
	if err != nil {
		return nil, err
	}
	msg := banktypes.NewQueryAllBalancesRequest(addr, nil, false)
	msgbz, err := k.cdc.MarshalJSON(msg)
	if err != nil {
		return nil, err
	}
	execmsg, err := json.Marshal(types.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		return nil, err
	}
	resp, err := k.query(ctx, bankAddress, sdk.AccAddress([]byte(types.ModuleName)), execmsg, nil, nil, false)
	if err != nil {
		return nil, err
	}
	var contractResp types.ContractResponse
	err = json.Unmarshal(resp.Data, &contractResp)
	if err != nil {
		return nil, err
	}
	var response banktypes.QueryAllBalancesResponse
	err = k.cdc.UnmarshalJSON(contractResp.Data, &response)
	if err != nil {
		return nil, err
	}
	return response.Balances, nil
}
