package keeper

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"mythos/v1/x/wasmx/ioutils"
	"mythos/v1/x/wasmx/types"
	wasmeth "mythos/v1/x/wasmx/vm"
)

func (k Keeper) Create(ctx sdk.Context, creator sdk.AccAddress, wasmByteCode []byte, deps []string, metadata types.CodeMetadata) (uint64, []byte, error) {
	return k.create(ctx, creator, wasmByteCode, deps, metadata)
}

func (k Keeper) Deploy(
	ctx sdk.Context,
	creator sdk.AccAddress,
	wasmByteCode []byte,
	deps []string,
	metadata types.CodeMetadata,
	initMsg types.RawContractMessage,
	funds sdk.Coins,
	label string,
) (uint64, []byte, sdk.AccAddress, error) {
	return k.CreateInterpreted(ctx, creator, nil, wasmByteCode, deps, metadata, initMsg, funds, label, []byte{})
}

func (k Keeper) PinCode(ctx sdk.Context, codeId uint64, compiledFolderPath string) error {
	fmt.Println("--PinCode")
	return k.pinCode(ctx, codeId, compiledFolderPath)
}

func (k Keeper) UnpinCode(ctx sdk.Context, codeId uint64) error {
	return k.unpinCode(ctx, codeId)
}

func (k Keeper) Instantiate(ctx sdk.Context, codeId uint64, creator sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins, label string) (sdk.AccAddress, []byte, error) {
	return k.instantiate(ctx, codeId, creator, msg, funds, label)
}

func (k Keeper) Instantiate2(ctx sdk.Context, codeId uint64, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins, salt []byte, fixMsg bool, label string) (sdk.AccAddress, []byte, error) {
	return k.instantiate2(ctx, codeId, senderAddr, msg, funds, salt, fixMsg, label)
}

func (k Keeper) Execute(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins, dependencies []string) ([]byte, error) {
	return k.execute(ctx, contractAddr, senderAddr, msg, funds, dependencies)
}

func (k Keeper) ExecuteWithOrigin(ctx sdk.Context, originAddr sdk.AccAddress, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins) ([]byte, error) {
	return k.executeWithOrigin(ctx, originAddr, contractAddr, senderAddr, msg, funds)
}

func (k Keeper) ExecuteDelegate(ctx sdk.Context, originAddr sdk.AccAddress, codeContractAddr sdk.AccAddress, storageContractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins) ([]byte, error) {
	return nil, nil
}

func (k Keeper) Query(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error) {
	return k.query(ctx, contractAddr, senderAddr, msg, funds, deps)
}

// QueryRaw returns the contract's state for give key. Returns `nil` when key is `nil`.
func (k Keeper) QueryRaw(ctx sdk.Context, contractAddress sdk.AccAddress, key []byte) []byte {
	defer telemetry.MeasureSince(time.Now(), "wasmx", "contract", "query-raw")
	if key == nil {
		return nil
	}
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	return prefixStore.Get(key)
}

func (k Keeper) GetContractDependency(ctx sdk.Context, addr sdk.AccAddress) (types.ContractDependency, error) {
	_, codeInfo, prefixStoreKey, err := k.ContractInstance(ctx, addr)
	if err != nil {
		return types.ContractDependency{}, err
	}
	var sdeps = k.SystemDepsFromCodeDeps(ctx, codeInfo.Deps)
	var filepath string
	if codeInfo.Pinned {
		filepath = k.wasmvm.build_path_pinned(k.wasmvm.DataDir, codeInfo.CodeHash)
	} else {
		filepath = k.wasmvm.build_path(k.wasmvm.DataDir, codeInfo.CodeHash)
	}
	if len(codeInfo.InterpretedBytecodeRuntime) > 0 {
		filepath = ""
	}

	return types.ContractDependency{
		Address:    addr,
		StoreKey:   prefixStoreKey,
		FilePath:   filepath,
		SystemDeps: sdeps,
		Bytecode:   codeInfo.InterpretedBytecodeRuntime,
	}, nil
}

