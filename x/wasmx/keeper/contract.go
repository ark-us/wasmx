package keeper

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmeth "wasmx/x/wasmx/ewasm"
	"wasmx/x/wasmx/ioutils"
	"wasmx/x/wasmx/types"
)

func (k Keeper) Create(ctx sdk.Context, creator sdk.AccAddress, wasmByteCode []byte) (uint64, []byte, error) {
	return k.create(ctx, creator, wasmByteCode)
}

func (k Keeper) Instantiate(ctx sdk.Context, codeId uint64, creator sdk.AccAddress, msg types.RawContractMessage, label string, funds sdk.Coins) (sdk.AccAddress, []byte, error) {
	return k.instantiate(ctx, codeId, creator, msg, label, funds)
}

func (k Keeper) Instantiate2(ctx sdk.Context, codeId uint64, senderAddr sdk.AccAddress, msg types.RawContractMessage, label string, funds sdk.Coins, salt []byte, fixMsg bool) (sdk.AccAddress, []byte, error) {
	return nil, nil, nil
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

func (k Keeper) create(ctx sdk.Context, creator sdk.AccAddress, wasmCode []byte) (codeID uint64, checksum []byte, err error) {
	if creator == nil {
		return 0, checksum, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "cannot be nil")
	}

	if ioutils.IsGzip(wasmCode) {
		// ctx.GasMeter().ConsumeGas(k.gasRegister.UncompressCosts(len(wasmCode)), "Uncompress gzip bytecode")
		wasmCode, err = ioutils.Uncompress(wasmCode, uint64(types.MaxWasmSize))
		if err != nil {
			return 0, checksum, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
		}
	}
	// ctx.GasMeter().ConsumeGas(k.gasRegister.CompileCosts(len(wasmCode)), "Compiling wasm bytecode")
	report, err := k.wasmvm.AnalyzeWasm(wasmCode)
	if err != nil {
		return 0, checksum, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	checksum, err = k.wasmvm.Create(wasmCode)
	if err != nil {
		return 0, checksum, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	codeID = k.autoIncrementID(ctx, types.KeyLastCodeID)
	k.Logger(ctx).Debug("storing new contract", "capabilities", report.Dependencies, "code_id", codeID)
	codeInfo := types.NewCodeInfo(checksum, creator, report.Dependencies)
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
	initMsg []byte,
	label string,
	deposit sdk.Coins,
	contractAddress sdk.AccAddress,
) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "instantiate_with_address")

	// get contact info
	codeInfo := k.GetCodeInfo(ctx, codeID)
	if codeInfo == nil {
		return nil, sdkerrors.Wrap(types.ErrNotFound, "code")
	}
	_, data, err := k.instantiateInternal(ctx, codeID, creator, initMsg, label, deposit, contractAddress, codeInfo)
	return data, err
}

func (k Keeper) instantiate(
	ctx sdk.Context,
	codeID uint64,
	creator sdk.AccAddress,
	initMsg []byte,
	label string,
	deposit sdk.Coins,
) (sdk.AccAddress, []byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "instantiate")

	// get contact info
	codeInfo := k.GetCodeInfo(ctx, codeID)
	if codeInfo == nil {
		return nil, nil, sdkerrors.Wrap(types.ErrNotFound, "code")
	}
	// TODO if we support multiple types of address generation
	// the type should be saved in CodeInfo
	contractAddress := k.EwasmClassicAddressGenerator(creator)(ctx, codeID, codeInfo.CodeHash)
	return k.instantiateInternal(ctx, codeID, creator, initMsg, label, deposit, contractAddress, codeInfo)
}

