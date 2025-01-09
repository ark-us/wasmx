package keeper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	sdkerr "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	cw8types "github.com/loredanacirstea/wasmx/x/wasmx/cw8/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/ioutils"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (k *Keeper) Create(ctx sdk.Context, creator mcodec.AccAddressPrefixed, wasmByteCode []byte, deps []string, metadata types.CodeMetadata, pinned bool, meteringOff bool) (uint64, []byte, error) {
	return k.create(ctx, &creator, wasmByteCode, deps, metadata, pinned, meteringOff)
}

func (k *Keeper) Deploy(
	ctx sdk.Context,
	creator mcodec.AccAddressPrefixed,
	wasmByteCode []byte,
	deps []string,
	metadata types.CodeMetadata,
	initMsg types.RawContractMessage,
	funds sdk.Coins,
	label string,
) (uint64, []byte, mcodec.AccAddressPrefixed, error) {
	return k.CreateInterpreted(ctx, &creator, nil, wasmByteCode, deps, metadata, initMsg, funds, label, []byte{})
}

func (k *Keeper) PinCode(ctx sdk.Context, codeHash []byte, compiledFolderPath string, meteringOff bool) error {
	return k.pinCode(ctx, codeHash, compiledFolderPath, meteringOff)
}

func (k *Keeper) PinCodeAndStore(ctx sdk.Context, codeId uint64, compiledFolderPath string, meteringOff bool) error {
	codeInfo, err := k.GetCodeInfo(ctx, codeId)
	if err != nil {
		return err
	}
	return k.pinCodeAndStore(ctx, codeInfo, compiledFolderPath, meteringOff)
}

func (k *Keeper) UnpinCode(ctx sdk.Context, codeId uint64) error {
	return k.unpinCode(ctx, codeId)
}

func (k *Keeper) Instantiate(ctx sdk.Context, codeId uint64, creator mcodec.AccAddressPrefixed, msg types.RawContractMessage, funds sdk.Coins, label string) (mcodec.AccAddressPrefixed, []byte, error) {
	return k.instantiate(ctx, codeId, creator, msg, funds, label)
}

func (k *Keeper) Instantiate2(ctx sdk.Context, codeId uint64, senderAddr mcodec.AccAddressPrefixed, msg types.RawContractMessage, funds sdk.Coins, salt []byte, fixMsg bool, label string) (mcodec.AccAddressPrefixed, []byte, error) {
	return k.instantiate2(ctx, codeId, senderAddr, msg, funds, salt, fixMsg, label)
}

func (k *Keeper) Execute(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg types.RawContractMessage, funds sdk.Coins, dependencies []string, inBackground bool) ([]byte, error) {
	return k.execute(ctx, contractAddr, senderAddr, msg, funds, dependencies, inBackground)
}

func (k *Keeper) ExecuteWithOrigin(ctx sdk.Context, originAddr mcodec.AccAddressPrefixed, contractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg types.RawContractMessage, funds sdk.Coins) ([]byte, error) {
	return k.executeWithOrigin(ctx, originAddr, contractAddr, senderAddr, msg, funds)
}

func (k *Keeper) ExecuteDelegate(ctx sdk.Context, originAddr mcodec.AccAddressPrefixed, codeContractAddr mcodec.AccAddressPrefixed, storageContractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg types.RawContractMessage, funds sdk.Coins) ([]byte, error) {
	return nil, nil
}

func (k *Keeper) Query(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg types.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error) {
	res, err := k.query(ctx, contractAddr, senderAddr, msg, funds, deps, false)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&types.WasmxQueryResponse{Data: res.Data})
}

func (k *Keeper) QueryDebug(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg types.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, []byte, string) {
	res, err := k.query(ctx, contractAddr, senderAddr, msg, funds, deps, true)
	if err != nil {
		errmsg := res.ErrorMessage
		if errmsg == "" {
			errmsg = err.Error()
		}
		return nil, res.MemorySnapshot, errmsg
	}
	data, err := json.Marshal(&types.WasmxQueryResponse{Data: res.Data})
	if err != nil {
		return nil, res.MemorySnapshot, err.Error()
	}
	return data, res.MemorySnapshot, res.ErrorMessage
}

// QueryRaw returns the contract's state for give key. Returns `nil` when key is `nil`.
func (k *Keeper) QueryRaw(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed, key []byte) []byte {
	defer telemetry.MeasureSince(time.Now(), "wasmx", "contract", "query-raw")
	if key == nil {
		return nil
	}
	prefixStoreKey := types.GetContractStorePrefix(contractAddress.Bytes())
	// TODO storage type in QueryRaw
	prefixStore := k.ContractStore(ctx, types.ContractStorageType_CoreConsensus, prefixStoreKey)
	return prefixStore.Get(key)
}

// QuerySmart queries the smart contract itself. cosmwasm compat.
func (k *Keeper) QuerySmart(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, req []byte) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "query-smart")
	senderAddr := contractAddr
	res, err := k.query(ctx, contractAddr, senderAddr, req, nil, nil, false)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (k *Keeper) GetContractDependency(ctx sdk.Context, addr mcodec.AccAddressPrefixed) (*types.ContractDependency, error) {
	contractInfo, codeInfo, prefixStoreKey, err := k.ContractInstance(ctx, addr)
	if err != nil {
		return nil, err
	}
	// there is not contract registered yet, we just have a simple address
	if codeInfo == nil || contractInfo == nil {
		return nil, nil
	}
	cdep, err := k.GetContractDependencyInner(ctx, addr, *contractInfo, *codeInfo, prefixStoreKey)
	if err != nil {
		return nil, err
	}
	return &cdep, nil
}

func (k *Keeper) GetContractDependencyInner(ctx sdk.Context, addr mcodec.AccAddressPrefixed, contractInfo types.ContractInfo, codeInfo types.CodeInfo, prefixStoreKey []byte) (types.ContractDependency, error) {
	cdep, err := k.GetContractDependencyInnerWithoutRoles(ctx, addr, contractInfo, codeInfo, prefixStoreKey)
	if err != nil {
		return types.ContractDependency{}, err
	}

	// shortcut for roles contract, to avoid cycle
	if k.GetRoleContractAddress(ctx).String() == addr.String() {
		cdep.Role = types.ROLE_ROLES
		cdep.Label = contractInfo.Label
		return cdep, nil
	}
	if k.GetCodeRegistryAddress(ctx).String() == addr.String() {
		cdep.Role = types.ROLE_STORAGE_CONTRACTS
		cdep.Label = contractInfo.Label
		return cdep, nil
	}

	label := k.GetRoleLabelByContract(ctx, addr)
	role := k.GetRoleByLabel(ctx, label)
	rolename := ""
	if role != nil {
		rolename = role.Role
	}

	cdep.Role = rolename
	cdep.Label = label
	return cdep, nil
}

