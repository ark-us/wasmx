package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"mythos/v1/x/cosmosmod/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

func (k KeeperAuth) SetParams(ctx sdk.Context, params authtypes.Params) error {
	k.Logger(ctx).Error("Auth.Params not implemented")
	return nil
}

func (k KeeperAuth) GetParams(goCtx context.Context) (params authtypes.Params) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	modaddr, err := k.WasmxKeeper.GetAddressOrRole(ctx, ctx.ChainID(), wasmxtypes.ROLE_AUTH)
	if err != nil {
		panic(err) // TODO catch this
	}
	msgbz := []byte(`{"GetParams":{}}`)
	execmsg, err := json.Marshal(wasmxtypes.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		panic(err)
	}
	resp, err := k.WasmxKeeper.Query(ctx, modaddr, sdk.AccAddress([]byte(types.ModuleName)), execmsg, nil, nil)
	if err != nil {
		panic(err)
	}
	var rresp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp, &rresp)
	if err != nil {
		panic(err)
	}
	var response authtypes.QueryParamsResponse
	err = k.cdc.UnmarshalJSON(rresp.Data, &response)
	if err != nil {
		panic(err)
	}
	return response.Params
}

func (k KeeperAuth) AddressCodec() address.Codec {
	return k.addressCodec
}

func (k KeeperAuth) HasAccount(goCtx context.Context, addr sdk.AccAddress) bool {
	ctx := sdk.UnwrapSDKContext(goCtx)
	modaddr, err := k.WasmxKeeper.GetAddressOrRole(ctx, ctx.ChainID(), wasmxtypes.ROLE_AUTH)
	if err != nil {
		panic(err) // TODO catch this
	}
	msgbz := []byte(fmt.Sprintf(`{"HasAccount":{"address":"%s"}}`, addr.String()))
	execmsg, err := json.Marshal(wasmxtypes.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		panic(err)
	}
	resp, err := k.WasmxKeeper.Query(ctx, modaddr, sdk.AccAddress([]byte(types.ModuleName)), execmsg, nil, nil)
	if err != nil {
		panic(err)
	}
	var rresp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp, &rresp)
	if err != nil {
		panic(err)
	}

	var response types.QueryHasAccountResponse
	err = k.cdc.UnmarshalJSON(rresp.Data, &response)
	if err != nil {
		panic(err)
	}
	return response.Found
}

func (k KeeperAuth) GetAccount(goCtx context.Context, addr sdk.AccAddress) sdk.AccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	modaddr, err := k.WasmxKeeper.GetAddressOrRole(ctx, ctx.ChainID(), wasmxtypes.ROLE_AUTH)
	if err != nil {
		panic(err) // TODO catch this
	}
	msgbz := []byte(fmt.Sprintf(`{"GetAccount":{"address":"%s"}}`, addr.String()))
	execmsg, err := json.Marshal(wasmxtypes.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		panic(err)
	}
	resp, err := k.WasmxKeeper.Query(ctx, modaddr, sdk.AccAddress([]byte(types.ModuleName)), execmsg, nil, nil)
	if err != nil {
		panic(err)
	}
	var rresp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp, &rresp)
	if err != nil {
		panic(err)
	}

	data := strings.ReplaceAll(string(rresp.Data), `{"@type":"","key":""}`, "null")
	var response types.QueryAccountResponse
	err = k.cdc.UnmarshalJSON([]byte(data), &response)
	if err != nil {
		panic(err)
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

func (k KeeperAuth) SetAccount(goCtx context.Context, acc sdk.AccountI) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	modaddr, err := k.WasmxKeeper.GetAddressOrRole(ctx, ctx.ChainID(), wasmxtypes.ROLE_AUTH)
	if err != nil {
		panic(err) // TODO eventually catch this
	}

	accbz, err := k.cdc.MarshalJSON(acc)
	if err != nil {
		panic(err) // TODO eventually catch this
	}
	msg := types.MsgSetAccount{Account: &types.AnyAccount{
		TypeUrl: sdk.MsgTypeURL(acc),
		Value:   accbz,
	}}
	accmsgbz, err := k.cdc.MarshalJSON(&msg)
	if err != nil {
		panic(err) // TODO eventually catch this
	}
	data := strings.ReplaceAll(string(accmsgbz), `"pub_key":null`, `"pub_key":{"@type":"","key":""}`)
	msgbz := []byte(fmt.Sprintf(`{"SetAccount":%s}`, data))
	execmsg, err := json.Marshal(wasmxtypes.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		panic(err) // TODO eventually catch this
	}
	_, err = k.WasmxKeeper.Execute(ctx, modaddr, modaddr, execmsg, nil, nil, false)
	if err != nil {
		panic(err) // TODO eventually catch this
	}
}

