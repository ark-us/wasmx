package vm

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	sdkerr "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/second-state/WasmEdge-go/wasmedge"

	cw8types "mythos/v1/x/wasmx/cw8/types"
	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/native"
)

// A kibi (kilo binary)
var KI = 1024

// A mibi (mega binary)
var MI = 1024 * 1024

// Max key length for db_write/db_read/db_remove/db_scan (when VM reads the key argument from Wasm memory)
var MAX_LENGTH_DB_KEY = 64 * KI

// Max value length for db_write (when VM reads the value argument from Wasm memory)
var MAX_LENGTH_DB_VALUE = 128 * KI

// Typically 20 (Cosmos SDK, Ethereum), 32 (Nano, Substrate) or 54 (MockApi)
var MAX_LENGTH_CANONICAL_ADDRESS = 64

// The max length of human address inputs (in bytes).
// The maximum allowed size for [bech32](https://github.com/bitcoin/bips/blob/master/bip-0173.mediawiki#bech32)
// is 90 characters and we're adding some safety margin around that for other formats.
var MAX_LENGTH_HUMAN_ADDRESS = 256
var MAX_LENGTH_QUERY_CHAIN_REQUEST = 64 * KI

// Length of a serialized Ed25519  signature
var MAX_LENGTH_ED25519_SIGNATURE = 64

// Max length of a Ed25519 message in bytes.
// This is an arbitrary value, for performance / memory contraints. If you need to verify larger
// messages, let us know.
var MAX_LENGTH_ED25519_MESSAGE = 128 * 1024

// Max number of batch Ed25519 messages / signatures / public_keys.
// This is an arbitrary value, for performance / memory contraints. If you need to batch-verify a
// larger number of signatures, let us know.
var MAX_COUNT_ED25519_BATCH = 256

// Max length for a debug message
var MAX_LENGTH_DEBUG = 2 * MI

// Max length for an abort message
var MAX_LENGTH_ABORT = 2 * MI

// should be the same as app.Secp256k1VerifyCost
var Secp256k1VerifyCost = uint64(21000)

// Return code (error code) for a valid signature
var SECP256K1_VERIFY_CODE_VALID = uint32(0)

// Return code (error code) for an invalid signature
var SECP256K1_VERIFY_CODE_INVALID = uint32(1)

// Return code (error code) for a valid signature
const ED25519_VERIFY_CODE_VALID = uint32(0)

// Return code (error code) for an invalid signature
const ED25519_VERIFY_CODE_INVALID = uint32(1)

// CosmWasm imports.
// A complete documentation those functions is available in the VM that provides them:
// https://github.com/CosmWasm/cosmwasm/blob/0dae968db91734045239f45e982cddb1e5f11115/packages/std/src/imports.rs