func (k *Keeper) GetContractDependencyInnerWithoutRoles(ctx sdk.Context, addr mcodec.AccAddressPrefixed, contractInfo types.ContractInfo, codeInfo types.CodeInfo, prefixStoreKey []byte) (types.ContractDependency, error) {
	var sdeps = k.SystemDepsFromCodeDeps(ctx, codeInfo.Deps)
	aotFilePath := ""
	if codeInfo.Pinned {
		aotFilePath = k.wasmvm.BuildPathPinned(codeInfo.CodeHash)
	}
	filepath := k.wasmvm.GetCodeFilePath(codeInfo)

	cdep := types.ContractDependency{
		Address:       addr,
		StoreKey:      prefixStoreKey,
		CodeFilePath:  filepath,
		AotFilePath:   aotFilePath,
		SystemDeps:    sdeps,
		Bytecode:      codeInfo.InterpretedBytecodeRuntime,
		CodeHash:      codeInfo.CodeHash,
		CodeId:        contractInfo.CodeId,
		SystemDepsRaw: codeInfo.Deps,
		StorageType:   contractInfo.StorageType,
		Pinned:        codeInfo.Pinned,
		MeteringOff:   codeInfo.MeteringOff,
	}
	return cdep, nil
}

func (k *Keeper) create(ctx sdk.Context, creator *mcodec.AccAddressPrefixed, wasmCode []byte, deps []string, metadata types.CodeMetadata, pinned bool, meteringOff bool) (codeID uint64, checksum []byte, err error) {
	if creator == nil {
		return 0, checksum, sdkerr.Wrap(sdkerrors.ErrInvalidAddress, "cannot be nil")
	}
	if ioutils.IsGzip(wasmCode) {
		ctx.GasMeter().ConsumeGas(k.gasRegister.UncompressCosts(len(wasmCode)), "Uncompress gzip bytecode")
		wasmCode, err = ioutils.Uncompress(wasmCode, uint64(types.MaxWasmSize))
		if err != nil {
			return 0, checksum, sdkerr.Wrap(types.ErrCreateFailed, err.Error())
		}
	}
	codeInfo, err := k.createCodeInfo(ctx, *creator, wasmCode, deps, metadata, pinned, meteringOff)
	if err != nil {
		return 0, checksum, err
	}
	codeID, err = k.storeNewCodeInfo(ctx, codeInfo)
	if err != nil {
		return 0, checksum, err
	}

	k.Logger(ctx).Debug("storing new contract", "deps", codeInfo.Deps, "code_id", codeID, "checksum", hex.EncodeToString(checksum))

	evt := sdk.NewEvent(
		types.EventTypeStoreCode,
		sdk.NewAttribute(types.AttributeKeyChecksum, hex.EncodeToString(checksum)),
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeID, 10)), // last element to be compatible with scripts
	)
	for _, d := range codeInfo.Deps {
		evt.AppendAttributes(sdk.NewAttribute(types.AttributeKeyRequiredCapability, d))
	}
	ctx.EventManager().EmitEvent(evt)

	k.Logger(ctx).Debug("created new contract", "code_id", codeID)

	return codeID, checksum, nil
}

func (k *Keeper) createCodeInfo(ctx sdk.Context, creator mcodec.AccAddressPrefixed, wasmCode []byte, deps []string, metadata types.CodeMetadata, pinned bool, meteringOff bool) (codeInfo types.CodeInfo, err error) {
	var checksum []byte
	var reportDeps = make([]string, 0)

	if ioutils.IsWasm(wasmCode) {
		checksum, reportDeps, err = k.createWasm(ctx, wasmCode)
	} else {
		if len(deps) > 0 && types.HasUtf8Dep(deps) {
			checksum, reportDeps, err = k.createSourceInterpreted(ctx, wasmCode, deps)
		} else if types.HasInterpreterDep(deps) {
			checksum = k.wasmvm.checksum(wasmCode)
		}
	}
	if err != nil {
		return types.CodeInfo{}, err
	}
	reportDeps = uniqueStrings(append(reportDeps, deps...))

	if len(checksum) == 0 {
		return types.CodeInfo{}, sdkerr.Wrap(types.ErrCreateFailed, "this is not wasm code, use deploy")
	}

	// TODO cache current supported deps (including interpreters) and verify these here
	// if !types.IsWasmDeps(deps) {
	// 	return 0, checksum, sdkerr.Wrap(types.ErrCreateFailed, "incorrect deps")
	// }

	codeInfo = types.NewCodeInfo(checksum, creator.String(), reportDeps, metadata, pinned, meteringOff)
	if types.HasInterpreterDep(deps) && !types.HasUtf8Dep(deps) {
		// TODO only store one
		codeInfo.InterpretedBytecodeDeployment = wasmCode
		codeInfo.InterpretedBytecodeRuntime = wasmCode
	}

	return codeInfo, nil
}

func (k *Keeper) createWasm(ctx sdk.Context, wasmCode []byte) (checksum []byte, reportDeps []string, err error) {
	ctx.GasMeter().ConsumeGas(k.gasRegister.CompileCosts(len(wasmCode)), "Compiling wasm bytecode")
	report, err := k.wasmvm.AnalyzeWasm(ctx, wasmCode)
	if err != nil {
		return checksum, nil, sdkerr.Wrap(types.ErrCreateFailed, err.Error())
	}

	checksum, err = k.wasmvm.Create(wasmCode)
	if err != nil {
		return checksum, report.Dependencies, sdkerr.Wrap(types.ErrCreateFailed, err.Error())
	}
	return checksum, report.Dependencies, nil
}

func (k *Keeper) createSourceInterpreted(_ sdk.Context, sourceCode []byte, deps []string) (checksum []byte, reportDeps []string, err error) {
	// TODO actually run the source code in the compiler
	// and verify that it is valid
	// maybe store the compiled bytecode

	// maybe check deps for supported interpreters (this can even be a contract address)

	extension := GetExtensionFromDeps(deps)
	checksum, err = k.wasmvm.CreateUtf8(sourceCode, extension)
	if err != nil {
		return checksum, deps, sdkerr.Wrap(types.ErrCreateFailed, err.Error())
	}

	return checksum, deps, nil
}

