package keeper

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdkerr "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"
)

func (k *Keeper) BootstrapSystemContracts(
	ctx sdk.Context,
	bootstrapAccountAddr mcodec.AccAddressPrefixed,
	contracts []types.SystemContract,
	compiledFolderPath string,
) error {
	if len(contracts) < 3 {
		return fmt.Errorf("not enough system contracts")
	}
	if contracts[0].Role != types.ROLE_STORAGE_CONTRACTS {
		return fmt.Errorf("genesis system contract 0 must be %s", types.ROLE_STORAGE_CONTRACTS)
	}
	if contracts[1].Role != types.ROLE_AUTH {
		return fmt.Errorf("genesis system contract 1 must be %s", types.ROLE_AUTH)
	}
	if contracts[2].Role != types.ROLE_ROLES {
		return fmt.Errorf("genesis system contract 2 must be %s", types.ROLE_ROLES)
	}

	var rolesAddress mcodec.AccAddressPrefixed
	var registryAddress mcodec.AccAddressPrefixed
	var registryId uint64
	var registryCodeInfo types.CodeInfo
	var registryContractInfo types.ContractInfo
	genesisRegistry := types.GenesisRegistryContract{
		CodeInfos:     make([]types.CodeInfo, len(contracts)),
		ContractInfos: make([]types.MsgSetContractInfoRequest, len(contracts)),
	}

	// initialize temporary roles first
	// and collect init msg data for the contract registry
	for i, contract := range contracts {
		contractAddress := k.accBech32Codec.BytesToAccAddressPrefixed(types.AccAddressFromHex(contract.Address))
		if contract.Role != "" {
			k.RegisterRoleInitial(ctx, contract.Role, contract.Label, contractAddress)
			if contract.Role == types.ROLE_ROLES {
				rolesAddress = contractAddress
			}
		}

		var codeInfo types.CodeInfo
		var contractInfo types.ContractInfo
		var err error
		codeID := uint64(i + 1)

		if contract.Native {
			codeInfo = types.NewCodeInfo([]byte(contract.Address), bootstrapAccountAddr.String(), contract.Deps, contract.Metadata.ToJson(), contract.Pinned, contract.MeteringOff)
		} else {
			wasmbin := precompiles.GetPrecompileByLabel(k.AddressCodec(), contract.Label)

			codeInfo, err = k.createCodeInfo(ctx, bootstrapAccountAddr, wasmbin, contract.Deps, contract.Metadata.ToJson(), contract.Pinned, contract.MeteringOff)
			if err != nil {
				return sdkerr.Wrap(err, "store system contract: "+contract.Label)
			}
		}

		contractInfo = types.NewContractInfo(codeID, bootstrapAccountAddr.String(), bootstrapAccountAddr.String(), contract.InitMessage, contract.Label)
		if !contract.Native {
			contractInfo.StorageType = contract.StorageType
		}

		k.Logger(ctx).Debug("core contract", "deps", codeInfo.Deps, "code_id", codeID, "checksum", hex.EncodeToString(codeInfo.CodeHash))

		if contract.Role == types.ROLE_STORAGE_CONTRACTS {
			registryAddress = contractAddress
			registryId = codeID
			registryCodeInfo = codeInfo
			registryContractInfo = contractInfo
		}

		genesisRegistry.CodeInfos[i] = codeInfo
		genesisRegistry.ContractInfos[i] = types.MsgSetContractInfoRequest{Address: contractAddress.String(), ContractInfo: contractInfo}
	}

	// initialize SystemBootstrap
	bootstrapData := &types.SystemBootstrap{RoleAddress: rolesAddress, CodeRegistryAddress: registryAddress, CodeRegistryId: registryId, CodeRegistryCodeInfo: &registryCodeInfo, CodeRegistryContractInfo: &registryContractInfo}
	err := k.SetSystemBootstrap(ctx, bootstrapData)
	if err != nil {
		return err
	}
	registryGenesisBz, err := json.Marshal(&genesisRegistry)
	if err != nil {
		return err
	}
	registryGenesisWrap, err := json.Marshal(&types.WasmxExecutionMessage{Data: registryGenesisBz})
	if err != nil {
		return err
	}

	for i, contract := range contracts {
		if contract.Role == types.ROLE_STORAGE_CONTRACTS {
			contract.InitMessage = registryGenesisWrap
		}
		err := k.ActivateSystemContract(ctx, bootstrapAccountAddr, contract, compiledFolderPath, uint64(i+1), genesisRegistry.CodeInfos[i], registryAddress)
		if err != nil {
			return sdkerr.Wrap(err, "bootstrap")
		}
	}
	return nil
}

