package vm

import (
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm/types"
)

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

type MigrateContractStateByStorageRequest struct {
	ContractAddress   string                    `json:"contract_address"`
	SourceStorageType types.ContractStorageType `json:"source_storage_type"`
	TargetStorageType types.ContractStorageType `json:"target_storage_type"`
}

// address -> codeInfo
func coreMigrateContractStateByStorageType(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	data, err := rnh.ReadMemFromPtr(params[0])
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

func setContractInfo(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	addr, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	address := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addr))
	data, err := rnh.ReadMemFromPtr(params[1])
	if err != nil {
		return nil, err
	}

	var contractInfo types.ContractInfo
	err = ctx.CosmosHandler.Codec().UnmarshalJSON(data, &contractInfo)
	if err != nil {
		return nil, fmt.Errorf("ContractInfo cannot be unmarshalled")
	}
	ctx.CosmosHandler.SetContractInfo(address.Bytes(), &contractInfo)
	returns := make([]interface{}, 0)
	return returns, nil
}

// call request -> call response
func coreExternalCall(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req vmtypes.CallRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		ctx.Ctx.Logger().Debug("unmarshalling CallRequest failed", "error", err)
		return nil, err
	}

	returns := make([]interface{}, 1)
	var success int32
	var returnData []byte

	to := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(req.To)
	from := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(req.From)
	commonReq := req.ToCommon(from, to)

	// Send funds
	if req.Value.BitLen() > 0 {
		err = BankSendCoin(ctx, ctx.Env.Contract.Address, to, sdk.NewCoins(sdk.NewCoin(ctx.Env.Chain.Denom, sdkmath.NewIntFromBigInt(req.Value))))
	}
	if err != nil {
		success = int32(2)
	} else {
		success, returnData = WasmxCall(ctx, commonReq)
	}

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	ptr, err := rnh.AllocateWriteMem(responsebz)
	if err != nil {
		return nil, err
	}
	returns[0] = ptr
	return returns, nil
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
	returns := make([]interface{}, 1)
	databz, err := rnh.ReadMemFromPtr(params[0])
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
	ptr, err := rnh.AllocateWriteMem(respbz)
	if err != nil {
		return nil, err
	}
	returns[0] = ptr
	return returns, nil
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
	reqbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req StartTimeoutRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}

	msgtosend := &networktypes.MsgStartTimeoutRequest{
		Id:       req.Id,
		Sender:   ctx.Env.Contract.Address.String(),
		Contract: req.Contract,
		Delay:    req.Delay,
		Args:     req.Args,
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
	reqbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req CancelTimeoutRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}

	msgtosend := &networktypes.MsgCancelTimeoutRequest{
		Id:     req.Id,
		Sender: ctx.Env.Contract.Address.String(),
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
	reqbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req StartBackgroundProcessRequest
	err = json.Unmarshal(reqbz, &req)
	if err != nil {
		return nil, err
	}

	msgtosend := &networktypes.MsgStartBackgroundProcessRequest{
		Sender:   ctx.Env.Contract.Address.String(), // TODO wasmx?
		Contract: req.Contract,
		Args:     req.Args,
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
	reqbz, err := rnh.ReadMemFromPtr(params[0])
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
	proc, ok := procc.Processes[contractAddress.String()]
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
	reqbz, err := rnh.ReadMemFromPtr(params[0])
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
	proc, ok := procc.Processes[contractAddress.String()]
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

func BuildWasmxCoreEnvi32(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("setContractInfo", setContractInfo, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("migrateContractStateByStorageType", coreMigrateContractStateByStorageType, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("externalCall", coreExternalCall, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("grpcRequest", coreWasmxGrpcRequest, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("startTimeout", coreWasmxStartTimeout, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("cancelTimeout", coreWasmxCancelTimeout, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("startBackgroundProcess", coreWasmxStartBackgroundProcess, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("writeToBackgroundProcess", coreWasmxWriteToBackgroundProcess, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("readFromBackgroundProcess", coreWasmxReadFromBackgroundProcess, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),

		// TODO
		// env.AddFunction("endBackgroundProcess", NewFunction(functype_i32_, wasmxEndBackgroundProcess, context, 0))
	}

	return vm.BuildModule(rnh, "wasmxcore", context, fndefs)
}