// this is for bytecode interpreters (e.g. for EVM)
func (k *Keeper) CreateInterpreted(
	ctx sdk.Context,
	creator *mcodec.AccAddressPrefixed,
	provenance *mcodec.AccAddressPrefixed,
	wasmCode []byte,
	deps []string,
	metadata types.CodeMetadata,
	initMsg types.RawContractMessage,
	deposit sdk.Coins,
	label string,
	salt []byte,
) (codeID uint64, checksum []byte, contractAddress mcodec.AccAddressPrefixed, err error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "createInterpreted")

	if creator == nil {
		return 0, nil, mcodec.AccAddressPrefixed{}, sdkerr.Wrap(sdkerrors.ErrInvalidAddress, "cannot be nil")
	}

	if ioutils.IsWasm(wasmCode) {
		return 0, nil, mcodec.AccAddressPrefixed{}, sdkerr.Wrap(types.ErrCreateFailed, "this is wasm code, use store")
	}

	// TODO cache current supported deps (including interpreters) and verify these here
	// if types.IsWasmDeps(deps) {
	// 	return 0, nil, nil, sdkerr.Wrap(types.ErrCreateFailed, "incorrect deps")
	// }

	checksum = k.wasmvm.checksum(wasmCode)
	codeID, err = k.GetNextCodeId(ctx)
	if err != nil {
		return 0, nil, mcodec.AccAddressPrefixed{}, err
	}
	k.Logger(ctx).Debug("storing new contract", "deps", deps, "code_id", codeID, "checksum", checksum)
	codeInfo := types.NewCodeInfo(checksum, creator.String(), deps, metadata, false, false)
	codeInfo.InterpretedBytecodeDeployment = wasmCode

	addressParent := provenance
	if addressParent == nil {
		addressParent = creator
	}
	if len(salt) == 0 {
		contractAddress = k.EwasmClassicAddressGenerator(*addressParent)(ctx, codeID, codeInfo.CodeHash)
	} else {
		contractAddress = k.EwasmPredictableAddressGenerator(*addressParent, salt, []byte{}, false)(ctx, codeID, codeInfo.CodeHash)
	}
	if provenance == nil {
		provenance = &mcodec.AccAddressPrefixed{}
	}

	_, runtimeCode, err := k.instantiateInternal(ctx, codeID, *creator, *provenance, types.ContractStorageType_CoreConsensus, initMsg, deposit, contractAddress, &codeInfo, label)
	if err != nil {
		return 0, checksum, contractAddress, sdkerr.Wrap(types.ErrCreateFailed, err.Error())
	}
	codeInfo.InterpretedBytecodeRuntime = runtimeCode
	// TODO the hash algo will depend on deps
	codeInfo.RuntimeHash = k.wasmvm.checksum(runtimeCode)
	newCodeId, err := k.storeNewCodeInfo(ctx, codeInfo)
	if err != nil {
		return 0, checksum, contractAddress, sdkerr.Wrap(types.ErrCreateFailed, err.Error())
	}
	if newCodeId != codeID {
		return 0, checksum, contractAddress, sdkerr.Wrap(types.ErrCreateFailed, "unexpected code id")
	}

	evt := sdk.NewEvent(
		types.EventTypeStoreCode,
		sdk.NewAttribute(types.AttributeKeyChecksum, hex.EncodeToString(checksum)),
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeID, 10)), // last element to be compatible with scripts
	)
	for _, d := range deps {
		evt.AppendAttributes(sdk.NewAttribute(types.AttributeKeyRequiredCapability, d))
	}
	ctx.EventManager().EmitEvent(evt)

	evt2 := sdk.NewEvent(
		types.EventTypeDeploy,
		sdk.NewAttribute(types.AttributeKeyContractAddrCreated, contractAddress.String()),
	)
	ctx.EventManager().EmitEvent(evt2)

	return codeID, checksum, contractAddress, nil
}

// func (k *Keeper) importCode(ctx sdk.Context, codeID uint64, codeInfo types.CodeInfo, wasmCode []byte) error {
// 	if ioutils.IsGzip(wasmCode) {
// 		var err error
// 		wasmCode, err = ioutils.Uncompress(wasmCode, uint64(types.MaxWasmSize))
// 		if err != nil {
// 			return sdkerr.Wrap(types.ErrCreateFailed, err.Error())
// 		}
// 	}
// 	newCodeHash, err := k.wasmvm.Create(wasmCode)
// 	if err != nil {
// 		return sdkerr.Wrap(types.ErrCreateFailed, err.Error())
// 	}
// 	if !bytes.Equal(codeInfo.CodeHash, newCodeHash) {
// 		return sdkerr.Wrap(types.ErrInvalid, "code hashes not same")
// 	}

// 	store := ctx.KVStore(k.storeKey)
// 	key := types.GetCodeKey(codeID)
// 	if store.Has(key) {
// 		return sdkerr.Wrapf(types.ErrDuplicate, "duplicate code: %d", codeID)
// 	}
// 	// 0x01 | codeID (uint64) -> ContractInfo
// 	store.Set(key, k.cdc.MustMarshal(&codeInfo))
// 	return nil
// }

// must be used only internally
func (k *Keeper) instantiateWithAddress(
	ctx sdk.Context,
	codeID uint64,
	creator mcodec.AccAddressPrefixed,
	contractAddress mcodec.AccAddressPrefixed,
	storageType types.ContractStorageType,
	initMsg []byte,
	deposit sdk.Coins,
	label string,
	role string,
	codeInfo types.CodeInfo,
) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "instantiate_with_address")

	if err := RequireNotSystemContract(contractAddress.Bytes(), codeInfo.Deps); err != nil {
		return nil, err
	}

	// res, gasUsed, err := k.ExecuteContractInstantiationInternal(ctx, codeID, &codeInfo, creator, contractAddress, storageType, initMsg, deposit, label)

	info := types.NewInfo(creator, creator, deposit)
	env, err := types.NewEnv(k.accBech32Codec, ctx, k.denom, contractAddress, codeInfo.CodeHash, codeInfo.InterpretedBytecodeDeployment, codeInfo.Deps, info)
	if err != nil {
		return nil, err
	}
	prefixStoreKey := types.GetContractStorePrefix(contractAddress.Bytes())
	prefixStore := k.ContractStore(ctx, storageType, prefixStoreKey)

	contractInfo, err := k.GetContractDependencyInnerWithoutRoles(ctx, contractAddress, types.ContractInfo{CodeId: codeID, StorageType: storageType}, codeInfo, prefixStoreKey)
	if err != nil {
		return nil, err
	}
	contractInfo.Role = role
	contractInfo.Label = label

	// prepare querier
	handler := k.newCosmosHandler(ctx, contractAddress)

	// instantiate wasm contract
	res, gasUsed, err := k.wasmvm.Instantiate(ctx, &codeInfo, env, initMsg, prefixStore, handler, k.gasMeter(ctx), contractInfo)

	k.consumeRuntimeGas(ctx, gasUsed)
	if err != nil {
		return nil, sdkerr.Wrap(types.ErrInstantiateFailed, err.Error())
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeInstantiate,
		sdk.NewAttribute(types.AttributeKeyContractAddrCreated, contractAddress.String()),
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeID, 10)),
	))

	err = k.handleResponseEvents(ctx, contractAddress.String(), "", res.Attributes, res.Events)
	if err != nil {
		return nil, sdkerr.Wrap(err, "instantiate dispatch")
	}
	return res.Data, err
}

