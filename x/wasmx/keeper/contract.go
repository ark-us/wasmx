package keeper

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"wasmx/x/wasmx/ioutils"
	"wasmx/x/wasmx/types"
)

func (k Keeper) Create(ctx sdk.Context, creator sdk.AccAddress, wasmByteCode []byte) (uint64, []byte, error) {
	return k.create(ctx, creator, wasmByteCode)
}

func (k Keeper) Instantiate(ctx sdk.Context, codeId uint64, senderAddr sdk.AccAddress, msg types.RawContractMessage, label string, funds sdk.Coins) (sdk.AccAddress, []byte, error) {
	return nil, nil, nil
}

func (k Keeper) Instantiate2(ctx sdk.Context, codeId uint64, senderAddr sdk.AccAddress, msg types.RawContractMessage, label string, funds sdk.Coins, salt []byte, fixMsg bool) (sdk.AccAddress, []byte, error) {
	return nil, nil, nil
}

func (k Keeper) Execute(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins) ([]byte, error) {
	return nil, nil
}

func (k Keeper) ExecuteWithOrigin(ctx sdk.Context, originAddr sdk.AccAddress, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins) ([]byte, error) {
	return nil, nil
}

func (k Keeper) ExecuteDelegate(ctx sdk.Context, originAddr sdk.AccAddress, codeContractAddr sdk.AccAddress, storageContractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins) ([]byte, error) {
	return nil, nil
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
	checksum, err = k.wasmvm.Create(wasmCode)
	if err != nil {
		return 0, checksum, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	report, err := k.wasmvm.AnalyzeCode(checksum)
	if err != nil {
		return 0, checksum, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	codeID = k.autoIncrementID(ctx, types.KeyLastCodeID)
	k.Logger(ctx).Debug("storing new contract", "capabilities", report.RequiredCapabilities, "code_id", codeID)
	codeInfo := types.NewCodeInfo(checksum, creator)
	k.storeCodeInfo(ctx, codeID, codeInfo)

	evt := sdk.NewEvent(
		types.EventTypeStoreCode,
		sdk.NewAttribute(types.AttributeKeyChecksum, hex.EncodeToString(checksum)),
		sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(codeID, 10)), // last element to be compatible with scripts
	)
	for _, f := range strings.Split(report.RequiredCapabilities, ",") {
		evt.AppendAttributes(sdk.NewAttribute(types.AttributeKeyRequiredCapability, strings.TrimSpace(f)))
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
	creator, admin sdk.AccAddress,
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
	_, data, err := k.instantiateInternal(ctx, codeID, creator, admin, initMsg, label, deposit, contractAddress, codeInfo)
	return data, err
}

func (k Keeper) instantiate(
	ctx sdk.Context,
	codeID uint64,
	creator, admin sdk.AccAddress,
	initMsg []byte,
	label string,
	deposit sdk.Coins,
	addressGenerator AddressGenerator,
) (sdk.AccAddress, []byte, error) {
	defer telemetry.MeasureSince(time.Now(), "wasm", "contract", "instantiate")

	// get contact info
	codeInfo := k.GetCodeInfo(ctx, codeID)
	if codeInfo == nil {
		return nil, nil, sdkerrors.Wrap(types.ErrNotFound, "code")
	}

	contractAddress := addressGenerator(ctx, codeID, codeInfo.CodeHash)
	return k.instantiateInternal(ctx, codeID, creator, admin, initMsg, label, deposit, contractAddress, codeInfo)
}

func (k Keeper) instantiateInternal(
	ctx sdk.Context,
	codeID uint64,
	creator, admin sdk.AccAddress,
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
	info := types.NewInfo(creator, deposit)

	// create prefixed data store
	// 0x03 | BuildContractAddressClassic (sdk.AccAddress)
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)

	// prepare querier
	// querier := k.newQueryHandler(ctx, contractAddress)

	// instantiate wasm contract
	// gas := k.runtimeGasForContract(ctx)
	// res, gasUsed, err := k.wasmvm.Instantiate(codeInfo.CodeHash, env, info, initMsg, prefixStore, cosmwasmAPI, querier, k.gasMeter(ctx), gas, costJSONDeserialization)
	// k.consumeRuntimeGas(ctx, gasUsed)

	res, _, err := k.wasmvm.Instantiate(codeInfo.CodeHash, env, info, initMsg, prefixStore)

	if err != nil {
		return nil, nil, sdkerrors.Wrap(types.ErrInstantiateFailed, err.Error())
	}

	// persist instance first
	contractInfo := types.NewContractInfo(codeID, creator, label)

	// check for IBC flag
	// report, err := k.wasmvm.AnalyzeCode(codeInfo.CodeHash)
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

func (k Keeper) autoIncrementID(ctx sdk.Context, lastIDKey []byte) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(lastIDKey)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	bz = sdk.Uint64ToBigEndian(id + 1)
	store.Set(lastIDKey, bz)
	return id
}

// PeekAutoIncrementID reads the current value without incrementing it.
func (k Keeper) PeekAutoIncrementID(ctx sdk.Context, lastIDKey []byte) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(lastIDKey)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	return id
}