func (k Keeper) create(ctx sdk.Context, creator sdk.AccAddress, wasmCode []byte, deps []string, metadata types.CodeMetadata) (codeID uint64, checksum []byte, err error) {
	if creator == nil {
		return 0, checksum, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot be nil")
	}

	if ioutils.IsGzip(wasmCode) {
		ctx.GasMeter().ConsumeGas(k.gasRegister.UncompressCosts(len(wasmCode)), "Uncompress gzip bytecode")
		wasmCode, err = ioutils.Uncompress(wasmCode, uint64(types.MaxWasmSize))
		if err != nil {
			return 0, checksum, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
		}
	}

	if !ioutils.IsWasm(wasmCode) {
		return 0, checksum, sdkerrors.Wrap(types.ErrCreateFailed, "this is not wasm code, use deploy")
	}

	// TODO cache current supported deps (including interpreters) and verify these here
	// if !types.IsWasmDeps(deps) {
	// 	return 0, checksum, sdkerrors.Wrap(types.ErrCreateFailed, "incorrect deps")
	// }

	ctx.GasMeter().ConsumeGas(k.gasRegister.CompileCosts(len(wasmCode)), "Compiling wasm bytecode")
	report, err := k.wasmvm.AnalyzeWasm(wasmCode)
	if err != nil {
		return 0, checksum, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	checksum, err = k.wasmvm.Create(wasmCode)
	if err != nil {
		return 0, checksum, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	codeID = k.autoIncrementID(ctx, types.KeyLastCodeID)
	k.Logger(ctx).Debug("storing new contract", "deps", report.Dependencies, "code_id", codeID, "checksum", checksum)
	codeInfo := types.NewCodeInfo(checksum, creator, report.Dependencies, metadata)
	k.storeCodeInfo(ctx, codeID, codeInfo)

	evt := sdk.NewEvent(
		types.EventTypeStoreCode,
		sdk.NewAttribute(types.AttributeKeyChecksum, hex.EncodeToString(checksum)),
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeID, 10)), // last element to be compatible with scripts
	)
	for _, d := range report.Dependencies {
		evt.AppendAttributes(sdk.NewAttribute(types.AttributeKeyRequiredCapability, d))
	}
	ctx.EventManager().EmitEvent(evt)

	return codeID, checksum, nil
}

func (k Keeper) CreateInterpreted(
	ctx sdk.Context,
	creator sdk.AccAddress,
	provenance sdk.AccAddress,
	wasmCode []byte,
	deps []string,
	metadata types.CodeMetadata,
	initMsg types.RawContractMessage,
	deposit sdk.Coins,
	label string,
	salt []byte,
) (codeID uint64, checksum []byte, contractAddress sdk.AccAddress, err error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "createInterpreted")
	// fmt.Println("--CreateInterpreted--")

	if creator == nil {
		return 0, nil, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot be nil")
	}

	if ioutils.IsWasm(wasmCode) {
		return 0, nil, nil, sdkerrors.Wrap(types.ErrCreateFailed, "this is wasm code, use store")
	}

	// TODO cache current supported deps (including interpreters) and verify these here
	// if types.IsWasmDeps(deps) {
	// 	return 0, nil, nil, sdkerrors.Wrap(types.ErrCreateFailed, "incorrect deps")
	// }

	checksum = k.wasmvm.checksum(wasmCode)
	codeID = k.autoIncrementID(ctx, types.KeyLastCodeID)
	k.Logger(ctx).Debug("storing new contract", "deps", deps, "code_id", codeID, "checksum", checksum)
	codeInfo := types.NewCodeInfo(checksum, creator, deps, metadata)
	codeInfo.InterpretedBytecodeDeployment = wasmCode

	addressParent := provenance
	if addressParent == nil {
		addressParent = creator
	}

	if len(salt) == 0 {
		contractAddress = k.EwasmClassicAddressGenerator(addressParent)(ctx, codeID, codeInfo.CodeHash)
	} else {
		contractAddress = k.EwasmPredictableAddressGenerator(addressParent, salt, []byte{}, false)(ctx, codeID, codeInfo.CodeHash)
	}

	_, runtimeCode, err := k.instantiateInternal(ctx, codeID, creator, provenance, initMsg, deposit, contractAddress, &codeInfo, label)
	if err != nil {
		return 0, checksum, contractAddress, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	codeInfo.InterpretedBytecodeRuntime = runtimeCode
	// TODO the hash algo will depend on deps
	codeInfo.RuntimeHash = k.wasmvm.checksum(runtimeCode)
	k.storeCodeInfo(ctx, codeID, codeInfo)

	evt := sdk.NewEvent(
		types.EventTypeStoreCode,
		sdk.NewAttribute(types.AttributeKeyChecksum, hex.EncodeToString(checksum)),
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeID, 10)), // last element to be compatible with scripts
	)
	for _, d := range deps {
		evt.AppendAttributes(sdk.NewAttribute(types.AttributeKeyRequiredCapability, d))
	}
	ctx.EventManager().EmitEvent(evt)

	return codeID, checksum, contractAddress, nil
}

