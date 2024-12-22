package vm

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	address "cosmossdk.io/core/address"
	sdkerr "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	crypto "github.com/loredanacirstea/wasmx/crypto"
	cw8types "github.com/loredanacirstea/wasmx/x/wasmx/cw8/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
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

var EDDSA_PUBKEY_LEN = 32

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
var SECP256K1_VERIFY_CODE_VALID = int32(0)

// Return code (error code) for an invalid signature
var SECP256K1_VERIFY_CODE_INVALID = int32(1)

// Return code (error code) for a valid signature
const ED25519_VERIFY_CODE_VALID = int32(0)

// Return code (error code) for an invalid signature
const ED25519_VERIFY_CODE_INVALID = int32(1)

// CosmWasm imports.
// A complete documentation those functions is available in the VM that provides them:
// https://github.com/CosmWasm/cosmwasm/blob/0dae968db91734045239f45e982cddb1e5f11115/packages/std/src/imports.rs

// db_read(key: u32) -> u32;
func cw_8_db_read(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	returns := make([]interface{}, 1)
	ctx := _context.(*Context)
	key, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	// TODO MAX_LENGTH_DB_KEY
	data := ctx.ContractStore.Get(key)
	if len(data) == 0 {
		returns[0] = int32(0)
		return returns, nil
	}
	region, err := writeMemCw(rnh.GetVm(), data)
	if err != nil {
		return nil, err
	}
	returns[0] = region.Pointer
	return returns, nil
}

// db_write(key: u32, value: u32);
func cw_8_db_write(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	// TODO env.is_storage_readonly
	key, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	data, err := readMemCw(rnh.GetVm(), params[1])
	if err != nil {
		return nil, err
	}
	ctx := _context.(*Context)
	ctx.GasMeter.ConsumeGas(uint64(SSTORE_GAS_WASMX), "cw_8_db_write")
	ctx.ContractStore.Set(key, data)
	returns := make([]interface{}, 0)
	return returns, nil
}

// db_remove(key: u32);
func cw_8_db_remove(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	key, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	ctx.ContractStore.Delete(key)
	returns := make([]interface{}, 0)
	return returns, nil
}

var (
	OrderAscending  = 1
	OrderDescending = 2
)

// scan creates an iterator, which can be read by consecutive next() calls
// db_scan(start_ptr: u32, end_ptr: u32, order: i32) -> u32;
func cw_8_db_scan(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	startKey, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	endKey, err := readMemCw(rnh.GetVm(), params[1])
	if err != nil {
		return nil, err
	}
	order := params[2].(int32)

	var iter types.Iterator
	if order == int32(OrderAscending) {
		iter = ctx.ContractStore.Iterator(startKey, endKey)
	} else {
		iter = ctx.ContractStore.ReverseIterator(startKey, endKey)
	}
	count := len(ctx.dbIterators)
	ctx.dbIterators[int32(count)] = iter
	returns := make([]interface{}, 1)
	returns[0] = count
	return returns, nil
}

// db_next(iterator_id: u32) -> u32;
func cw_8_db_next(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	iterid := params[0].(int32)
	iterator, ok := ctx.dbIterators[iterid]
	if !ok {
		return nil, fmt.Errorf("cw_8_db_next out of bounds: %d", iterid)
	}

	values := make([][]byte, 2)
	if iterator.Valid() {
		values[0] = iterator.Key()
		values[1] = ctx.ContractStore.Get(values[0])
		iterator.Next()
	} else {
		iterator.Close()
		delete(ctx.dbIterators, iterid)
	}

	out_data := encode_sections(values)
	returns := make([]interface{}, 1)
	region, err := writeMemCw(rnh.GetVm(), out_data)
	if err != nil {
		return nil, err
	}
	returns[0] = region.Pointer
	return returns, nil
}

// addr_validate(source_ptr: u32) -> u32;
func cw_8_addr_validate(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	addrBz, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	addr, err := addr_canonicalize(ctx.CosmosHandler.AddressCodec(), addrBz)
	if err != nil {
		return cwError(rnh, err.Error())
	}
	addrstr, err := ctx.CosmosHandler.AddressCodec().BytesToString(addr)
	if err != nil {
		return nil, err
	}
	if string(addrBz) != addrstr {
		return cwError(rnh, "address validation failed")
	}
	_, err = writeMemCw(rnh.GetVm(), []byte(addr.Bytes()))
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, nil
}

