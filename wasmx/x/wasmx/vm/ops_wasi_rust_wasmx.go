package vm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path"
	"path/filepath"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	wasimem "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/wasi"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm/types"
)

// This is used for the python & js interpreters
// TODO do a full fledged Rust memory adapter for any Rust contracts
// use the i64 ptr for reading data, so we have a consistent implementation with wasmx api

// getEnv(): ArrayBuffer
func wasi_getEnv(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	data, err := json.Marshal(ctx.Env)
	if err != nil {
		return nil, err
	}
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), data)
	if err != nil {
		return returns, err
	}
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(data)))
	return returns, nil
}

func wasiStorageStore(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	keybz, err := mem.ReadRaw(params[0], params[1])
	if err != nil {
		return nil, err
	}
	fmt.Println("--wasiStorageStore key--", string(keybz), keybz)
	valuebz, err := mem.ReadRaw(params[2], params[3])
	if err != nil {
		return nil, err
	}
	fmt.Println("--wasiStorageStore value--", string(valuebz), valuebz)
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_EWASM), "wasiStorageStore")
	ctx.ContractStore.Set(keybz, valuebz)

	returns := make([]interface{}, 0)
	return returns, nil
}

func wasiStorageLoad(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	keybz, err := mem.ReadRaw(params[0], params[1])
	if err != nil {
		return nil, err
	}
	fmt.Println("--wasiStorageLoad key--", string(keybz), keybz)
	data := ctx.ContractStore.Get(keybz)
	if len(data) == 0 {
		data = types.EMPTY_BYTES32
	}
	fmt.Println("--wasiStorageLoad data--", string(data), data)
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), data)
	if err != nil {
		return returns, err
	}
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(data)))
	return returns, nil
}

// getCallData(ptr): ArrayBuffer
func wasiGetCallData(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	ctx := _context.(*Context)
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), ctx.Env.CurrentCall.CallData)
	if err != nil {
		return returns, err
	}
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(ctx.Env.CurrentCall.CallData)))
	return returns, nil
}

func wasiSetFinishData(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	data, err := mem.ReadRaw(params[0], params[1])
	if err != nil {
		return nil, err
	}
	ctx.FinishData = data
	ctx.ReturnData = data
	return returns, nil
}

func wasiSetExitCode(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	ctx := _context.(*Context)
	code := params[0].(int32)
	if code == 0 {
		return returns, nil
	}
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	errorMsg, err := mem.ReadRaw(params[1], params[2])
	if err != nil {
		return nil, err
	}
	ctx.FinishData = errorMsg
	ctx.ReturnData = errorMsg
	return returns, fmt.Errorf(string(errorMsg))
}

func wasiCallClassic(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	gasLimit := params[0].(int64)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	addrbz, err := mem.ReadRaw(params[1], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}
	addr := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addrbz))
	value, err := memc.ReadBigInt(mem, params[2], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}
	calldata, err := mem.ReadRaw(params[3], params[4])
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}

	var success int32
	var returnData []byte

	// Send funds
	if value.BitLen() > 0 {
		err = BankSendCoin(ctx, ctx.Env.Contract.Address, addr, sdk.NewCoins(sdk.NewCoin(ctx.Env.Chain.Denom, sdkmath.NewIntFromBigInt(value))))
	}
	if err != nil {
		success = int32(2)
	} else {
		contractInfo := GetContractDependency(ctx, addr)
		if contractInfo == nil {
			// ! we return success here in case the contract does not exist
			success = int32(0)
		} else {
			req := vmtypes.CallRequestCommon{
				To:           addr,
				From:         ctx.Env.Contract.Address,
				Value:        value,
				GasLimit:     big.NewInt(gasLimit),
				Calldata:     calldata,
				Bytecode:     contractInfo.Bytecode,
				CodeHash:     contractInfo.CodeHash,
				CodeFilePath: contractInfo.CodeFilePath,
				AotFilePath:  contractInfo.AotFilePath,
				CodeId:       contractInfo.CodeId,
				SystemDeps:   contractInfo.SystemDepsRaw,
				IsQuery:      false,
			}
			success, returnData = WasmxCall(ctx, req)
		}
	}

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), responsebz)
	if err != nil {
		return returns, err
	}
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(responsebz)))
	return returns, nil
}