func (k Keeper) storeCodeInfo(ctx sdk.Context, codeID uint64, codeInfo types.CodeInfo) {
	store := ctx.KVStore(k.storeKey)
	// 0x01 | codeID (uint64) -> ContractInfo
	store.Set(types.GetCodeKey(codeID), k.cdc.MustMarshal(&codeInfo))
}

func (k Keeper) importCode(ctx sdk.Context, codeID uint64, codeInfo types.CodeInfo, wasmCode []byte) error {
	if ioutils.IsGzip(wasmCode) {
		var err error
		wasmCode, err = ioutils.Uncompress(wasmCode, uint64(types.MaxWasmSize))
		if err != nil {
			return sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
		}
	}
	newCodeHash, err := k.wasmvm.Create(wasmCode)
	if err != nil {
		return sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	if !bytes.Equal(codeInfo.CodeHash, newCodeHash) {
		return sdkerrors.Wrap(types.ErrInvalid, "code hashes not same")
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetCodeKey(codeID)
	if store.Has(key) {
		return sdkerrors.Wrapf(types.ErrDuplicate, "duplicate code: %d", codeID)
	}
	// 0x01 | codeID (uint64) -> ContractInfo
	store.Set(key, k.cdc.MustMarshal(&codeInfo))
	return nil
}

func (k Keeper) instantiateWithAddress(
	ctx sdk.Context,
	codeID uint64,
	creator sdk.AccAddress,
	contractAddress sdk.AccAddress,
	initMsg []byte,
	deposit sdk.Coins,
	label string,
) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "instantiate_with_address")

	// get contact info
	codeInfo := k.GetCodeInfo(ctx, codeID)
	if codeInfo == nil {
		return nil, sdkerrors.Wrap(types.ErrNotFound, "code")
	}
	_, data, err := k.instantiateInternal(ctx, codeID, creator, nil, initMsg, deposit, contractAddress, codeInfo, label)
	return data, err
}

func (k Keeper) instantiate(
	ctx sdk.Context,
	codeID uint64,
	creator sdk.AccAddress,
	initMsg []byte,
	deposit sdk.Coins,
	label string,
) (sdk.AccAddress, []byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "instantiate")

	// get contact info
	codeInfo := k.GetCodeInfo(ctx, codeID)
	if codeInfo == nil {
		return nil, nil, sdkerrors.Wrap(types.ErrNotFound, "code")
	}
	// TODO deps: support multiple types of address generation
	contractAddress := k.EwasmClassicAddressGenerator(creator)(ctx, codeID, codeInfo.CodeHash)
	return k.instantiateInternal(ctx, codeID, creator, nil, initMsg, deposit, contractAddress, codeInfo, label)
}