// addr_canonicalize(source_ptr: u32, destination_ptr: u32) -> u32;
func cw_8_addr_canonicalize(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	addrBz, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	data, err := addr_canonicalize(ctx.CosmosHandler.AddressCodec(), addrBz)
	if err != nil {
		return cwError(rnh, err.Error())
	}
	_, err = writeMemToDestinationCw(rnh.GetVm(), data.Bytes(), params[1])
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, nil
}

// addr_humanize(source_ptr: u32, destination_ptr: u32) -> u32;
func cw_8_addr_humanize(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	addrBz, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	addr := sdk.AccAddress(addrBz)
	addrstr, err := ctx.CosmosHandler.AddressCodec().BytesToString(addr)
	if err != nil {
		return nil, err
	}
	_, err = writeMemToDestinationCw(rnh.GetVm(), []byte(addrstr), params[1])
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = int32(0)
	return returns, nil
}

// / Verifies message hashes against a signature with a public key, using the
// / secp256k1 ECDSA parametrization.
// / Returns 0 on verification success, 1 on verification failure, and values
// / greater than 1 in case of error.
// secp256k1_verify(message_hash_ptr: u32, signature_ptr: u32, public_key_ptr: u32) -> u32;
func cw_8_secp256k1_verify(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	msgHash, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	signature, err := readMemCw(rnh.GetVm(), params[1])
	if err != nil {
		return nil, err
	}
	publicKeyBz, err := readMemCw(rnh.GetVm(), params[2])
	if err != nil {
		return nil, err
	}
	ctx.GasMeter.ConsumeGas(uint64(Secp256k1VerifyCost), "cw_8_secp256k1_verify")

	publicKey := secp256k1.PubKey{Key: publicKeyBz}
	valid := crypto.VerifySignature(&publicKey, msgHash, signature)

	returns := make([]interface{}, 1)
	if valid {
		returns[0] = SECP256K1_VERIFY_CODE_VALID
	} else {
		returns[0] = SECP256K1_VERIFY_CODE_INVALID
	}
	return returns, nil
}

// secp256k1_recover_pubkey(message_hash_ptr: u32, signature_ptr: u32, recovery_param: u32) -> u64;
func cw_8_secp256k1_recover_pubkey(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	errmsg := "cosmwasm8: secp256k1_recover_pubkey: Not implemented"
	ctx.Logger(ctx.Ctx).Error(errmsg)
	// TODO
	return nil, fmt.Errorf(errmsg)
	msgHash, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	signature, err := readMemCw(rnh.GetVm(), params[1])
	if err != nil {
		return nil, err
	}
	// TODO use this
	recoveryParam := params[2].(int32)

	signature = append(signature, byte(recoveryParam))
	ctx.GasMeter.ConsumeGas(uint64(Secp256k1VerifyCost), "cw_8_secp256k1_recover_pubkey")
	recoveredPublicKey, err := crypto.Secp256k1Recover(msgHash, signature)
	if err != nil {
		return cwError(rnh, err.Error())
	}
	region, err := writeMemCw(rnh.GetVm(), recoveredPublicKey)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = uint64(region.Pointer)
	return returns, nil
}

// / Verifies a message against a signature with a public key, using the
// / ed25519 EdDSA scheme.
// / Returns 0 on verification success, 1 on verification failure, and values
// / greater than 1 in case of error.
// ed25519_verify(message_ptr: u32, signature_ptr: u32, public_key_ptr: u32) -> u32;
func cw_8_ed25519_verify(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	msg, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	if len(msg) > MAX_LENGTH_ED25519_MESSAGE {
		return nil, fmt.Errorf("cw_8_ed25519_verify: message length exceeds %d bytes", MAX_LENGTH_ED25519_MESSAGE)
	}
	signature, err := readMemCw(rnh.GetVm(), params[1])
	if err != nil {
		return nil, err
	}
	if len(signature) != MAX_LENGTH_ED25519_SIGNATURE {
		return nil, fmt.Errorf("cw_8_ed25519_verify: signature length exceeds %d bytes", MAX_LENGTH_ED25519_SIGNATURE)
	}
	publicKeyBz, err := readMemCw(rnh.GetVm(), params[2])
	if err != nil {
		return nil, err
	}
	ctx.GasMeter.ConsumeGas(uint64(Secp256k1VerifyCost), "cw_8_ed25519_verify")

	publicKey := ed25519.PubKey{Key: publicKeyBz}
	valid := publicKey.VerifySignature(msg, signature)

	returns := make([]interface{}, 1)
	if valid {
		returns[0] = ED25519_VERIFY_CODE_VALID
	} else {
		returns[0] = ED25519_VERIFY_CODE_INVALID
	}
	return returns, nil
}

