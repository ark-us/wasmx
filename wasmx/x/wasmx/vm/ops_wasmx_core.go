package vm

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm/types"

	mcfg "github.com/loredanacirstea/wasmx/config"
)

type MigrateContractStateByStorageRequest struct {
	ContractAddress   string                    `json:"contract_address"`
	SourceStorageType types.ContractStorageType `json:"source_storage_type"`
	TargetStorageType types.ContractStorageType `json:"target_storage_type"`
}

type MigrateContractStateByAddressRequest struct {
	SourceContractAddress string                    `json:"source_contract_address"`
	TargetContractAddress string                    `json:"target_contract_address"`
	SourceStorageType     types.ContractStorageType `json:"source_storage_type"`
	TargetStorageType     types.ContractStorageType `json:"target_storage_type"`
}

func migrateContractStateByStorageType(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	sourceStorage types.ContractStorageType,
	targetStorage types.ContractStorageType,
	getContractStore func(ctx sdk.Context, storageType types.ContractStorageType, prefixStoreKey []byte) prefix.Store,
) {
	prefixStoreKey := types.GetContractStorePrefix(contractAddress)
	prefixStoreSource := getContractStore(ctx, sourceStorage, prefixStoreKey)
	prefixStoreTarget := getContractStore(ctx, targetStorage, prefixStoreKey)
	iter := prefixStoreSource.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		prefixStoreTarget.Set(iter.Key(), iter.Value())
		prefixStoreSource.Delete(iter.Key())
	}
}

func migrateContractStateByAddress(
	ctx sdk.Context,
	sourceContractAddress sdk.AccAddress,
	targetContractAddress sdk.AccAddress,
	sourceStorageType types.ContractStorageType,
	targetStorageType types.ContractStorageType,
	getContractStore func(ctx sdk.Context, storageType types.ContractStorageType, prefixStoreKey []byte) prefix.Store,
) {
	prefixStoreKeySource := types.GetContractStorePrefix(sourceContractAddress)
	prefixStoreKeyTarget := types.GetContractStorePrefix(targetContractAddress)
	prefixStoreSource := getContractStore(ctx, sourceStorageType, prefixStoreKeySource)
	prefixStoreTarget := getContractStore(ctx, targetStorageType, prefixStoreKeyTarget)
	iter := prefixStoreSource.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		prefixStoreTarget.Set(iter.Key(), iter.Value())
		prefixStoreSource.Delete(iter.Key())
	}
}

func coreMigrateContractStateByStorageType(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	data, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req MigrateContractStateByStorageRequest
	err = json.Unmarshal(data, &req)
	if err != nil {
		return nil, fmt.Errorf("MigrateContractStateByStorageRequest cannot be unmarshalled")
	}
	addr, err := ctx.CosmosHandler.AccBech32Codec().StringToAccAddressPrefixed(req.ContractAddress)
	if err != nil {
		return nil, fmt.Errorf("MigrateContractStateByStorageRequest cannot unmarshal contract address")
	}

	migrateContractStateByStorageType(ctx.Ctx, addr.Bytes(), req.SourceStorageType, req.TargetStorageType, ctx.GetCosmosHandler().ContractStore)

	returns := make([]interface{}, 0)
	return returns, nil
}

func coreMigrateContractStateByAddress(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	data, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req MigrateContractStateByAddressRequest
	err = json.Unmarshal(data, &req)
	if err != nil {
		return nil, fmt.Errorf("MigrateContractStateByAddressRequest cannot be unmarshalled")
	}
	sourceAddr, err := ctx.CosmosHandler.AccBech32Codec().StringToAccAddressPrefixed(req.SourceContractAddress)
	if err != nil {
		return nil, fmt.Errorf("MigrateContractStateByAddressRequest cannot unmarshal source contract address")
	}
	targetAddr, err := ctx.CosmosHandler.AccBech32Codec().StringToAccAddressPrefixed(req.TargetContractAddress)
	if err != nil {
		return nil, fmt.Errorf("MigrateContractStateByAddressRequest cannot unmarshal target contract address")
	}

	migrateContractStateByAddress(ctx.Ctx, sourceAddr.Bytes(), targetAddr.Bytes(), req.SourceStorageType, req.TargetStorageType, ctx.GetCosmosHandler().ContractStore)

	returns := make([]interface{}, 0)
	return returns, nil
}

