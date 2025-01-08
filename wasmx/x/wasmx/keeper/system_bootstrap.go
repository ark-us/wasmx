package keeper

import (
	"encoding/json"
	"fmt"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