func wasiCallStatic(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	gasLimit := params[0].(int64)
	addrbz, err := mem.ReadRaw(params[1], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}
	addr := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addrbz))
	calldata, err := mem.ReadRaw(params[2], params[3])
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}

	var success int32
	var returnData []byte

	contractInfo := GetContractDependency(ctx, addr)
	if contractInfo == nil {
		// ! we return success here in case the contract does not exist
		success = int32(0)
	} else {
		req := vmtypes.CallRequestCommon{
			To:           addr,
			From:         ctx.Env.Contract.Address,
			Value:        big.NewInt(0),
			GasLimit:     big.NewInt(gasLimit),
			Calldata:     calldata,
			Bytecode:     contractInfo.Bytecode,
			CodeHash:     contractInfo.CodeHash,
			CodeFilePath: contractInfo.CodeFilePath,
			AotFilePath:  contractInfo.AotFilePath,
			CodeId:       contractInfo.CodeId,
			SystemDeps:   contractInfo.SystemDepsRaw,
			IsQuery:      true,
		}
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
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), responsebz)
	if err != nil {
		return returns, err
	}
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(responsebz)))
	return returns, nil
}

// address -> account
func wasi_getAccount(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	addr, err := mem.ReadRaw(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	address := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addr))
	code := types.EnvContractInfo{
		Address:    address,
		CodeHash:   types.EMPTY_BYTES32,
		CodeId:     0,
		SystemDeps: []string{},
	}
	contractInfo, codeInfo, _, err := ctx.CosmosHandler.GetContractInstance(address.Bytes())
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
		return nil, err
	}
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), codebz)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(codebz)))
	return returns, nil
}

func wasi_keccak256(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	data, err := mem.ReadRaw(params[0], params[1])
	if err != nil {
		return nil, err
	}
	if ctx.ContractRouter["keccak256"] == nil {
		return nil, fmt.Errorf("missing keccak256 wasm module")
	}
	keccakRnh := ctx.ContractRouter["keccak256"].RuntimeHandler
	input_offset := int32(0)
	input_length := int32(len(data))
	output_offset := input_length
	context_offset := output_offset + int32(32)

	keccakVm := keccakRnh.GetVm()
	keccakMem, err := keccakVm.GetMemory()
	if err != nil {
		return nil, err
	}
	if keccakMem == nil {
		return nil, fmt.Errorf("missing keccak256 wasm memory")
	}
	err = keccakMem.Write(input_offset, data)
	if err != nil {
		return nil, err
	}

	_, err = keccakVm.Call("keccak", []interface{}{context_offset, input_offset, input_length, output_offset}, ctx.GasMeter)
	if err != nil {
		return nil, err
	}
	result, err := keccakMem.Read(output_offset, 32)
	if err != nil {
		return nil, err
	}
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), result)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(result)))
	return returns, nil
}

// getBalance(address): i256
func wasi_getBalance(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	addr, err := mem.ReadRaw(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	address := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addr))
	balance, err := BankGetBalance(ctx, address, ctx.Env.Chain.Denom)
	if err != nil {
		return nil, err
	}
	balancebz := balance.Amount.BigInt().FillBytes(make([]byte, 32))
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), balancebz)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(balancebz)))
	return returns, nil
}

// getBlockHash(i64): bytes32
func wasi_getBlockHash(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	blockNumber := params[0].(int64)
	data := ctx.CosmosHandler.GetBlockHash(uint64(blockNumber))
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), data)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(data)))
	return returns, nil
}

// create(bytecode, balance): address
// instantiateAccount(codeId, msgInit, balance): address
func wasi_instantiateAccount(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	codeId := params[0].(int64)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	initMsg, err := mem.ReadRaw(params[1], params[2])
	if err != nil {
		return nil, err
	}
	balance, err := memc.ReadBigInt(mem, params[3], params[4])
	if err != nil {
		return nil, err
	}
	// TODO label?
	contractAddress, err := ctx.CosmosHandler.Create(uint64(codeId), ctx.Env.Contract.Address, initMsg, "", balance, nil)
	if err != nil {
		return nil, err
	}
	contractbz := memc.PaddLeftTo32(contractAddress.Bytes())
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), contractbz)
	if err != nil {
		return nil, err
	}
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(contractbz)))
	return returns, nil
}

// create2(address, salt, bytes): address
// instantiateAccount(codeId, salt, msgInit, balance): address
func wasi_instantiateAccount2(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	codeId := params[0].(int64)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	salt, err := mem.ReadRaw(params[1], int32(32))
	if err != nil {
		return nil, err
	}
	initMsg, err := mem.ReadRaw(params[2], params[3])
	if err != nil {
		return nil, err
	}
	balance, err := memc.ReadBigInt(mem, params[4], params[5])
	if err != nil {
		return nil, err
	}
	// TODO label?
	contractAddress, err := ctx.CosmosHandler.Create2(uint64(codeId), ctx.Env.Contract.Address, initMsg, salt, "", balance, nil)
	if err != nil {
		return nil, err
	}
	contractbz := memc.PaddLeftTo32(contractAddress.Bytes())
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), contractbz)
	if err != nil {
		return nil, err
	}
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(contractbz)))
	return returns, nil
}