// / Verifies a batch of messages against a batch of signatures and public keys, using the
// / ed25519 EdDSA scheme.
// / Returns 0 on verification success, 1 on verification failure, and values
// / greater than 1 in case of error.
// ed25519_batch_verify(messages_ptr: u32, signatures_ptr: u32, public_keys_ptr: u32) -> u32;
// / Three Variants are suppported in the input for convenience:
// /  - Equal number of messages, signatures, and public keys: Standard, generic functionality.
// /  - One message, and an equal number of signatures and public keys: Multiple digital signature
// / (multisig) verification of a single message.
// /  - One public key, and an equal number of messages and signatures: Verification of multiple
// / messages, all signed with the same private key.
// /
// / Any other variants of input vectors result in an error.
// /
// / Notes:
// /  - The "one-message, with zero signatures and zero public keys" case, is considered the empty
// / case.
// /  - The "one-public key, with zero messages and zero signatures" case, is considered the empty
// / case.
// /  - The empty case (no messages, no signatures and no public keys) returns true.
func cw_8_ed25519_batch_verify(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	errmsg := "cosmwasm8: ed25519_batch_verify: Not implemented"
	ctx.Logger(ctx.Ctx).Error(errmsg)
	// TODO
	return nil, fmt.Errorf(errmsg)

	msgs, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	signatures, err := readMemCw(rnh.GetVm(), params[1])
	if err != nil {
		return nil, err
	}
	if len(signatures) == 0 {
		return nil, fmt.Errorf("cw_8_ed25519_batch_verify: no signatures")
	}
	publicKeysBz, err := readMemCw(rnh.GetVm(), params[2])
	if err != nil {
		return nil, err
	}

	count := len(signatures) / MAX_LENGTH_ED25519_SIGNATURE
	if count > MAX_COUNT_ED25519_BATCH {
		return nil, fmt.Errorf("cw_8_ed25519_batch_verify: too many signatures")
	}
	// countPublicKeys := len(publicKeysBz) / EDDSA_PUBKEY_LEN

	ctx.GasMeter.ConsumeGas(uint64(Secp256k1VerifyCost), "cw_8_ed25519_batch_verify")

	publicKey := ed25519.PubKey{Key: publicKeysBz}
	valid := publicKey.VerifySignature(msgs, signatures)

	returns := make([]interface{}, 1)
	if valid {
		returns[0] = ED25519_VERIFY_CODE_VALID
	} else {
		returns[0] = ED25519_VERIFY_CODE_INVALID
	}
	return returns, nil
}

// / Writes a debug message (UFT-8 encoded) to the host for debugging purposes.
// / The host is free to log or process this in any way it considers appropriate.
// / In production environments it is expected that those messages are discarded.
// debug(source_ptr: u32);
func cw_8_debug(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	msgBz, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	ctx.Logger(ctx.Ctx).Debug("cosmwasm8: debug: " + string(msgBz))
	returns := make([]interface{}, 0)
	return returns, nil
}

// / Executes a query on the chain (import). Not to be confused with the
// / query export, which queries the state of the contract.
// query_chain(request: u32) -> u32;
func cw_8_query_chain(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	databz, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	if len(databz) > MAX_LENGTH_QUERY_CHAIN_REQUEST {
		return nil, fmt.Errorf("cw_8_query_chain: query too big")
	}
	var req cw8types.QueryRequest
	err = json.Unmarshal(databz, &req)
	if err != nil {
		return nil, err
	}
	result, err := ctx.CosmosHandler.WasmVMQueryHandler(ctx.Env.Contract.Address, req)
	response := cw8types.ToQuerierResult(result, err)
	responseBz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	region, err := writeMemCw(rnh.GetVm(), responseBz)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = uint64(region.Pointer)
	return returns, nil
}

// abort(source_ptr: u32);
func cw_8_abort(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	data, err := readMemCw(rnh.GetVm(), params[0])
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 0)
	ctx.FinishData = data
	ctx.ReturnData = data
	ctx.Logger(ctx.Ctx).Debug("cosmwasm8: abort: " + string(data))
	return returns, fmt.Errorf("abort")
}

