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
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	memc "mythos/v1/x/wasmx/vm/memory/common"
	wasimem "mythos/v1/x/wasmx/vm/memory/wasi"
	vmtypes "mythos/v1/x/wasmx/vm/types"
)

// getEnv(): ArrayBuffer
func wasi_getEnv(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	data, err := json.Marshal(ctx.Env)
	if err != nil {
		return nil, err
	}
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), data)
	if err != nil {
		return returns, err
	}
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(data)))
	return returns, nil
}

func wasiStorageStore(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keybz, err := rnh.GetMemory().Read(params[0], params[1])
	if err != nil {
		return nil, err
	}
	valuebz, err := rnh.GetMemory().Read(params[2], params[3])
	if err != nil {
		return nil, err
	}
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_EWASM), "wasmx")
	ctx.ContractStore.Set(keybz, valuebz)

	returns := make([]interface{}, 0)
	return returns, nil
}

func wasiStorageLoad(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	keybz, err := rnh.GetMemory().Read(params[0], params[1])
	if err != nil {
		return nil, err
	}
	data := ctx.ContractStore.Get(keybz)
	if len(data) == 0 {
		data = types.EMPTY_BYTES32
	}
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), data)
	if err != nil {
		return returns, err
	}
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(data)))
	return returns, nil
}

// getCallData(ptr): ArrayBuffer
func wasiGetCallData(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	ctx := _context.(*Context)
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), ctx.Env.CurrentCall.CallData)
	if err != nil {
		return returns, err
	}
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(ctx.Env.CurrentCall.CallData)))
	return returns, nil
}

func wasiSetFinishData(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	ctx := _context.(*Context)
	data, err := rnh.GetMemory().Read(params[0], params[1])
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
	errorMsg, err := rnh.GetMemory().Read(params[1], params[2])
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
	addrbz, err := rnh.GetMemory().Read(params[1], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}
	addr := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addrbz))
	value, err := memc.ReadBigInt(rnh.GetMemory(), params[2], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}
	calldata, err := rnh.GetMemory().Read(params[3], params[4])
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
		contractContext := GetContractContext(ctx, addr.Bytes())
		if contractContext == nil {
			// ! we return success here in case the contract does not exist
			success = int32(0)
		} else {
			req := vmtypes.CallRequestCommon{
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

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), responsebz)
	if err != nil {
		return returns, err
	}
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(responsebz)))
	return returns, nil
}

func wasiCallStatic(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)

	gasLimit := params[0].(int64)
	addrbz, err := rnh.GetMemory().Read(params[1], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}
	addr := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addrbz))
	calldata, err := rnh.GetMemory().Read(params[2], params[3])
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}

	var success int32
	var returnData []byte

	contractContext := GetContractContext(ctx, addr.Bytes())
	if contractContext == nil {
		// ! we return success here in case the contract does not exist
		success = int32(0)
	} else {
		req := vmtypes.CallRequestCommon{
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

	response := vmtypes.CallResponse{
		Success: uint8(success),
		Data:    returnData,
	}
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), responsebz)
	if err != nil {
		return returns, err
	}
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(responsebz)))
	return returns, nil
}

// address -> account
func wasi_getAccount(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	addr, err := rnh.GetMemory().Read(params[0], int32(32))
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
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), codebz)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(codebz)))
	return returns, nil
}

func wasi_keccak256(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	data, err := rnh.GetMemory().Read(params[0], params[1])
	if err != nil {
		return nil, err
	}
	if ctx.ContractRouter["keccak256"] == nil {
		return nil, fmt.Errorf("missing keccak256 wasm module")
	}
	keccakVm := ctx.ContractRouter["keccak256"].Vm
	input_offset := int32(0)
	input_length := int32(len(data))
	output_offset := input_length
	context_offset := output_offset + int32(32)

	keccakMem := keccakVm.GetActiveModule().FindMemory("memory")
	if keccakMem == nil {
		return nil, fmt.Errorf("missing keccak256 wasm memory")
	}
	err = keccakMem.SetData(data, uint(input_offset), uint(input_length))
	if err != nil {
		return nil, err
	}

	_, err = keccakVm.Execute("keccak", context_offset, input_offset, input_length, output_offset)
	if err != nil {
		return nil, err
	}
	result, err := keccakMem.GetData(uint(output_offset), uint(32))
	if err != nil {
		return nil, err
	}
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), result)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(result)))
	return returns, nil
}