func (k Keeper) instantiate2(
	ctx sdk.Context,
	codeID uint64,
	creator sdk.AccAddress,
	initMsg []byte,
	deposit sdk.Coins,
	salt []byte,
	fixMsg bool,
	label string,
) (sdk.AccAddress, []byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "instantiate")

	// get contact info
	codeInfo := k.GetCodeInfo(ctx, codeID)
	if codeInfo == nil {
		return nil, nil, sdkerrors.Wrap(types.ErrNotFound, "code")
	}
	// TODO if we support multiple types of address generation
	// the type should be saved in CodeInfo
	contractAddress := k.EwasmPredictableAddressGenerator(creator, salt, initMsg, fixMsg)(ctx, codeID, codeInfo.CodeHash)
	return k.instantiateInternal(ctx, codeID, creator, nil, initMsg, deposit, contractAddress, codeInfo, label)
}

func (k Keeper) instantiateInternal(
	ctx sdk.Context,
	codeID uint64,
	creator sdk.AccAddress,
	provenance sdk.AccAddress,
	initMsg []byte,
	deposit sdk.Coins,
	contractAddress sdk.AccAddress,
	codeInfo *types.CodeInfo,
	label string,
) (sdk.AccAddress, []byte, error) {
	if creator == nil {
		return nil, nil, types.ErrEmpty.Wrap("creator")
	}
	instanceCosts := k.gasRegister.NewContractInstanceCosts(k.IsPinnedCode(ctx, codeID), len(initMsg))
	ctx.GasMeter().ConsumeGas(instanceCosts, "Loading wasm module: instantiate")

	if k.HasContractInfo(ctx, contractAddress) {
		return nil, nil, types.ErrDuplicate.Wrap("instance with this contract address already exists")
	}

	// check account
	// every cosmos module can define custom account types when needed. The cosmos-sdk comes with extension points
	// to support this and a set of base and vesting account types that we integrated in our default lists.
	// But not all account types of other modules are known or may make sense for contracts, therefore we kept this
	// decision logic also very flexible and extendable. We provide new options to overwrite the default settings via WithAcceptedAccountTypesOnContractInstantiation and
	// WithPruneAccountTypesOnContractInstantiation as constructor arguments
	existingAcct := k.accountKeeper.GetAccount(ctx, contractAddress)
	if existingAcct != nil {
		if existingAcct.GetSequence() != 0 || existingAcct.GetPubKey() != nil {
			return nil, nil, types.ErrAccountExists.Wrap("address is claimed by external account")
		}

		// consider an account in the wasmx namespace spam and overwrite it.
		k.Logger(ctx).Info("pruning existing account for contract instantiation", "address", contractAddress.String())
		contractAccount := k.accountKeeper.NewAccountWithAddress(ctx, contractAddress)
		k.accountKeeper.SetAccount(ctx, contractAccount)
	} else {
		// create an empty account (so we don't have issues later)
		contractAccount := k.accountKeeper.NewAccountWithAddress(ctx, contractAddress)
		k.accountKeeper.SetAccount(ctx, contractAccount)
	}
	// deposit initial contract funds
	if !deposit.IsZero() {
		if err := k.TransferCoins(ctx, creator, contractAddress, deposit); err != nil {
			return nil, nil, err
		}
	}

	// prepare params for contract instantiate call
	info := types.NewInfo(creator, creator, deposit)
	env := types.NewEnv(ctx, k.denom, contractAddress, codeInfo.CodeHash, codeInfo.InterpretedBytecodeDeployment, info)

	// create prefixed data store
	// 0x03 | BuildContractAddressClassic (sdk.AccAddress)
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)

	// prepare querier
	handler := k.newCosmosHandler(ctx, contractAddress)
	var systemDeps = k.SystemDepsFromCodeDeps(ctx, codeInfo.Deps)

	// instantiate wasm contract
	res, gasUsed, err := k.wasmvm.Instantiate(ctx, codeInfo, env, initMsg, prefixStoreKey, prefixStore, handler, k.gasMeter(ctx), systemDeps)
	k.consumeRuntimeGas(ctx, gasUsed)

	if err != nil {
		return nil, nil, sdkerrors.Wrap(types.ErrInstantiateFailed, err.Error())
	}

	// persist instance first
	contractInfo := types.NewContractInfo(codeID, creator, provenance, initMsg, label)

	// check for IBC flag - TODO use codeInfo.Dependencies
	// report, err := k.wasmvm.AnalyzeWasm(codeInfo.CodeHash)
	// if err != nil {
	// 	return nil, nil, sdkerrors.Wrap(types.ErrInstantiateFailed, err.Error())
	// }
	// if report.HasIBCEntryPoints {
	// 	// register IBC port
	// 	ibcPort, err := k.ensureIbcPort(ctx, contractAddress)
	// 	if err != nil {
	// 		return nil, nil, err
	// 	}
	// 	contractInfo.IbcPortId = ibcPort
	// }

	// store contract before dispatch so that contract could be called back
	k.storeContractInfo(ctx, contractAddress, &contractInfo)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeInstantiate,
		sdk.NewAttribute(types.AttributeKeyContractAddr, contractAddress.String()),
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeID, 10)),
	))

	return contractAddress, res.Data, nil
}

