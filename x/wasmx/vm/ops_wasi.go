package vm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm/types"
)

// getEnv(): ArrayBuffer
func wasi_getEnv(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	data, err := json.Marshal(ctx.Env)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, data)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = buildPtr64(ptr, int32(len(data)))
	return returns, wasmedge.Result_Success
}

func wasiStorageStore(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	keybz, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	valuebz, err := readMem(callframe, params[2], params[3])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_EWASM), "wasmx")
	ctx.ContractStore.Set(keybz, valuebz)

	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func wasiStorageLoad(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	keybz, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data := ctx.ContractStore.Get(keybz)
	if len(data) == 0 {
		data = types.EMPTY_BYTES32
	}
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, data)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = buildPtr64(ptr, int32(len(data)))
	return returns, wasmedge.Result_Success
}

// getCallData(ptr): ArrayBuffer
func wasiGetCallData(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	ctx := context.(*Context)
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, ctx.Env.CurrentCall.CallData)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = buildPtr64(ptr, int32(len(ctx.Env.CurrentCall.CallData)))
	return returns, wasmedge.Result_Success
}

func wasiSetReturnData(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	ctx := context.(*Context)
	data, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.ReturnData = data
	return returns, wasmedge.Result_Success
}

func wasiSetExitCode(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	ctx := context.(*Context)
	code := params[0].(int32)
	if code == 0 {
		return returns, wasmedge.Result_Success
	}
	errorMsg, err := readMem(callframe, params[1], params[2])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.ReturnData = errorMsg
	fmt.Println("--wasiSetExitCode", string(errorMsg))
	return returns, wasmedge.Result_Fail
}

func wasiCallClassic(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--wasiCallClassic--")
	ctx := context.(*Context)
	returns := make([]interface{}, 1)

	gasLimit := params[0].(int64)
	addrbz, err := readMem(callframe, params[1], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrbz))
	value, err := readBigInt(callframe, params[2], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	calldata, err := readMem(callframe, params[3], params[4])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}

	var success int32
	var returnData []byte

	// Send funds
	if value.BitLen() > 0 {
		err = ctx.CosmosHandler.SendCoin(addr, value)
	}
	if err != nil {
		success = int32(2)
	} else {
		contractContext := GetContractContext(ctx, addr)
		if contractContext == nil {
			// ! we return success here in case the contract does not exist
			success = int32(0)
		} else {
			req := vmtypes.CallRequest{
				To:         addr,
				From:       ctx.Env.Contract.Address,
				Value:      value,
				GasLimit:   big.NewInt(gasLimit),
				Calldata:   calldata,
				Bytecode:   contractContext.ContractInfo.Bytecode,
				CodeHash:   contractContext.ContractInfo.CodeHash,
				FilePath:   contractContext.ContractInfo.FilePath,
				CodeId:     contractContext.ContractInfo.CodeId,
				SystemDeps: contractContext.ContractInfo.SystemDepsRaw,
				IsQuery:    false,
			}
			success, returnData = WasmxCall(ctx, req)
		}
	}
	fmt.Println("--wasiCallClassic-success, returnData-", success, returnData)

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = buildPtr64(ptr, int32(len(responsebz)))
	return returns, wasmedge.Result_Success
}

func wasiCallStatic(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--wasiCallStatic--")
	ctx := context.(*Context)
	returns := make([]interface{}, 1)

	gasLimit := params[0].(int64)
	addrbz, err := readMem(callframe, params[1], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrbz))
	calldata, err := readMem(callframe, params[2], params[3])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}

	var success int32
	var returnData []byte

	contractContext := GetContractContext(ctx, addr)
	if contractContext == nil {
		// ! we return success here in case the contract does not exist
		success = int32(0)
	} else {
		req := vmtypes.CallRequest{
			To:         addr,
			From:       ctx.Env.Contract.Address,
			Value:      big.NewInt(0),
			GasLimit:   big.NewInt(gasLimit),
			Calldata:   calldata,
			Bytecode:   contractContext.ContractInfo.Bytecode,
			CodeHash:   contractContext.ContractInfo.CodeHash,
			FilePath:   contractContext.ContractInfo.FilePath,
			CodeId:     contractContext.ContractInfo.CodeId,
			SystemDeps: contractContext.ContractInfo.SystemDepsRaw,
			IsQuery:    true,
		}
		success, returnData = WasmxCall(ctx, req)
	}
	fmt.Println("--wasiCallStatic-success, returnData-", success, returnData)

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, responsebz)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	returns[0] = buildPtr64(ptr, int32(len(responsebz)))
	return returns, wasmedge.Result_Success
}