func (k *Keeper) instantiate(
	ctx sdk.Context,
	codeID uint64,
	creator mcodec.AccAddressPrefixed,
	initMsg []byte,
	deposit sdk.Coins,
	label string,
) (mcodec.AccAddressPrefixed, []byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "instantiate")

	// get contact info
	codeInfo, err := k.GetCodeInfo(ctx, codeID)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, nil, sdkerr.Wrapf(err, "instantiate: cannot get code info")
	}
	if codeInfo == nil {
		return mcodec.AccAddressPrefixed{}, nil, sdkerr.Wrap(types.ErrNotFound, "code")
	}
	// TODO deps: support multiple types of address generation
	contractAddress := k.EwasmClassicAddressGenerator(creator)(ctx, codeID, codeInfo.CodeHash)
	return k.instantiateInternal(ctx, codeID, creator, mcodec.AccAddressPrefixed{}, types.ContractStorageType_CoreConsensus, initMsg, deposit, contractAddress, codeInfo, label)
}

func (k *Keeper) instantiate2(
	ctx sdk.Context,
	codeID uint64,
	creator mcodec.AccAddressPrefixed,
	initMsg []byte,
	deposit sdk.Coins,
	salt []byte,
	fixMsg bool,
	label string,
) (mcodec.AccAddressPrefixed, []byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "instantiate")

	// get contact info
	codeInfo, err := k.GetCodeInfo(ctx, codeID)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, nil, sdkerr.Wrapf(err, "instantiate2: cannot get code info")
	}
	if codeInfo == nil {
		return mcodec.AccAddressPrefixed{}, nil, sdkerr.Wrap(types.ErrNotFound, "code")
	}
	// TODO if we support multiple types of address generation
	// the type should be saved in CodeInfo
	contractAddress := k.EwasmPredictableAddressGenerator(creator, salt, initMsg, fixMsg)(ctx, codeID, codeInfo.CodeHash)
	return k.instantiateInternal(ctx, codeID, creator, mcodec.AccAddressPrefixed{}, types.ContractStorageType_CoreConsensus, initMsg, deposit, contractAddress, codeInfo, label)
}

func (k *Keeper) instantiateInternal(
	ctx sdk.Context,
	codeID uint64,
	creator mcodec.AccAddressPrefixed,
	provenance mcodec.AccAddressPrefixed,
	storageType types.ContractStorageType,
	initMsg []byte,
	deposit sdk.Coins,
	contractAddress mcodec.AccAddressPrefixed,
	codeInfo *types.CodeInfo,
	label string,
) (mcodec.AccAddressPrefixed, []byte, error) {
	if err := RequireNotSystemContract(contractAddress.Bytes(), codeInfo.Deps); err != nil {
		return mcodec.AccAddressPrefixed{}, nil, err
	}
	instanceCosts := k.gasRegister.NewContractInstanceCosts(k.IsPinnedCode(ctx, codeID), len(initMsg))
	ctx.GasMeter().ConsumeGas(instanceCosts, "Loading wasm module: instantiate")

	if k.HasContractInfo(ctx, contractAddress) {
		return mcodec.AccAddressPrefixed{}, nil, types.ErrDuplicate.Wrap("instance with this contract address already exists")
	}

	// deposit initial contract funds
	if !deposit.IsZero() {
		if err := k.GetBankKeeper().TransferCoins(ctx, creator, contractAddress, deposit); err != nil {
			return mcodec.AccAddressPrefixed{}, nil, err
		}
	}
	res, gasUsed, err := k.ExecuteContractInstantiationInternal(ctx, codeID, codeInfo, creator, contractAddress, storageType, initMsg, deposit, label)
	k.consumeRuntimeGas(ctx, gasUsed)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, nil, sdkerr.Wrap(types.ErrInstantiateFailed, err.Error())
	}

	// persist instance first
	contractInfo := types.NewContractInfo(codeID, creator.String(), provenance.String(), initMsg, label)
	contractInfo.StorageType = storageType

	// check for IBC flag - TODO use codeInfo.Dependencies
	// report, err := k.wasmvm.AnalyzeWasm(codeInfo.CodeHash)
	// if err != nil {
	// 	return nil, nil, sdkerr.Wrap(types.ErrInstantiateFailed, err.Error())
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
	err = k.storeContractInfo(ctx, contractAddress, contractInfo)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, nil, sdkerr.Wrap(err, "instantiate store contract info")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeInstantiate,
		sdk.NewAttribute(types.AttributeKeyContractAddrCreated, contractAddress.String()),
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeID, 10)),
	))

	err = k.handleResponseEvents(ctx, contractAddress.String(), contractInfo.IbcPortId, res.Attributes, res.Events)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, nil, sdkerr.Wrap(err, "instantiate dispatch")
	}

	// we run this logic after instantiation, to be able to initiate the auth contract at genesis
	err = k.instantiateNewContractAccount(ctx, contractAddress)
	if err != nil {
		return mcodec.AccAddressPrefixed{}, nil, sdkerr.Wrap(err, "instantiate create account")
	}

	k.Logger(ctx).Debug("instantiated new contract", "address", contractAddress.String())

	return contractAddress, res.Data, nil
}

func (k *Keeper) instantiateNewContractAccount(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed) error {
	// check account
	// every cosmos module can define custom account types when needed. The cosmos-sdk comes with extension points
	// to support this and a set of base and vesting account types that we integrated in our default lists.
	// But not all account types of other modules are known or may make sense for contracts, therefore we kept this
	// decision logic also very flexible and extendable. We provide new options to overwrite the default settings via WithAcceptedAccountTypesOnContractInstantiation and
	// WithPruneAccountTypesOnContractInstantiation as constructor arguments
	existingAcct, err := k.GetAccountKeeper().GetAccountPrefixed(ctx, contractAddress)
	if err != nil && existingAcct != nil {
		if existingAcct.GetSequence() != 0 || existingAcct.GetPubKey() != nil {
			return types.ErrAccountExists.Wrap("address is claimed by external account")
		}

		// consider an account in the wasmx namespace spam and overwrite it.
		k.Logger(ctx).Info("pruning existing account for contract instantiation", "address", contractAddress.String())
		// this also stores the account
		k.GetAccountKeeper().NewAccountWithAddressPrefixed(ctx, contractAddress)
	} else {
		if err != nil {
			k.Logger(ctx).Debug("account not found, creating a new one", "error", err.Error())
		}
		// create an empty account (so we don't have issues later)
		// this also stores the account
		_, err = k.GetAccountKeeper().NewAccountWithAddressPrefixed(ctx, contractAddress)
		if err != nil {
			panic(fmt.Sprintf("account not created: %s: %s", contractAddress.String(), err.Error()))
		}
	}
	return nil
}