// call request -> call response
func coreExternalCall(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req vmtypes.CallRequestCommon
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Debug("unmarshalling coreExternalCall failed", "error", err)
		return nil, err
	}

	var success int32
	var returnData []byte

	// Send funds
	if req.Value.GT(sdkmath.ZeroInt()) {
		err = BankSendCoin(ctx, req.From, req.To, sdk.NewCoins(sdk.NewCoin(ctx.Env.Chain.Denom, req.Value)))
	}
	if err != nil {
		success = int32(2)
	} else {
		success, returnData = WasmxCall(ctx, req)
	}

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

type GrpcRequest struct {
	IpAddress string `json:"ip_address"`
	Contract  []byte `json:"contract"`
	Data      []byte `json:"data"` // should be []byte (base64 encoded)
}

type GrpcResponse struct {
	Data  []byte `json:"data"`
	Error string `json:"error"`
}

func coreWasmxGrpcRequest(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	databz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var data GrpcRequest
	err = json.Unmarshal(databz, &data)
	if err != nil {
		return nil, err
	}

	contractAddress := sdk.AccAddress(vmtypes.CleanupAddress(data.Contract))
	contractAddressStr, err := ctx.CosmosHandler.AddressCodec().BytesToString(contractAddress)
	if err != nil {
		return nil, err
	}

	msg := &networktypes.MsgGrpcSendRequest{
		Authority: ctx.WasmxAuthority,
		IpAddress: data.IpAddress,
		Contract:  contractAddressStr,
		Data:      []byte(data.Data),
		Sender:    ctx.Env.Contract.Address.String(),
	}
	evs, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(msg)
	errmsg := ""
	if err != nil {
		errmsg = err.Error()
	} else {
		ctx.Ctx.EventManager().EmitEvents(evs)
	}
	rres := networktypes.MsgGrpcSendRequestResponse{Data: make([]byte, 0)}
	if res != nil {
		err = rres.Unmarshal(res)
		if err != nil {
			return nil, err
		}
	}
	resp := GrpcResponse{
		Data:  rres.Data,
		Error: errmsg,
	}
	respbz, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(respbz)
}

type StartTimeoutRequest struct {
	Id       string `json:"id"`
	Contract string `json:"contract"`
	Delay    int64  `json:"delay"`
	Args     []byte `json:"args"`
}

type CancelTimeoutRequest struct {
	Id string `json:"id"`
}

type StartBackgroundProcessRequest struct {
	Contract string `json:"contract"`
	Args     []byte `json:"args"`
}

type StartBackgroundProcessResponse struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

type WriteToBackgroundProcessRequest struct {
	Contract string `json:"contract"`
	Data     []byte `json:"data"`
	PtrFunc  string `json:"ptrFunc"`
}

type WriteToBackgroundProcessResponse struct {
	Error string `json:"error"`
}

type ReadFromBackgroundProcessRequest struct {
	Contract string `json:"contract"`
	PtrFunc  string `json:"ptrFunc"`
	LenFunc  string `json:"lenFunc"`
}

type ReadFromBackgroundProcessResponse struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

// TODO move this to a restricted role
// startTimeout(req: ArrayBuffer): i32
func coreWasmxStartTimeout(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req StartTimeoutRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}

	msgtosend := &networktypes.MsgStartTimeoutRequest{
		Authority: ctx.WasmxAuthority,
		Id:        req.Id,
		Sender:    ctx.Env.Contract.Address.String(),
		Contract:  req.Contract,
		Delay:     req.Delay,
		Args:      req.Args,
	}
	evs, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error())
		return nil, err
	}
	ctx.Ctx.EventManager().EmitEvents(evs)
	var resp networktypes.MsgStartTimeoutResponse
	err = resp.Unmarshal(res)
	if err != nil {
		return nil, err
	}
	return returns, nil
}

// TODO move this to a restricted role
func coreWasmxCancelTimeout(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req CancelTimeoutRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}

	msgtosend := &networktypes.MsgCancelTimeoutRequest{
		Authority: ctx.WasmxAuthority,
		Id:        req.Id,
		Sender:    ctx.Env.Contract.Address.String(),
	}
	evs, _, err := ctx.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error())
		return nil, err
	}
	ctx.Ctx.EventManager().EmitEvents(evs)
	return returns, nil
}

