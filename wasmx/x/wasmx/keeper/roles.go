package keeper

import (
	"encoding/json"
	"fmt"
	"strings"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// Note!! the set functions must only be used at chain initialization
// during chain lifecycle, just interact with the ROLES contract

// role => contractAddress
// label => role
// contractAddress => label

// RegisterRoleInitial registers a contract
// this must only be used at chain initialization
// it stores directly in the ROLES contract
func (k *Keeper) RegisterRoleInitial(
	ctx sdk.Context,
	role string,
	label string,
	contractAddress mcodec.AccAddressPrefixed,
) error {
	roleObj := &types.RoleChanged{
		Role:            role,
		Label:           label,
		ContractAddress: contractAddress.String(),
	}
	k.SetContractAddressByRole(ctx, role, contractAddress)
	k.SetRoleByLabel(ctx, roleObj)
	k.SetRoleLabelByContract(ctx, contractAddress.Bytes(), label)
	return nil
}

func (k *Keeper) GetAddressOrRoleInitial(ctx sdk.Context, addressOrRole string) (mcodec.AccAddressPrefixed, error) {
	addr, found := k.GetContractAddressByRoleInitial(ctx, addressOrRole)
	if found {
		return addr, nil
	}
	role := k.GetRoleByLabelInitial(ctx, addressOrRole)
	if role != nil {
		contractAddr, err := k.accBech32Codec.StringToAccAddressPrefixed(role.ContractAddress)
		if err != nil {
			return mcodec.AccAddressPrefixed{}, fmt.Errorf("role address decoding failed: %s: %s", role.ContractAddress, err.Error())
		}
		return contractAddr, nil
	}
	contractAddr, err := k.accBech32Codec.StringToAccAddressPrefixed(addressOrRole)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, fmt.Errorf("no role found and address invalid: %s: %s", addressOrRole, err.Error())
	}
	return contractAddr, nil
}

// GetContractAddressByRoleInitial
func (k *Keeper) GetContractAddressByRoleInitial(ctx sdk.Context, role string) (mcodec.AccAddressPrefixed, bool) {
	fmt.Println("--GetContractAddressByRoleInitial--", role)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRolePrefix(types.AccAddressFromHex(types.ADDR_ROLES), role))
	if bz == nil {
		return mcodec.AccAddressPrefixed{}, false
	}
	addr, err := k.accBech32Codec.StringToAccAddressPrefixed(string(bz))
	if err != nil {
		return mcodec.AccAddressPrefixed{}, false
	}
	fmt.Println("--GetContractAddressByRoleInitial.addr--", addr.String())
	return addr, true
}

// SetContractAddressByRole
func (k *Keeper) SetContractAddressByRole(ctx sdk.Context, role string, contractAddress mcodec.AccAddressPrefixed) {
	fmt.Println("--SetContractAddressByRole--", role, contractAddress.String())
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRolePrefix(types.AccAddressFromHex(types.ADDR_ROLES), role), []byte(contractAddress.String()))
}

// GetRoleByLabelInitial
func (k *Keeper) GetRoleByLabelInitial(ctx sdk.Context, label string) *types.RoleChanged {
	fmt.Println("--GetRoleByLabelInitial--", label)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRoleLabelPrefix(types.AccAddressFromHex(types.ADDR_ROLES), label))
	if bz == nil {
		return nil
	}
	rolej := &types.RoleChanged{}
	err := k.cdc.Unmarshal(bz, rolej)
	if err != nil {
		return nil
	}
	fmt.Println("--GetRoleByLabelInitial.rolej--", rolej)
	return rolej
}

// SetRoleByLabel
func (k *Keeper) SetRoleByLabel(ctx sdk.Context, role *types.RoleChanged) {
	if role == nil {
		return
	}
	fmt.Println("--SetRoleByLabel--", role)
	store := ctx.KVStore(k.storeKey)
	value, err := k.cdc.Marshal(&types.RoleChanged{Role: role.Role, Label: role.Label, ContractAddress: role.ContractAddress})
	if err != nil {
		return
	}
	store.Set(types.GetRoleLabelPrefix(types.AccAddressFromHex(types.ADDR_ROLES), role.Label), value)
}

