package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	// secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	verifysig "mythos/v1/crypto/verifysig"
	cosmosmod "mythos/v1/x/cosmosmod/types"
	"mythos/v1/x/wasmx/types"
)

func (k *Keeper) SetAccount(ctx sdk.Context, acc sdk.AccountI) error {
	modaddr, err := k.GetAddressOrRole(ctx, types.ROLE_AUTH)
	if err != nil {
		panic(err)
	}

	accbz, err := k.cdc.MarshalJSON(acc)
	if err != nil {
		panic(err) // TODO eventually catch this
	}
	msg := cosmosmod.MsgSetAccount{Account: &cosmosmod.AnyAccount{
		TypeUrl: sdk.MsgTypeURL(acc),
		Value:   accbz,
	}}

	msgbzz, err := k.cdc.MarshalJSON(&msg)
	if err != nil {
		return err
	}
	msgbz := []byte(fmt.Sprintf(`{"SetAccount":%s}`, string(msgbzz)))
	execmsg, err := json.Marshal(types.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		return err
	}
	_, err = k.execute(ctx, modaddr, modaddr, execmsg, nil, nil)
	return err
}

func (k *Keeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	modaddr, err := k.GetAddressOrRole(ctx, types.ROLE_AUTH)
	if err != nil {
		panic(err)
	}
	msgbz := []byte(fmt.Sprintf(`{"GetAccount":{"address":"%s"}}`, addr.String()))
	execmsg, err := json.Marshal(types.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		return nil
	}
	resp, err := k.query(ctx, modaddr, sdk.AccAddress([]byte(types.ModuleName)), execmsg, nil, nil, false)
	if err != nil {
		return nil
	}
	data := strings.ReplaceAll(string(resp.Data), `{"@type":"","key":""}`, "null")
	var response cosmosmod.QueryAccountResponse
	err = k.cdc.UnmarshalJSON([]byte(data), &response)
	if err != nil {
		return nil
	}
	if response.Account == nil {
		return nil
	}
	if response.Account.TypeUrl == sdk.MsgTypeURL(&authtypes.BaseAccount{}) {
		var acc authtypes.BaseAccount
		err = k.cdc.UnmarshalJSON(response.Account.Value, &acc)
		if err != nil {
			panic(err)
		}
		return &acc
	} else if response.Account.TypeUrl == sdk.MsgTypeURL(&authtypes.ModuleAccount{}) {
		var acc authtypes.ModuleAccount
		err = k.cdc.UnmarshalJSON(response.Account.Value, &acc)
		if err != nil {
			panic(err)
		}
		return &acc
	}
	return nil
}

func (k *Keeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	modaddr, err := k.GetAddressOrRole(ctx, types.ROLE_AUTH)
	if err != nil {
		panic(err)
	}

	acc := authtypes.BaseAccount{Address: addr.String()}
	accbz, err := k.cdc.MarshalJSON(&acc)
	if err != nil {
		panic(err) // TODO eventually catch this
	}
	msg := cosmosmod.MsgSetAccount{Account: &cosmosmod.AnyAccount{
		TypeUrl: sdk.MsgTypeURL(&acc),
		Value:   accbz,
	}}
	accmsgbz, err := k.cdc.MarshalJSON(&msg)

	msgbz := []byte(fmt.Sprintf(`{"SetAccount":%s}`, string(accmsgbz)))
	execmsg, err := json.Marshal(types.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		panic(err)
	}
	_, err = k.execute(ctx, modaddr, modaddr, execmsg, nil, nil)
	if err != nil {
		panic(err)
	}
	return k.GetAccount(ctx, addr)
}

type AccountKeeperVerifySig struct {
	k *Keeper
}

func NewAccountKeeperVerifySig(keeper *Keeper) verifysig.AccountKeeper {
	return AccountKeeperVerifySig{k: keeper}
}

func (ak AccountKeeperVerifySig) GetAccount(goCtx context.Context, addr sdk.AccAddress) sdk.AccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return ak.k.GetAccount(ctx, addr)
}