// startBackgroundProcess(ArrayBuffer): ArrayBuffer
func coreWasmxStartBackgroundProcess(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req StartBackgroundProcessRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}

	msgtosend := &networktypes.MsgStartBackgroundProcessRequest{
		Authority: ctx.WasmxAuthority,
		Sender:    ctx.Env.Contract.Address.String(),
		Contract:  req.Contract,
		Args:      req.Args,
	}
	evs, res, err := ctx.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error())
		return nil, err
	}
	ctx.Ctx.EventManager().EmitEvents(evs)
	var resp networktypes.MsgStartBackgroundProcessResponse
	err = resp.Unmarshal(res)
	if err != nil {
		return nil, err
	}
	return returns, nil
}

// writeToBackgroundProcess(ArrayBuffer): ArrayBuffer
func coreWasmxWriteToBackgroundProcess(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req WriteToBackgroundProcessRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	resp := WriteToBackgroundProcessResponse{
		Error: "",
	}

	contractAddress, err := ctx.CosmosHandler.GetAddressOrRole(ctx.Ctx, req.Contract)
	if err != nil {
		return nil, err
	}

	procc, err := types.GetBackgroundProcesses(ctx.GoContextParent)
	if err != nil {
		resp.Error = err.Error()
		ptr, err := prepareResponse(ctx, rnh, &resp)
		if err != nil {
			return nil, err
		}
		returns[0] = ptr
		return returns, nil
	}
	proc, ok := procc.GetProcess(contractAddress.String())
	if !ok {
		resp.Error = "process not existent"
		ptr, err := prepareResponse(ctx, rnh, &resp)
		if err != nil {
			return nil, err
		}
		returns[0] = ptr
		return returns, nil
	}

	vm := proc.RuntimeHandler.GetVm()
	ptrGlobal := vm.FindGlobal(req.PtrFunc)
	activeMemory, err := proc.RuntimeHandler.GetMemory()
	if err != nil {
		return nil, err
	}
	err = activeMemory.WriteRaw(ptrGlobal, req.Data)
	if err != nil {
		resp.Error = err.Error()
	}

	ptr, err := prepareResponse(ctx, rnh, &resp)
	if err != nil {
		return nil, err
	}
	returns[0] = ptr
	return returns, nil
}

// readFromBackgroundProcess(ArrayBuffer): ArrayBuffer
func coreWasmxReadFromBackgroundProcess(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ReadFromBackgroundProcessRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	resp := ReadFromBackgroundProcessResponse{
		Data:  []byte{},
		Error: "",
	}

	contractAddress, err := ctx.CosmosHandler.GetAddressOrRole(ctx.Ctx, req.Contract)
	if err != nil {
		return nil, err
	}

	procc, err := types.GetBackgroundProcesses(ctx.GoContextParent)
	if err != nil {
		resp.Error = err.Error()
		ptr, err := prepareResponse(ctx, rnh, &resp)
		if err != nil {
			return nil, err
		}
		returns[0] = ptr
		return returns, nil
	}
	proc, ok := procc.GetProcess(contractAddress.String())
	if !ok {
		resp.Error = "process not existent"
		ptr, err := prepareResponse(ctx, rnh, &resp)
		if err != nil {
			return nil, err
		}
		returns[0] = ptr
		return returns, nil
	}
	vm := proc.RuntimeHandler.GetVm()
	lengthGlobal := vm.FindGlobal(req.LenFunc)
	ptrGlobal := vm.FindGlobal(req.PtrFunc)
	activeMemory, err := proc.RuntimeHandler.GetMemory()
	if err != nil {
		return nil, err
	}
	byteArray, err := activeMemory.ReadRaw(ptrGlobal, lengthGlobal)
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Data = byteArray
	}
	ptr, err := prepareResponse(ctx, rnh, &resp)
	if err != nil {
		return nil, err
	}
	returns[0] = ptr
	return returns, nil
}

type GlobalStorageStoreRequest struct {
	StoreKey string `json:"store_key"`
	Key      []byte `json:"key"`
	Value    []byte `json:"value"`
}

