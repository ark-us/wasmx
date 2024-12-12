package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	mcfg "github.com/loredanacirstea/wasmx/config"
	"github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (k KeeperAuth) SetParams(ctx sdk.Context, params authtypes.Params) error {
	k.Logger(ctx).Error("Auth.Params not implemented")
	return nil
}

func (k KeeperAuth) GetParams(goCtx context.Context) (params authtypes.Params) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	modaddr, err := k.WasmxKeeper.GetAddressOrRole(ctx, wasmxtypes.ROLE_AUTH)
	if err != nil {
		panic(err) // TODO catch this
	}
	msgbz := []byte(`{"GetParams":{}}`)
	execmsg, err := json.Marshal(wasmxtypes.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		panic(err)
	}
	resp, err := k.WasmxKeeper.Query(ctx, modaddr, k.accBech32Codec.BytesToAccAddressPrefixed([]byte(types.ModuleName)), execmsg, nil, nil)
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

func (k KeeperAuth) HasAccount(goCtx context.Context, addr sdk.AccAddress) bool {
	ctx := sdk.UnwrapSDKContext(goCtx)
	modaddr, err := k.WasmxKeeper.GetAddressOrRole(ctx, wasmxtypes.ROLE_AUTH)
	if err != nil {
		panic(err) // TODO catch this
	}
	addrstr, err := k.AddressCodec().BytesToString(addr)
	if err != nil {
		panic(fmt.Errorf("address: %s", mcfg.ERRORMSG_ACC_TOSTRING))
	}
	msgbz := []byte(fmt.Sprintf(`{"HasAccount":{"address":"%s"}}`, addrstr))
	execmsg, err := json.Marshal(wasmxtypes.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		panic(err)
	}
	resp, err := k.WasmxKeeper.Query(ctx, modaddr, k.accBech32Codec.BytesToAccAddressPrefixed([]byte(types.ModuleName)), execmsg, nil, nil)
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
	addrp := k.accBech32Codec.BytesToAccAddressPrefixed(addr)
	acc, err := k.GetAccountPrefixed(goCtx, addrp)
	if err != nil {
		panic(err)
	}
	if acc == nil {
		return nil
	}
	return acc.ToCosmosAccountI()
}

func (k KeeperAuth) GetAccountPrefixed(goCtx context.Context, addr mcodec.AccAddressPrefixed) (mcodec.AccountI, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	modaddr, err := k.WasmxKeeper.GetAddressOrRole(ctx, wasmxtypes.ROLE_AUTH)
	if err != nil {
		return nil, err
	}
	msgbz := []byte(fmt.Sprintf(`{"GetAccount":{"address":"%s"}}`, addr.String()))
	execmsg, err := json.Marshal(wasmxtypes.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		return nil, err
	}
	resp, err := k.WasmxKeeper.Query(ctx, modaddr, k.accBech32Codec.BytesToAccAddressPrefixed([]byte(types.ModuleName)), execmsg, nil, nil)
	if err != nil {
		return nil, err
	}
	var rresp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp, &rresp)
	if err != nil {
		return nil, err
	}
	var response types.QueryAccountResponse
	err = k.cdc.UnmarshalJSON(rresp.Data, &response)
	if err != nil {
		return nil, err
	}
	if response.Account == nil {
		return nil, nil
	}

	sdkmsg, err := mcodec.AnyToSdkMsg(k.cdc, response.Account)
	if err != nil {
		return nil, err
	}
	return sdkmsg.(mcodec.AccountI), nil
}

// TODO eventually remove after we replace all cosmos modules
func (k KeeperAuth) SetAccount(goCtx context.Context, authacc sdk.AccountI) {
	// TODO we assume all are BaseAccounts or ModuleAccounts
	// cosmos.sdk also has BaseVestingAccount, but we do not support this yet
	acc := types.NewBaseAccount(
		k.accBech32Codec.BytesToAccAddressPrefixed(authacc.GetAddress()),
		authacc.GetPubKey(),
		authacc.GetAccountNumber(),
		authacc.GetSequence(),
	)
	modacc, ok := authacc.(sdk.ModuleAccountI)
	if !ok {
		// store BaseAccount account
		err := k.SetAccountPrefixed(goCtx, acc)
		if err != nil {
			panic(err)
		}
		return
	}
	// store ModuleAccount
	newmodacc := types.NewModuleAccount(
		acc,
		modacc.GetName(),
		modacc.GetPermissions()...,
	)
	err := k.SetAccountPrefixed(goCtx, newmodacc)
	if err != nil {
		panic(err)
	}
}

func (k KeeperAuth) SetAccountPrefixed(goCtx context.Context, acc mcodec.AccountI) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	modaddr, err := k.WasmxKeeper.GetAddressOrRole(ctx, wasmxtypes.ROLE_AUTH)
	if err != nil {
		return err
	}
	accany, err := cdctypes.NewAnyWithValue(acc)
	if err != nil {
		return err
	}
	msg := types.MsgSetAccount{Account: accany}
	accmsgbz, err := k.cdc.MarshalJSON(&msg)
	if err != nil {
		return err
	}
	msgbz := []byte(fmt.Sprintf(`{"SetAccount":%s}`, accmsgbz))
	execmsg, err := json.Marshal(wasmxtypes.WasmxExecutionMessage{Data: msgbz})
	if err != nil {
		return err
	}
	_, err = k.WasmxKeeper.Execute(ctx, modaddr, modaddr, execmsg, nil, nil, false)
	if err != nil {
		return err
	}
	return nil
}

func (k KeeperAuth) NewAccountWithAddress(goCtx context.Context, addr sdk.AccAddress) sdk.AccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addrstr, err := k.AddressCodec().BytesToString(addr)
	if err != nil {
		panic(fmt.Errorf("address: %s", mcfg.ERRORMSG_ACC_TOSTRING))
	}
	acc := &types.BaseAccount{Address: addrstr}
	err = k.SetAccountPrefixed(goCtx, acc)
	if err != nil {
		panic(err) // TODO eventually catch this
	}
	return k.GetAccount(ctx, addr)
}

func (k KeeperAuth) NewAccountWithAddressPrefixed(goCtx context.Context, addr mcodec.AccAddressPrefixed) (mcodec.AccountI, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	acc := &types.BaseAccount{Address: addr.String()}
	err := k.SetAccountPrefixed(goCtx, acc)
	if err != nil {
		return nil, err
	}
	return k.GetAccountPrefixed(ctx, addr)
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

func (k KeeperAuth) SetModuleAccountPrefixed(goCtx context.Context, macc mcodec.ModuleAccountI) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetAccountPrefixed(ctx, macc)
}

func (k KeeperAuth) NewAccount(goCtx context.Context, acc sdk.AccountI) sdk.AccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetAccount(goCtx, acc)
	return k.GetAccount(ctx, acc.GetAddress())
}

func (k KeeperAuth) NewAccountPrefixed(goCtx context.Context, acc mcodec.AccountI) mcodec.AccountI {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetAccountPrefixed(goCtx, acc)
	acc, err := k.GetAccountPrefixed(ctx, acc.GetAddressPrefixed())
	if err != nil {
		panic(fmt.Errorf("NewAccountPrefixed: %s", err.Error()))
	}
	return acc
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
	macc := types.NewEmptyModuleAccount(k.accBech32Codec, moduleName, perms...)
	maccIP := (k.NewAccountPrefixed(ctx, macc)).(mcodec.ModuleAccountI) // set the account number
	k.SetModuleAccountPrefixed(ctx, maccIP)
	maccI := (maccIP.ToCosmosAccountI()).(sdk.ModuleAccountI)

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
