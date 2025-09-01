package vm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm/types"
)

var (
	SSTORE_GAS_EWASM = 20_000
	LOG_TYPE_EWASM   = "ewasm"
)

func useGas(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	gasToConsume := params[0].(int64)
	// panics with out of gas error when out of gas
	ctx.GasMeter.ConsumeGas(uint64(gasToConsume), "ewasm")
	return returns, nil
}

// GAS -> i64
func getGasLeft(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int64(ctx.GasMeter.GasRemaining())
	return returns, nil
}

// SLOAD key_ptr: i32, result_ptr: i32
func storageLoad(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	keybz, err := mem.ReadRaw(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	data := ctx.ContractStore.Get(keybz)
	if len(data) == 0 {
		data = types.EMPTY_BYTES32
	}
	err = mem.WriteRaw(params[1], data)
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// SSTORE key_ptr: i32, value_ptr: i32,
func storageStore(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	keybz, err := mem.ReadRaw(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	valuebz, err := mem.ReadRaw(params[1], int32(32))
	if err != nil {
		return nil, err
	}
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_EWASM), "ewasm_storageStore")
	ctx.ContractStore.Set(keybz, valuebz)
	returns := make([]interface{}, 0)
	return returns, nil
}

// SELFBALANCE result_ptr: i32
func getBalance(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	balance, err := BankGetBalance(ctx, ctx.Env.Contract.Address, ctx.Env.Chain.Denom)
	if err != nil {
		return nil, err
	}
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = memc.WriteBigInt(mem, balance.Amount.BigInt(), params[0])
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 0)
	return returns, nil
}

// BALANCE value_ptr: i32, result_ptr: i32,
func getExternalBalance(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	addressbz, err := mem.ReadRaw(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	addr := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addressbz))
	balance, err := BankGetBalance(ctx, addr, ctx.Env.Chain.Denom)
	if err != nil {
		return nil, err
	}
	err = memc.WriteBigInt(mem, balance.Amount.BigInt(), params[1])
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 0)
	return returns, nil
}