func addr_canonicalize(addressCodec address.Codec, addrUtf8 []byte) (sdk.AccAddress, error) {
	if len(addrUtf8) == 0 {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "empty address")
	}
	addr, err := addressCodec.StringToBytes(string(addrUtf8))
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "invalid address")
	}
	if len(addr) > cw8types.MAX_LENGTH_HUMAN_ADDRESS {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "address too long")
	}
	return addr, nil
}

func cwError(rnh memc.RuntimeHandler, cwErr string) ([]interface{}, error) {
	_, err := writeMemCw(rnh.GetVm(), []byte(cwErr))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func BuildCosmWasm_8(context *Context, rnh memc.RuntimeHandler) (interface{}, error) {
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		vm.BuildFn("db_read", cw_8_db_read, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("db_write", cw_8_db_write, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("db_remove", cw_8_db_remove, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("db_scan", cw_8_db_scan, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("db_next", cw_8_db_next, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("addr_validate", cw_8_addr_validate, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("addr_canonicalize", cw_8_addr_canonicalize, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("addr_humanize", cw_8_addr_humanize, []interface{}{vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("secp256k1_verify", cw_8_secp256k1_verify, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("secp256k1_recover_pubkey", cw_8_secp256k1_recover_pubkey, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I64()}, 0),
		vm.BuildFn("ed25519_verify", cw_8_ed25519_verify, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("ed25519_batch_verify", cw_8_ed25519_batch_verify, []interface{}{vm.ValType_I32(), vm.ValType_I32(), vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("debug", cw_8_debug, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
		vm.BuildFn("query_chain", cw_8_query_chain, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("abort", cw_8_abort, []interface{}{vm.ValType_I32()}, []interface{}{}, 0),
	}

	return vm.BuildModule(rnh, "env", context, fndefs)
}

func BuildArgsCw(context *Context) ([]byte, []byte, []byte, error) {
	addrstr := context.Env.Contract.Address.String()
	senderstr := context.Env.CurrentCall.Sender.String()
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
			Address: addrstr,
		},
	}
	info := cw8types.MessageInfo{
		Sender: senderstr,
		Funds:  cw8types.Coins{cw8types.Coin{Denom: context.Env.Chain.Denom, Amount: sdkmath.NewIntFromBigInt(context.Env.CurrentCall.Funds).String()}},
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
func ExecuteCw8Execute(context *Context, vm memc.IVm, funcName string) ([]int32, error) {
	envBz, infoBz, msgBz, err := BuildArgsCw(context)
	if err != nil {
		return nil, err
	}
	envRegion, err := writeMemCw(vm, envBz)
	if err != nil {
		return nil, err
	}
	infoRegion, err := writeMemCw(vm, infoBz)
	if err != nil {
		return nil, err
	}
	msgRegion, err := writeMemCw(vm, msgBz)
	if err != nil {
		return nil, err
	}
	res, execErr := vm.Call(funcName, []interface{}{envRegion.Pointer, infoRegion.Pointer, msgRegion.Pointer}, context.GasMeter)
	if len(res) == 0 {
		return nil, execErr
	}
	data, err := readMemCw(vm, res[0])
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

	if result.Ok != nil {
		context.Messages = result.Ok.Messages
		context.ReturnData = result.Ok.Data
		context.FinishData = result.Ok.Data
		// TODO make these wasmx logs, of type cosmwasm
		// result.Ok.Attributes
		// result.Ok.Events
		// this Response data can be returned by Reply
		return nil, nil
	}

	err = sdkerr.Wrapf(sdkerr.Error{}, "%s %s", cw8types.ERROR_FLAG_EXECUTION, result.Err)
	if execErr != nil {
		return nil, sdkerr.Wrapf(err, execErr.Error())
	}
	return nil, err
}

// reply(env_ptr: u32, msg_ptr: u32)
func ExecuteCw8Reply(context *Context, vm memc.IVm, funcName string) ([]int32, error) {
	envBz, _, msgBz, err := BuildArgsCw(context)
	if err != nil {
		return nil, err
	}
	envRegion, err := writeMemCw(vm, envBz)
	if err != nil {
		return nil, err
	}
	msgRegion, err := writeMemCw(vm, msgBz)
	if err != nil {
		return nil, err
	}
	res, execErr := vm.Call(funcName, []interface{}{envRegion.Pointer, msgRegion.Pointer}, context.GasMeter)
	if len(res) == 0 {
		return nil, execErr
	}
	data, err := readMemCw(vm, res[0])
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

	if result.Ok != nil {
		context.Messages = result.Ok.Messages
		context.ReturnData = result.Ok.Data
		context.FinishData = result.Ok.Data
		// TODO make these wasmx logs, of type cosmwasm
		// result.Ok.Attributes
		// result.Ok.Events
		return nil, nil
	}

	err = sdkerr.Wrapf(sdkerr.Error{}, "%s %s", cw8types.ERROR_FLAG_REPLY, result.Err)
	if execErr != nil {
		return nil, sdkerr.Wrapf(err, execErr.Error())
	}
	return nil, err
}

func ExecuteCw8Query(context *Context, vm memc.IVm, funcName string) ([]int32, error) {
	envBz, _, msgBz, err := BuildArgsCw(context)
	if err != nil {
		return nil, err
	}
	envRegion, err := writeMemCw(vm, envBz)
	if err != nil {
		return nil, err
	}
	msgRegion, err := writeMemCw(vm, msgBz)
	if err != nil {
		return nil, err
	}
	res, execErr := vm.Call(funcName, []interface{}{envRegion.Pointer, msgRegion.Pointer}, context.GasMeter)
	if len(res) == 0 {
		return nil, execErr
	}
	data, err := readMemCw(vm, res[0])
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
		context.FinishData = result.Ok
		return nil, nil
	}
	err = sdkerr.Wrapf(sdkerr.Error{}, "%s %s", cw8types.ERROR_FLAG_QUERY, result.Err)
	if execErr != nil {
		return nil, sdkerr.Wrapf(err, execErr.Error())
	}
	return nil, err
}

// instantiate/execute(env_ptr: u32, info_ptr: u32, msg_ptr: u32)
func ExecuteCw8(context *Context, vm memc.IVm, funcName string, args []interface{}) ([]int32, error) {
	switch funcName {
	case types.ENTRY_POINT_QUERY:
		return ExecuteCw8Query(context, vm, funcName)
	case types.ENTRY_POINT_REPLY:
		return ExecuteCw8Reply(context, vm, funcName)
	}
	// instantiate, execute
	return ExecuteCw8Execute(context, vm, funcName)
}

func writeMemToDestinationCw(vm memc.IVm, data []byte, ptr interface{}) (*Region, error) {
	mem, err := vm.GetMemory()
	if err != nil {
		return nil, err
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

func writeMemCw(vm memc.IVm, data []byte) (*Region, error) {
	mem, err := vm.GetMemory()
	if err != nil {
		return nil, err
	}
	res, err := vm.Call(types.MEMORY_EXPORT_ALLOCATE, []interface{}{int32(len(data))}, nil) // TODO gas meter allocate
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

func readMemCw(vm memc.IVm, ptr interface{}) ([]byte, error) {
	mem, err := vm.GetMemory()
	if err != nil {
		return nil, err
	}
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

func NewRegion(mem memc.IMemory, pointer interface{}) (*Region, error) {
	ptr := pointer.(int32)
	offsetBz, err := mem.Read(ptr, 4)
	if err != nil {
		return nil, err
	}
	capacityBz, err := mem.Read(ptr+4, 4)
	if err != nil {
		return nil, err
	}
	lengthBz, err := mem.Read(ptr+8, 4)
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

func (r *Region) Write(mem memc.IMemory, data []byte) error {
	// writing nothing in memory gives the error "wrong VM workflow"
	if len(data) == 0 {
		return nil
	}
	if int(r.Capacity) < len(data) {
		return sdkerr.Wrapf(sdkerr.Error{}, "region write too small")
	}
	err := mem.Write(int32(r.Offset), data)
	if err != nil {
		return err
	}

	// Update the region reference
	r.Length = uint32(len(data))
	lengthBz := binary.LittleEndian.AppendUint32([]byte{}, r.Length)
	err = mem.Write(r.Pointer+8, lengthBz)
	return err
}

func (r *Region) Read(mem memc.IMemory) ([]byte, error) {
	if r.Length == 0 {
		return nil, nil
	}
	return mem.Read(int32(r.Offset), int32(r.Length))
}

func encode_sections(sections [][]byte) []byte {
	out_data := []byte{}
	for _, section := range sections {
		lengthbz := binary.BigEndian.AppendUint32([]byte{}, uint32(len(section)))
		out_data = append(out_data, section...)
		out_data = append(out_data, lengthbz...)
	}
	return out_data
}
