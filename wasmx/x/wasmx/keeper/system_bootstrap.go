package keeper

import (
	"encoding/json"
	"fmt"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/cometbft/cometbft/abci/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (k *Keeper) SetSystemBootstrap(ctx sdk.Context, data *types.SystemBootstrap) error {
	internalData := &types.SystemBootstrapData{
		RoleAddress:              data.RoleAddress.String(),
		CodeRegistryAddress:      data.CodeRegistryAddress.String(),
		CodeRegistryId:           data.CodeRegistryId,
		CodeRegistryCodeInfo:     data.CodeRegistryCodeInfo.ToProto(),
		CodeRegistryContractInfo: data.CodeRegistryContractInfo.ToProto(),
	}
	err := k.SetSystemBootstrapData(ctx, internalData)
	if err != nil {
		return err
	}
	return types.SetSystemBootstrap(k.wasmvm.goContextParent, data)
}

func (k *Keeper) SetSystemBootstrapData(ctx sdk.Context, data *types.SystemBootstrapData) error {
	store := ctx.KVStore(k.storeKey)
	databz, err := json.Marshal(data)
	if err != nil {
		return sdkerr.Wrapf(err, "cannot marshal system bootstrap data")
	}
	store.Set(types.GetCacheSystemBootstrapPrefix(), databz)
	return nil
}

func (k *Keeper) GetSystemBootstrapData(ctx sdk.Context) (*types.SystemBootstrapData, error) {
	store := ctx.KVStore(k.storeKey)
	var data types.SystemBootstrapData
	databz := store.Get(types.GetCacheSystemBootstrapPrefix())
	if databz == nil {
		return nil, fmt.Errorf("system bootstrap data missing")
	}
	err := json.Unmarshal(databz, &data)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "cannot unmarshal system bootstrap data")
	}
	return &data, nil
}

func (k *Keeper) GetRoleContractAddress(ctx sdk.Context) mcodec.AccAddressPrefixed {
	data, err := types.GetSystemBootstrap(k.wasmvm.goContextParent)
	if err != nil {
		panic(err)
	}
	return data.RoleAddress
}

func (k *Keeper) GetCodeRegistryAddress(ctx sdk.Context) mcodec.AccAddressPrefixed {
	data, err := types.GetSystemBootstrap(k.wasmvm.goContextParent)
	if err != nil {
		panic(err)
	}
	return data.CodeRegistryAddress
}

func (k *Keeper) GetSystemBootstrap(ctx sdk.Context) *types.SystemBootstrap {
	data, err := types.GetSystemBootstrap(k.wasmvm.goContextParent)
	if err != nil || data == nil {
		// read from storage
		_data, err := k.GetSystemBootstrapData(ctx)
		if err != nil {
			panic(err)
		}
		roleAddress, err := k.accBech32Codec.StringToAccAddressPrefixed(_data.RoleAddress)
		if err != nil {
			panic(sdkerr.Wrapf(err, "role address invalid address"))
		}
		registryAddress, err := k.accBech32Codec.StringToAccAddressPrefixed(_data.CodeRegistryAddress)
		if err != nil {
			panic(sdkerr.Wrapf(err, "role address invalid address"))
		}
		data = &types.SystemBootstrap{
			RoleAddress:              roleAddress,
			CodeRegistryAddress:      registryAddress,
			CodeRegistryId:           _data.CodeRegistryId,
			CodeRegistryCodeInfo:     _data.CodeRegistryCodeInfo.ToJson(),
			CodeRegistryContractInfo: _data.CodeRegistryContractInfo.ToJson(),
		}
	}
	return data
}

func (k *Keeper) FinalizeBlockResultHandler(ctx sdk.Context, resp *abci.ResponseFinalizeBlock) error {
	for _, txr := range resp.TxResults {
		for _, ev := range txr.GetEvents() {
			// [{"type":"register_role","attributes":[{"key":"role","value":"storage_contracts","index":true},{"key":"label","value":"storage_contracts_0.0.1","index":true},{"key":"contract_address","value":"mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqrpq5kw64","index":true}]}]
			if ev.Type == types.EventTypeRegisterRole {
				var role string
				var contractAddrStr string
				for _, attr := range ev.Attributes {
					if attr.Key == types.AttributeKeyRole {
						role = attr.Value
					}
					if attr.Key == types.AttributeKeyContractAddress {
						contractAddrStr = attr.Value
					}
				}
				// upgrade cache for roles, contract registry

				// format errors should never happen, but we error
				cache := k.GetSystemBootstrap(ctx)
				changed := false
				if role == types.ROLE_ROLES {
					addr, err := k.accBech32Codec.StringToAccAddressPrefixed(contractAddrStr)
					if err != nil {
						return sdkerr.Wrapf(err, "roles upgrade invalid address %s", contractAddrStr)
					}
					cache.RoleAddress = addr
					changed = true
				}
				if role == types.ROLE_STORAGE_CONTRACTS {
					addr, err := k.accBech32Codec.StringToAccAddressPrefixed(contractAddrStr)
					if err != nil {
						return sdkerr.Wrapf(err, "contract registry upgrade invalid address %s", contractAddrStr)
					}
					cache.CodeRegistryAddress = addr
					changed = true
					// get code & contract info for address
					contractInfo, codeInfo, _, err := k.ContractInstance(ctx, cache.CodeRegistryAddress)
					if err != nil {
						return sdkerr.Wrapf(err, "contract registry code info upgrade not found %s", contractAddrStr)
					}
					if codeInfo == nil || contractInfo == nil {
						return fmt.Errorf("contract registry code info upgrade not found %s", contractAddrStr)
					}
					cache.CodeRegistryCodeInfo = codeInfo
					cache.CodeRegistryContractInfo = contractInfo
				}
				if changed {
					err := k.SetSystemBootstrap(ctx, cache)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