func (k *Keeper) ExecuteContractInstantiationInternal(
	ctx sdk.Context,
	codeID uint64,
	codeInfo *types.CodeInfo,
	creator mcodec.AccAddressPrefixed,
	contractAddress mcodec.AccAddressPrefixed,
	storageType types.ContractStorageType,
	initMsg []byte,
	deposit sdk.Coins,
	label string,
) (types.ContractResponse, uint64, error) {
	// prepare params for contract instantiate call
	info := types.NewInfo(creator, creator, deposit)
	env, err := types.NewEnv(k.accBech32Codec, ctx, k.denom, contractAddress, codeInfo.CodeHash, codeInfo.InterpretedBytecodeDeployment, codeInfo.Deps, info)
	if err != nil {
		return types.ContractResponse{}, 0, err
	}

	// create prefixed data store
	// 0x03 | BuildContractAddressClassic (sdk.AccAddress)
	prefixStoreKey := types.GetContractStorePrefix(contractAddress.Bytes())
	prefixStore := k.ContractStore(ctx, storageType, prefixStoreKey)

	contractInfo, err := k.GetContractDependencyInner(ctx, contractAddress, types.ContractInfo{CodeId: codeID, StorageType: storageType}, *codeInfo, prefixStoreKey)
	if err != nil {
		return types.ContractResponse{}, 0, err
	}

	// prepare querier
	handler := k.newCosmosHandler(ctx, contractAddress)

	// instantiate wasm contract
	return k.wasmvm.Instantiate(ctx, codeInfo, env, initMsg, prefixStore, handler, k.gasMeter(ctx), contractInfo)
}

// PinCode pins the wasm contract in wasmvm cache
func (k *Keeper) pinCode(ctx sdk.Context, codeHash []byte, compiledFolderPath string, meteringOff bool) error {
	if err := k.wasmvm.Pin(ctx, codeHash, compiledFolderPath, meteringOff); err != nil {
		return sdkerr.Wrap(types.ErrPinContractFailed, err.Error())
	}
	k.Logger(ctx).Info("contract is AOT compiled", "code_hash", hex.EncodeToString(codeHash))
	return nil
}

// PinCode pins the wasm contract in wasmvm cache
func (k *Keeper) pinCodeAndStore(ctx sdk.Context, codeInfo *types.CodeInfo, compiledFolderPath string, meteringOff bool) error {
	codeInfo.Pinned = true
	codeId, err := k.storeNewCodeInfo(ctx, *codeInfo)
	if err != nil {
		return sdkerr.Wrap(types.ErrCreateFailed, err.Error())
	}
	return k.pinCodeWithEvent(ctx, codeId, codeInfo.CodeHash, compiledFolderPath, meteringOff)
}

func (k *Keeper) pinCodeWithEvent(ctx sdk.Context, codeId uint64, codeHash []byte, compiledFolderPath string, meteringOff bool) error {
	err := k.pinCode(ctx, codeHash, compiledFolderPath, meteringOff)
	if err != nil {
		return sdkerr.Wrap(types.ErrPinContractFailed, err.Error())
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePinCode,
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeId, 10)),
		sdk.NewAttribute(types.AttributeKeyChecksum, hex.EncodeToString(codeHash)),
	))
	return nil
}

// UnpinCode removes the wasm contract from wasmvm cache
func (k *Keeper) unpinCode(ctx sdk.Context, codeId uint64) error {
	codeInfo, err := k.GetCodeInfo(ctx, codeId)
	if err != nil {
		return err
	}
	if codeInfo == nil {
		return sdkerr.Wrap(types.ErrNotFound, "code info")
	}
	if err := k.wasmvm.Unpin(codeInfo.CodeHash); err != nil {
		return sdkerr.Wrap(types.ErrUnpinContractFailed, err.Error())
	}

	codeInfo.Pinned = false
	err = k.storeCodeInfo(ctx, codeId, *codeInfo)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info("contract AOT compiled module removed ", "codeId", codeId, "code_hash", hex.EncodeToString(codeInfo.CodeHash))

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUnpinCode,
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeId, 10)),
	))
	return nil
}

// Execute executes the contract instance
func (k *Keeper) execute(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed, caller mcodec.AccAddressPrefixed, msg []byte, coins sdk.Coins, dependencies []string, inBackground bool) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasmx", "contract", "execute")
	contractInfo, codeInfo, prefixStoreKey, err := k.ContractInstance(ctx, contractAddress)
	if err != nil || contractInfo == nil || codeInfo == nil {
		// This can be just an ethcall sending value
		// we do not fail, to maintain compatibility with EVM
		if coins.IsZero() {
			return nil, nil
		}
		err := k.GetBankKeeper().SendCoinsPrefixed(ctx, caller, contractAddress, coins)
		return nil, err
	}

	if err := RequireNotSystemContract(contractAddress.Bytes(), codeInfo.Deps); err != nil {
		return nil, err
	}

	// add known dependencies for that codeId
	// TODO system deps in the form of smart contracts
	// e.g. dep = {value, type}
	// if we cannot just load all the modules in the same VM
	allDeps := append(codeInfo.Deps, dependencies...)
	contractDeps, err := k.ContractDepsFromCodeDeps(ctx, allDeps)
	if err != nil {
		return nil, err
	}

	// TODO panic if coin is not the correct denomination
	// add denom param for wasmx

	executeCosts := k.gasRegister.InstantiateContractCosts(k.IsPinnedCode(ctx, contractInfo.CodeId), len(msg))
	ctx.GasMeter().ConsumeGas(executeCosts, "WasmX module execution: execute")

	// add more funds
	if !coins.IsZero() {
		if err := k.GetBankKeeper().TransferCoins(ctx, caller, contractAddress, coins); err != nil {
			return nil, err
		}
	}
	// TODO execute with origin
	info := types.NewInfo(caller, caller, coins)
	env, err := types.NewEnv(k.accBech32Codec, ctx, k.denom, contractAddress, codeInfo.CodeHash, codeInfo.InterpretedBytecodeRuntime, codeInfo.Deps, info)
	if err != nil {
		return nil, err
	}

	extendedContractInfo, err := k.GetContractDependencyInner(ctx, contractAddress, *contractInfo, *codeInfo, prefixStoreKey)
	if err != nil {
		return nil, err
	}

	// prepare querier
	handler := k.newCosmosHandler(ctx, contractAddress)
	store := k.ContractStore(ctx, contractInfo.StorageType, prefixStoreKey)
	res, gasUsed, execErr := k.wasmvm.Execute(ctx, codeInfo, env, msg, store, handler, k.gasMeter(ctx), extendedContractInfo, contractDeps, inBackground)
	k.consumeRuntimeGas(ctx, gasUsed)

	if execErr != nil {
		return nil, sdkerr.Wrap(types.ErrExecuteFailed, execErr.Error())
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeExecute,
		sdk.NewAttribute(types.AttributeKeyContractAddr, contractAddress.String()),
	))

	err = k.handleResponseEvents(ctx, contractAddress.String(), contractInfo.IbcPortId, res.Attributes, res.Events)
	if err != nil {
		return nil, sdkerr.Wrap(err, "dispatch events")
	}

	data, err := k.handleResponseMessages(ctx, contractAddress, contractInfo.IbcPortId, res.Messages, res.Data)
	if err != nil {
		return nil, sdkerr.Wrap(err, "dispatch message")
	}
	// TODO return data or res.Data
	return data, nil
}