// ActivateSystemContract
func (k *Keeper) ActivateSystemContract(
	ctx sdk.Context,
	bootstrapAccountAddr mcodec.AccAddressPrefixed,
	contract types.SystemContract,
	compiledFolderPath string,
	codeID uint64,
	codeInfo types.CodeInfo,
	// contractInfo types.ContractInfo,
	registryAddress mcodec.AccAddressPrefixed,
) error {
	var err error
	k.SetSystemContract(ctx, contract)

	if contract.Pinned {
		if err := k.pinCodeWithEvent(ctx, codeID, codeInfo.CodeHash, compiledFolderPath, contract.MeteringOff); err != nil {
			return sdkerr.Wrap(err, "pin system contract: "+contract.Label)
		}
	}
	// no address, we just need to create a code id
	if contract.Address == "" {
		k.Logger(ctx).Info("created system contract", "label", contract.Label, "code_id", codeID)
		return nil
	}

	contractAddress := k.accBech32Codec.BytesToAccAddressPrefixed(types.AccAddressFromHex(contract.Address))

	if !contract.Native {
		_, err = k.instantiateWithAddress(
			ctx,
			codeID,
			bootstrapAccountAddr,
			contractAddress,
			contract.StorageType,
			contract.InitMessage,
			nil,
			contract.Label,
			contract.Role,
			codeInfo,
		)
		if err != nil {
			return sdkerr.Wrap(err, "instantiate system contract: "+contract.Label)
		}
	}

	k.ImportContractState(ctx, contractAddress.Bytes(), contract.StorageType, contract.ContractState)

	// this must be stored separately, as the gateway for other roles
	// do this after instantiation to avoid a cycle for the ROLES contract instantiation
	// if contract.Role == types.ROLE_ROLES {
	// 	k.SetRoleContractAddress(ctx, contractAddress)
	// }

	if contract.Role != types.ROLE_STORAGE_CONTRACTS {
		// register the auth account
		err = k.instantiateNewContractAccount(ctx, contractAddress)
		if err != nil {
			return sdkerr.Wrapf(err, "create auth account for contract %s", contract.Label)
		}
	}

	if contract.Role == types.ROLE_AUTH {
		// only now we create the account for the first precompile - contract storage
		err = k.instantiateNewContractAccount(ctx, registryAddress)
		if err != nil {
			return sdkerr.Wrap(err, "create auth account for contracts registry")
		}
	}

	k.Logger(ctx).Info("activated system contract", "label", contract.Label, "address", contractAddress.String(), "hex_address", contract.Address, "code_id", codeID, "role", contract.Role)
	return nil
}

// SetSystemContract
func (k *Keeper) SetSystemContract(ctx sdk.Context, contract types.SystemContract) {
	// for contracts where we just need the code id and are not deployed
	// TODO better, because these contracts will not be exported
	if contract.Address == "" {
		return
	}
	addr := types.AccAddressFromHex(contract.Address)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixSystemContract)
	bz := k.cdc.MustMarshal(&contract)
	prefixStore.Set(addr.Bytes(), bz)
}

// GetSystemContracts
func (k *Keeper) GetSystemContracts(ctx sdk.Context) (contracts []types.SystemContract) {
	k.IterateSystemContracts(ctx, func(contract types.SystemContract) bool {
		contracts = append(contracts, contract)
		return false
	})
	return
}

// IterateSystemContracts
// When the callback returns true, the loop is aborted early.
func (k *Keeper) IterateSystemContracts(ctx sdk.Context, cb func(types.SystemContract) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixSystemContract)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		// cb returns true to stop early
		var value types.SystemContract
		k.cdc.MustUnmarshal(iter.Value(), &value)
		if cb(value) {
			return
		}
	}
}