// ADDRESS result_ptr: i32
func getAddress(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	addr := types.Evm32AddressFromAcc(ctx.Env.Contract.Address.Bytes())
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = mem.WriteRaw(params[0], addr.Bytes())
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// CALLER result_ptr: i32
func getCaller(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	addr := types.Evm32AddressFromAcc(ctx.Env.CurrentCall.Sender.Bytes())
	mem, err := rnh.GetMemory()
	if err != nil {
		return returns, err
	}
	err = mem.WriteRaw(params[0], addr.Bytes())
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// CALLVALUE  result_ptr: i32
func getCallValue(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	mem, err := rnh.GetMemory()
	if err != nil {
		return returns, err
	}
	err = memc.WriteBigInt(mem, ctx.Env.CurrentCall.Funds, params[0])
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// CALLDATASIZE -> i32
func getCallDataSize(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int32(len(ctx.Env.CurrentCall.CallData))
	return returns, nil
}

// CALLDATACOPY result_ptr: i32, data_ptr: i32, data_len: i32,
func callDataCopy(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	ctx := _context.(*Context)
	dataStart := params[1].(int32)
	dataLen := params[2].(int32)
	part := memc.ReadAndFillWithZero(ctx.Env.CurrentCall.CallData, dataStart, dataLen)
	mem, err := rnh.GetMemory()
	if err != nil {
		return returns, err
	}
	err = mem.WriteRaw(params[0], part)
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// RETURNDATASIZE -> i32
func getReturnDataSize(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int32(len(ctx.ReturnData))
	return returns, nil
}

// RETURNDATACOPY result_ptr: i32, data_ptr: i32, data_len: i32
func returnDataCopy(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	dataStart := params[1].(int32)
	dataLen := params[2].(int32)
	part := ctx.ReturnData[dataStart:dataLen]
	mem, err := rnh.GetMemory()
	if err != nil {
		return returns, err
	}
	err = mem.WriteRaw(params[0], part)
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// CODESIZE -> i32
func getCodeSize(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int32(len(ctx.Env.Contract.Bytecode))
	return returns, nil
}

// EXTCODESIZE address_ptr: i32 -> i32
func getExternalCodeSize(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	returns[0] = int32(100000)
	return returns, nil
}

// CODECOPY result_ptr: i32, code_ptr: i32, data_len: i32
// works only for constructor args that need to be copied at deployment time
func codeCopy(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	codePtr := params[1].(int32)
	dataLen := params[2].(int32)
	part := memc.ReadAndFillWithZero(ctx.Env.Contract.Bytecode, codePtr, dataLen)
	mem, err := rnh.GetMemory()
	if err != nil {
		return returns, err
	}
	err = mem.WriteRaw(params[0], part)
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// EXTCODECOPY address_ptr: i32, result_ptr: i32, code_ptr: i32, data_len: i32
func externalCodeCopy(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	return returns, nil
}

// EXTCODEHASH address_ptr: i32, result_ptr: i32
func getExternalCodeHash(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	addressbz, err := mem.ReadRaw(params[0], int32(32))
	if err != nil {
		return nil, err
	}
	addrPrefixed := ctx.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(vmtypes.CleanupAddress(addressbz))
	data := ctx.CosmosHandler.GetCodeHash(addrPrefixed)
	err = mem.WriteRaw(params[1], data)
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// GASPRICE result_ptr: i32
func getTxGasPrice(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = memc.WriteBigInt(mem, ctx.Env.Transaction.GasPrice, params[0])
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// ORIGIN result_ptr: i32
func getTxOrigin(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	addr := types.Evm32AddressFromAcc(ctx.Env.CurrentCall.Origin.Bytes())
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = mem.WriteRaw(params[0], addr.Bytes())
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// NUMBER -> i64
func getBlockNumber(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int64(ctx.Env.Block.Height)
	return returns, nil
}

// COINBASE result_ptr: i32
func getBlockCoinbase(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	addr := types.Evm32AddressFromAcc(ctx.Env.Block.Proposer.Bytes())
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = mem.WriteRaw(params[0], addr.Bytes())
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// BLOCKHASH block_number: i64, result_ptr: i32
func getBlockHash(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	blockNumber := params[0].(int64)
	data := ctx.CosmosHandler.GetBlockHash(uint64(blockNumber))
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = mem.WriteRaw(params[1], data)
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// GASLIMIT -> i64
func getBlockGasLimit(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int64(ctx.Env.Block.GasLimit)
	return returns, nil
}

// TIMESTAMP -> i64
func getBlockTimestamp(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 1)
	// EVM time is in seconds since unix epoch
	// ctx.Env.Block.Time is in nanoseconds
	timestamp := time.Unix(0, int64(ctx.Env.Block.Timestamp))
	returns[0] = timestamp.Unix()
	return returns, nil
}

// DIFFICULTY result_ptr: i32
func getBlockDifficulty(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	data := types.EMPTY_BYTES32
	returns := make([]interface{}, 0)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = mem.WriteRaw(params[0], data)
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// PREVRANDAO result_ptr: i32
func prevrandao(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	// TODO random
	data := types.EMPTY_BYTES32
	returns := make([]interface{}, 0)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = mem.WriteRaw(params[0], data)
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// CHAINID result_ptr: i32
func getChainId(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = memc.WriteBigInt(mem, ctx.Env.Chain.ChainId, params[0])
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// BASEFEE result_ptr: i32
func getBaseFee(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	data := types.EMPTY_BYTES32
	returns := make([]interface{}, 0)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	err = mem.WriteRaw(params[0], data)
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// CALL gas_limit: i64, address_ptr: i32, value_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
// Returns 0 on success, 1 on failure and 2 on revert
func call(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
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
				To:       addr,
				From:     ctx.Env.Contract.Address,
				Value:    value,
				GasLimit: big.NewInt(gasLimit),
				Calldata: calldata,
				Bytecode: contractInfo.Bytecode,
				CodeHash: contractInfo.CodeHash,
				IsQuery:  false,
			}
			success, returnData = WasmxCall(ctx, req)
			ctx.ReturnData = returnData
		}
	}
	returns[0] = success
	err = memc.WriteMemBoundBySize(mem, returnData, params[5], params[6])
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}
	return returns, nil
}

// CALLCODE gas_limit: i64, address_ptr: i32, value_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
func callCode(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
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
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrbz))
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

	// We don't need to send funds because it would be sending funds
	// from the current contract to itself
	var success int32
	var returnData []byte
	addrPrefixed := ctx.GetCosmosHandler().AccBech32Codec().BytesToAccAddressPrefixed(addr)
	contractInfo := GetContractDependency(ctx, addrPrefixed)
	if contractInfo == nil {
		// ! we return success here in case the contract does not exist
		success = int32(0)
	} else {
		req := vmtypes.CallRequestCommon{
			To:       ctx.Env.Contract.Address,
			From:     ctx.Env.Contract.Address,
			Value:    value,
			GasLimit: big.NewInt(gasLimit),
			Calldata: calldata,
			Bytecode: contractInfo.Bytecode,
			CodeHash: contractInfo.CodeHash,
			IsQuery:  false,
		}
		success, returnData = WasmxCall(ctx, req)
		ctx.ReturnData = returnData
	}
	returns[0] = success

	err = memc.WriteMemBoundBySize(mem, returnData, params[5], params[6])
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}
	return returns, nil
}

// CALLDELEGATE gas_limit: i64, address_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
func callDelegate(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
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
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrbz))
	calldata, err := mem.ReadRaw(params[2], params[3])
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}

	// We don't need to send funds because it would be sending funds
	// from the current contract to itself
	var success int32
	var returnData []byte
	addrPrefixed := ctx.GetCosmosHandler().AccBech32Codec().BytesToAccAddressPrefixed(addr)
	contractInfo := GetContractDependency(ctx, addrPrefixed)
	if contractInfo == nil {
		// ! we return success here in case the contract does not exist
		success = int32(0)
	} else {
		req := vmtypes.CallRequestCommon{
			To:       ctx.Env.Contract.Address,
			From:     ctx.Env.CurrentCall.Sender,
			Value:    ctx.Env.CurrentCall.Funds,
			GasLimit: big.NewInt(gasLimit),
			Calldata: calldata,
			Bytecode: contractInfo.Bytecode,
			CodeHash: contractInfo.CodeHash,
			IsQuery:  false,
		}
		success, returnData = WasmxCall(ctx, req)
		ctx.ReturnData = returnData
	}
	returns[0] = success

	err = memc.WriteMemBoundBySize(mem, returnData, params[4], params[5])
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}
	return returns, nil
}

// STATICCALL gas_limit: i64, address_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
// Returns 0 on success, 1 on failure and 2 on revert
func callStatic(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	// TODO static
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
			To:       addr,
			From:     ctx.Env.Contract.Address,
			Value:    big.NewInt(0),
			GasLimit: big.NewInt(gasLimit),
			Calldata: calldata,
			Bytecode: contractInfo.Bytecode,
			CodeHash: contractInfo.CodeHash,
			IsQuery:  true,
		}
		success, returnData = WasmxCall(ctx, req)
		ctx.ReturnData = returnData
	}
	returns[0] = success

	err = memc.WriteMemBoundBySize(mem, returnData, params[4], params[5])
	if err != nil {
		returns[0] = int32(1)
		return returns, nil
	}
	return returns, nil
}

// CREATE value_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32
func create(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	value, err := memc.ReadBigInt(mem, params[0], int32(32))
	if err != nil {
		return returns, err
	}
	data, err := mem.ReadRaw(params[1], params[2])
	if err != nil {
		return returns, err
	}
	metadata := types.CodeMetadata{}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, err
	}
	var sdeps []string
	contractstr := ctx.Env.Contract.Address.String()

	for _, dep := range ctx.ContractRouter[contractstr].ContractInfo.SystemDeps {
		sdeps = append(sdeps, dep.Label)
	}
	_, _, contractAddress, err := ctx.CosmosHandler.Deploy(
		data,
		&ctx.Env.CurrentCall.Origin,
		&ctx.Env.Contract.Address,
		initMsg,
		value,
		sdeps,
		metadata,
		"", // TODO label?
		[]byte{},
		[]byte{},
	)
	if err != nil {
		return returns, err
	}
	err = mem.WriteRaw(params[3], memc.PaddLeftTo32(contractAddress.Bytes()))
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// CREATE2 value_ptr: i32, data_ptr: i32, data_len: i32, salt_ptr: i32, result_ptr: i32
func create2(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	returns := make([]interface{}, 0)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	value, err := memc.ReadBigInt(mem, params[0], int32(32))
	if err != nil {
		return returns, err
	}
	data, err := mem.ReadRaw(params[1], params[2])
	if err != nil {
		return returns, err
	}
	salt, err := mem.ReadRaw(params[3], int32(32))
	if err != nil {
		return returns, err
	}
	metadata := types.CodeMetadata{}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, err
	}
	var sdeps []string
	contractstr := ctx.Env.Contract.Address.String()

	for _, dep := range ctx.ContractRouter[contractstr].ContractInfo.SystemDeps {
		sdeps = append(sdeps, dep.Label)
	}
	_, _, contractAddress, err := ctx.CosmosHandler.Deploy(
		data,
		&ctx.Env.CurrentCall.Origin,
		&ctx.Env.Contract.Address,
		initMsg,
		value,
		sdeps,
		metadata,
		"", // TODO label?
		salt,
		[]byte{},
	)
	if err != nil {
		return returns, err
	}
	err = mem.WriteRaw(params[3], memc.PaddLeftTo32(contractAddress.Bytes()))
	if err != nil {
		return returns, err
	}
	return returns, nil
}