// db_read(key: u32) -> u32;
func cw_8_db_read(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_db_read-", params)
	returns := make([]interface{}, 1)
	ctx := context.(*Context)
	key, err := readMemFromPtrCw(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// TODO MAX_LENGTH_DB_KEY
	data := ctx.ContractStore.Get(key)
	if len(data) == 0 {
		returns[0] = 0
		return returns, wasmedge.Result_Success
	}
	region, err := allocateWriteMemCw(ctx, callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns[0] = region.Pointer
	return returns, wasmedge.Result_Success
}

// db_write(key: u32, value: u32);
func cw_8_db_write(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_db_write-", params)
	// TODO env.is_storage_readonly
	key, err := readMemFromPtrCw(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data, err := readMemFromPtrCw(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx := context.(*Context)
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_WASMX), "wasmx")
	ctx.ContractStore.Set(key, data)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// db_remove(key: u32);
func cw_8_db_remove(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_db_remove-", params)
	ctx := context.(*Context)
	key, err := readMemFromPtrCw(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.ContractStore.Delete(key)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// scan creates an iterator, which can be read by consecutive next() calls
// db_scan(start_ptr: u32, end_ptr: u32, order: i32) -> u32;
func cw_8_db_scan(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_db_scan--", params)
	// TODO
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// db_next(iterator_id: u32) -> u32;
func cw_8_db_next(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_db_next--", params)
	// TODO
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// addr_validate(source_ptr: u32) -> u32;
func cw_8_addr_validate(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_addr_validate--", params)
	ctx := context.(*Context)
	addrBz, err := readMemFromPtrCw(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr, err := addr_canonicalize(addrBz)
	if err != nil {
		return cwError(ctx, callframe, err.Error())
	}
	if string(addrBz) != addr.String() {
		return cwError(ctx, callframe, "address validation failed")
	}
	_, err = allocateWriteMemCw(ctx, callframe, []byte(addr.Bytes()))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
}

// addr_canonicalize(source_ptr: u32, destination_ptr: u32) -> u32;
func cw_8_addr_canonicalize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_addr_canonicalize--", params)
	ctx := context.(*Context)
	addrBz, err := readMemFromPtrCw(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data, err := addr_canonicalize(addrBz)
	if err != nil {
		return cwError(ctx, callframe, err.Error())
	}
	_, err = writeMemToDestinationCw(ctx, callframe, data.Bytes(), params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
}

// addr_humanize(source_ptr: u32, destination_ptr: u32) -> u32;
func cw_8_addr_humanize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_addr_humanize--", params)
	ctx := context.(*Context)
	addrBz, err := readMemFromPtrCw(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	addr, err := addr_canonicalize(addrBz)
	if err != nil {
		return cwError(ctx, callframe, err.Error())
	}
	_, err = writeMemToDestinationCw(ctx, callframe, []byte(addr.String()), params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = 0
	return returns, wasmedge.Result_Success
}

// / Verifies message hashes against a signature with a public key, using the
// / secp256k1 ECDSA parametrization.
// / Returns 0 on verification success, 1 on verification failure, and values
// / greater than 1 in case of error.
// secp256k1_verify(message_hash_ptr: u32, signature_ptr: u32, public_key_ptr: u32) -> u32;
func cw_8_secp256k1_verify(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_secp256k1_verify--", params)
	ctx := context.(*Context)
	msgHash, err := readMemFromPtrCw(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	signature, err := readMemFromPtrCw(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	publicKeyBz, err := readMemFromPtrCw(callframe, params[2])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	ctx.GasMeter.ConsumeGas(uint64(Secp256k1VerifyCost), "cosmwasm8")

	publicKey := secp256k1.PubKey{Key: publicKeyBz}
	valid := publicKey.VerifySignature(msgHash, signature)
	returns := make([]interface{}, 1)
	if valid {
		returns[0] = SECP256K1_VERIFY_CODE_VALID
	} else {
		returns[0] = SECP256K1_VERIFY_CODE_INVALID
	}
	return returns, wasmedge.Result_Success
}

// secp256k1_recover_pubkey(message_hash_ptr: u32, signature_ptr: u32, recovery_param: u32) -> u64;
func cw_8_secp256k1_recover_pubkey(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_secp256k1_recover_pubkey--", params)
	ctx := context.(*Context)
	msgHash, err := readMemFromPtrCw(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	signature, err := readMemFromPtrCw(callframe, params[1])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// TODO use this
	// recoveryParam, err := readMemFromPtrCw(callframe, params[2])
	// if err != nil {
	// 	return nil, wasmedge.Result_Fail
	// }
	ctx.GasMeter.ConsumeGas(uint64(Secp256k1VerifyCost), "cosmwasm8")
	recoveredPublicKey, err := native.Secp256k1Recover(msgHash, signature)
	if err != nil {
		return cwError(ctx, callframe, err.Error())
	}
	region, err := allocateWriteMemCw(ctx, callframe, recoveredPublicKey)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = uint64(region.Pointer)
	return returns, wasmedge.Result_Success
}

// / Verifies a message against a signature with a public key, using the
// / ed25519 EdDSA scheme.
// / Returns 0 on verification success, 1 on verification failure, and values
// / greater than 1 in case of error.
// ed25519_verify(message_ptr: u32, signature_ptr: u32, public_key_ptr: u32) -> u32;
func cw_8_ed25519_verify(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_ed25519_verify--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// / Verifies a batch of messages against a batch of signatures and public keys, using the
// / ed25519 EdDSA scheme.
// / Returns 0 on verification success, 1 on verification failure, and values
// / greater than 1 in case of error.
// ed25519_batch_verify(messages_ptr: u32, signatures_ptr: u32, public_keys_ptr: u32) -> u32;
func cw_8_ed25519_batch_verify(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_ed25519_batch_verify--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// / Writes a debug message (UFT-8 encoded) to the host for debugging purposes.
// / The host is free to log or process this in any way it considers appropriate.
// / In production environments it is expected that those messages are discarded.
// debug(source_ptr: u32);
func cw_8_debug(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_debug--", params)
	// TODO only print if in debug mode
	msgBz, err := readMemFromPtrCw(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	fmt.Println(string(msgBz))
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// / Executes a query on the chain (import). Not to be confused with the
// / query export, which queries the state of the contract.
// query_chain(request: u32) -> u32;
func cw_8_query_chain(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_query_chain--", params)
	ctx := context.(*Context)
	databz, err := readMemFromPtrCw(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	// TODO MAX_LENGTH_QUERY_CHAIN_REQUEST
	// do query
	result := databz
	region, err := allocateWriteMemCw(ctx, callframe, result)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = uint64(region.Pointer)
	return returns, wasmedge.Result_Success
}

// abort(source_ptr: u32);
func cw_8_abort(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_abort--", params)
	ctx := context.(*Context)
	data, err := readMemFromPtrCw(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 0)
	ctx.ReturnData = data
	return returns, wasmedge.Result_Fail
}

func addr_canonicalize(addrUtf8 []byte) (sdk.AccAddress, error) {
	if len(addrUtf8) == 0 {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "empty address")
	}
	fmt.Println("--cw_8_addr_validate-addrBz-str-", string(addrUtf8))
	addr, err := sdk.AccAddressFromBech32(string(addrUtf8))
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "invalid address")
	}
	fmt.Println("--cw_8_addr_validate-addr-", addr, addr.String())
	if len(addr.Bytes()) > cw8types.MAX_LENGTH_HUMAN_ADDRESS {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "address too long")
	}
	return addr, nil
}

func cwError(ctx *Context, callframe *wasmedge.CallingFrame, cwErr string) ([]interface{}, wasmedge.Result) {
	_, err := allocateWriteMemCw(ctx, callframe, []byte(cwErr))
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	return nil, wasmedge.Result_Success
}

func BuildCosmWasm_8(context *Context) *wasmedge.Module {
	env := wasmedge.NewModule("env")
	functype_i32i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32_ := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32},
		[]wasmedge.ValType{},
	)
	functype_i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32_i32 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I32},
	)
	functype_i32i32i32_i64 := wasmedge.NewFunctionType(
		[]wasmedge.ValType{wasmedge.ValType_I32, wasmedge.ValType_I32, wasmedge.ValType_I32},
		[]wasmedge.ValType{wasmedge.ValType_I64},
	)

	env.AddFunction("db_read", wasmedge.NewFunction(functype_i32_i32, cw_8_db_read, context, 0))
	env.AddFunction("db_write", wasmedge.NewFunction(functype_i32i32_, cw_8_db_write, context, 0))
	env.AddFunction("db_remove", wasmedge.NewFunction(functype_i32_, cw_8_db_remove, context, 0))
	env.AddFunction("db_scan", wasmedge.NewFunction(functype_i32i32i32_i32, cw_8_db_scan, context, 0))
	env.AddFunction("db_next", wasmedge.NewFunction(functype_i32_i32, cw_8_db_next, context, 0))
	env.AddFunction("addr_validate", wasmedge.NewFunction(functype_i32_i32, cw_8_addr_validate, context, 0))
	env.AddFunction("addr_canonicalize", wasmedge.NewFunction(functype_i32i32_i32, cw_8_addr_canonicalize, context, 0))
	env.AddFunction("addr_humanize", wasmedge.NewFunction(functype_i32i32_i32, cw_8_addr_humanize, context, 0))
	env.AddFunction("secp256k1_verify", wasmedge.NewFunction(functype_i32i32i32_i32, cw_8_secp256k1_verify, context, 0))
	env.AddFunction("secp256k1_recover_pubkey", wasmedge.NewFunction(functype_i32i32i32_i64, cw_8_secp256k1_recover_pubkey, context, 0))
	env.AddFunction("ed25519_verify", wasmedge.NewFunction(functype_i32i32i32_i32, cw_8_ed25519_verify, context, 0))
	env.AddFunction("ed25519_batch_verify", wasmedge.NewFunction(functype_i32i32i32_i32, cw_8_ed25519_batch_verify, context, 0))
	env.AddFunction("debug", wasmedge.NewFunction(functype_i32_, cw_8_debug, context, 0))
	env.AddFunction("query_chain", wasmedge.NewFunction(functype_i32_i32, cw_8_query_chain, context, 0))
	env.AddFunction("abort", wasmedge.NewFunction(functype_i32_, cw_8_abort, context, 0))

	return env
}

func BuildArgsCw(context *Context, contractVm *wasmedge.VM) ([]byte, []byte, []byte, error) {
	env := cw8types.Env{
		Block: cw8types.BlockInfo{
			Height:  context.Env.Block.Height,
			Time:    context.Env.Block.Timestamp,
			ChainID: context.Env.Chain.ChainIdFull,
		},
		Transaction: &cw8types.TransactionInfo{
			Index: context.Env.Transaction.Index,
		},
		Contract: cw8types.ContractInfo{
			Address: context.Env.Contract.Address.String(),
		},
	}
	info := cw8types.MessageInfo{
		Sender: context.Env.CurrentCall.Sender.String(),
		Funds:  cw8types.Coins{cw8types.Coin{Denom: context.Env.Chain.Denom, Amount: sdk.NewIntFromBigInt(context.Env.CurrentCall.Funds).String()}},
	}
	msgBz := context.Env.CurrentCall.CallData
	envBz, err := json.Marshal(env)
	if err != nil {
		return nil, nil, nil, err
	}
	infoBz, err := json.Marshal(info)
	if err != nil {
		return nil, nil, nil, err
	}
	return envBz, infoBz, msgBz, nil
}

// instantiate(env_ptr: u32, info_ptr: u32, msg_ptr: u32)
func ExecuteCw8Execute(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	envBz, infoBz, msgBz, err := BuildArgsCw(context, contractVm)
	if err != nil {
		return nil, err
	}
	activeMemory := contractVm.GetActiveModule().FindMemory("memory")

	envRegion, err := writeMemCw(contractVm, activeMemory, envBz)
	if err != nil {
		return nil, err
	}
	infoRegion, err := writeMemCw(contractVm, activeMemory, infoBz)
	if err != nil {
		return nil, err
	}
	msgRegion, err := writeMemCw(contractVm, activeMemory, msgBz)
	if err != nil {
		return nil, err
	}
	res, execErr := contractVm.Execute(funcName, envRegion.Pointer, infoRegion.Pointer, msgRegion.Pointer)
	if len(res) == 0 {
		return nil, execErr
	}
	data, err := readMemCw(activeMemory, res[0])
	if err != nil {
		if execErr != nil {
			return nil, sdkerr.Wrapf(execErr, "cannot read returned data: %s", err.Error())
		}
		return nil, err
	}

	var result cw8types.ContractResult
	err = json.Unmarshal(data, &result)
	if err != nil {
		if execErr != nil {
			return nil, sdkerr.Wrapf(execErr, "cannot unmarshal returned data: %s", err.Error())
		}
		return nil, err
	}
	// context.ReturnData = data
	fmt.Println("---executed", string(data))

	if result.Ok != nil {
		context.Messages = result.Ok.Messages
		context.ReturnData = result.Ok.Data
		// TODO make these wasmx logs, of type cosmwasm
		// result.Ok.Attributes
		// result.Ok.Events
		return nil, nil
	}

	err = sdkerr.Wrapf(sdkerr.Error{}, result.Err)
	if execErr != nil {
		return nil, sdkerr.Wrapf(err, execErr.Error())
	}
	return nil, err
}

func ExecuteCw8Query(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	envBz, _, msgBz, err := BuildArgsCw(context, contractVm)
	if err != nil {
		return nil, err
	}
	activeMemory := contractVm.GetActiveModule().FindMemory("memory")

	envRegion, err := writeMemCw(contractVm, activeMemory, envBz)
	if err != nil {
		return nil, err
	}
	msgRegion, err := writeMemCw(contractVm, activeMemory, msgBz)
	if err != nil {
		return nil, err
	}
	res, execErr := contractVm.Execute(funcName, envRegion.Pointer, msgRegion.Pointer)
	if len(res) == 0 {
		return nil, execErr
	}
	data, err := readMemCw(activeMemory, res[0])
	if err != nil {
		if execErr != nil {
			return nil, sdkerr.Wrapf(execErr, "cannot read returned data: %s", err.Error())
		}
		return nil, err
	}

	var result cw8types.QueryResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		if execErr != nil {
			return nil, sdkerr.Wrapf(execErr, "cannot unmarshal returned data: %s", err.Error())
		}
		return nil, err
	}

	if result.Ok != nil {
		context.ReturnData = result.Ok
		return nil, nil
	}

	err = sdkerr.Wrapf(sdkerr.Error{}, result.Err)
	if execErr != nil {
		return nil, sdkerr.Wrapf(err, execErr.Error())
	}
	return nil, err
}

// instantiate/execute(env_ptr: u32, info_ptr: u32, msg_ptr: u32)
func ExecuteCw8(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	if funcName == types.ENTRY_POINT_QUERY {
		return ExecuteCw8Query(context, contractVm, funcName)
	}
	return ExecuteCw8Execute(context, contractVm, funcName)
}

func writeMemToDestinationCw(ctx *Context, callframe *wasmedge.CallingFrame, data []byte, ptr interface{}) (*Region, error) {
	mem := callframe.GetMemoryByIndex(0)
	if mem == nil {
		return nil, fmt.Errorf("could not find memory")
	}
	region, err := NewRegion(mem, ptr)
	if err != nil {
		return nil, err
	}
	err = region.Write(mem, data)
	if err != nil {
		return nil, err
	}
	return region, nil
}

func allocateWriteMemCw(ctx *Context, callframe *wasmedge.CallingFrame, data []byte) (*Region, error) {
	addr := ctx.Env.Contract.Address
	contractCtx, ok := ctx.ContractRouter[addr.String()]
	if !ok {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "contract context not found for address %s", addr.String())
	}
	mem := callframe.GetMemoryByIndex(0)
	if mem == nil {
		return nil, fmt.Errorf("could not find memory")
	}
	return writeMemCw(contractCtx.Vm, mem, data)
}

func writeMemCw(vm *wasmedge.VM, mem *wasmedge.Memory, data []byte) (*Region, error) {
	res, err := vm.Execute("allocate", int32(len(data)))
	if err != nil {
		return nil, err
	}
	ptr := res[0]
	region, err := NewRegion(mem, ptr)
	if err != nil {
		return nil, err
	}
	err = region.Write(mem, data)
	if err != nil {
		return nil, err
	}
	return region, nil
}

func readMemCw(mem *wasmedge.Memory, ptr interface{}) ([]byte, error) {
	region, err := NewRegion(mem, ptr)
	if err != nil {
		return nil, err
	}
	data, err := region.Read(mem)
	if err != nil {
		return nil, err
	}
	// byte slices are references, so we need to copy the data outside wasm memory
	dest := make([]byte, len(data))
	copy(dest, data)
	return dest, nil
}

func readMemFromPtrCw(callframe *wasmedge.CallingFrame, pointer interface{}) ([]byte, error) {
	mem := callframe.GetMemoryByIndex(0)
	if mem == nil {
		return nil, fmt.Errorf("could not find memory")
	}
	return readMemCw(mem, pointer)
}

type Region struct {
	// The pointer to where the region is defined in memory
	Pointer int32
	// The beginning of the region expressed as bytes from the beginning of the linear memory
	Offset uint32
	// The number of bytes available in this region
	Capacity uint32
	// The number of bytes used in this region
	Length uint32
}

func NewRegion(mem *wasmedge.Memory, pointer interface{}) (*Region, error) {
	ptr := pointer.(int32)
	offsetBz, err := mem.GetData(uint(ptr), 4)
	if err != nil {
		return nil, err
	}
	capacityBz, err := mem.GetData(uint(ptr)+4, 4)
	if err != nil {
		return nil, err
	}
	lengthBz, err := mem.GetData(uint(ptr)+8, 4)
	if err != nil {
		return nil, err
	}
	offset := binary.LittleEndian.Uint32(offsetBz)
	capacity := binary.LittleEndian.Uint32(capacityBz)
	length := binary.LittleEndian.Uint32(lengthBz)

	region := &Region{
		Pointer:  ptr,
		Offset:   offset,
		Capacity: capacity,
		Length:   length,
	}
	err = region.Validate()
	if err != nil {
		return nil, err
	}
	return region, nil
}

func (r *Region) Validate() error {
	if r.Offset == 0 {
		return sdkerr.Wrapf(sdkerr.Error{}, "invalid region offset zero")
	}
	if r.Length > r.Capacity {
		return sdkerr.Wrapf(sdkerr.Error{}, "invalid region length exceeds capacity")
	}
	if r.Capacity > (math.MaxUint32 - r.Offset) {
		return sdkerr.Wrapf(sdkerr.Error{}, "invalid region capacity out of range")
	}
	return nil
}

func (r *Region) Write(mem *wasmedge.Memory, data []byte) error {
	// writing nothing in memory gives the error "wrong VM workflow"
	if len(data) == 0 {
		return nil
	}
	if int(r.Capacity) < len(data) {
		return sdkerr.Wrapf(sdkerr.Error{}, "region write too small")
	}
	err := mem.SetData(data, uint(r.Offset), uint(len(data)))
	if err != nil {
		return err
	}

	// Update the region reference
	r.Length = uint32(len(data))
	lengthBz := binary.LittleEndian.AppendUint32([]byte{}, r.Length)
	err = mem.SetData(lengthBz, uint(r.Pointer)+8, 4)
	return err
}

func (r *Region) Read(mem *wasmedge.Memory) ([]byte, error) {
	if r.Length == 0 {
		return nil, nil
	}
	return mem.GetData(uint(r.Offset), uint(r.Length))
}