// getCodeHash(address): bytes32
func wasi_getCodeHash(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	addr, err := mem.ReadRaw(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	checksum := ctx.CosmosHandler.GetCodeHash(address)
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), checksum)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(checksum)))
	return returns, nil
}

// getCode(address): []byte
func wasi_getCode(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	addr, err := mem.ReadRaw(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	code := ctx.CosmosHandler.GetCode(address)
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), code)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(code)))
	return returns, nil
}

// sendCosmosMsg(req): bytes
func wasi_sendCosmosMsg(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	// TODO
	return returns, fmt.Errorf("wasi_sendCosmosMsg not implemented")
}

// sendCosmosQuery(req): bytes
func wasi_sendCosmosQuery(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	// TODO
	return returns, fmt.Errorf("wasi_sendCosmosQuery not implemented")
}

// getGasLeft(): i64
func wasi_getGasLeft(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	returns[0] = int32(0)
	return returns, nil
}

// bech32StringToBytes(ptr, len): i64
func wasi_bech32StringToBytes(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	addrbz, err := mem.ReadRaw(params[0], params[1])
	if err != nil {
		return nil, err
	}
	addr, err := ctx.CosmosHandler.AddressCodec().StringToBytes(string(addrbz))
	if err != nil {
		return nil, err
	}
	data := types.PaddLeftTo32(addr)
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), data)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	// all addresses are 32 bytes
	returns[0] = ptr
	return returns, nil
}

// bech32BytesToString(ptr): i64
func wasi_bech32BytesToString(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	addrbz, err := mem.ReadRaw(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrbz))

	addrstr, err := ctx.CosmosHandler.AddressCodec().BytesToString(addr)
	if err != nil {
		return nil, err
	}

	data := []byte(addrstr)
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), data)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtrI64(ptr, int32(len(data)))
	return returns, nil
}

func wasiLog(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	data, err := mem.ReadRaw(params[0], params[1])
	if err != nil {
		return nil, err
	}
	ctx.Ctx.Logger().Debug("wasi log", "message", string(data))
	var wlog WasmxJsonLog
	err = json.Unmarshal(data, &wlog)
	if err != nil {
		return nil, err
	}
	// fmt.Println("LOG data: ", string(wlog.Data))
	dependency := types.DEFAULT_SYS_DEP
	if len(ctx.Env.Contract.SystemDeps) > 0 {
		dependency = ctx.Env.Contract.SystemDeps[0]
	}
	log := WasmxLog{
		Type:             LOG_TYPE_WASMX,
		ContractAddress:  ctx.Env.Contract.Address,
		SystemDependency: dependency,
		Data:             wlog.Data,
		Topics:           wlog.Topics,
	}
	ctx.Logs = append(ctx.Logs, log)

	returns := make([]interface{}, 0)
	return returns, nil
}