// SELFDESTRUCT address_ptr: i32
func selfDestruct(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	return returns, nil
}

// LOG data_ptr: i32, data_len: i32, topic_count: i32, topic_ptr1: i32, topic_ptr2: i32, topic_ptr3: i32, topic_ptr4: i32
func ewasmLog(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	data, err := mem.ReadRaw(params[0], params[1])
	if err != nil {
		return nil, err
	}
	dependency := types.DEFAULT_SYS_DEP
	if len(ctx.Env.Contract.SystemDeps) > 0 {
		dependency = ctx.Env.Contract.SystemDeps[0]
	}

	log := WasmxLog{Type: LOG_TYPE_EWASM, Data: data, ContractAddress: ctx.Env.Contract.Address, SystemDependency: dependency}
	topicCount := int(params[2].(int32))
	topicPtrs := []interface{}{params[3], params[4], params[5], params[6]}

	for i := 0; i < topicCount; i++ {
		topic, err := mem.ReadRaw(topicPtrs[i], int32(32))
		if err != nil {
			return nil, err
		}
		var topic_ [32]byte
		copy(topic_[:], topic)
		log.Topics = append(log.Topics, topic_)
	}
	ctx.Logs = append(ctx.Logs, log)
	returns := make([]interface{}, 0)
	return returns, nil
}