// PinCode pins the wasm contract in wasmvm cache
func (k Keeper) pinCode(ctx sdk.Context, codeId uint64, compiledFolderPath string) error {
	codeInfo := k.GetCodeInfo(ctx, codeId)
	if codeInfo == nil {
		return sdkerrors.Wrap(types.ErrNotFound, "code info")
	}
	fmt.Println("--before k.wasmvm.Pin")

	if err := k.wasmvm.Pin(codeInfo.CodeHash, compiledFolderPath); err != nil {
		return sdkerrors.Wrap(types.ErrPinContractFailed, err.Error())
	}

	codeInfo.Pinned = true
	k.storeCodeInfo(ctx, codeId, *codeInfo)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePinCode,
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeId, 10)),
		sdk.NewAttribute(types.AttributeKeyChecksum, hex.EncodeToString(codeInfo.CodeHash)),
	))
	return nil
}

// UnpinCode removes the wasm contract from wasmvm cache
func (k Keeper) unpinCode(ctx sdk.Context, codeId uint64) error {
	codeInfo := k.GetCodeInfo(ctx, codeId)
	if codeInfo == nil {
		return sdkerrors.Wrap(types.ErrNotFound, "code info")
	}
	if err := k.wasmvm.Unpin(codeInfo.CodeHash); err != nil {
		return sdkerrors.Wrap(types.ErrUnpinContractFailed, err.Error())
	}

	codeInfo.Pinned = false
	k.storeCodeInfo(ctx, codeId, *codeInfo)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUnpinCode,
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeId, 10)),
	))
	return nil
}