func BuildWasiWasmxEnv(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("getEnv", wasi_getEnv, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("storageStore", wasiStorageStore, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("storageLoad", wasiStorageLoad, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getCallData", wasiGetCallData, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("setFinishData", wasiSetFinishData, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		// TODO some precompiles use setReturnData instead of setFinishData
		vm.BuildFn("setReturnData", wasiSetFinishData, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("setExitCode", wasiSetExitCode, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("callClassic", wasiCallClassic, []interface{}{vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("callStatic", wasiCallStatic, []interface{}{vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getBlockHash", wasi_getBlockHash, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getAccount", wasi_getAccount, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getBalance", wasi_getBalance, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getCodeHash", wasi_getCodeHash, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getCode", wasi_getCode, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("keccak256", wasi_keccak256, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),

		// env.AddFunction("createAccount", NewFunction(functype_i32i32i32i32_i64, wasi_createAccount, context, 0))
		// env.AddFunction("createAccount2", NewFunction(functype_i32i32i32i32i32_i64, wasi_createAccount2, context, 0))

		vm.BuildFn("instantiateAccount", wasi_instantiateAccount, []interface{}{vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("instantiateAccount2", wasi_instantiateAccount2, []interface{}{vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("sendCosmosMsg", wasi_sendCosmosMsg, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("sendCosmosQuery", wasi_sendCosmosQuery, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getGasLeft", wasi_getGasLeft, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("bech32StringToBytes", wasi_bech32StringToBytes, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("bech32BytesToString", wasi_bech32BytesToString, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("log", wasiLog, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
	}

	return vm.BuildModule(rnh, "wasmx", context, fndefs)
}

func ExecuteWasi(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]int32, error) {
	var res []int32
	var err error

	fmt.Println("---ExecuteWasi---", funcName, args)

	// WASI standard does not have instantiate
	// this is only for wasmx contracts (e.g. compiled with tinygo, javy)
	// TODO consider extracting this in a dependency
	if funcName == types.ENTRY_POINT_INSTANTIATE {
		fnNames := contractVm.GetFunctionList()
		fmt.Println("---ExecuteWasi fnNames---", fnNames)
		found := false
		for _, name := range fnNames {
			// WASI reactor
			if name == "_initialize" {
				found = true
				fmt.Println("---ExecuteWasi _initialize---")
				res, err = contractVm.Call("_initialize", []interface{}{}, context.GasMeter)
				break
			}
			// note that custom entries do not have access to WASI endpoints at this time
			if name == "main.instantiate" {
				found = true
				fmt.Println("---ExecuteWasi main.instantiate---")
				res, err = contractVm.Call("main.instantiate", []interface{}{}, context.GasMeter)
				break
			}
			if name == funcName {
				found = true
				fmt.Println("---ExecuteWasi main.instantiate---")
				res, err = contractVm.Call(funcName, []interface{}{}, context.GasMeter)
				break
			}
		}
		if !found {
			return nil, nil
		}
	} else if funcName == types.ENTRY_POINT_TIMED || funcName == types.ENTRY_POINT_P2P_MSG {
		fmt.Println("---ExecuteWasi---", funcName)
		res, err = contractVm.Call(funcName, []interface{}{}, context.GasMeter)
	} else {
		fmt.Println("---ExecuteWasi _start---")
		// WASI command - no args, no return
		res, err = contractVm.Call("_start", []interface{}{}, context.GasMeter)

		// res, err = contractVm.Execute("_initialize")
		// fmt.Println("--testtime-main, err", res, err)

		// res, err = contractVm.Execute("main")
		// fmt.Println("--testtime-main, err", res, err)

		// res, err = contractVm.Execute("testtime")
		// fmt.Println("--testtime-res, err", res, err)
	}
	if err != nil {
		return nil, err
	}
	fmt.Println("---ExecuteWasi END---", funcName, args)

	return res, nil
}

func ExecutePythonInterpreter(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]int32, error) {
	if funcName == "execute" || funcName == "query" {
		funcName = "main"
	}
	fileName := "__main__.py"
	dir := filepath.Dir(context.ContractInfo.CodeFilePath)
	inputFile := path.Join(dir, fileName)
	content, err := os.ReadFile(context.ContractInfo.CodeFilePath)
	if err != nil {
		return nil, err
	}
	// TODO import the module or make sure json & sys are not imported one more time
	// TODO!! set_finishdata instead of set_returndata
	strcontent := fmt.Sprintf(`import sys
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

	fileMap := map[string][]byte{}
	fileMap[inputFile] = []byte(strcontent)
	contractVm.InstantiateWasi(
		[]string{
			``,
			fileName,
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
		fileMap,
	)
	return ExecuteWasi(context, contractVm, types.ENTRY_POINT_EXECUTE, make([]interface{}, 0))
}

func ExecuteJsInterpreter(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]int32, error) {
	if funcName == "execute" || funcName == "query" {
		funcName = "main"
	}

	dir := filepath.Dir(context.ContractInfo.CodeFilePath)
	fileName := filepath.Base(context.ContractInfo.CodeFilePath)
	inputFileName := "main.js"
	inputFile := path.Join(dir, inputFileName)

	content, err := os.ReadFile(context.ContractInfo.CodeFilePath)
	if err != nil {
		return nil, err
	}

	// TODO setReturnData to setFinishData
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

	fileMap := map[string][]byte{}
	fileMap[inputFile] = []byte(strcontent)
	fileMap[context.ContractInfo.CodeFilePath] = content

	contractVm.InstantiateWasi(
		[]string{
			``,
			inputFileName,
			string(context.Env.CurrentCall.CallData),
		},
		[]string{},
		[]string{
			fmt.Sprintf(`.:%s`, dir),
		},
		fileMap,
	)
	return ExecuteWasi(context, contractVm, types.ENTRY_POINT_EXECUTE, make([]interface{}, 0))
}

func ExecuteWasiWrap(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]int32, error) {
	// if funcName == "execute" || funcName == "query" {
	// 	funcName = "main"
	// }

	fileMap := map[string][]byte{}

	contractVm.InstantiateWasi(
		[]string{
			``,
		},
		// os.Environ(), // The envs
		[]string{},
		// The mapping preopens
		[]string{
			// fmt.Sprintf(`.:%s`, dir),
		},
		fileMap,
	)
	return ExecuteWasi(context, contractVm, funcName, args)
}