// Execute executes the contract instance
func (k *Keeper) ExecuteEntryPoint(ctx sdk.Context, contractEntryPoint string, contractAddress mcodec.AccAddressPrefixed, caller mcodec.AccAddressPrefixed, msg []byte, dependencies []string, inBackground bool) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasmx", "contract", "ExecuteEntryPoint")
	contractInfo, codeInfo, prefixStoreKey, err := k.ContractInstance(ctx, contractAddress)
	if err != nil {
		return nil, err
	}
	if contractInfo == nil || codeInfo == nil {
		return nil, fmt.Errorf("contract not found %s", contractAddress.String())
	}

	if err := RequireNotSystemContract(contractAddress.Bytes(), codeInfo.Deps); err != nil {
		return nil, err
	}

	// add known dependencies for that codeId
	// TODO system deps in the form of smart contracts
	// e.g. dep = {value, type}
	// if we cannot just load all the modules in the same VM
	allDeps := append(codeInfo.Deps, dependencies...)
	contractDeps, err := k.ContractDepsFromCodeDeps(ctx, allDeps)
	if err != nil {
		return nil, err
	}

	// TODO panic if coin is not the correct denomination
	// add denom param for wasmx

	executeCosts := k.gasRegister.InstantiateContractCosts(k.IsPinnedCode(ctx, contractInfo.CodeId), len(msg))
	ctx.GasMeter().ConsumeGas(executeCosts, "WasmX module execution: execute entry point")

	info := types.NewInfo(caller, caller, nil)
	env, err := types.NewEnv(k.accBech32Codec, ctx, k.denom, contractAddress, codeInfo.CodeHash, codeInfo.InterpretedBytecodeRuntime, codeInfo.Deps, info)
	if err != nil {
		return nil, err
	}

	extendedContractInfo, err := k.GetContractDependencyInner(ctx, contractAddress, *contractInfo, *codeInfo, prefixStoreKey)
	if err != nil {
		return nil, err
	}

	// prepare querier
	handler := k.newCosmosHandler(ctx, contractAddress)
	store := k.ContractStore(ctx, contractInfo.StorageType, prefixStoreKey)
	res, gasUsed, execErr := k.wasmvm.ExecuteEntryPoint(ctx, contractEntryPoint, codeInfo, env, msg, store, handler, k.gasMeter(ctx), extendedContractInfo, contractDeps, inBackground)
	k.consumeRuntimeGas(ctx, gasUsed)

	if execErr != nil {
		return nil, sdkerr.Wrap(types.ErrExecuteFailed, execErr.Error())
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeExecute,
		sdk.NewAttribute(types.AttributeKeyContractAddr, contractAddress.String()),
	))

	err = k.handleResponseEvents(ctx, contractAddress.String(), contractInfo.IbcPortId, res.Attributes, res.Events)
	if err != nil {
		return nil, sdkerr.Wrap(err, "dispatch events")
	}

	data, err := k.handleResponseMessages(ctx, contractAddress, contractInfo.IbcPortId, res.Messages, res.Data)
	if err != nil {
		return nil, sdkerr.Wrap(err, "dispatch message")
	}
	// TODO return data or res.Data
	return data, nil
}

// For CosmWasm compatibility
// reply is only called from keeper internal functions (dispatchSubmessages) after processing the submessage
func (k *Keeper) Reply(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed, reply cw8types.Reply) ([]byte, error) {
	contractInfo, codeInfo, prefixStoreKey, err := k.ContractInstance(ctx, contractAddress)
	if err != nil {
		return nil, err
	}
	if contractInfo == nil || codeInfo == nil {
		return nil, fmt.Errorf("contract not found %s", contractAddress.String())
	}

	// TODO
	// for CosmWasm compat
	// always consider this pinned
	// replyCosts := k.gasRegister.ReplyCosts(true, reply)
	// ctx.GasMeter().ConsumeGas(replyCosts, "Loading WasmX module: reply")

	env, err := types.NewEnv(k.accBech32Codec, ctx, k.denom, contractAddress, codeInfo.CodeHash, codeInfo.InterpretedBytecodeDeployment, codeInfo.Deps, types.MessageInfo{})
	if err != nil {
		return nil, err
	}

	// prepare querier
	handler := k.newCosmosHandler(ctx, contractAddress)
	// gas := k.runtimeGasForContract(ctx)

	replyBz, err := json.Marshal(reply)
	if err != nil {
		return nil, sdkerr.Wrap(err, "marshal reply failed")
	}

	extendedContractInfo, err := k.GetContractDependencyInner(ctx, contractAddress, *contractInfo, *codeInfo, prefixStoreKey)
	if err != nil {
		return nil, err
	}
	store := k.ContractStore(ctx, contractInfo.StorageType, prefixStoreKey)

	// TODO costJSONDeserialization
	res, gasUsed, execErr := k.wasmvm.Reply(ctx, codeInfo, env, replyBz, store, handler, k.gasMeter(ctx), extendedContractInfo, nil)
	k.consumeRuntimeGas(ctx, gasUsed)
	if execErr != nil {
		return nil, sdkerr.Wrap(types.ErrExecuteFailed, execErr.Error())
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		cw8types.EventTypeReply,
		sdk.NewAttribute(types.AttributeKeyContractAddr, contractAddress.String()),
	))

	err = k.handleResponseEvents(ctx, contractAddress.String(), contractInfo.IbcPortId, res.Attributes, res.Events)
	if err != nil {
		return nil, sdkerr.Wrap(err, "dispatch events")
	}

	data, err := k.handleResponseMessages(ctx, contractAddress, contractInfo.IbcPortId, res.Messages, res.Data)
	if err != nil {
		return nil, sdkerr.Wrap(err, "dispatch message")
	}

	return data, nil
}

