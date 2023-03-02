package ewasm

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"wasmx/x/wasmx/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/second-state/WasmEdge-go/wasmedge"
)

var EMPTY_BYTES32 = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func useGas(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	returns := make([]interface{}, 1)
	return returns, wasmedge.Result_Success
}

// GAS -> i64
func getGasLeft(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getGasLeft")
	returns := make([]interface{}, 1)
	returns[0] = int64(1000000)
	return returns, wasmedge.Result_Success
}

// SLOAD key_ptr: i32, result_ptr: i32
func storageLoad(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: storageLoad")
	ctx := context.(*Context)
	keybz, err := readMem(callframe, params[0], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data := ctx.ContractStore.Get(keybz)
	writeMem(callframe, data, params[1])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// SSTORE key_ptr: i32, value_ptr: i32,
func storageStore(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: storageLoad")
	ctx := context.(*Context)
	keybz, err := readMem(callframe, params[0], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	valuebz, err := readMem(callframe, params[1], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.ContractStore.Set(keybz, valuebz)
	// fmt.Println("Go: storageLoad", keybz, valuebz)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// SELFBALANCE result_ptr: i32
func getBalance(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getBalance")
	data := EMPTY_BYTES32
	writeMem(callframe, data, params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// BALANCE value_ptr: i32, result_ptr: i32,
func getExternalBalance(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getExternalBalance")
	// addressbz, err := readMem(callframe, params[0], int32(32))
	// if err != nil {
	// 	return nil, wasmedge.Result_Fail
	// }
	data := EMPTY_BYTES32
	writeMem(callframe, data, params[1])
	// fmt.Println("Go: getExternalBalance", addressbz)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// ADDRESS result_ptr: i32
func getAddress(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getAddress")
	ctx := context.(*Context)
	// fmt.Println("Go: getAddress *sender", ctx.CallContext.Sender.String())
	addr := Evm32AddressFromAcc(ctx.Env.Contract.Address)
	// fmt.Println("Go: getAddress addr", addr, ctx.Env.Contract.Address.String(), addr.Bytes())
	// fmt.Println("Go: getAddress addr2", EvmAddressFromAcc(ctx.Env.Contract.Address))
	writeMem(callframe, addr.Bytes(), params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// CALLER result_ptr: i32
func getCaller(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getCaller")
	ctx := context.(*Context)
	addr := Evm32AddressFromAcc(ctx.CallContext.Sender)
	writeMem(callframe, addr.Bytes(), params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// CALLVALUE  result_ptr: i32
func getCallValue(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getCallValue")
	ctx := context.(*Context)
	// fmt.Println("Go: getCallValue", ctx.Callvalue)
	writeMem(callframe, ctx.Callvalue.FillBytes(make([]byte, 32)), params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// CALLDATASIZE -> i32
func getCallDataSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getCallDataSize")
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = len(ctx.Calldata)
	return returns, wasmedge.Result_Success
}

// CALLDATACOPY result_ptr: i32, data_ptr: i32, data_len: i32,
func callDataCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: callDataCopy")
	returns := make([]interface{}, 0)
	ctx := context.(*Context)
	// fmt.Println("Go: callDataCopy ctx.Calldata", ctx.Calldata)
	dataStart := params[1].(int32)
	dataLen := params[2].(int32)
	calldLen := int32(len(ctx.Calldata))
	var part []byte

	if dataStart >= calldLen {
		return returns, wasmedge.Result_Success
	}

	if dataStart+dataLen > calldLen {
		part = make([]byte, dataLen)
		copy(part, ctx.Calldata[dataStart:dataStart+calldLen])
	} else {
		part = ctx.Calldata[dataStart : dataStart+dataLen]
	}

	writeMem(callframe, part, params[0])
	return returns, wasmedge.Result_Success
}

// RETURNDATASIZE -> i32
func getReturnDataSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getReturnDataSize")
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = len(ctx.ReturnData)
	// fmt.Println("Go: getReturnDataSize", returns[0])
	return returns, wasmedge.Result_Success
}

// RETURNDATACOPY result_ptr: i32, data_ptr: i32, data_len: i32
func returnDataCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: returnDataCopy", params[1].(int32), params[2].(int32))
	ctx := context.(*Context)
	dataStart := params[1].(int32)
	dataLen := params[2].(int32)
	part := ctx.ReturnData[dataStart:dataLen]
	writeMem(callframe, part, params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// CODESIZE -> i32
func getCodeSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getCodeSize")
	returns := make([]interface{}, 1)
	returns[0] = int32(100000)
	return returns, wasmedge.Result_Success
}

// EXTCODESIZE address_ptr: i32 -> i32
func getExternalCodeSize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getExternalCodeSize")
	// addressbz, err := readMem(callframe, params[0], int32(32))
	// if err != nil {
	// 	return nil, wasmedge.Result_Fail
	// }
	// fmt.Println("Go: getExternalCodeSize", addressbz)
	returns := make([]interface{}, 1)
	returns[0] = int32(100000)
	return returns, wasmedge.Result_Success
}

// CODECOPY result_ptr: i32, code_ptr: i32, data_len: i32
func codeCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: codeCopy")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// EXTCODECOPY address_ptr: i32, result_ptr: i32, code_ptr: i32, data_len: i32
func externalCodeCopy(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: externalCodeCopy")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// EXTCODEHASH address_ptr: i32, result_ptr: i32
func getExternalCodeHash(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getExternalCodeHash")
	// addressbz, err := readMem(callframe, params[0], int32(32))
	// if err != nil {
	// 	return nil, wasmedge.Result_Fail
	// }
	data := EMPTY_BYTES32
	writeMem(callframe, data, params[1])
	// fmt.Println("Go: getExternalCodeHash", addressbz)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// GASPRICE result_ptr: i32
func getTxGasPrice(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getTxGasPrice")
	data := EMPTY_BYTES32
	writeMem(callframe, data, params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// ORIGIN result_ptr: i32
func getTxOrigin(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getTxOrigin")
	ctx := context.(*Context)
	addr := Evm32AddressFromAcc(ctx.CallContext.Origin)
	writeMem(callframe, addr.Bytes(), params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// NUMBER -> i64
func getBlockNumber(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getBlockNumber")
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int64(ctx.Env.Block.Height)
	return returns, wasmedge.Result_Success
}

// COINBASE result_ptr: i32
func getBlockCoinbase(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getBlockCoinbase")
	ctx := context.(*Context)
	addr := Evm32AddressFromAcc(ctx.Env.Block.Proposer)
	writeMem(callframe, addr.Bytes(), params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// BLOCKHASH block_number: i64, result_ptr: i32
func getBlockHash(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getBlockHash")
	// blockNumber, err := readI64(callframe, params[0], int32(32))
	// if err != nil {
	// 	return nil, wasmedge.Result_Fail
	// }
	// fmt.Println("Go: getBlockHash", blockNumber)
	data := EMPTY_BYTES32
	writeMem(callframe, data, params[1])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// GASLIMIT -> i64
func getBlockGasLimit(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getBlockGasLimit")
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int64(ctx.Env.Block.GasLimit)
	return returns, wasmedge.Result_Success
}

// TIMESTAMP -> i64
func getBlockTimestamp(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getBlockTimestamp")
	ctx := context.(*Context)
	returns := make([]interface{}, 1)
	returns[0] = int64(ctx.Env.Block.Time)
	return returns, wasmedge.Result_Success
}

// DIFFICULTY result_ptr: i32
func getBlockDifficulty(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getBlockDifficulty")
	data := EMPTY_BYTES32
	writeMem(callframe, data, params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// CHAINID result_ptr: i32
func getChainId(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getChainId")
	ctx := context.(*Context)
	data := ctx.Env.Chain.ChainId.FillBytes(make([]byte, 32))
	writeMem(callframe, data, params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// BASEFEE result_ptr: i32
func getBaseFee(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: getBaseFee")
	data := EMPTY_BYTES32
	writeMem(callframe, data, params[0])
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// CALL gas_limit: i64, address_ptr: i32, value_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
func call(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: call")
	ctx := context.(*Context)
	returns := make([]interface{}, 1)

	addrbz, err := readMem(callframe, params[1], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// fmt.Println("Go: call addrbz", addrbz)
	addr := sdk.AccAddress(cleanupAddress(addrbz))
	// fmt.Println("Go: call addr", addr, addr.String())
	value, err := readBigInt(callframe, params[2], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// fmt.Println("Go: call value", value, valuebz)
	calldata, err := readMem(callframe, params[3], params[4])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// fmt.Println("Go: call calldata", calldata)

	// fmt.Println("---ctx.ContractRouter", ctx.ContractRouter)

	_, ok := ctx.ContractRouter[addr.String()]
	// fmt.Println("---ctx.ContractRouter ok", ok)
	if !ok {
		return nil, wasmedge.Result_Fail
	}

	callContext := types.MessageInfo{
		Origin:   ctx.CallContext.Origin,
		Sender:   ctx.Env.Contract.Address,
		Funds:    value,
		IsQuery:  false,
		ReadOnly: false,
	}
	// fmt.Println("--callContext sender", callContext.Sender.String())

	newctx := &Context{
		Callvalue:      value,
		Calldata:       calldata,
		ContractRouter: ctx.ContractRouter,
		CallContext:    callContext,
		Env: &types.Env{
			Block:       ctx.Env.Block,
			Transaction: ctx.Env.Transaction,
			Chain:       ctx.Env.Chain,
			Contract: types.EnvContractInfo{
				Address: addr,
			},
		},
	}

	// ctx.ContractRouter[addr.String()].Context.Callvalue = value
	// ctx.ContractRouter[addr.String()].Context.Calldata = calldata
	// ctx.ContractRouter[addr.String()].Context.ContractRouter = ctx.ContractRouter
	// ctx.ContractRouter[addr.String()].Context.CallContext = callContext
	// ctx.ContractRouter[addr.String()].Context.Env.Contract.Address = addr

	_, err = ctx.ContractRouter[addr.String()].Execute(newctx)
	// fmt.Println("Go: call result err", err)
	if err != nil {
		returns[0] = int32(0)
	} else {
		returns[0] = int32(1) // TODO revert with 2
	}
	ctx.ReturnData = newctx.ReturnData
	// fmt.Println("Go: call result ReturnData", ctx.ReturnData)

	returns[0] = int32(0)
	return returns, wasmedge.Result_Success
}

// CALLCODE gas_limit: i64, address_ptr: i32, value_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
func callCode(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: callCode")
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, wasmedge.Result_Success
}

// CALLDELEGATE gas_limit: i64, address_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
func callDelegate(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: callDelegate")
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, wasmedge.Result_Success
}

// STATICCALL gas_limit: i64, address_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32, result_len: i32 -> i32
func callStatic(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: callStatic")
	// TODO static
	ctx := context.(*Context)
	returns := make([]interface{}, 1)

	addrbz, err := readMem(callframe, params[1], int32(32))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// fmt.Println("Go: call addrbz", addrbz)
	addr := sdk.AccAddress(cleanupAddress(addrbz))
	// fmt.Println("Go: call addr", addr, addr.String())
	calldata, err := readMem(callframe, params[2], params[3])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// fmt.Println("Go: call calldata", calldata)
	// fmt.Println("---ctx.ContractRouter", ctx.ContractRouter)

	_, ok := ctx.ContractRouter[addr.String()]
	// fmt.Println("---ctx.ContractRouter ok", ok)
	if !ok {
		return nil, wasmedge.Result_Fail
	}

	callContext := types.MessageInfo{
		Origin:   ctx.CallContext.Origin,
		Sender:   ctx.Env.Contract.Address,
		IsQuery:  false,
		ReadOnly: true,
	}
	// fmt.Println("--callContext sender", callContext.Sender.String())

	newctx := &Context{
		Callvalue:      big.NewInt(0),
		Calldata:       calldata,
		ContractRouter: ctx.ContractRouter,
		CallContext:    callContext,
		Env: &types.Env{
			Block:       ctx.Env.Block,
			Transaction: ctx.Env.Transaction,
			Chain:       ctx.Env.Chain,
			Contract: types.EnvContractInfo{
				Address: addr,
			},
		},
	}
	_, err = ctx.ContractRouter[addr.String()].Execute(newctx)
	// fmt.Println("Go: call result err", err)
	if err != nil {
		returns[0] = int32(0)
	} else {
		returns[0] = int32(1) // TODO revert with 2
	}
	ctx.ReturnData = newctx.ReturnData
	// fmt.Println("Go: call result ReturnData", ctx.ReturnData)

	returns[0] = int32(0)
	return returns, wasmedge.Result_Success
}

// CREATE value_ptr: i32, data_ptr: i32, data_len: i32, result_ptr: i32
func create(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: create")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// CREATE2 value_ptr: i32, data_ptr: i32, data_len: i32, salt_ptr: i32, result_ptr: i32
func create2(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: create2")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// SELFDESTRUCT address_ptr: i32
func selfDestruct(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: selfDestruct")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// LOG data_ptr: i32, data_len: i32, topic_count: i32, topic_ptr1: i32, topic_ptr2: i32, topic_ptr3: i32, topic_ptr4: i32
func log(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: log")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// RETURN data_ptr: i32, data_len: i32
func finish(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: finish")
	ctx := context.(*Context)
	result, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = result
	ctx.ReturnData = result
	// fmt.Println("Go: finish", result)
	return returns, wasmedge.Result_Success
}

// STOP data_ptr: i32, data_len: i32
func stop(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: stop")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// REVERT data_ptr: i32, data_len: i32
func revert(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: revert")
	ctx := context.(*Context)
	result, err := readMem(callframe, params[0], params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = result
	ctx.ReturnData = result
	// fmt.Println("Go: revert", result)
	return returns, wasmedge.Result_Fail
}

// msg_ptr: i32, _msg_len: i32
func sendCosmosMsg(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: sendCosmosMsg")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// msg_ptr: i32, _msg_len: i32
func sendCosmosQuery(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// fmt.Println("Go: sendCosmosQuery")
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// value: i32
func debugPrinti32(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: debugPrinti32", params[0].(int32))
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// value: i64
func debugPrinti64(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("Go: debugPrinti64", params[0].(int64))
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// value_ptr: i32, value_len: i32
func debugPrintMemHex(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	pointer := params[0].(int32)
	size := params[1].(int32)
	mem := callframe.GetMemoryByIndex(0)
	data, _ := mem.GetData(uint(pointer), uint(size))
	fmt.Println("Go: debugPrintMemHex", hex.EncodeToString(data))
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
	ewasmEnv.AddFunction("ethereum_getBlockHash", wasmedge.NewFunction(functype_i32i32_, getBlockHash, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockGasLimit", wasmedge.NewFunction(functype__i64, getBlockGasLimit, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockTimestamp", wasmedge.NewFunction(functype__i64, getBlockTimestamp, context, 0))
	ewasmEnv.AddFunction("ethereum_getBlockDifficulty", wasmedge.NewFunction(functype_i32_, getBlockDifficulty, context, 0))
	ewasmEnv.AddFunction("ethereum_getChainId", wasmedge.NewFunction(functype_i32_, getChainId, context, 0))
	ewasmEnv.AddFunction("ethereum_getBaseFee", wasmedge.NewFunction(functype_i32_, getBaseFee, context, 0))
	ewasmEnv.AddFunction("ethereum_call", wasmedge.NewFunction(functype_i64i32i32i32i32i32i32_i32, call, context, 0))
	ewasmEnv.AddFunction("ethereum_callCode", wasmedge.NewFunction(functype_i64i32i32i32i32i32i32_i32, callCode, context, 0))
	ewasmEnv.AddFunction("ethereum_callDelegate", wasmedge.NewFunction(functype_i64i32i32i32i32i32_i32, callDelegate, context, 0))
	ewasmEnv.AddFunction("ethereum_callStatic", wasmedge.NewFunction(functype_i64i32i32i32i32i32_i32, callStatic, context, 0))
	ewasmEnv.AddFunction("ethereum_create", wasmedge.NewFunction(functype_i32i32i32i32_, create, context, 0))
	ewasmEnv.AddFunction("ethereum_create2", wasmedge.NewFunction(functype_i32i32i32i32i32_, create2, context, 0))
	ewasmEnv.AddFunction("ethereum_selfDestruct", wasmedge.NewFunction(functype_i32_, selfDestruct, context, 0))
	ewasmEnv.AddFunction("ethereum_log", wasmedge.NewFunction(functype_i32i32i32i32i32i32i32_, log, context, 0))
	ewasmEnv.AddFunction("ethereum_finish", wasmedge.NewFunction(functype_i32i32_, finish, context, 0))
	ewasmEnv.AddFunction("ethereum_stop", wasmedge.NewFunction(functype__, stop, context, 0))
	ewasmEnv.AddFunction("ethereum_revert", wasmedge.NewFunction(functype_i32i32_, revert, context, 0))
	ewasmEnv.AddFunction("ethereum_sendCosmosMsg", wasmedge.NewFunction(functype_i32i32_i32, sendCosmosMsg, context, 0))
	ewasmEnv.AddFunction("ethereum_sendCosmosQuery", wasmedge.NewFunction(functype_i32i32_i32, sendCosmosQuery, context, 0))
	ewasmEnv.AddFunction("ethereum_debugPrinti32", wasmedge.NewFunction(functype_i32_, debugPrinti32, context, 0))
	ewasmEnv.AddFunction("ethereum_debugPrinti64", wasmedge.NewFunction(functype_i64_, debugPrinti64, context, 0))
	ewasmEnv.AddFunction("ethereum_debugPrintMemHex", wasmedge.NewFunction(functype_i32i32_, debugPrintMemHex, context, 0))

	return ewasmEnv
}

func readMem(callframe *wasmedge.CallingFrame, pointer interface{}, size interface{}) ([]byte, error) {
	ptr := pointer.(int32)
	length := size.(int32)
	mem := callframe.GetMemoryByIndex(0)
	data, err := mem.GetData(uint(ptr), uint(length))
	if err != nil {
		return nil, err
	}
	result := make([]byte, length)
	copy(result, data)
	return result, nil
}

func writeMem(callframe *wasmedge.CallingFrame, data []byte, pointer interface{}) error {
	ptr := pointer.(int32)
	length := len(data)
	mem := callframe.GetMemoryByIndex(0)
	err := mem.SetData(data, uint(ptr), uint(length))
	return err
}

func readBigInt(callframe *wasmedge.CallingFrame, pointer interface{}, size interface{}) (*big.Int, error) {
	data, err := readMem(callframe, pointer, size)
	if err != nil {
		return nil, err
	}
	x := new(big.Int)
	x.SetBytes(data)
	return x, nil
}

func readI64(callframe *wasmedge.CallingFrame, pointer interface{}, size interface{}) (int64, error) {
	x, err := readBigInt(callframe, pointer, size)
	if err != nil {
		return 0, err
	}
	if !x.IsInt64() {
		return 0, fmt.Errorf("readI32 overflow")
	}
	return x.Int64(), nil
}

func readI32(callframe *wasmedge.CallingFrame, pointer interface{}, size interface{}) (int32, error) {
	xi64, err := readI64(callframe, pointer, size)
	if err != nil {
		return 0, err
	}
	xi32 := int32(xi64)
	if xi64 > int64(xi32) {
		return 0, fmt.Errorf("readI32 overflow")
	}
	return xi32, nil
}

func isEvmAddress(addr AddressCW) bool {
	return hex.EncodeToString(addr.Bytes()[:12]) == hex.EncodeToString(make([]byte, 12))
}

func cleanupAddress(addr []byte) []byte {
	if isEvmAddress(BytesToAddressCW(addr)) {
		return addr[12:]
	}
	return addr
}