// GetRoleLabelByContractInitial
func (k *Keeper) GetRoleLabelByContractInitial(ctx sdk.Context, contractAddress sdk.AccAddress) string {
	fmt.Println("--GetRoleLabelByContractInitial--", contractAddress)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRoleContractPrefix(types.AccAddressFromHex(types.ADDR_ROLES), contractAddress))
	fmt.Println("--GetRoleLabelByContractInitial.res--", string(bz))
	return string(bz)
}

// SetRoleLabelByContract
func (k *Keeper) SetRoleLabelByContract(ctx sdk.Context, contractAddress sdk.AccAddress, label string) {
	fmt.Println("--SetRoleLabelByContract--", label, contractAddress.String())
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRoleContractPrefix(types.AccAddressFromHex(types.ADDR_ROLES), contractAddress), []byte(label))
}

// TODO cache roles contract in SystemBootstrap
// GetRoleByContractAddress ; used in WasmxCall
func (k *Keeper) GetRoleByContractAddress(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed) *types.RoleJSON {
	cache := k.MustGetSystemBootstrap(ctx)
	if cache.RoleAddress.String() == contractAddress.String() {
		role, err := k.GetRoleByRoleName(ctx, types.ROLE_ROLES)
		if err != nil {
			return nil
		}
		return role
	}
	if cache.CodeRegistryAddress.String() == contractAddress.String() {
		role, err := k.GetRoleByRoleName(ctx, types.ROLE_STORAGE_CONTRACTS)
		if err != nil {
			return nil
		}
		return role
	}
	label := k.GetRoleLabelByContract(ctx, contractAddress)
	if label == "" {
		return nil
	}
	return k.GetRoleByLabel(ctx, label)
}

func (k *Keeper) IsInternalContract(ctx sdk.Context, addressOrRole string) (bool, error) {
	contractAddr := k.GetRoleContractAddress(ctx)
	msg := fmt.Sprintf(`{"IsInternalContract":{"addressOrRole":"%s"}}`, addressOrRole)
	data, err := k.internalQuery(ctx, contractAddr, msg)
	if err != nil {
		return false, err
	}
	if len(data) != 1 {
		return false, nil
	}
	return data[0] == 1, nil
}

func (k *Keeper) GetAddressOrRole(ctx sdk.Context, addressOrRole string) (mcodec.AccAddressPrefixed, error) {
	if addressOrRole == "" {
		return mcodec.AccAddressPrefixed{}, fmt.Errorf("addressOrRole is empty")
	}
	cache := k.MustGetSystemBootstrap(ctx)
	if addressOrRole == types.ROLE_ROLES {
		return cache.RoleAddress, nil
	}
	if addressOrRole == types.ROLE_STORAGE_CONTRACTS {
		return cache.CodeRegistryAddress, nil
	}
	contractAddr := cache.RoleAddress
	msg := fmt.Sprintf(`{"GetAddressOrRole":{"addressOrRole":"%s"}}`, addressOrRole)
	data, err := k.internalQuery(ctx, contractAddr, msg)
	if err != nil {
		// this happens only at chain instantiation, so we read directly from storage
		if strings.Contains(err.Error(), `contract: not found`) {
			k.Logger(ctx).Info("reading directly from storage, must be instantiating the chain: get address or role", "addressOrRole", addressOrRole)
			fmt.Println("---GetAddressOrRoleInitial--", addressOrRole, err.Error())
			resultAddr, err := k.GetAddressOrRoleInitial(ctx, addressOrRole)
			fmt.Println("---GetAddressOrRoleInitial--", err, resultAddr)
			if err != nil {
				return mcodec.AccAddressPrefixed{}, err
			}
			return resultAddr, nil
		}
		return mcodec.AccAddressPrefixed{}, err
	}
	resultAddr, err := k.accBech32Codec.StringToAccAddressPrefixed(string(data))
	if err != nil {
		return mcodec.AccAddressPrefixed{}, err
	}
	return resultAddr, nil
}