// executeWithOrigin executes the contract instance
func (k *Keeper) executeWithOrigin(ctx sdk.Context, origin mcodec.AccAddressPrefixed, contractAddress mcodec.AccAddressPrefixed, caller mcodec.AccAddressPrefixed, msg []byte, coins sdk.Coins) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "executeWithOrigin")

	// fail if caller is not a contract
	contractInfo, codeInfo, _, err := k.ContractInstance(ctx, caller)
	if err != nil || contractInfo == nil || codeInfo == nil {
		return nil, sdkerr.Wrap(types.ErrExecuteFailed, "cannot executeWithOrigin from EOA")
	}

	contractInfo, codeInfo, prefixStoreKey, err := k.ContractInstance(ctx, contractAddress)
	if err != nil {
		return nil, err
	}
	if err := RequireNotSystemContract(contractAddress.Bytes(), codeInfo.Deps); err != nil {
		return nil, err
	}

	executeCosts := k.gasRegister.InstantiateContractCosts(k.IsPinnedCode(ctx, contractInfo.CodeId), len(msg))
	ctx.GasMeter().ConsumeGas(executeCosts, "WasmX module execution: executeWithOrigin")

	// add more funds
	if !coins.IsZero() {
		if err := k.GetBankKeeper().TransferCoins(ctx, caller, contractAddress, coins); err != nil {
			return nil, err
		}
	}

	info := types.NewInfo(origin, caller, coins)
	env, err := types.NewEnv(k.accBech32Codec, ctx, k.denom, contractAddress, codeInfo.CodeHash, codeInfo.InterpretedBytecodeRuntime, codeInfo.Deps, info)
	if err != nil {
		return nil, err
	}

	handler := k.newCosmosHandler(ctx, contractAddress)
	extendedContractInfo, err := k.GetContractDependencyInner(ctx, contractAddress, *contractInfo, *codeInfo, prefixStoreKey)
	if err != nil {
		return nil, err
	}
	store := k.ContractStore(ctx, contractInfo.StorageType, prefixStoreKey)

	res, gasUsed, execErr := k.wasmvm.Execute(ctx, codeInfo, env, msg, store, handler, k.gasMeter(ctx), extendedContractInfo, nil, false)
	k.consumeRuntimeGas(ctx, gasUsed)

	// res, _, execErr = k.handleExecutionRerun(ctx, codeInfo.CodeHash, env, info, msg, prefixStore, cosmwasmAPI, querier, gas, costJSONDeserialization, contractAddress, contractInfo, res, gasUsed, execErr, k.wasmVM.Execute)

	if execErr != nil {
		return nil, sdkerr.Wrap(types.ErrExecuteFailed, execErr.Error())
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeExecute,
		sdk.NewAttribute(types.AttributeKeyContractAddr, contractAddress.String()),
	))

	err = k.handleResponseEvents(ctx, contractAddress.String(), contractInfo.IbcPortId, res.Attributes, res.Events)
	if err != nil {
		return nil, sdkerr.Wrap(err, "dispatch")
	}

	return res.Data, nil
}

// Execute executes the contract instance
func (k *Keeper) query(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed, caller mcodec.AccAddressPrefixed, msg []byte, coins sdk.Coins, dependencies []string, isdebug bool) (*types.ContractResponse, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "query")
	contractInfo, codeInfo, prefixStoreKey, err := k.ContractInstance(ctx, contractAddress)
	if err != nil {
		return nil, err
	}
	if contractInfo == nil || codeInfo == nil {
		return nil, fmt.Errorf("contract not found %s", contractAddress.String())
	}
	if err := RequireNotSystemContract(contractAddress.Bytes(), codeInfo.Deps); err != nil {
		return nil, err
	}

	// TODO panic if coin is not the correct denomination
	// add denom param for wasmx

	executeCosts := k.gasRegister.InstantiateContractCosts(k.IsPinnedCode(ctx, contractInfo.CodeId), len(msg))
	ctx.GasMeter().ConsumeGas(executeCosts, "WasmX module execution: query")

	// add known dependencies for that codeId
	// TODO system deps in the form of smart contracts
	// e.g. dep = {value, type}
	// if we cannot just load all the modules in the same VM
	allDeps := append(codeInfo.Deps, dependencies...)
	contractDeps, err := k.ContractDepsFromCodeDeps(ctx, allDeps)
	if err != nil {
		return nil, err
	}

	// add more funds
	if !coins.IsZero() {
		if err := k.GetBankKeeper().TransferCoins(ctx, caller, contractAddress, coins); err != nil {
			return nil, err
		}
	}

	info := types.NewInfo(caller, caller, coins)
	env, err := types.NewEnv(k.accBech32Codec, ctx, k.denom, contractAddress, codeInfo.CodeHash, codeInfo.InterpretedBytecodeRuntime, codeInfo.Deps, info)
	if err != nil {
		return nil, err
	}

	extendedContractInfo, err := k.GetContractDependencyInner(ctx, contractAddress, *contractInfo, *codeInfo, prefixStoreKey)
	if err != nil {
		return nil, err
	}

	handler := k.newCosmosHandler(ctx, contractAddress)
	store := k.ContractStore(ctx, contractInfo.StorageType, prefixStoreKey)
	res, gasUsed, execErr := k.wasmvm.QueryExecute(ctx, codeInfo, env, msg, store, handler, k.gasMeter(ctx), extendedContractInfo, contractDeps, isdebug)
	k.consumeRuntimeGas(ctx, gasUsed)

	// res, _, execErr = k.handleExecutionRerun(ctx, codeInfo.CodeHash, env, info, msg, prefixStore, cosmwasmAPI, querier, gas, costJSONDeserialization, contractAddress, contractInfo, res, gasUsed, execErr, k.wasmVM.Execute)
	if execErr != nil {
		return nil, sdkerr.Wrap(types.ErrExecuteFailed, execErr.Error())
	}

	// data, err := k.handleContractResponse(ctx, contractAddress, contractInfo.IbcPortId, res.Attributes, res.Data, res.Events)
	// if err != nil {
	// 	return nil, sdkerr.Wrap(err, "dispatch")
	// }

	return &res, nil
}