// RETURN data_ptr: i32, data_len: i32
func finish(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	result, err := mem.ReadRaw(params[0], params[1])
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 0)
	ctx.FinishData = result
	ctx.ReturnData = result
	// terminate the WASM execution
	return returns, fmt.Errorf(memc.VM_TERMINATE_ERROR)
}

// STOP data_ptr: i32, data_len: i32
func stop(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	return returns, fmt.Errorf(memc.VM_TERMINATE_ERROR)
}

// REVERT data_ptr: i32, data_len: i32
func revert(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	result, err := mem.ReadRaw(params[0], params[1])
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 0)
	ctx.FinishData = result
	ctx.ReturnData = result
	return returns, fmt.Errorf("revert")
}

// msg_ptr: i32, _msg_len: i32
func sendCosmosMsg(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	return returns, nil
}

// msg_ptr: i32, _msg_len: i32
func sendCosmosQuery(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 0)
	return returns, nil
}

// value: i32
func debugPrinti32(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Logger(ctx.Ctx).Debug(fmt.Sprintf("Go: debugPrinti32: %d, %d", params[0].(int32), params[1].(int32)))
	returns := make([]interface{}, 1)
	returns[0] = params[0]
	return returns, nil
}

// value: i64
func debugPrinti64(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	ctx.Logger(ctx.Ctx).Debug(fmt.Sprintf("Go: debugPrinti64: %d, %d", params[0].(int64), params[1].(int32)))
	returns := make([]interface{}, 1)
	returns[0] = params[0]
	return returns, nil
}

