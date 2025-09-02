package keeper

import (
	"encoding/json"
	"fmt"
	"os"

	sdkerr "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
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

// IMPORTANT!
// setting system cache is part of the deterministic process and happens only at:
// * at chain bootstrap for system contracts
// * governance upgrades for cached contracts (roles, codes)
func (k *Keeper) SetSystemBootstrapData(ctx sdk.Context, data *types.SystemBootstrapData) error {
	store := ctx.KVStore(k.storeKey)
	databz, err := json.Marshal(data)
	if err != nil {
		return sdkerr.Wrapf(err, "cannot marshal system bootstrap data")
	}
	store.Set(types.GetCacheSystemBootstrapPrefix(), databz)
	return nil
}

// IMPORTANT! we exclude this caching from consuming gas or changing store commits!!
// writing the state is deterministic
// but reading is NOT deterministic: it can happen any time the node is restarted or after statesync
// reading does not change state root hashes, and we make it not consume gas
// otherwise, we would get results hash mismatch in blocks due to gas consumption
func (k *Keeper) GetSystemBootstrapData(ctx_ sdk.Context) (*types.SystemBootstrapData, error) {
	ctx := sdk.NewContext(ctx_.MultiStore().CacheMultiStore(), ctx_.BlockHeader(), ctx_.IsCheckTx(), ctx_.Logger())
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(k.queryGasLimit))

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
	data := k.GetSystemBootstrap(ctx)
	if data == nil {
		panic("cannot find system bootstrap data")
	}
	return data.RoleAddress
}

func (k *Keeper) GetCodeRegistryAddress(ctx sdk.Context) mcodec.AccAddressPrefixed {
	data := k.GetSystemBootstrap(ctx)
	if data == nil {
		panic("cannot find system bootstrap data")
	}
	return data.CodeRegistryAddress
}

func (k *Keeper) GetSystemBootstrap(ctx sdk.Context) *types.SystemBootstrap {
	data, err := types.GetSystemBootstrap(k.wasmvm.goContextParent)
	if err != nil || data == nil || len(data.RoleAddress.Bytes()) == 0 {
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
		types.SetSystemBootstrap(k.wasmvm.goContextParent, data)
	}
	return data
}

func (k *Keeper) UpdateSystemCache(ctx sdk.Context, req *types.SystemBootstrap) error {
	cache := k.GetSystemBootstrap(ctx)
	if req == nil {
		return nil
	}
	if len(req.RoleAddress.Bytes()) > 0 {
		cache.RoleAddress = req.RoleAddress
		k.Logger(ctx).Info("system cache updated roles", "contract_address", req.RoleAddress.String())
	}
	if len(req.CodeRegistryAddress.Bytes()) > 0 {
		k.Logger(ctx).Info("system cache updated codes", "contract_address", req.CodeRegistryAddress.String())
		cache.CodeRegistryAddress = req.CodeRegistryAddress
		cache.CodeRegistryId = req.CodeRegistryId
		// if these are missing, we should just stop the node
		// TODO test if this is the right approach
		if req.CodeRegistryCodeInfo == nil {
			k.Logger(ctx).Error("system cache update: tried to update code registry: missing code info", "contract_address", req.CodeRegistryAddress.String())
			os.Exit(1)
		}
		if req.CodeRegistryContractInfo == nil {
			k.Logger(ctx).Error("system cache update: tried to update code registry: missing contract info", "contract_address", req.CodeRegistryAddress.String())
			os.Exit(1)
		}
		cache.CodeRegistryCodeInfo = req.CodeRegistryCodeInfo
		cache.CodeRegistryContractInfo = req.CodeRegistryContractInfo
	}

	err := k.SetSystemBootstrap(ctx, cache)
	if err != nil {
		return err
	}
	return nil
}

func (k *Keeper) FinalizeBlockResultHandler(ctx sdk.Context, resp *abci.ResponseFinalizeBlock) error {
	return nil
}

func (k *Keeper) EndBlockResultHandler(ctx sdk.Context, resp *abci.ResponseFinalizeBlock) error {
	evs := resp.Events
	for _, txr := range resp.TxResults {
		evs = append(evs, txr.GetEvents()...)
	}

	// TODO (security) check contract address has correct role!!!
	for _, ev := range evs {
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
				err = k.UpdateSystemCacheCodes(ctx, cache, addr)
				if err != nil {
					return err
				}
			}
			if changed {
				err := k.SetSystemBootstrap(ctx, cache)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (k *Keeper) UpdateSystemCacheCodes(ctx sdk.Context, cache *types.SystemBootstrap, addr mcodec.AccAddressPrefixed) error {
	cache.CodeRegistryAddress = addr
	// get code & contract info for address
	contractInfo, codeInfo, _, err := k.ContractInstance(ctx, cache.CodeRegistryAddress)
	if err != nil {
		return sdkerr.Wrapf(err, "contract registry code info upgrade not found %s", addr.String())
	}
	if codeInfo == nil || contractInfo == nil {
		return fmt.Errorf("contract registry code info upgrade not found %s", addr.String())
	}
	cache.CodeRegistryCodeInfo = codeInfo
	cache.CodeRegistryContractInfo = contractInfo
	return nil
}