type GlobalStorageLoadRequest struct {
	StoreKey string `json:"store_key"`
	Key      []byte `json:"key"`
}

type GlobalStorageResetRequest struct {
	StoreKey string `json:"store_key"`
}

type GlobalStorageResetResponse struct {
	Error string `json:"error"`
}

type UpdateSystemCacheResponse struct {
	Error string `json:"error"`
}

type BroadcastTxAsyncResponse struct {
	Error    string                      `json:"error"`
	Response *rpctypes.ResultBroadcastTx `json:"response"`
}

func coreWasmxStorageStoreGlobal(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req GlobalStorageStoreRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	if req.Key == nil {
		return make([]interface{}, 0), fmt.Errorf("storage key must not be nil")
	}
	storeKey := storetypes.NewKVStoreKey(req.StoreKey)
	store := ctx.Ctx.KVStore(storeKey)
	store.Set(req.Key, req.Value)
	returns := make([]interface{}, 0)
	return returns, nil
}

func coreWasmxStorageLoadGlobal(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req GlobalStorageLoadRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	if req.Key == nil {
		return make([]interface{}, 0), fmt.Errorf("storage key must not be nil")
	}
	storeKey := storetypes.NewKVStoreKey(req.StoreKey)
	store := ctx.Ctx.KVStore(storeKey)
	data := store.Get(req.Key)
	return rnh.AllocateWriteMem(data)
}

func coreWasmxStorageDeleteGlobal(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req GlobalStorageLoadRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	if req.Key == nil {
		return make([]interface{}, 0), fmt.Errorf("storage key must not be nil")
	}
	storeKey := storetypes.NewKVStoreKey(req.StoreKey)
	store := ctx.Ctx.KVStore(storeKey)
	store.Delete(req.Key)
	returns := make([]interface{}, 0)
	return returns, nil
}

func coreWasmxStorageHasGlobal(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req GlobalStorageLoadRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	if req.Key == nil {
		return make([]interface{}, 0), fmt.Errorf("storage key must not be nil")
	}
	storeKey := storetypes.NewKVStoreKey(req.StoreKey)
	store := ctx.Ctx.KVStore(storeKey)
	haskey := store.Has(req.Key)
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	if haskey {
		returns[0] = int32(1)
	}
	return returns, nil
}