// address -> account
func wasi_getAccount(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addr, err := readMem(callframe, params[0], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	code := types.EnvContractInfo{
		Address:    address,
		CodeHash:   types.EMPTY_BYTES32,
		CodeId:     0,
		SystemDeps: []string{},
	}
	contractInfo, codeInfo, _, err := ctx.CosmosHandler.GetContractInstance(address)
	if err == nil {
		code = types.EnvContractInfo{
			Address:    address,
			CodeHash:   codeInfo.CodeHash,
			CodeId:     contractInfo.CodeId,
			SystemDeps: codeInfo.Deps,
		}
	}

	codebz, err := json.Marshal(code)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, codebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = buildPtr64(ptr, int32(len(codebz)))
	return returns, wasmedge.Result_Success
}

func wasi_keccak256(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	data, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	if ctx.ContractRouter["keccak256"] == nil {
		return nil, wasmedge.Result_Fail
	}
	keccakVm := ctx.ContractRouter["keccak256"].Vm
	input_offset := int32(0)
	input_length := int32(len(data))
	output_offset := input_length
	context_offset := output_offset + int32(32)

	keccakMem := keccakVm.GetActiveModule().FindMemory("memory")
	if keccakMem == nil {
		return nil, wasmedge.Result_Fail
	}
	err = keccakMem.SetData(data, uint(input_offset), uint(input_length))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}

	_, err = keccakVm.Execute("keccak", context_offset, input_offset, input_length, output_offset)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	result, err := keccakMem.GetData(uint(output_offset), uint(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, result)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = buildPtr64(ptr, int32(len(result)))
	return returns, wasmedge.Result_Success
}

// getBalance(address): i256
func wasi_getBalance(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addr, err := readMem(callframe, params[0], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	balance := ctx.CosmosHandler.GetBalance(address)
	balancebz := balance.FillBytes(make([]byte, 32))
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, balancebz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = buildPtr64(ptr, int32(len(balancebz)))
	return returns, wasmedge.Result_Success
}

// getBlockHash(i64): bytes32
func wasi_getBlockHash(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	blockNumber := params[0].(int64)
	data := ctx.CosmosHandler.GetBlockHash(uint64(blockNumber))
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = buildPtr64(ptr, int32(len(data)))
	return returns, wasmedge.Result_Success
}

// create(bytecode, balance): address
// instantiateAccount(codeId, msgInit, balance): address
func wasi_instantiateAccount(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	codeId := params[0].(int64)
	initMsg, err := readMem(callframe, params[1], params[2])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	balance, err := readBigInt(callframe, params[3], params[4])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// TODO label?
	contractAddress, err := ctx.CosmosHandler.Create(uint64(codeId), ctx.Env.Contract.Address, initMsg, "", balance)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	contractbz := paddLeftTo32(contractAddress.Bytes())
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, contractbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = buildPtr64(ptr, int32(len(contractbz)))
	return returns, wasmedge.Result_Success
}

// create2(address, salt, bytes): address
// instantiateAccount(codeId, salt, msgInit, balance): address
func wasi_instantiateAccount2(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	codeId := params[0].(int64)
	salt, err := readMem(callframe, params[1], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	initMsg, err := readMem(callframe, params[2], params[3])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	balance, err := readBigInt(callframe, params[4], params[5])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// TODO label?
	contractAddress, err := ctx.CosmosHandler.Create2(uint64(codeId), ctx.Env.Contract.Address, initMsg, salt, "", balance)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	contractbz := paddLeftTo32(contractAddress.Bytes())
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, contractbz)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = buildPtr64(ptr, int32(len(contractbz)))
	return returns, wasmedge.Result_Success
}

// getCodeHash(address): bytes32
func wasi_getCodeHash(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addr, err := readMem(callframe, params[0], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	checksum := ctx.CosmosHandler.GetCodeHash(address)
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, checksum)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = buildPtr64(ptr, int32(len(checksum)))
	return returns, wasmedge.Result_Success
}

// getCode(address): []byte
func wasi_getCode(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addr, err := readMem(callframe, params[0], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	code := ctx.CosmosHandler.GetCode(address)
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, code)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = buildPtr64(ptr, int32(len(code)))
	return returns, wasmedge.Result_Success
}

// sendCosmosMsg(req): bytes
func wasi_sendCosmosMsg(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	// TODO
	return returns, wasmedge.Result_Fail
}

// sendCosmosQuery(req): bytes
func wasi_sendCosmosQuery(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	// TODO
	return returns, wasmedge.Result_Fail
}

// getGasLeft(): i64
func wasi_getGasLeft(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	returns[0] = 0
	return returns, wasmedge.Result_Success
}

// bech32StringToBytes(ptr, len): i64
func wasi_bech32StringToBytes(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addrbz, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr, err := sdk.AccAddressFromBech32(string(addrbz))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data := types.PaddLeftTo32(addr.Bytes())
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	// all addresses are 32 bytes
	returns[0] = ptr
	return returns, wasmedge.Result_Success
}

// bech32BytesToString(ptr): i64
func wasi_bech32BytesToString(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addrbz, err := readMem(callframe, params[0], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrbz))
	data := []byte(addr.String())
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = buildPtr64(ptr, int32(len(data)))
	return returns, wasmedge.Result_Success
}

func wasiLog(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	data, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// // TODO only in debug mode
	fmt.Println("LOG: ", string(data))
	var wlog WasmxJsonLog
	err = json.Unmarshal(data, &wlog)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println("LOG data: ", string(wlog.Data))
	dependency := types.DEFAULT_SYS_DEP
	if len(ctx.Env.Contract.SystemDeps) > 0 {
		dependency = ctx.Env.Contract.SystemDeps[0]
	}
	fmt.Println("--WasmxLog--", ctx.Env.Contract.SystemDeps, dependency)
	log := WasmxLog{
		Type:             LOG_TYPE_WASMX,
		ContractAddress:  ctx.Env.Contract.Address,
		SystemDependency: dependency,
		Data:             wlog.Data,
		Topics:           wlog.Topics,
	}
	ctx.Logs = append(ctx.Logs, log)

	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func BuildWasiWasmxEnv(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("wasmx")

	functype__i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i64_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i64i32i32i32_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i64i32i32i32i32_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i64i32i32i32i32i32_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)

	env.AddFunction("getEnv", wasmedge.NewFunction(functype__i64, wasi_getEnv, context, 0))
	env.AddFunction("storageStore", wasmedge.NewFunction(functype_i32i32i32i32_, wasiStorageStore, context, 0))
	env.AddFunction("storageLoad", wasmedge.NewFunction(functype_i32i32_i64, wasiStorageLoad, context, 0))
	env.AddFunction("getCallData", wasmedge.NewFunction(functype__i64, wasiGetCallData, context, 0))
	env.AddFunction("setReturnData", wasmedge.NewFunction(functype_i32i32_, wasiSetReturnData, context, 0))
	env.AddFunction("setExitCode", wasmedge.NewFunction(functype_i32i32i32_, wasiSetExitCode, context, 0))

	env.AddFunction("callClassic", wasmedge.NewFunction(functype_i64i32i32i32i32_i64, wasiCallClassic, context, 0))
	env.AddFunction("callStatic", wasmedge.NewFunction(functype_i64i32i32i32_i64, wasiCallStatic, context, 0))

	env.AddFunction("getBlockHash", wasmedge.NewFunction(functype_i64_i64, wasi_getBlockHash, context, 0))
	env.AddFunction("getAccount", wasmedge.NewFunction(functype_i32_i64, wasi_getAccount, context, 0))
	env.AddFunction("getBalance", wasmedge.NewFunction(functype_i32_i64, wasi_getBalance, context, 0))
	env.AddFunction("getCodeHash", wasmedge.NewFunction(functype_i32_i64, wasi_getCodeHash, context, 0))
	env.AddFunction("getCode", wasmedge.NewFunction(functype_i32_i64, wasi_getCode, context, 0))
	env.AddFunction("keccak256", wasmedge.NewFunction(functype_i32i32_i64, wasi_keccak256, context, 0))

	// env.AddFunction("createAccount", wasmedge.NewFunction(functype_i32i32i32i32_i64, wasi_createAccount, context, 0))
	// env.AddFunction("createAccount2", wasmedge.NewFunction(functype_i32i32i32i32i32_i64, wasi_createAccount2, context, 0))

	env.AddFunction("instantiateAccount", wasmedge.NewFunction(functype_i64i32i32i32i32_i64, wasi_instantiateAccount, context, 0))
	env.AddFunction("instantiateAccount2", wasmedge.NewFunction(functype_i64i32i32i32i32i32_i64, wasi_instantiateAccount2, context, 0))

	env.AddFunction("sendCosmosMsg", wasmedge.NewFunction(functype_i32i32_i64, wasi_sendCosmosMsg, context, 0))
	env.AddFunction("sendCosmosQuery", wasmedge.NewFunction(functype_i32i32_i64, wasi_sendCosmosQuery, context, 0))
	env.AddFunction("getGasLeft", wasmedge.NewFunction(functype__i64, wasi_getGasLeft, context, 0))
	env.AddFunction("bech32StringToBytes", wasmedge.NewFunction(functype_i32i32_i32, wasi_bech32StringToBytes, context, 0))
	env.AddFunction("bech32BytesToString", wasmedge.NewFunction(functype_i32_i64, wasi_bech32BytesToString, context, 0))

	env.AddFunction("log", wasmedge.NewFunction(functype_i32i32_, wasiLog, context, 0))

	return env
}

func ExecuteWasi(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	var res []interface{}
	var err error

	// WASI standard does not have instantiate
	// this is only for wasmx contracts (e.g. compiled with tinygo, javy)
	// TODO consider extracting this in a dependency
	if funcName == types.ENTRY_POINT_INSTANTIATE {
		fnNames, _ := contractVm.GetFunctionList()
		found := false
		for _, name := range fnNames {
			// note that custom entries do not have access to WASI endpoints at this time
			if name == "main.instantiate" {
				found = true
				res, err = contractVm.Execute("main.instantiate")
			}
		}
		if !found {
			return nil, nil
		}
	} else {
		// WASI standard - no args, no return
		res, err = contractVm.Execute("_start")
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func ExecutePythonInterpreter(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	if funcName == "execute" || funcName == "query" {
		funcName = "main"
	}

	wasimodule := contractVm.GetImportModule(wasmedge.WASI)
	dir := filepath.Dir(context.Env.Contract.FilePath)
	inputFile := path.Join(dir, "main.py")
	content, err := os.ReadFile(context.Env.Contract.FilePath)
	if err != nil {
		return nil, err
	}
	// TODO import the module or make sure json & sys are not imported one more time
	strcontent := fmt.Sprintf(`
import sys
import json
from wasmx import set_returndata

%s

res = b''
entrypoint = %s
if len(sys.argv) > 1 and sys.argv[1] != "":
	inputObject = sys.argv[1]
	input = json.loads(inputObject)
	res = entrypoint(input)
else:
	res = entrypoint()

set_returndata(res or b'')
`, string(content), funcName)

	err = os.WriteFile(inputFile, []byte(strcontent), 0644)
	if err != nil {
		return nil, err
	}

	wasimodule.InitWasi(
		[]string{
			``,
			"main.py",
			string(context.Env.CurrentCall.CallData),
			// --quiet
		},
		// os.Environ(), // The envs
		[]string{},
		// The mapping preopens
		[]string{
			// ".:.",
			// fmt.Sprintf(`%s:.`, dir),
			// fmt.Sprintf(`%s:%s`, dir, dir),
			fmt.Sprintf(`.:%s`, dir),
			// fmt.Sprintf(`/:%s`, dir),
		},
	)
	return ExecuteWasi(context, contractVm, types.ENTRY_POINT_EXECUTE)
}

func ExecuteJsInterpreter(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	if funcName == "execute" || funcName == "query" {
		funcName = "main"
	}
	wasimodule := contractVm.GetImportModule(wasmedge.WASI)
	dir := filepath.Dir(context.Env.Contract.FilePath)
	fileName := filepath.Base(context.Env.Contract.FilePath)
	inputFileName := "main.js"
	inputFile := path.Join(dir, inputFileName)
	strcontent := fmt.Sprintf(`
import * as std from "std";
import * as os from "os";
import * as wasmx from "wasmx";
import * as contract from "./%s";

let inputData;

if (scriptArgs.length > 1 && scriptArgs[1] != "") {
	inputData = std.parseExtJSON(scriptArgs[1]);
}
const res = contract.%s(inputData);
wasmx.setReturnData(res || new ArrayBuffer(0));

	`, fileName, funcName)

	err := os.WriteFile(inputFile, []byte(strcontent), 0644)
	if err != nil {
		return nil, err
	}

	wasimodule.InitWasi(
		[]string{
			``,
			inputFileName,
			string(context.Env.CurrentCall.CallData),
		},
		[]string{},
		[]string{
			fmt.Sprintf(`.:%s`, dir),
		},
	)
	return ExecuteWasi(context, contractVm, types.ENTRY_POINT_EXECUTE)
}

func writeMemDefaultMalloc(vm *wasmedge.VM, callframe *wasmedge.CallingFrame, data []byte) (int32, error) {
	datalen := int32(len(data))
	ptr, err := allocateMemDefaultMalloc(vm, datalen)
	if err != nil {
		return 0, err
	}
	err = writeMem(callframe, data, ptr)
	if err != nil {
		return 0, err
	}
	return ptr, nil
}

func writeDynMemDefaultMalloc(ctx *Context, callframe *wasmedge.CallingFrame, data []byte) (uint64, error) {
	ptr, err := writeMemDefaultMalloc(ctx.MustGetVmFromContext(), callframe, data)
	if err != nil {
		return 0, err
	}
	return buildPtr64(ptr, int32(len(data))), nil
}

func allocateMemDefaultMalloc(vm *wasmedge.VM, size int32) (int32, error) {
	result, err := vm.Execute(types.MEMORY_EXPORT_MALLOC, size)
	if err != nil {
		return 0, err
	}
	return result[0].(int32), nil
}

func buildPtr64(ptr int32, datalen int32) uint64 {
	return (uint64(ptr) << uint64(32)) | uint64(datalen)
}