// Execute executes the contract instance
func (k Keeper) execute(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins, dependencies []string) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasmx", "contract", "execute")
	contractInfo, codeInfo, prefixStoreKey, err := k.ContractInstance(ctx, contractAddress)
	if err != nil {
		return nil, err
	}

	// add known dependencies for that codeId
	// TODO system deps in the form of smart contracts
	// e.g. dep = {value, type}
	// if we cannot just load all the modules in the same VM
	allDeps := append(codeInfo.Deps, dependencies...)

	var systemDeps = k.SystemDepsFromCodeDeps(ctx, allDeps)
	var contractDeps []types.ContractDependency

	for _, hexaddr := range allDeps {
		if hexaddr[0:2] != "0x" {
			continue
		}
		addr := wasmeth.AccAddressFromHex(hexaddr)
		contractDep, err := k.GetContractDependency(ctx, addr)
		if err != nil {
			return nil, err
		}
		contractDeps = append(contractDeps, contractDep)
	}

	// TODO panic if coin is not the correct denomination
	// add denom param for wasmx

	executeCosts := k.gasRegister.InstantiateContractCosts(k.IsPinnedCode(ctx, contractInfo.CodeId), len(msg))
	ctx.GasMeter().ConsumeGas(executeCosts, "Loading CosmWasm module: execute")

	// add more funds
	if !coins.IsZero() {
		if err := k.TransferCoins(ctx, caller, contractAddress, coins); err != nil {
			return nil, err
		}
	}

	info := types.NewInfo(caller, caller, coins)
	env := types.NewEnv(ctx, k.denom, contractAddress, codeInfo.CodeHash, codeInfo.InterpretedBytecodeRuntime, info)

	// prepare querier
	handler := k.newCosmosHandler(ctx, contractAddress)
	res, gasUsed, execErr := k.wasmvm.Execute(ctx, &codeInfo, env, msg, prefixStoreKey, k.ContractStore(ctx, prefixStoreKey), handler, k.gasMeter(ctx), systemDeps, contractDeps)
	k.consumeRuntimeGas(ctx, gasUsed)

	// res, _, execErr = k.handleExecutionRerun(ctx, codeInfo.CodeHash, env, info, msg, prefixStore, cosmwasmAPI, querier, gas, costJSONDeserialization, contractAddress, contractInfo, res, gasUsed, execErr, k.wasmVM.Execute)
	if execErr != nil {
		return nil, sdkerrors.Wrap(types.ErrExecuteFailed, execErr.Error())
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeExecute,
		sdk.NewAttribute(types.AttributeKeyContractAddr, contractAddress.String()),
	))

	data, err := k.handleContractResponse(ctx, contractAddress, contractInfo.IbcPortId, res.Attributes, res.Data, res.Events)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "dispatch")
	}

	return data, nil
}

// executeWithOrigin executes the contract instance
func (k Keeper) executeWithOrigin(ctx sdk.Context, origin sdk.AccAddress, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "executeWithOrigin")

	// fail if caller is not a contract
	_, _, _, err := k.ContractInstance(ctx, caller)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrExecuteFailed, "cannot executeWithOrigin from EOA")
	}

	contractInfo, codeInfo, prefixStoreKey, err := k.ContractInstance(ctx, contractAddress)
	if err != nil {
		return nil, err
	}

	executeCosts := k.gasRegister.InstantiateContractCosts(k.IsPinnedCode(ctx, contractInfo.CodeId), len(msg))
	ctx.GasMeter().ConsumeGas(executeCosts, "Loading CosmWasm module: execute")

	// add more funds
	if !coins.IsZero() {
		if err := k.TransferCoins(ctx, caller, contractAddress, coins); err != nil {
			return nil, err
		}
	}

	info := types.NewInfo(origin, caller, coins)
	env := types.NewEnv(ctx, k.denom, contractAddress, codeInfo.CodeHash, codeInfo.InterpretedBytecodeRuntime, info)
	handler := k.newCosmosHandler(ctx, contractAddress)
	var systemDeps = k.SystemDepsFromCodeDeps(ctx, codeInfo.Deps)

	res, gasUsed, execErr := k.wasmvm.Execute(ctx, &codeInfo, env, msg, prefixStoreKey, k.ContractStore(ctx, prefixStoreKey), handler, k.gasMeter(ctx), systemDeps, nil)
	k.consumeRuntimeGas(ctx, gasUsed)

	// res, _, execErr = k.handleExecutionRerun(ctx, codeInfo.CodeHash, env, info, msg, prefixStore, cosmwasmAPI, querier, gas, costJSONDeserialization, contractAddress, contractInfo, res, gasUsed, execErr, k.wasmVM.Execute)

	if execErr != nil {
		return nil, sdkerrors.Wrap(types.ErrExecuteFailed, execErr.Error())
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeExecute,
		sdk.NewAttribute(types.AttributeKeyContractAddr, contractAddress.String()),
	))

	data, err := k.handleContractResponse(ctx, contractAddress, contractInfo.IbcPortId, res.Attributes, res.Data, res.Events)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "dispatch")
	}

	return data, nil
}