// handleResponseEvents processes the contract response data by emitting events
func (k *Keeper) handleResponseEvents(
	ctx sdk.Context,
	contractAddrBech32 string,
	_ string, // ibcPort string,
	attrs []types.EventAttribute,
	evts types.Events,
) error {
	attributeGasCost := k.gasRegister.EventCosts(attrs, evts)
	ctx.GasMeter().ConsumeGas(attributeGasCost, "Custom contract event attributes")
	// emit all events from this contract itself
	// these are not used
	if len(attrs) != 0 {
		wasmEvents, err := newWasmModuleEvent(attrs, contractAddrBech32)
		if err != nil {
			return err
		}
		ctx.EventManager().EmitEvents(wasmEvents)
	}
	if len(evts) > 0 {
		customEvents, err := newCustomEvents(evts, contractAddrBech32)
		if err != nil {
			return err
		}
		ctx.EventManager().EmitEvents(customEvents)
	}
	return nil
}

func (k *Keeper) handleResponseMessages(
	ctx sdk.Context,
	contractAddr mcodec.AccAddressPrefixed,
	ibcPort string,
	msgs []cw8types.SubMsg,
	data []byte,
) ([]byte, error) {
	if len(msgs) == 0 {
		return data, nil
	}
	// TODO do we still need this?
	if k.wasmVMResponseHandler == nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "no wasmVMResponseHandler found")
	}
	return k.wasmVMResponseHandler.Handle(ctx, contractAddr, ibcPort, msgs, data)
}

// Calculate how much gas can we use for the wasmx execution
func (k *Keeper) runtimeGasForContract(ctx sdk.Context) uint64 {
	meter := ctx.GasMeter()
	if meter.IsOutOfGas() {
		return 0
	}
	if meter.Limit() == 0 { // infinite gas meter with limit=0 and not out of gas
		return math.MaxUint64
	}
	return k.gasRegister.ToWasmVMGas(meter.Limit() - meter.GasConsumedToLimit())
}

func (k *Keeper) consumeRuntimeGas(ctx sdk.Context, gas uint64) {
	consumed := k.gasRegister.FromWasmVMGas(gas)
	ctx.GasMeter().ConsumeGas(consumed, "wasm contract")
	// throw OutOfGas error if we ran out (got exactly to zero due to better limit enforcing)
	if ctx.GasMeter().IsOutOfGas() {
		panic(storetypes.ErrorOutOfGas{Descriptor: "WasmX function execution"})
	}
}

func (k *Keeper) ContractDepsFromCodeDeps(ctx sdk.Context, allDeps []string) ([]types.ContractDependency, error) {
	var contractDeps []types.ContractDependency
	for _, hexaddr := range allDeps {
		if hexaddr[0:2] != "0x" {
			continue
		}
		hexaddr, role := types.ParseDep(hexaddr)
		addr := types.AccAddressFromHex(hexaddr)
		addrPrefixed := k.accBech32Codec.BytesToAccAddressPrefixed(addr)
		contractDep, err := k.GetContractDependency(ctx, addrPrefixed)
		if err != nil {
			return nil, err
		}
		if contractDep == nil {
			return nil, fmt.Errorf("contract dependency not found %s", addrPrefixed.String())
		}
		contractDep.Role = role
		contractDeps = append(contractDeps, *contractDep)
	}
	return contractDeps, nil
}

func (k *Keeper) SystemDepsFromCodeDeps(ctx sdk.Context, depLabels []string) []types.SystemDep {
	var sdeps []types.SystemDep
	for _, dep := range depLabels {
		if dep[0:2] != "0x" {
			sdep := types.SystemDep{
				Label: dep,
				Role:  dep,
			}
			// supported host interfaces are not roles for contracts
			if _, ok := types.SUPPORTED_HOST_INTERFACES[dep]; !ok {
				_sdep, err := k.SystemDepFromLabel(ctx, dep)
				if err == nil {
					sdep = _sdep
				}
			}
			sdeps = append(sdeps, sdep)
		}
	}
	return sdeps
}

func (k *Keeper) SystemDepFromLabel(ctx sdk.Context, label string) (types.SystemDep, error) {
	role := k.GetRoleByLabel(ctx, label)
	if role == nil {
		return types.SystemDep{}, fmt.Errorf("no role from label")
	}
	contractAddress, err := k.accBech32Codec.StringToAccAddressPrefixed(role.ContractAddress)
	if err != nil {
		return types.SystemDep{}, err
	}
	_, codeInfo, _, err := k.ContractInstance(ctx, contractAddress)
	if err != nil {
		return types.SystemDep{}, err
	}
	if codeInfo == nil {
		return types.SystemDep{}, fmt.Errorf("contract dependency not found %s", contractAddress.String())
	}
	aotFilePath := ""
	if codeInfo.Pinned {
		aotFilePath = k.wasmvm.BuildPathPinned(codeInfo.CodeHash)
	}
	filepath := k.wasmvm.GetCodeFilePath(*codeInfo)
	dep := types.SystemDep{
		Role:         role.Role,
		Label:        label,
		CodeFilePath: filepath,
		AotFilePath:  aotFilePath,
		Pinned:       codeInfo.Pinned,
		Deps:         k.SystemDepsFromCodeDeps(ctx, codeInfo.Deps),
	}
	return dep, nil
}

// MultipliedGasMeter wraps the GasMeter from context and multiplies all reads by out defined multiplier
type MultipliedGasMeter struct {
	originalMeter storetypes.GasMeter
	GasRegister   GasRegister
}

func NewMultipliedGasMeter(originalMeter storetypes.GasMeter, gr GasRegister) MultipliedGasMeter {
	return MultipliedGasMeter{originalMeter: originalMeter, GasRegister: gr}
}

var _ types.GasMeter = MultipliedGasMeter{}

// gas consumed in wasm VM units
func (m MultipliedGasMeter) GasConsumed() storetypes.Gas {
	return m.GasRegister.ToWasmVMGas(m.originalMeter.GasConsumed())
}

// consume gas in wasm VM units
func (m MultipliedGasMeter) ConsumeGas(gas storetypes.Gas, descriptor string) {
	descriptor = fmt.Sprintf("wasmx: %s", descriptor)
	sdkgas := m.GasRegister.FromWasmVMGas(gas)
	m.originalMeter.ConsumeGas(sdkgas, descriptor)
}

// gas limit in wasm VM units
func (m MultipliedGasMeter) GasLimit() storetypes.Gas {
	return m.GasRegister.ToWasmVMGas(m.originalMeter.Limit())
}

func (m MultipliedGasMeter) GasRemaining() storetypes.Gas {
	return m.GasRegister.ToWasmVMGas(m.originalMeter.GasRemaining())
}

func (k *Keeper) gasMeter(ctx sdk.Context) MultipliedGasMeter {
	return NewMultipliedGasMeter(ctx.GasMeter(), k.gasRegister)
}

func GetExtensionFromDeps(deps []string) string {
	extension := ""
	for _, dep := range deps {
		ext, found := types.FILE_EXTENSIONS[dep]
		if found {
			extension = ext
		}
	}
	return extension
}