func (k *Keeper) GetRoleByRoleName(ctx sdk.Context, roleName string) (*types.RoleJSON, error) {
	if roleName == "" {
		return nil, fmt.Errorf("host.GetRoleByRoleName: empty role")
	}
	contractAddr := k.GetRoleContractAddress(ctx)
	msg := fmt.Sprintf(`{"GetRoleByRoleName":{"role":"%s"}}`, roleName)
	data, err := k.internalQuery(ctx, contractAddr, msg)
	if err != nil {
		return nil, err
	}
	var role types.RoleJSON
	err = json.Unmarshal(data, &role)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "could not decode role")
	}
	return &role, nil
}

// GetRoleLabelByContract
func (k *Keeper) GetRoleLabelByContract(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed) string {
	roleAddress := k.GetRoleContractAddress(ctx)
	msg := fmt.Sprintf(`{"GetRoleLabelByContract":{"address":"%s"}}`, contractAddress.String())
	data, err := k.internalQuery(ctx, roleAddress, msg)
	if err != nil {
		// this happens only at chain instantiation, so we read directly from storage
		if strings.Contains(err.Error(), `contract: not found`) {
			k.Logger(ctx).Info("reading directly from storage, must be instantiating the chain: get role label from address", "address", contractAddress.String())
			fmt.Println("---GetRoleLabelByContractInitial--", contractAddress.String(), err.Error())
			return k.GetRoleLabelByContractInitial(ctx, contractAddress.Bytes())
		}
		return ""
	}
	return string(data)
}

// GetRoleByLabel
func (k *Keeper) GetRoleByLabel(ctx sdk.Context, label string) *types.RoleJSON {
	if label == "" {
		return nil
	}
	contractAddr := k.GetRoleContractAddress(ctx)
	msg := fmt.Sprintf(`{"GetRoleByLabel":{"label":"%s"}}`, label)
	// Note! role contract should not have any other depedencies aside from the host import interface
	data, err := k.internalQuery(ctx, contractAddr, msg)
	if err != nil {
		// this happens only at chain instantiation, so we read directly from storage
		if strings.Contains(err.Error(), `contract: not found`) {
			k.Logger(ctx).Info("reading directly from storage, must be instantiating the chain: get role by label", "label", label)
			fmt.Println("---GetRoleByLabelInitial--", label, err.Error())
			role := k.GetRoleByLabelInitial(ctx, label)
			fmt.Println("---GetRoleByLabelInitial--", role)
			return &types.RoleJSON{Role: role.Role, Primary: 0, Labels: []string{role.Label}, Addresses: []string{role.ContractAddress}}
		}
		return nil
	}
	var role types.RoleJSON
	err = json.Unmarshal(data, &role)
	if err != nil {
		return nil
	}
	return &role
}

func (k *Keeper) internalQuery(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, msg string) ([]byte, error) {
	msgbz, err := json.Marshal(&types.WasmxExecutionMessage{Data: []byte(msg)})
	if err != nil {
		return nil, err
	}
	data, err := k.Query(ctx, contractAddr, contractAddr, msgbz, nil, nil)
	if err != nil {
		return nil, err
	}
	var rresp types.ContractResponse
	err = json.Unmarshal(data, &rresp)
	if err != nil {
		return nil, err
	}
	return rresp.Data, nil
}

func (k *Keeper) internalExecute(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, msg string) ([]byte, error) {
	msgbz, err := json.Marshal(&types.WasmxExecutionMessage{Data: []byte(msg)})
	if err != nil {
		return nil, err
	}
	data, err := k.Execute(ctx, contractAddr, contractAddr, msgbz, nil, nil, false)
	if err != nil {
		return nil, err
	}
	return data, nil
}