func coreWasmxStorageResetGlobal(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req GlobalStorageResetRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	storeKey := storetypes.NewKVStoreKey(req.StoreKey)
	store := ctx.Ctx.KVStore(storeKey)
	err = store.Reset()
	resp := GlobalStorageResetResponse{Error: ""}
	if err != nil {
		resp.Error = err.Error()
	}
	responsebz, err := json.Marshal(&resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

func coreUpdateSystemCache(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	resp := UpdateSystemCacheResponse{Error: ""}
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req types.SystemBootstrapPerChain
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}
	err = ctx.CosmosHandler.UpdateSystemCache(ctx.Ctx, &req)
	if err != nil {
		// TODO maybe we should shut down the system here, if the cache was not updated correctly
		resp.Error = err.Error()
	}
	responsebz, err := json.Marshal(&resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

func coreBroadcastTxAsync(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	resp := BroadcastTxAsyncResponse{Error: ""}
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	txbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	fmt.Println("--coreBroadcastTxAsync--", hex.EncodeToString((txbz)))
	mapp, ok := ctx.App.(mcfg.MythosApp)
	if !ok {
		return nil, fmt.Errorf("app not MythosApp interface in coreBroadcastTxAsync")
	}
	rpcClient := mapp.GetRpcClient()
	fmt.Println("--coreBroadcastTxAsync.rpcClient--", rpcClient)
	if rpcClient == nil {
		return nil, fmt.Errorf("rpcClient nil in coreBroadcastTxAsync")
	}
	// we use goroutines because the ActionExecutor now is not paralelizable and we will end up waiting for this tx to be broadcasted, which will happen only after the current transaction is finished
	ctx.GoRoutineGroup.Go(func() error {
		res, err := rpcClient.BroadcastTxAsync(ctx.GoContextParent, txbz)
		fmt.Println("--coreBroadcastTxAsync.BroadcastTxAsync--", err, res)
		return err
	})
	// if err != nil {
	// 	resp.Error = err.Error()
	// }
	// if res != nil {
	// 	resp.Response = res
	// }
	responsebz, err := json.Marshal(&resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

type ExecuteCliCommandRequest struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	WorkDir string            `json:"work_dir,omitempty"`
	Env     map[string]string `json:"env,omitempty"` // Environment variables to set for the command
}

type ExecuteCliCommandResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error"`
}

type PrepareTxRequest struct {
	ToAddress    string    `json:"to_address"`
	Data         []byte    `json:"data"`
	Funds        sdk.Coins `json:"funds"`
	Dependencies []string  `json:"dependencies"`
	GasLimit     uint64    `json:"gas_limit"`
	GasPrice     sdk.Coin  `json:"gas_price"`
	PrivateKey   []byte    `json:"private_key"`
}

type PrepareTxResponse struct {
	Error   string `json:"error"`
	TxBytes []byte `json:"tx_bytes"`
}

func coreExecuteCliCommand(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	resp := ExecuteCliCommandResponse{ExitCode: -1}

	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		resp.Error = err.Error()
		responsebz, _ := json.Marshal(&resp)
		return rnh.AllocateWriteMem(responsebz)
	}

	var req ExecuteCliCommandRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		resp.Error = err.Error()
		responsebz, _ := json.Marshal(&resp)
		return rnh.AllocateWriteMem(responsebz)
	}

	// Execute the CLI command
	cmd := exec.CommandContext(ctx.Ctx, req.Command, req.Args...)
	if req.WorkDir != "" {
		cmd.Dir = req.WorkDir
	}

	// Set environment variables if provided
	if len(req.Env) > 0 {
		// Start with current environment
		cmd.Env = os.Environ()
		// Add/override with custom environment variables
		for key, value := range req.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else {
			resp.Error = err.Error()
		}
	} else {
		resp.ExitCode = 0
	}

	responsebz, err := json.Marshal(&resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

func corePrepareTx(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	resp := PrepareTxResponse{Error: "", TxBytes: nil}

	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	reqbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		resp.Error = err.Error()
		responsebz, _ := json.Marshal(&resp)
		return rnh.AllocateWriteMem(responsebz)
	}

	fmt.Println("--corePrepareTx.req--", string(reqbz))

	var req PrepareTxRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		resp.Error = err.Error()
		responsebz, _ := json.Marshal(&resp)
		return rnh.AllocateWriteMem(responsebz)
	}

	privKey := &secp256k1.PrivKey{Key: req.PrivateKey}
	pubKey := privKey.PubKey()
	fromAddr := sdk.AccAddress(pubKey.Address())
	fromAddrPrefixed := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(fromAddr)

	// Parse to address
	toAddr, err := ctx.CosmosHandler.AccBech32Codec().StringToAccAddressPrefixed(req.ToAddress)
	if err != nil {
		resp.Error = fmt.Sprintf("failed to parse to address: %s", err.Error())
		responsebz, _ := json.Marshal(&resp)
		return rnh.AllocateWriteMem(responsebz)
	}

	// Determine message type: use bank.MsgSend for simple coin transfers, MsgExecuteContract for contract calls
	var msg sdk.Msg
	isSimpleTransfer := len(req.Data) == 0 && (req.Dependencies == nil || len(req.Dependencies) == 0)

	if isSimpleTransfer && len(req.Funds) > 0 {
		// Simple bank transfer
		msgSend := &banktypes.MsgSend{
			FromAddress: fromAddrPrefixed.String(),
			ToAddress:   toAddr.String(),
			Amount:      req.Funds,
		}
		msg = msgSend
		fmt.Println("--corePrepareTx.msgSend--", fromAddrPrefixed.String(), toAddr.String(), req.Funds)
	} else {
		// Contract execution
		msgExecute := &types.MsgExecuteContract{
			Contract:     toAddr.String(),
			Sender:       fromAddrPrefixed.String(),
			Msg:          req.Data,
			Funds:        req.Funds,
			Dependencies: req.Dependencies,
		}
		msg = msgExecute
		fmt.Println("--corePrepareTx.msgExecute--", fromAddrPrefixed.String(), toAddr.String(), string(req.Data), req.Funds)
	}

	// Get TxConfig
	txConfig := ctx.App.(mcfg.MythosApp).TxConfig()
	txBuilder := txConfig.NewTxBuilder()

	// Set gas limit
	txBuilder.SetGasLimit(req.GasLimit)

	feeAmount := req.GasPrice.Amount.MulRaw(int64(req.GasLimit))
	fees := sdk.NewCoins(sdk.NewCoin(req.GasPrice.Denom, feeAmount))
	txBuilder.SetFeeAmount(fees)

	// Set message
	err = txBuilder.SetMsgs(msg)
	if err != nil {
		resp.Error = fmt.Sprintf("failed to set messages: %s", err.Error())
		responsebz, _ := json.Marshal(&resp)
		return rnh.AllocateWriteMem(responsebz)
	}

	// Get account and sequence (nonce)
	accP, err := ctx.CosmosHandler.GetAccount(fromAddrPrefixed)
	if err != nil {
		resp.Error = fmt.Sprintf("failed to get account: %s", err.Error())
		responsebz, _ := json.Marshal(&resp)
		return rnh.AllocateWriteMem(responsebz)
	}
	if accP == nil {
		accP, err = ctx.CosmosHandler.NewAccountWithAddressPrefixed(fromAddrPrefixed)
		if err != nil {
			resp.Error = fmt.Sprintf("failed to create account: %s", err.Error())
			responsebz, _ := json.Marshal(&resp)
			return rnh.AllocateWriteMem(responsebz)
		}
	}
	seq := accP.GetSequence()

	fmt.Println("--corePrepareTx.acc--", seq)

	// First round: set empty signature
	sigV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode(txConfig.SignModeHandler().DefaultMode()),
			Signature: nil,
		},
		Sequence: seq,
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		resp.Error = fmt.Sprintf("failed to set empty signatures: %s", err.Error())
		responsebz, _ := json.Marshal(&resp)
		return rnh.AllocateWriteMem(responsebz)
	}

	// Second round: sign transaction
	signerData := authsigning.SignerData{
		ChainID:       ctx.Ctx.ChainID(),
		AccountNumber: accP.GetAccountNumber(),
		Sequence:      seq,
		PubKey:        pubKey,
		Address:       accP.String(),
	}

	fmt.Println("--corePrepareTx.signerData--", signerData)

	sigV2, err = tx.SignWithPrivKey(
		ctx.Ctx.Context(),
		signing.SignMode(txConfig.SignModeHandler().DefaultMode()),
		signerData,
		txBuilder,
		privKey,
		txConfig,
		seq,
	)
	if err != nil {
		resp.Error = fmt.Sprintf("failed to sign transaction: %s", err.Error())
		responsebz, _ := json.Marshal(&resp)
		return rnh.AllocateWriteMem(responsebz)
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		resp.Error = fmt.Sprintf("failed to set final signatures: %s", err.Error())
		responsebz, _ := json.Marshal(&resp)
		return rnh.AllocateWriteMem(responsebz)
	}

	// Encode transaction
	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		resp.Error = fmt.Sprintf("failed to encode transaction: %s", err.Error())
		responsebz, _ := json.Marshal(&resp)
		return rnh.AllocateWriteMem(responsebz)
	}

	fmt.Println("--corePrepareTx.txBytes--", hex.EncodeToString(txBytes))

	resp.TxBytes = txBytes
	responsebz, err := json.Marshal(&resp)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

type GenerateKeyPairResponse struct {
	Error      string `json:"error"`
	PublicKey  []byte `json:"public_key"`
	PrivateKey []byte `json:"private_key"`
}

// coreGenerateKeyPair generates a secp256k1 key pair
func coreGenerateKeyPair(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	resp := GenerateKeyPairResponse{Error: "", PublicKey: nil, PrivateKey: nil}

	// Generate secp256k1 private key
	privKey := secp256k1.GenPrivKey()

	// Get the private key bytes (32 bytes)
	resp.PrivateKey = privKey.Bytes()

	// Get the public key bytes (33 bytes compressed)
	resp.PublicKey = privKey.PubKey().Bytes()

	responsebz, err := json.Marshal(&resp)
	if err != nil {
		resp.Error = fmt.Sprintf("failed to marshal response: %s", err.Error())
		responsebz, _ = json.Marshal(&resp)
	}

	return rnh.AllocateWriteMem(responsebz)
}