func (k Keeper) importAutoIncrementID(ctx sdk.Context, lastIDKey []byte, val uint64) error {
	store := ctx.KVStore(k.storeKey)
	if store.Has(lastIDKey) {
		return sdkerrors.Wrapf(types.ErrDuplicate, "autoincrement id: %s", string(lastIDKey))
	}
	bz := sdk.Uint64ToBigEndian(val)
	store.Set(lastIDKey, bz)
	return nil
}

func (k Keeper) GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *types.ContractInfo {
	store := ctx.KVStore(k.storeKey)
	var contract types.ContractInfo
	contractBz := store.Get(types.GetContractAddressKey(contractAddress))
	if contractBz == nil {
		return nil
	}
	k.cdc.MustUnmarshal(contractBz, &contract)
	return &contract
}

func (k Keeper) HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetContractAddressKey(contractAddress))
}

// storeContractInfo persists the ContractInfo. No secondary index updated here.
func (k Keeper) storeContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress, contract *types.ContractInfo) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetContractAddressKey(contractAddress), k.cdc.MustMarshal(contract))
}

// Temporary function, used for testing. TODO: remove
// StoreContractInfo persists the ContractInfo. No secondary index updated here.
func (k Keeper) StoreContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress, contract *types.ContractInfo) {
	k.storeContractInfo(ctx, contractAddress, contract)
}

func (k Keeper) IterateContractInfo(ctx sdk.Context, cb func(sdk.AccAddress, types.ContractInfo) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ContractKeyPrefix)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var contract types.ContractInfo
		k.cdc.MustUnmarshal(iter.Value(), &contract)
		// cb returns true to stop early
		if cb(iter.Key(), contract) {
			break
		}
	}
}

// IterateContractState iterates through all elements of the key value store for the given contract address and passes
// them to the provided callback function. The callback method can return true to abort early.
func (k Keeper) IterateContractState(ctx sdk.Context, contractAddress sdk.AccAddress, cb func(key, value []byte) bool) {
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if cb(iter.Key(), iter.Value()) {
			break
		}
	}
}

func (k Keeper) importContractState(ctx sdk.Context, contractAddress sdk.AccAddress, models []types.ContractStorage) error {
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	for _, model := range models {
		if model.Value == nil {
			model.Value = []byte{}
		}
		if prefixStore.Has(model.Key) {
			return sdkerrors.Wrapf(types.ErrDuplicate, "duplicate key: %x", model.Key)
		}
		prefixStore.Set(model.Key, model.Value)
	}
	return nil
}

func (k Keeper) GetCodeInfo(ctx sdk.Context, codeID uint64) *types.CodeInfo {
	store := ctx.KVStore(k.storeKey)
	var codeInfo types.CodeInfo
	codeInfoBz := store.Get(types.GetCodeKey(codeID))
	if codeInfoBz == nil {
		return nil
	}
	k.cdc.MustUnmarshal(codeInfoBz, &codeInfo)
	return &codeInfo
}

func (k Keeper) containsCodeInfo(ctx sdk.Context, codeID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetCodeKey(codeID))
}

func (k Keeper) IterateCodeInfos(ctx sdk.Context, cb func(uint64, types.CodeInfo) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.CodeKeyPrefix)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var c types.CodeInfo
		k.cdc.MustUnmarshal(iter.Value(), &c)
		// cb returns true to stop early
		if cb(binary.BigEndian.Uint64(iter.Key()), c) {
			return
		}
	}
}

func (k Keeper) GetByteCode(ctx sdk.Context, codeID uint64) ([]byte, error) {
	store := ctx.KVStore(k.storeKey)
	var codeInfo types.CodeInfo
	codeInfoBz := store.Get(types.GetCodeKey(codeID))
	if codeInfoBz == nil {
		return nil, nil
	}
	k.cdc.MustUnmarshal(codeInfoBz, &codeInfo)
	return k.wasmvm.GetCode(codeInfo.CodeHash)
}
