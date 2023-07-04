package vm

import (
	"fmt"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

// CosmWasm imports.
// A complete documentation those functions is available in the VM that provides them:
// https://github.com/CosmWasm/cosmwasm/blob/0dae968db91734045239f45e982cddb1e5f11115/packages/std/src/imports.rs

// db_read(key: u32) -> u32;
func cw_8_db_read(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_db_read--", params)
	ctx := context.(*Context)
	keybz, err := readMemFromPtr(callframe, params[0])
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	data := ctx.ContractStore.Get(keybz)
	if len(data) == 0 {
		data = make([]byte, 32)
	}
	newptr, err := allocateWriteMem(ctx, callframe, data)
	if err != nil {
		return nil, wasmedge.Result_Fail
	}
	returns := make([]interface{}, 1)
	returns[0] = newptr
	return returns, wasmedge.Result_Success
}

// db_write(key: u32, value: u32);
func cw_8_db_write(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_db_write--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// db_remove(key: u32);
func cw_8_db_remove(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_db_remove--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// scan creates an iterator, which can be read by consecutive next() calls
// db_scan(start_ptr: u32, end_ptr: u32, order: i32) -> u32;
func cw_8_db_scan(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_db_scan--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// db_next(iterator_id: u32) -> u32;
func cw_8_db_next(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_db_next--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// addr_validate(source_ptr: u32) -> u32;
func cw_8_addr_validate(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_addr_validate--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// addr_canonicalize(source_ptr: u32, destination_ptr: u32) -> u32;
func cw_8_addr_canonicalize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_addr_canonicalize--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// addr_humanize(source_ptr: u32, destination_ptr: u32) -> u32;
func cw_8_addr_humanize(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_addr_humanize--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// / Verifies message hashes against a signature with a public key, using the
// / secp256k1 ECDSA parametrization.
// / Returns 0 on verification success, 1 on verification failure, and values
// / greater than 1 in case of error.
// secp256k1_verify(message_hash_ptr: u32, signature_ptr: u32, public_key_ptr: u32) -> u32;
func cw_8_secp256k1_verify(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_secp256k1_verify--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// secp256k1_recover_pubkey(message_hash_ptr: u32, signature_ptr: u32, recovery_param: u32) -> u64;
func cw_8_secp256k1_recover_pubkey(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_secp256k1_recover_pubkey--", params)
	returns := make([]interface{}, 0)
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
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// / Executes a query on the chain (import). Not to be confused with the
// / query export, which queries the state of the contract.
// query_chain(request: u32) -> u32;
func cw_8_query_chain(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_query_chain--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
}

// abort(source_ptr: u32);
func cw_8_abort(context interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	fmt.Println("--cw_8_abort--", params)
	returns := make([]interface{}, 0)
	return returns, wasmedge.Result_Success
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