// getBalance(address): i256
func wasi_getBalance(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	addr, err := rnh.GetMemory().Read(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	address := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addr))
	balance, err := BankGetBalance(ctx, address, ctx.Env.Chain.Denom)
	if err != nil {
		return nil, err
	}
	balancebz := balance.Amount.BigInt().FillBytes(make([]byte, 32))
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), balancebz)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(balancebz)))
	return returns, nil
}

// getBlockHash(i64): bytes32
func wasi_getBlockHash(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	blockNumber := params[0].(int64)
	data := ctx.CosmosHandler.GetBlockHash(uint64(blockNumber))
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), data)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(data)))
	return returns, nil
}

// create(bytecode, balance): address
// instantiateAccount(codeId, msgInit, balance): address
func wasi_instantiateAccount(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	codeId := params[0].(int64)
	initMsg, err := rnh.GetMemory().Read(params[1], params[2])
	if err != nil {
		return nil, err
	}
	balance, err := memc.ReadBigInt(rnh.GetMemory(), params[3], params[4])
	if err != nil {
		return nil, err
	}
	// TODO label?
	contractAddress, err := ctx.CosmosHandler.Create(uint64(codeId), ctx.Env.Contract.Address, initMsg, "", balance, nil)
	if err != nil {
		return nil, err
	}
	contractbz := memc.PaddLeftTo32(contractAddress.Bytes())
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), contractbz)
	if err != nil {
		return nil, err
	}
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(contractbz)))
	return returns, nil
}

// create2(address, salt, bytes): address
// instantiateAccount(codeId, salt, msgInit, balance): address
func wasi_instantiateAccount2(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	codeId := params[0].(int64)
	salt, err := rnh.GetMemory().Read(params[1], int32(32))
	if err != nil {
		return nil, err
	}
	initMsg, err := rnh.GetMemory().Read(params[2], params[3])
	if err != nil {
		return nil, err
	}
	balance, err := memc.ReadBigInt(rnh.GetMemory(), params[4], params[5])
	if err != nil {
		return nil, err
	}
	// TODO label?
	contractAddress, err := ctx.CosmosHandler.Create2(uint64(codeId), ctx.Env.Contract.Address, initMsg, salt, "", balance, nil)
	if err != nil {
		return nil, err
	}
	contractbz := memc.PaddLeftTo32(contractAddress.Bytes())
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), contractbz)
	if err != nil {
		return nil, err
	}
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(contractbz)))
	return returns, nil
}

// getCodeHash(address): bytes32
func wasi_getCodeHash(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	addr, err := rnh.GetMemory().Read(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	checksum := ctx.CosmosHandler.GetCodeHash(address)
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), checksum)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(checksum)))
	return returns, nil
}

// getCode(address): []byte
func wasi_getCode(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	addr, err := rnh.GetMemory().Read(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	address := sdk.AccAddress(vmtypes.CleanupAddress(addr))
	code := ctx.CosmosHandler.GetCode(address)
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), code)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(code)))
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
	returns[0] = 0
	return returns, nil
}

// bech32StringToBytes(ptr, len): i64
func wasi_bech32StringToBytes(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	addrbz, err := rnh.GetMemory().Read(params[0], params[1])
	if err != nil {
		return nil, err
	}
	addr, err := ctx.CosmosHandler.AddressCodec().StringToBytes(string(addrbz))
	if err != nil {
		return nil, err
	}
	data := types.PaddLeftTo32(addr)
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), data)
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
	addrbz, err := rnh.GetMemory().Read(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrbz))

	addrstr, err := ctx.CosmosHandler.AddressCodec().BytesToString(addr)
	if err != nil {
		return nil, err
	}

	data := []byte(addrstr)
	ptr, err := wasimem.WriteMemDefaultMalloc(rnh.GetVm(), rnh.GetMemory(), data)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = wasimem.BuildPtr64(ptr, int32(len(data)))
	return returns, nil
}