func (k Keeper) instantiateInternal(
	ctx sdk.Context,
	codeID uint64,
	creator sdk.AccAddress,
	initMsg []byte,
	label string,
	deposit sdk.Coins,
	contractAddress sdk.AccAddress,
	codeInfo *types.CodeInfo,
) (sdk.AccAddress, []byte, error) {
	if creator == nil {
		return nil, nil, types.ErrEmpty.Wrap("creator")
	}
	// instanceCosts := k.gasRegister.NewContractInstanceCosts(k.IsPinnedCode(ctx, codeID), len(initMsg))
	// ctx.GasMeter().ConsumeGas(instanceCosts, "Loading CosmWasm module: instantiate")

	if k.HasContractInfo(ctx, contractAddress) {
		return nil, nil, types.ErrDuplicate.Wrap("instance with this code id, sender and label exists: try a different label")
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
		// TODO see alternatives in cosmwasm
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
		if err := k.bank.SendCoins(ctx, creator, contractAddress, deposit); err != nil {
			return nil, nil, err
		}
	}

	// prepare params for contract instantiate call
	env := types.NewEnv(ctx, contractAddress)
	info := types.NewInfo(creator, creator, deposit, false, false)

	// create prefixed data store
	// 0x03 | BuildContractAddressClassic (sdk.AccAddress)
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)

	// prepare querier
	handler := k.newCosmosHandler(ctx, contractAddress)

	// instantiate wasm contract
	// gas := k.runtimeGasForContract(ctx)
	// res, gasUsed, err := k.wasmvm.Instantiate(codeInfo.CodeHash, env, info, initMsg, prefixStore, cosmwasmAPI, querier, k.gasMeter(ctx), gas, costJSONDeserialization)
	// k.consumeRuntimeGas(ctx, gasUsed)

	res, _, err := k.wasmvm.Instantiate(codeInfo.CodeHash, env, info, initMsg, prefixStore, handler, 0, codeInfo.Deps)

	if err != nil {
		return nil, nil, sdkerrors.Wrap(types.ErrInstantiateFailed, err.Error())
	}

	// persist instance first
	contractInfo := types.NewContractInfo(codeID, creator, label)

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

	// data, err := k.handleContractResponse(ctx, contractAddress, contractInfo.IBCPortID, res.Messages, res.Attributes, res.Data, res.Events)
	// if err != nil {
	// 	return nil, nil, sdkerrors.Wrap(err, "dispatch")
	// }

	// return contractAddress, data, nil

	return contractAddress, res.Data, nil
}