func (k KeeperAuth) NewAccountWithAddress(goCtx context.Context, addr sdk.AccAddress) sdk.AccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	modaddr, err := k.WasmxKeeper.GetAddressOrRole(ctx, ctx.ChainID(), wasmxtypes.ROLE_AUTH)
	if err != nil {
		panic(err)
	}

	acc := authtypes.BaseAccount{Address: addr.String()}
	accbz, err := k.cdc.MarshalJSON(&acc)
	if err != nil {
		panic(err) // TODO eventually catch this
	}
	msg := types.MsgSetAccount{Account: &types.AnyAccount{
		TypeUrl: sdk.MsgTypeURL(&acc),
		Value:   accbz,
	}}
	bankmsgbz, err := k.cdc.MarshalJSON(&msg)
	msgbz := []byte(fmt.Sprintf(`{"SetAccount":%s}`, string(bankmsgbz)))

	execmsg, err := json.Marshal(wasmxtypes.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		panic(err)
	}
	_, err = k.WasmxKeeper.Execute(ctx, modaddr, modaddr, execmsg, nil, nil, false)
	if err != nil {
		panic(err)
	}
	return k.GetAccount(ctx, addr)
}

func (k KeeperAuth) GetModuleAddress(moduleName string) sdk.AccAddress {
	permAddr, ok := k.permAddrs[moduleName]
	if !ok {
		return nil
	}
	return permAddr.GetAddress()
}

func (k KeeperAuth) GetModuleAccount(goCtx context.Context, moduleName string) sdk.ModuleAccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	acc, _ := k.GetModuleAccountAndPermissions(ctx, moduleName)
	return acc
}

func (k KeeperAuth) SetModuleAccount(goCtx context.Context, macc sdk.ModuleAccountI) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetAccount(ctx, macc)
}

func (k KeeperAuth) NewAccount(goCtx context.Context, acc sdk.AccountI) sdk.AccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetAccount(goCtx, acc)
	return k.GetAccount(ctx, acc.GetAddress())
}

func (k KeeperAuth) RemoveAccount(goCtx context.Context, acc sdk.AccountI) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.RemoveAccount not implemented")
}

func (k KeeperAuth) IterateAccounts(goCtx context.Context, cb func(account sdk.AccountI) (stop bool)) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Error("Auth.IterateAccounts not implemented")
}

func (k KeeperAuth) GetSequence(goCtx context.Context, addr sdk.AccAddress) (uint64, error) {
	acc := k.GetAccount(goCtx, addr)
	if acc == nil {
		return 0, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}
	return acc.GetSequence(), nil
}

// GetModuleAddressAndPermissions returns an address and permissions based on the module name
func (k KeeperAuth) GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string) {
	permAddr, ok := k.permAddrs[moduleName]
	if !ok {
		return addr, permissions
	}

	return permAddr.GetAddress(), permAddr.GetPermissions()
}

// GetModuleAccountAndPermissions gets the module account from the auth account store and its
// registered permissions
func (k KeeperAuth) GetModuleAccountAndPermissions(ctx context.Context, moduleName string) (sdk.ModuleAccountI, []string) {
	addr, perms := k.GetModuleAddressAndPermissions(moduleName)
	if addr == nil {
		return nil, []string{}
	}

	acc := k.GetAccount(ctx, addr)
	if acc != nil {
		macc, ok := acc.(sdk.ModuleAccountI)
		if !ok {
			panic("account is not a module account")
		}
		return macc, perms
	}

	// create a new module account
	macc := authtypes.NewEmptyModuleAccount(moduleName, perms...)
	maccI := (k.NewAccount(ctx, macc)).(sdk.ModuleAccountI) // set the account number
	k.SetModuleAccount(ctx, maccI)

	return maccI, perms
}

// func (authkeeper.AccountKeeper).AddressCodec() address.Codec
// func (authkeeper.AccountKeeper).ExportGenesis(ctx sdk.Context) *authtypes.GenesisState
// func (authkeeper.AccountKeeper).GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
// func (authkeeper.AccountKeeper).GetAllAccounts(ctx context.Context) (accounts []sdk.AccountI)
// func (authkeeper.AccountKeeper).GetAuthority() string
// func (authkeeper.AccountKeeper).GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
// func (authkeeper.AccountKeeper).GetModuleAccountAndPermissions(ctx context.Context, moduleName string) (sdk.ModuleAccountI, []string)
// func (authkeeper.AccountKeeper).GetModuleAddress(moduleName string) sdk.AccAddress
// func (authkeeper.AccountKeeper).GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string)
// func (authkeeper.AccountKeeper).GetModulePermissions() map[string]authtypes.PermissionsForAddress
// func (authkeeper.AccountKeeper).GetParams(ctx context.Context) (params authtypes.Params)
// func (authkeeper.AccountKeeper).GetPubKey(ctx context.Context, addr sdk.AccAddress) (types.PubKey, error)
// func (authkeeper.AccountKeeper).GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error)
// func (authkeeper.AccountKeeper).HasAccount(ctx context.Context, addr sdk.AccAddress) bool
// func (authkeeper.AccountKeeper).InitGenesis(ctx sdk.Context, data authtypes.GenesisState)
// func (authkeeper.AccountKeeper).IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) (stop bool))
// func (authkeeper.AccountKeeper).Logger(ctx context.Context) log.Logger
// func (authkeeper.AccountKeeper).NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
// func (authkeeper.AccountKeeper).NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
// func (authkeeper.AccountKeeper).NextAccountNumber(ctx context.Context) uint64
// func (authkeeper.AccountKeeper).RemoveAccount(ctx context.Context, acc sdk.AccountI)
// func (authkeeper.AccountKeeper).SetAccount(ctx context.Context, acc sdk.AccountI)
// func (authkeeper.AccountKeeper).SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI)
// func (authkeeper.AccountKeeper).ValidatePermissions(macc sdk.ModuleAccountI) error