// Execute executes the contract instance
func (k Keeper) query(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins, dependencies []string) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "execute")
	contractInfo, codeInfo, prefixStoreKey, err := k.ContractInstance(ctx, contractAddress)
	if err != nil {
		return nil, err
	}

	// TODO panic if coin is not the correct denomination
	// add denom param for wasmx

	executeCosts := k.gasRegister.InstantiateContractCosts(k.IsPinnedCode(ctx, contractInfo.CodeId), len(msg))
	ctx.GasMeter().ConsumeGas(executeCosts, "Loading CosmWasm module: execute")

	// add known dependencies for that codeId
	// TODO system deps in the form of smart contracts
	// e.g. dep = {value, type}
	// if we cannot just load all the modules in the same VM
	allDeps := append(codeInfo.Deps, dependencies...)
	var systemDeps = k.SystemDepsFromCodeDeps(ctx, allDeps)

	var contractDeps []types.ContractDependency
	for _, hexaddr := range allDeps {
		if hexaddr[0:2] != "0x" {
			continue
		}
		addr := wasmeth.AccAddressFromHex(hexaddr)
		contractDep, err := k.GetContractDependency(ctx, addr)
		if err != nil {
			return nil, err
		}
		contractDeps = append(contractDeps, contractDep)
	}

	// add more funds
	if !coins.IsZero() {
		if err := k.TransferCoins(ctx, caller, contractAddress, coins); err != nil {
			return nil, err
		}
	}

	info := types.NewInfo(caller, caller, coins)
	env := types.NewEnv(ctx, k.denom, contractAddress, codeInfo.CodeHash, codeInfo.InterpretedBytecodeRuntime, info)
	handler := k.newCosmosHandler(ctx, contractAddress)
	res, gasUsed, execErr := k.wasmvm.QueryExecute(ctx, &codeInfo, env, msg, prefixStoreKey, k.ContractStore(ctx, prefixStoreKey), handler, k.gasMeter(ctx), systemDeps, contractDeps)
	k.consumeRuntimeGas(ctx, gasUsed)

	// res, _, execErr = k.handleExecutionRerun(ctx, codeInfo.CodeHash, env, info, msg, prefixStore, cosmwasmAPI, querier, gas, costJSONDeserialization, contractAddress, contractInfo, res, gasUsed, execErr, k.wasmVM.Execute)
	if execErr != nil {
		return nil, sdkerrors.Wrap(types.ErrExecuteFailed, execErr.Error())
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeExecute,
		sdk.NewAttribute(types.AttributeKeyContractAddr, contractAddress.String()),
	))

	// data, err := k.handleContractResponse(ctx, contractAddress, contractInfo.IbcPortId, res.Attributes, res.Data, res.Events)
	// if err != nil {
	// 	return nil, sdkerrors.Wrap(err, "dispatch")
	// }

	return json.Marshal(res)
}

// handleContractResponse processes the contract response data by emitting events and sending sub-/messages.
func (k *Keeper) handleContractResponse(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	ibcPort string,
	attrs []types.EventAttribute,
	data []byte,
	evts types.Events,
) ([]byte, error) {
	attributeGasCost := k.gasRegister.EventCosts(attrs, evts)
	ctx.GasMeter().ConsumeGas(attributeGasCost, "Custom contract event attributes")
	// emit all events from this contract itself
	// these are not used
	if len(attrs) != 0 {
		wasmEvents, err := newWasmModuleEvent(attrs, contractAddr)
		if err != nil {
			return nil, err
		}
		ctx.EventManager().EmitEvents(wasmEvents)
	}
	if len(evts) > 0 {
		customEvents, err := newCustomEvents(evts, contractAddr)
		if err != nil {
			return nil, err
		}
		ctx.EventManager().EmitEvents(customEvents)
	}

	return data, nil
}