// Execute executes the contract instance
func (k Keeper) execute(ctx sdk.Context, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins, dependencies []string) ([]byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasmx", "contract", "execute")
	contractInfo, codeInfo, prefixStore, err := k.contractInstance(ctx, contractAddress)
	if err != nil {
		return nil, err
	}

	var deps []types.ContractDependency
	// TODO get deps also from codeInfo.Dependencies
	for _, hexaddr := range dependencies {
		addr := wasmeth.AccAddressFromHex(hexaddr)
		_, codeInfo, prefixStore, err := k.contractInstance(ctx, addr)
		if err != nil {
			return nil, err
		}
		deps = append(deps, types.ContractDependency{
			Address:  addr,
			Store:    prefixStore,
			FilePath: k.wasmvm.build_path(k.wasmvm.DataDir, codeInfo.CodeHash),
		})
	}

	// TODO panic if coin is not the correct denomination
	// add denom param for ewasm

	// executeCosts := k.gasRegister.InstantiateContractCosts(k.IsPinnedCode(ctx, contractInfo.CodeID), len(msg))
	// ctx.GasMeter().ConsumeGas(executeCosts, "Loading CosmWasm module: execute")

	// add more funds
	if !coins.IsZero() {
		if err := k.TransferCoins(ctx, caller, contractAddress, coins); err != nil {
			return nil, err
		}
	}

	env := types.NewEnv(ctx, contractAddress)
	info := types.NewInfo(caller, caller, coins, false, false)

	// prepare querier
	handler := k.newCosmosHandler(ctx, contractAddress)
	// gas := k.runtimeGasForContract(ctx)
	// res, gasUsed, execErr := k.wasmVM.Execute(codeInfo.CodeHash, env, info, msg, prefixStore, cosmwasmAPI, querier, k.gasMeter(ctx), gas, costJSONDeserialization)

	res, _, execErr := k.wasmvm.Execute(codeInfo.CodeHash, env, info, msg, prefixStore, handler, 0, codeInfo.Deps, deps)

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
	_, _, _, err := k.contractInstance(ctx, caller)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrExecuteFailed, "cannot executeWithOrigin from EOA")
	}

	contractInfo, codeInfo, prefixStore, err := k.contractInstance(ctx, contractAddress)
	if err != nil {
		return nil, err
	}

	// executeCosts := k.gasRegister.InstantiateContractCosts(k.IsPinnedCode(ctx, contractInfo.CodeID), len(msg))
	// ctx.GasMeter().ConsumeGas(executeCosts, "Loading CosmWasm module: execute")

	// add more funds
	if !coins.IsZero() {
		if err := k.TransferCoins(ctx, caller, contractAddress, coins); err != nil {
			return nil, err
		}
	}

	env := types.NewEnv(ctx, contractAddress)
	info := types.NewInfo(origin, caller, coins, false, false)
	handler := k.newCosmosHandler(ctx, contractAddress)
	// gas := k.runtimeGasForContract(ctx)
	// res, gasUsed, execErr := k.wasmVM.Execute(codeInfo.CodeHash, env, info, msg, prefixStore, cosmwasmAPI, querier, k.gasMeter(ctx), gas, costJSONDeserialization)

	res, _, execErr := k.wasmvm.Execute(codeInfo.CodeHash, env, info, msg, prefixStore, handler, 0, codeInfo.Deps, nil)

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
	_, codeInfo, prefixStore, err := k.contractInstance(ctx, contractAddress)
	if err != nil {
		return nil, err
	}

	// TODO panic if coin is not the correct denomination
	// add denom param for ewasm

	// executeCosts := k.gasRegister.InstantiateContractCosts(k.IsPinnedCode(ctx, contractInfo.CodeID), len(msg))
	// ctx.GasMeter().ConsumeGas(executeCosts, "Loading CosmWasm module: execute")

	var deps []types.ContractDependency
	// TODO get deps also from codeInfo.Dependencies
	for _, hexaddr := range dependencies {
		addr := wasmeth.AccAddressFromHex(hexaddr)
		_, codeInfo, prefixStore, err := k.contractInstance(ctx, addr)
		if err != nil {
			return nil, err
		}
		deps = append(deps, types.ContractDependency{
			Address:  addr,
			Store:    prefixStore,
			FilePath: k.wasmvm.build_path(k.wasmvm.DataDir, codeInfo.CodeHash),
		})
	}

	// add more funds
	if !coins.IsZero() {
		if err := k.TransferCoins(ctx, caller, contractAddress, coins); err != nil {
			return nil, err
		}
	}

	env := types.NewEnv(ctx, contractAddress)
	info := types.NewInfo(caller, caller, coins, true, true)
	handler := k.newCosmosHandler(ctx, contractAddress)
	// gas := k.runtimeGasForContract(ctx)
	// res, gasUsed, execErr := k.wasmVM.Execute(codeInfo.CodeHash, env, info, msg, prefixStore, cosmwasmAPI, querier, k.gasMeter(ctx), gas, costJSONDeserialization)

	res, _, execErr := k.wasmvm.QueryExecute(codeInfo.CodeHash, env, info, msg, prefixStore, handler, 0, dependencies, deps)
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
	// msgs []wasmvmtypes.SubMsg,
	attrs []types.EventAttribute,
	data []byte,
	evts types.Events,
) ([]byte, error) {
	// attributeGasCost := k.gasRegister.EventCosts(attrs, evts)
	// ctx.GasMeter().ConsumeGas(attributeGasCost, "Custom contract event attributes")
	// emit all events from this contract itself
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
