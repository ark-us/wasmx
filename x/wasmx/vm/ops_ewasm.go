package vm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	mem "mythos/v1/x/wasmx/vm/memory/common"
	vmtypes "mythos/v1/x/wasmx/vm/types"
)

var (
	SSTORE_GAS_EWASM = 20_000
	LOG_TYPE_EWASM   = "ewasm"
)

func useGas(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	gasToConsume := params[0].(int64)
	// panics with out of gas error when out of gas
	ctx.GasMeter.ConsumeGas(uint64(gasToConsume), "ewasm")
	return returns, wasmedge.Result_Success
}

// GAS -> i64
func getGasLeft(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int64(ctx.GasMeter.GasConsumed())
	return returns, wasmedge.Result_Success
}

// SLOAD key_ptr: i32, result_ptr: i32
func storageLoad(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	keybz, err := mem.ReadMem(callframe, params[0], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data := ctx.ContractStore.Get(keybz)
	if len(data) == 0 {
		data = types.EMPTY_BYTES32
	}
	err = mem.WriteMem(callframe, data, params[1])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// SSTORE key_ptr: i32, value_ptr: i32,
func storageStore(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	keybz, err := mem.ReadMem(callframe, params[0], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	valuebz, err := mem.ReadMem(callframe, params[1], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_EWASM), "ewasm")
	ctx.ContractStore.Set(keybz, valuebz)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// SELFBALANCE result_ptr: i32
func getBalance(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	balance, err := BankGetBalance(ctx, ctx.Env.Contract.Address, ctx.Env.Chain.Denom)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	mem.WriteBigInt(callframe, balance.Amount.BigInt(), params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// BALANCE value_ptr: i32, result_ptr: i32,
func getExternalBalance(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	addressbz, err := mem.ReadMem(callframe, params[0], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	balance, err := BankGetBalance(ctx, vmtypes.CleanupAddress(addressbz), ctx.Env.Chain.Denom)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	mem.WriteBigInt(callframe, balance.Amount.BigInt(), params[1])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// ADDRESS result_ptr: i32
func getAddress(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	addr := types.Evm32AddressFromAcc(ctx.Env.Contract.Address)
	err := mem.WriteMem(callframe, addr.Bytes(), params[0])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// CALLER result_ptr: i32
func getCaller(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	addr := types.Evm32AddressFromAcc(ctx.Env.CurrentCall.Sender)
	err := mem.WriteMem(callframe, addr.Bytes(), params[0])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// CALLVALUE  result_ptr: i32
func getCallValue(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	mem.WriteBigInt(callframe, ctx.Env.CurrentCall.Funds, params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// CALLDATASIZE -> i32
func getCallDataSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = len(ctx.Env.CurrentCall.CallData)
	return returns, wasmedge.Result_Success
}

// CALLDATACOPY result_ptr: i32, data_ptr: i32, data_len: i32,
func callDataCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	ctx := context.(*Context)
	dataStart := params[1].(int32)
	dataLen := params[2].(int32)
	part := mem.ReadAndFillWithZero(ctx.Env.CurrentCall.CallData, dataStart, dataLen)
	err := mem.WriteMem(callframe, part, params[0])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// RETURNDATASIZE -> i32
func getReturnDataSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = len(ctx.ReturnData)
	return returns, wasmedge.Result_Success
}

// RETURNDATACOPY result_ptr: i32, data_ptr: i32, data_len: i32
func returnDataCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	dataStart := params[1].(int32)
	dataLen := params[2].(int32)
	part := ctx.ReturnData[dataStart:dataLen]
	err := mem.WriteMem(callframe, part, params[0])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// CODESIZE -> i32
func getCodeSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = len(ctx.Env.Contract.Bytecode)
	return returns, wasmedge.Result_Success
}

// EXTCODESIZE address_ptr: i32 -> i32
func getExternalCodeSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	returns[0] = int32(100000)
	return returns, wasmedge.Result_Success
}

// CODECOPY result_ptr: i32, code_ptr: i32, data_len: i32
// works only for constructor args that need to be copied at deployment time
func codeCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	codePtr := params[1].(int32)
	dataLen := params[2].(int32)
	part := mem.ReadAndFillWithZero(ctx.Env.Contract.Bytecode, codePtr, dataLen)
	err := mem.WriteMem(callframe, part, params[0])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// EXTCODECOPY address_ptr: i32, result_ptr: i32, code_ptr: i32, data_len: i32
func externalCodeCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// EXTCODEHASH address_ptr: i32, result_ptr: i32
func getExternalCodeHash(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	addressbz, err := mem.ReadMem(callframe, params[0], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data := ctx.CosmosHandler.GetCodeHash(vmtypes.CleanupAddress(addressbz))
	err = mem.WriteMem(callframe, data, params[1])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// GASPRICE result_ptr: i32
func getTxGasPrice(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	mem.WriteBigInt(callframe, ctx.Env.Transaction.GasPrice, params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// ORIGIN result_ptr: i32
func getTxOrigin(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	addr := types.Evm32AddressFromAcc(ctx.Env.CurrentCall.Origin)
	err := mem.WriteMem(callframe, addr.Bytes(), params[0])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// NUMBER -> i64
func getBlockNumber(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int64(ctx.Env.Block.Height)
	return returns, wasmedge.Result_Success
}

// COINBASE result_ptr: i32
func getBlockCoinbase(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	addr := types.Evm32AddressFromAcc(ctx.Env.Block.Proposer)
	err := mem.WriteMem(callframe, addr.Bytes(), params[0])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// BLOCKHASH block_number: i64, result_ptr: i32
func getBlockHash(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	blockNumber := params[0].(int64)
	data := ctx.CosmosHandler.GetBlockHash(uint64(blockNumber))
	err := mem.WriteMem(callframe, data, params[1])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// GASLIMIT -> i64
func getBlockGasLimit(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int64(ctx.Env.Block.GasLimit)
	return returns, wasmedge.Result_Success
}

// TIMESTAMP -> i64
func getBlockTimestamp(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	// EVM time is in seconds since unix epoch
	// ctx.Env.Block.Time is in nanoseconds
	timestamp := time.Unix(0, int64(ctx.Env.Block.Timestamp))
	returns[0] = timestamp.Unix()
	return returns, wasmedge.Result_Success
}

// DIFFICULTY result_ptr: i32
func getBlockDifficulty(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	data := types.EMPTY_BYTES32
	returns := make([]interface{}, 0)
	err := mem.WriteMem(callframe, data, params[0])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// PREVRANDAO result_ptr: i32
func prevrandao(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// TODO random
	data := types.EMPTY_BYTES32
	returns := make([]interface{}, 0)
	err := mem.WriteMem(callframe, data, params[0])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// CHAINID result_ptr: i32
func getChainId(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	mem.WriteBigInt(callframe, ctx.Env.Chain.ChainId, params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// BASEFEE result_ptr: i32
func getBaseFee(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	data := types.EMPTY_BYTES32
	returns := make([]interface{}, 0)
	err := mem.WriteMem(callframe, data, params[0])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// CALL gas_limit: i64, address_ptr: i32, value_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
// Returns 0 on success, 1 on failure and 2 on revert
func call(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)

	gasLimit := params[0].(int64)
	addrbz, err := mem.ReadMem(callframe, params[1], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrbz))
	value, err := mem.ReadBigInt(callframe, params[2], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	calldata, err := mem.ReadMem(callframe, params[3], params[4])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
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
		contractContext := GetContractContext(ctx, addr)
		if contractContext == nil {
			// ! we return success here in case the contract does not exist
			success = int32(0)
		} else {
			req := vmtypes.CallRequest{
				To:       addr,
				From:     ctx.Env.Contract.Address,
				Value:    value,
				GasLimit: big.NewInt(gasLimit),
				Calldata: calldata,
				Bytecode: contractContext.ContractInfo.Bytecode,
				CodeHash: contractContext.ContractInfo.CodeHash,
				IsQuery:  false,
			}
			success, returnData = WasmxCall(ctx, req)
			ctx.ReturnData = returnData
		}
	}
	returns[0] = success
	err = mem.WriteMemBoundBySize(callframe, returnData, params[5], params[6])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	return returns, wasmedge.Result_Success
}

// CALLCODE gas_limit: i64, address_ptr: i32, value_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
func callCode(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)

	gasLimit := params[0].(int64)
	addrbz, err := mem.ReadMem(callframe, params[1], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrbz))
	value, err := mem.ReadBigInt(callframe, params[2], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	calldata, err := mem.ReadMem(callframe, params[3], params[4])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}

	// We don't need to send funds because it would be sending funds
	// from the current contract to itself
	var success int32
	var returnData []byte
	contractContext := GetContractContext(ctx, addr)
	if contractContext == nil {
		// ! we return success here in case the contract does not exist
		success = int32(0)
	} else {
		req := vmtypes.CallRequest{
			To:       ctx.Env.Contract.Address,
			From:     ctx.Env.Contract.Address,
			Value:    value,
			GasLimit: big.NewInt(gasLimit),
			Calldata: calldata,
			Bytecode: contractContext.ContractInfo.Bytecode,
			CodeHash: contractContext.ContractInfo.CodeHash,
			IsQuery:  false,
		}
		success, returnData = WasmxCall(ctx, req)
		ctx.ReturnData = returnData
	}
	returns[0] = success

	err = mem.WriteMemBoundBySize(callframe, returnData, params[5], params[6])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	return returns, wasmedge.Result_Success
}

// CALLDELEGATE gas_limit: i64, address_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
func callDelegate(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 1)

	gasLimit := params[0].(int64)
	addrbz, err := mem.ReadMem(callframe, params[1], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrbz))
	calldata, err := mem.ReadMem(callframe, params[2], params[3])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}

	// We don't need to send funds because it would be sending funds
	// from the current contract to itself
	var success int32
	var returnData []byte
	contractContext := GetContractContext(ctx, addr)
	if contractContext == nil {
		// ! we return success here in case the contract does not exist
		success = int32(0)
	} else {
		req := vmtypes.CallRequest{
			To:       ctx.Env.Contract.Address,
			From:     ctx.Env.CurrentCall.Sender,
			Value:    ctx.Env.CurrentCall.Funds,
			GasLimit: big.NewInt(gasLimit),
			Calldata: calldata,
			Bytecode: contractContext.ContractInfo.Bytecode,
			CodeHash: contractContext.ContractInfo.CodeHash,
			IsQuery:  false,
		}
		success, returnData = WasmxCall(ctx, req)
		ctx.ReturnData = returnData
	}
	returns[0] = success

	err = mem.WriteMemBoundBySize(callframe, returnData, params[4], params[5])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	return returns, wasmedge.Result_Success
}

// STATICCALL gas_limit: i64, address_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
// Returns 0 on success, 1 on failure and 2 on revert
func callStatic(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// TODO static
	ctx := context.(*Context)
	returns := make([]interface{}, 1)

	gasLimit := params[0].(int64)
	addrbz, err := mem.ReadMem(callframe, params[1], int32(32))
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	addr := sdk.AccAddress(vmtypes.CleanupAddress(addrbz))
	calldata, err := mem.ReadMem(callframe, params[2], params[3])
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
			To:       addr,
			From:     ctx.Env.Contract.Address,
			Value:    big.NewInt(0),
			GasLimit: big.NewInt(gasLimit),
			Calldata: calldata,
			Bytecode: contractContext.ContractInfo.Bytecode,
			CodeHash: contractContext.ContractInfo.CodeHash,
			IsQuery:  true,
		}
		success, returnData = WasmxCall(ctx, req)
		ctx.ReturnData = returnData
	}
	returns[0] = success

	err = mem.WriteMemBoundBySize(callframe, returnData, params[4], params[5])
	if err != nil {
		returns[0] = int32(1)
		return returns, wasmedge.Result_Success
	}
	return returns, wasmedge.Result_Success
}

// CREATE value_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32
func create(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	value, err := mem.ReadBigInt(callframe, params[0], int32(32))
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	data, err := mem.ReadMem(callframe, params[1], params[2])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	metadata := types.CodeMetadata{}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	var sdeps []string
	contractstr, err := ctx.CosmosHandler.AddressCodec().BytesToString(ctx.Env.Contract.Address)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	for _, dep := range ctx.ContractRouter[contractstr].ContractInfo.SystemDeps {
		sdeps = append(sdeps, dep.Label)
	}
	_, _, contractAddress, err := ctx.CosmosHandler.Deploy(
		data,
		ctx.Env.CurrentCall.Origin,
		ctx.Env.Contract.Address,
		initMsg,
		value,
		sdeps,
		metadata,
		"", // TODO label?
		[]byte{},
	)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	err = mem.WriteMem(callframe, mem.PaddLeftTo32(contractAddress.Bytes()), params[3])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// CREATE2 value_ptr: i32, data_ptr: i32, data_len: i32, salt_ptr: i32, result_ptr: i32
func create2(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	returns := make([]interface{}, 0)
	value, err := mem.ReadBigInt(callframe, params[0], int32(32))
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	data, err := mem.ReadMem(callframe, params[1], params[2])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	salt, err := mem.ReadMem(callframe, params[3], int32(32))
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	metadata := types.CodeMetadata{}
	// TODO info from provenance ?
	initMsg, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	var sdeps []string
	contractstr, err := ctx.CosmosHandler.AddressCodec().BytesToString(ctx.Env.Contract.Address)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}

	for _, dep := range ctx.ContractRouter[contractstr].ContractInfo.SystemDeps {
		sdeps = append(sdeps, dep.Label)
	}
	_, _, contractAddress, err := ctx.CosmosHandler.Deploy(
		data,
		ctx.Env.CurrentCall.Origin,
		ctx.Env.Contract.Address,
		initMsg,
		value,
		sdeps,
		metadata,
		"", // TODO label?
		salt,
	)
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	err = mem.WriteMem(callframe, mem.PaddLeftTo32(contractAddress.Bytes()), params[3])
	if err != nil {
		return returns, wasmedge.Result_Fail
	}
	return returns, wasmedge.Result_Success
}

// SELFDESTRUCT address_ptr: i32
func selfDestruct(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// LOG data_ptr: i32, data_len: i32, topic_count: i32, topic_ptr1: i32, topic_ptr2: i32, topic_ptr3: i32, topic_ptr4: i32
func ewasmLog(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	data, err := mem.ReadMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	dependency := types.DEFAULT_SYS_DEP
	if len(ctx.Env.Contract.SystemDeps) > 0 {
		dependency = ctx.Env.Contract.SystemDeps[0]
	}

	log := WasmxLog{Type: LOG_TYPE_EWASM, Data: data, ContractAddress: ctx.Env.Contract.Address, SystemDependency: dependency}
	topicCount := int(params[2].(int32))
	topicPtrs := []interface{}{params[3], params[4], params[5], params[6]}

	for i := 0; i < topicCount; i++ {
		topic, err := mem.ReadMem(callframe, topicPtrs[i], int32(32))
		if err != nil {
			return nil, wasmedge.Result_Fail
		}
		var topic_ [32]byte
		copy(topic_[:], topic)
		log.Topics = append(log.Topics, topic_)
	}
	ctx.Logs = append(ctx.Logs, log)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// RETURN data_ptr: i32, data_len: i32
func finish(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	result, err := mem.ReadMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = result
	ctx.FinishData = result
	ctx.ReturnData = result
	// terminate the WASM execution
	return returns, wasmedge.Result_Terminate
}

// STOP data_ptr: i32, data_len: i32
func stop(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Terminate
}

// REVERT data_ptr: i32, data_len: i32
func revert(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	result, err := mem.ReadMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = result
	ctx.FinishData = result
	ctx.ReturnData = result
	return returns, wasmedge.Result_Fail
}

// msg_ptr: i32, _msg_len: i32
func sendCosmosMsg(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// msg_ptr: i32, _msg_len: i32
func sendCosmosQuery(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// value: i32
func debugPrinti32(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	ctx.Logger(ctx.Ctx).Debug(fmt.Sprintf("Go: debugPrinti32: %d, %d", params[0].(int32), params[1].(int32)))
	returns := make([]interface{}, 1)
	returns[0] = params[0]
	return returns, wasmedge.Result_Success
}

// value: i64
func debugPrinti64(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	ctx := context.(*Context)
	ctx.Logger(ctx.Ctx).Debug(fmt.Sprintf("Go: debugPrinti64: %d, %d", params[0].(int64), params[1].(int32)))
	returns := make([]interface{}, 1)
	returns[0] = params[0]
	return returns, wasmedge.Result_Success
}

// value_ptr: i32, value_len: i32
func debugPrintMemHex(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	pointer := params[0].(int32)
	size := params[1].(int32)
	mem := callframe.GetMemoryByIndex(0)
	data, _ := mem.GetData(uint(pointer), uint(size))
	ctx := context.(*Context)
	ctx.Logger(ctx.Ctx).Debug(fmt.Sprintf("Go: debugPrintMemHex: %s", hex.EncodeToString(data)))
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

func BuildEwasmEnv(context *Context) *wasmedge.Module {
	ewasmEnv := wasmedge.NewModule("env")

	functype__ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{},
	)
	functype_i64_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64},
		[]wasmedge.ValType{},
	)
	functype__i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype__i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i64i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i64i32_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)
	functype_i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i64i32i32i32i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i64i32i32i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I64, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32i32i32i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)

	ewasmEnv.AddFunction("ethereum_useGas", wasmedge.NewFunction(functype_i64_, useGas, context, 0))
	ewasmEnv.AddFunction("ethereum_getGasLeft", wasmedge.NewFunction(functype__i64, getGasLeft, context, 0))
	ewasmEnv.AddFunction("ethereum_storageLoad", wasmedge.NewFunction(functype_i32i32_, storageLoad, context, 0))
	ewasmEnv.AddFunction("ethereum_storageStore", wasmedge.NewFunction(functype_i32i32_, storageStore, context, 0))
	ewasmEnv.AddFunction("ethereum_getBalance", wasmedge.NewFunction(functype_i32_, getBalance, context, 0))
	ewasmEnv.AddFunction("ethereum_getExternalBalance", wasmedge.NewFunction(functype_i32i32_, getExternalBalance, context, 0))
	ewasmEnv.AddFunction("ethereum_getAddress", wasmedge.NewFunction(functype_i32_, getAddress, context, 0))
	ewasmEnv.AddFunction("ethereum_getCaller", wasmedge.NewFunction(functype_i32_, getCaller, context, 0))
	ewasmEnv.AddFunction("ethereum_getCallValue", wasmedge.NewFunction(functype_i32_, getCallValue, context, 0))
	ewasmEnv.AddFunction("ethereum_getCallDataSize", wasmedge.NewFunction(functype__i32, getCallDataSize, context, 0))
	ewasmEnv.AddFunction("ethereum_callDataCopy", wasmedge.NewFunction(functype_i32i32i32_, callDataCopy, context, 0))
	ewasmEnv.AddFunction("ethereum_getReturnDataSize", wasmedge.NewFunction(functype__i32, getReturnDataSize, context, 0))
	ewasmEnv.AddFunction("ethereum_returnDataCopy", wasmedge.NewFunction(functype_i32i32i32_, returnDataCopy, context, 0))
	ewasmEnv.AddFunction("ethereum_getCodeSize", wasmedge.NewFunction(functype__i32, getCodeSize, context, 0))
	ewasmEnv.AddFunction("ethereum_getExternalCodeSize", wasmedge.NewFunction(functype_i32_i32, getExternalCodeSize, context, 0))
	ewasmEnv.AddFunction("ethereum_codeCopy", wasmedge.NewFunction(functype_i32i32i32_, codeCopy, context, 0))
	ewasmEnv.AddFunction("ethereum_externalCodeCopy", wasmedge.NewFunction(functype_i32i32i32i32_, externalCodeCopy, context, 0))
	ewasmEnv.AddFunction("ethereum_getExternalCodeHash", wasmedge.NewFunction(functype_i32i32_, getExternalCodeHash, context, 0))
	ewasmEnv.AddFunction("ethereum_getTxGasPrice", wasmedge.NewFunction(functype_i32_, getTxGasPrice, context, 0))
	ewasmEnv.AddFunction("ethereum_getTxOrigin", wasmedge.NewFunction(functype_i32_, getTxOrigin, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockNumber", wasmedge.NewFunction(functype__i64, getBlockNumber, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockCoinbase", wasmedge.NewFunction(functype_i32_, getBlockCoinbase, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockHash", wasmedge.NewFunction(functype_i64i32_, getBlockHash, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockGasLimit", wasmedge.NewFunction(functype__i64, getBlockGasLimit, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockTimestamp", wasmedge.NewFunction(functype__i64, getBlockTimestamp, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockDifficulty", wasmedge.NewFunction(functype_i32_, getBlockDifficulty, context, 0))
	ewasmEnv.AddFunction("ethereum_prevrandao", wasmedge.NewFunction(functype_i32_, prevrandao, context, 0))
	ewasmEnv.AddFunction("ethereum_getChainId", wasmedge.NewFunction(functype_i32_, getChainId, context, 0))
	ewasmEnv.AddFunction("ethereum_getBaseFee", wasmedge.NewFunction(functype_i32_, getBaseFee, context, 0))
	ewasmEnv.AddFunction("ethereum_call", wasmedge.NewFunction(functype_i64i32i32i32i32i32i32_i32, call, context, 0))
	ewasmEnv.AddFunction("ethereum_callCode", wasmedge.NewFunction(functype_i64i32i32i32i32i32i32_i32, callCode, context, 0))
	ewasmEnv.AddFunction("ethereum_callDelegate", wasmedge.NewFunction(functype_i64i32i32i32i32i32_i32, callDelegate, context, 0))
	ewasmEnv.AddFunction("ethereum_callStatic", wasmedge.NewFunction(functype_i64i32i32i32i32i32_i32, callStatic, context, 0))
	ewasmEnv.AddFunction("ethereum_create", wasmedge.NewFunction(functype_i32i32i32i32_, create, context, 0))
	ewasmEnv.AddFunction("ethereum_create2", wasmedge.NewFunction(functype_i32i32i32i32i32_, create2, context, 0))
	ewasmEnv.AddFunction("ethereum_selfDestruct", wasmedge.NewFunction(functype_i32_, selfDestruct, context, 0))
	ewasmEnv.AddFunction("ethereum_log", wasmedge.NewFunction(functype_i32i32i32i32i32i32i32_, ewasmLog, context, 0))
	ewasmEnv.AddFunction("ethereum_finish", wasmedge.NewFunction(functype_i32i32_, finish, context, 0))
	ewasmEnv.AddFunction("ethereum_stop", wasmedge.NewFunction(functype__, stop, context, 0))
	ewasmEnv.AddFunction("ethereum_revert", wasmedge.NewFunction(functype_i32i32_, revert, context, 0))
	ewasmEnv.AddFunction("ethereum_sendCosmosMsg", wasmedge.NewFunction(functype_i32i32_i32, sendCosmosMsg, context, 0))
	ewasmEnv.AddFunction("ethereum_sendCosmosQuery", wasmedge.NewFunction(functype_i32i32_i32, sendCosmosQuery, context, 0))
	ewasmEnv.AddFunction("ethereum_debugPrinti32", wasmedge.NewFunction(functype_i32i32_i32, debugPrinti32, context, 0))
	ewasmEnv.AddFunction("ethereum_debugPrinti64", wasmedge.NewFunction(functype_i64i32_i64, debugPrinti64, context, 0))
	ewasmEnv.AddFunction("ethereum_debugPrintMemHex", wasmedge.NewFunction(functype_i32i32_, debugPrintMemHex, context, 0))

	return ewasmEnv
}