// Calculate how much gas can we use for the wasmx execution
func (k Keeper) runtimeGasForContract(ctx sdk.Context) uint64 {
	meter := ctx.GasMeter()
	if meter.IsOutOfGas() {
		return 0
	}
	if meter.Limit() == 0 { // infinite gas meter with limit=0 and not out of gas
		return math.MaxUint64
	}
	return k.gasRegister.ToWasmVMGas(meter.Limit() - meter.GasConsumedToLimit())
}

func (k Keeper) consumeRuntimeGas(ctx sdk.Context, gas uint64) {
	consumed := k.gasRegister.FromWasmVMGas(gas)
	ctx.GasMeter().ConsumeGas(consumed, "wasm contract")
	// throw OutOfGas error if we ran out (got exactly to zero due to better limit enforcing)
	if ctx.GasMeter().IsOutOfGas() {
		panic(sdk.ErrorOutOfGas{Descriptor: "Wasmer function execution"})
	}
}

func (k Keeper) SystemDepsFromCodeDeps(ctx sdk.Context, depLabels []string) []types.SystemDep {
	var sdeps []types.SystemDep
	for _, dep := range depLabels {
		if dep[0:2] != "0x" {
			sdep, err := k.SystemDepFromLabel(ctx, dep)
			if err != nil {
				sdep = types.SystemDep{
					Label: dep,
					Role:  dep,
				}
			}
			sdeps = append(sdeps, sdep)
		}
	}
	return sdeps
}

func (k Keeper) SystemDepFromLabel(ctx sdk.Context, label string) (types.SystemDep, error) {
	role := k.GetRoleByLabel(ctx, label)
	if role == nil {
		return types.SystemDep{}, fmt.Errorf("no role from label")
	}
	contractAddress, err := sdk.AccAddressFromBech32(role.ContractAddress)
	if err != nil {
		return types.SystemDep{}, err
	}
	_, codeInfo, _, err := k.ContractInstance(ctx, contractAddress)
	if err != nil {
		return types.SystemDep{}, err
	}
	var path string
	if codeInfo.Pinned {
		path = k.wasmvm.build_path_pinned(k.wasmvm.DataDir, codeInfo.CodeHash)
	} else {
		path = k.wasmvm.build_path(k.wasmvm.DataDir, codeInfo.CodeHash)
	}
	dep := types.SystemDep{
		Role:     role.Role,
		Label:    label,
		FilePath: path,
	}
	return dep, nil
}

// MultipliedGasMeter wraps the GasMeter from context and multiplies all reads by out defined multiplier
type MultipliedGasMeter struct {
	originalMeter sdk.GasMeter
	GasRegister   GasRegister
}

func NewMultipliedGasMeter(originalMeter sdk.GasMeter, gr GasRegister) MultipliedGasMeter {
	return MultipliedGasMeter{originalMeter: originalMeter, GasRegister: gr}
}

var _ types.GasMeter = MultipliedGasMeter{}

func (m MultipliedGasMeter) GasConsumed() sdk.Gas {
	return m.GasRegister.ToWasmVMGas(m.originalMeter.GasConsumed())
}

func (m MultipliedGasMeter) ConsumeGas(gas sdk.Gas, descriptor string) {
	sdkgas := m.GasRegister.FromWasmVMGas(gas)
	m.originalMeter.ConsumeGas(sdkgas, descriptor)
	if m.originalMeter.IsOutOfGas() {
		panic(sdk.ErrorOutOfGas{Descriptor: "Wasmer function execution"})
	}
}

func (k Keeper) gasMeter(ctx sdk.Context) MultipliedGasMeter {
	return NewMultipliedGasMeter(ctx.GasMeter(), k.gasRegister)
}
