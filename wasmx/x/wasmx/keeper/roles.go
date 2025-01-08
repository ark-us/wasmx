package keeper

import (
	"encoding/json"
	"fmt"
	"strings"

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
	roleObj := &types.Role{
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
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRolePrefix(types.AccAddressFromHex(types.ADDR_ROLES), role))
	if bz == nil {
		return mcodec.AccAddressPrefixed{}, false
	}
	addr, err := k.accBech32Codec.StringToAccAddressPrefixed(string(bz))
	if err != nil {
		return mcodec.AccAddressPrefixed{}, false
	}
	return addr, true
}

// SetContractAddressByRole
func (k *Keeper) SetContractAddressByRole(ctx sdk.Context, role string, contractAddress mcodec.AccAddressPrefixed) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRolePrefix(types.AccAddressFromHex(types.ADDR_ROLES), role), []byte(contractAddress.String()))
}

// GetRoleByLabelInitial
func (k *Keeper) GetRoleByLabelInitial(ctx sdk.Context, label string) *types.Role {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRoleLabelPrefix(types.AccAddressFromHex(types.ADDR_ROLES), label))
	if bz == nil {
		return nil
	}
	rolej := &types.RoleJSON{}
	err := json.Unmarshal(bz, &rolej)
	if err != nil {
		return nil
	}
	return &types.Role{Role: rolej.Role, Label: rolej.Label, ContractAddress: rolej.ContractAddress}
}

// SetRoleByLabel
func (k *Keeper) SetRoleByLabel(ctx sdk.Context, role *types.Role) {
	if role == nil {
		return
	}
	store := ctx.KVStore(k.storeKey)
	value, err := json.Marshal(&types.RoleJSON{Role: role.Role, Label: role.Label, ContractAddress: role.ContractAddress})
	if err != nil {
		return
	}
	store.Set(types.GetRoleLabelPrefix(types.AccAddressFromHex(types.ADDR_ROLES), role.Label), value)
}

// GetRoleLabelByContractInitial
func (k *Keeper) GetRoleLabelByContractInitial(ctx sdk.Context, contractAddress sdk.AccAddress) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRoleContractPrefix(types.AccAddressFromHex(types.ADDR_ROLES), contractAddress))
	return string(bz)
}

// SetRoleLabelByContract
func (k *Keeper) SetRoleLabelByContract(ctx sdk.Context, contractAddress sdk.AccAddress, label string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRoleContractPrefix(types.AccAddressFromHex(types.ADDR_ROLES), contractAddress), []byte(label))
}

// GetRoleByContractAddress
func (k *Keeper) GetRoleByContractAddress(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed) string {
	label := k.GetRoleLabelByContract(ctx, contractAddress)
	if label == "" {
		return ""
	}
	role := k.GetRoleByLabel(ctx, label)
	if role == nil {
		return ""
	}
	return role.Role
}

func (k *Keeper) GetAddressOrRole(ctx sdk.Context, addressOrRole string) (mcodec.AccAddressPrefixed, error) {
	contractAddr := k.GetRoleContractAddress(ctx)
	msg := fmt.Sprintf(`{"GetAddressOrRole":{"addressOrRole":"%s"}}`, addressOrRole)
	data, err := k.internalQuery(ctx, contractAddr, msg)
	if err != nil {
		// this happens only at chain instantiation, so we read directly from storage
		if strings.Contains(err.Error(), `contract: not found`) {
			resultAddr, err := k.GetAddressOrRoleInitial(ctx, addressOrRole)
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

// GetRoleLabelByContract
func (k *Keeper) GetRoleLabelByContract(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed) string {
	contractAddr := k.GetRoleContractAddress(ctx)
	msg := fmt.Sprintf(`{"GetRoleLabelByContract":{"address":"%s"}}`, contractAddress.String())
	data, err := k.internalQuery(ctx, contractAddr, msg)
	if err != nil {
		// this happens only at chain instantiation, so we read directly from storage
		if strings.Contains(err.Error(), `contract: not found`) {
			return k.GetRoleLabelByContractInitial(ctx, contractAddress.Bytes())
		}
		return ""
	}
	return string(data)
}

// GetRoleByLabel
func (k *Keeper) GetRoleByLabel(ctx sdk.Context, label string) *types.Role {
	contractAddr := k.GetRoleContractAddress(ctx)
	msg := fmt.Sprintf(`{"GetRoleByLabel":{"label":"%s"}}`, label)
	// Note! role contract should not have any other depedencies aside from the host import interface
	data, err := k.internalQuery(ctx, contractAddr, msg)
	if err != nil {
		// this happens only at chain instantiation, so we read directly from storage
		if strings.Contains(err.Error(), `contract: not found`) {
			return k.GetRoleByLabelInitial(ctx, label)
		}
		return nil
	}
	var role types.Role
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