// value_ptr: i32, value_len: i32
func debugPrintMemHex(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	pointer := params[0].(int32)
	size := params[1].(int32)
	mem, err := rnh.GetMemory()
	if err != nil {
		return nil, err
	}
	data, _ := mem.Read(pointer, size)
	ctx := _context.(*Context)
	ctx.Logger(ctx.Ctx).Debug(fmt.Sprintf("Go: debugPrintMemHex: %s", hex.EncodeToString(data)))
	returns := make([]interface{}, 0)
	return returns, nil
}

func BuildEwasmEnv(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("ethereum_useGas", useGas, []interface{}{vm.ValType_I64()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getGasLeft", getGasLeft, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ethereum_storageLoad", storageLoad, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_storageStore", storageStore, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getBalance", getBalance, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getExternalBalance", getExternalBalance, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getAddress", getAddress, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getCaller", getCaller, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getCallValue", getCallValue, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getCallDataSize", getCallDataSize, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ethereum_callDataCopy", callDataCopy, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getReturnDataSize", getReturnDataSize, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ethereum_returnDataCopy", returnDataCopy, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getCodeSize", getCodeSize, []interface{}{}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ethereum_getExternalCodeSize", getExternalCodeSize, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ethereum_codeCopy", codeCopy, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_externalCodeCopy", externalCodeCopy, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getExternalCodeHash", getExternalCodeHash, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getTxGasPrice", getTxGasPrice, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getTxOrigin", getTxOrigin, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getBlockNumber", getBlockNumber, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ethereum_getBlockCoinbase", getBlockCoinbase, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getBlockHash", getBlockHash, []interface{}{vm.ValType_I64(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getBlockGasLimit", getBlockGasLimit, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ethereum_getBlockTimestamp", getBlockTimestamp, []interface{}{}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ethereum_getBlockDifficulty", getBlockDifficulty, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_prevrandao", prevrandao, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getChainId", getChainId, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_getBaseFee", getBaseFee, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_call", call, []interface{}{vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ethereum_callCode", callCode, []interface{}{vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ethereum_callDelegate", callDelegate, []interface{}{vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ethereum_callStatic", callStatic, []interface{}{vm.ValType_I64(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ethereum_create", create, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_create2", create2, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_selfDestruct", selfDestruct, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_log", ewasmLog, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_finish", finish, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_stop", stop, []interface{}{}, []interface{}{}, 0),
		vm.BuildFn("ethereum_revert", revert, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("ethereum_sendCosmosMsg", sendCosmosMsg, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ethereum_sendCosmosQuery", sendCosmosQuery, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ethereum_debugPrinti32", debugPrinti32, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ethereum_debugPrinti64", debugPrinti64, []interface{}{vm.ValType_I64(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ethereum_debugPrintMemHex", debugPrintMemHex, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
	}

	return vm.BuildModule(rnh, "env", context, fndefs)
}