func wasiLog(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	data, err := rnh.GetMemory().Read(params[0], params[1])
	if err != nil {
		return nil, err
	}
	// // TODO show in debug mode
	// fmt.Println("LOG: ", string(data))
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
		vm.BuildFn("setReturnData", wasmxSetFinishData, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("setExitCode", wasiSetExitCode, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("callClassic", wasiCallClassic, []interface{}{vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("callStatic", wasiCallStatic, []interface{}{vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getBlockHash", wasi_getBlockHash, []interface{}{vm.ValType_I64()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getAccount", wasi_getAccount, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getBalance", wasi_getBalance, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getCodeHash", wasi_getCodeHash, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("getCode", wasi_getCode, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("keccak256", wasi_keccak256, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),

		// env.AddFunction("createAccount", wasmedge.NewFunction(functype_i32i32i32i32_i64, wasi_createAccount, context, 0))
		// env.AddFunction("createAccount2", wasmedge.NewFunction(functype_i32i32i32i32i32_i64, wasi_createAccount2, context, 0))

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

func ExecuteWasi(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]interface{}, error) {
	var res []interface{}
	var err error

	// WASI standard does not have instantiate
	// this is only for wasmx contracts (e.g. compiled with tinygo, javy)
	// TODO consider extracting this in a dependency
	if funcName == types.ENTRY_POINT_INSTANTIATE {
		fnNames, _ := contractVm.GetFunctionList()
		found := false
		for _, name := range fnNames {
			// WASI reactor
			if name == "_initialize" {
				found = true
				res, err = contractVm.Execute("_initialize")
			}
			// note that custom entries do not have access to WASI endpoints at this time
			if name == "main.instantiate" {
				found = true
				res, err = contractVm.Execute("main.instantiate")
			}
		}
		if !found {
			return nil, nil
		}
	} else if funcName == types.ENTRY_POINT_TIMED || funcName == types.ENTRY_POINT_P2P_MSG {
		res, err = contractVm.Execute(funcName)
	} else {
		// WASI command - no args, no return
		res, err = contractVm.Execute("_start")
		// fmt.Println("--ExecuteWasi-_start, err", res, err)

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
	return res, nil
}

func ExecutePythonInterpreter(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]interface{}, error) {
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
from wasmx import set_finishdata

%s

res = b''
entrypoint = %s
if len(sys.argv) > 1 and sys.argv[1] != "":
	inputObject = sys.argv[1]
	input = json.loads(inputObject)
	res = entrypoint(input)
else:
	res = entrypoint()

set_finishdata(res or b'')
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
	return ExecuteWasi(context, contractVm, types.ENTRY_POINT_EXECUTE, make([]interface{}, 0))
}

func ExecuteJsInterpreter(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]interface{}, error) {
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
wasmx.setFinishData(res || new ArrayBuffer(0));

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
	return ExecuteWasi(context, contractVm, types.ENTRY_POINT_EXECUTE, make([]interface{}, 0))
}

func ExecuteWasiWrap(context *Context, contractVm memc.IVm, funcName string, args []interface{}) ([]interface{}, error) {
	// if funcName == "execute" || funcName == "query" {
	// 	funcName = "main"
	// }

	wasimodule := contractVm.GetImportModule(wasmedge.WASI)
	wasimodule.InitWasi(
		[]string{
			``,
		},
		// os.Environ(), // The envs
		[]string{},
		// The mapping preopens
		[]string{
			// fmt.Sprintf(`.:%s`, dir),
		},
	)
	return ExecuteWasi(context, contractVm, funcName, args)
}
