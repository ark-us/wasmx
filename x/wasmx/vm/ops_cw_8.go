package vm

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	sdkerr "cosmossdk.io/errors"

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

// instantiate(env_ptr: u32, info_ptr: u32, msg_ptr: u32)
func ExecuteCw8(context *Context, contractVm *wasmedge.VM, funcName string) ([]interface{}, error) {
	env := Env{
		Block: BlockInfo{
			Height:  context.Env.Block.Height,
			Time:    context.Env.Block.Timestamp,
			ChainID: context.Env.Chain.ChainIdFull,
		},
		Transaction: &TransactionInfo{
			Index: context.Env.Transaction.Index,
		},
		Contract: ContractInfo{
			Address: context.Env.Contract.Address.String(),
		},
	}
	info := MessageInfo{
		Sender: context.Env.CurrentCall.Sender.String(),
	}
	msgBz := context.Env.CurrentCall.CallData
	envBz, err := json.Marshal(env)
	if err != nil {
		return nil, err
	}
	infoBz, err := json.Marshal(info)
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
	res, err := contractVm.Execute(funcName, envRegion.Pointer, infoRegion.Pointer, msgRegion.Pointer)
	data, err := readMemCw(contractVm, activeMemory, res[0])
	if err != nil {
		return nil, err
	}
	context.ReturnData = data
	return res, nil
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

func readMemCw(vm *wasmedge.VM, mem *wasmedge.Memory, ptr interface{}) ([]byte, error) {
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
	/// The beginning of the region expressed as bytes from the beginning of the linear memory
	Offset uint32
	/// The number of bytes available in this region
	Capacity uint32
	/// The number of bytes used in this region
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
	return mem.SetData(lengthBz, uint(r.Pointer), 4)
}

func (r *Region) Read(mem *wasmedge.Memory) ([]byte, error) {
	if r.Length == 0 {
		return nil, nil
	}
	return mem.GetData(uint(r.Offset), uint(r.Length))
}

// CosmWasm Types

type HumanAddress = string

// Coin is a string representation of the sdk.Coin type (more portable than sdk.Int)
type Coin struct {
	Denom  string `json:"denom"`  // type, eg. "ATOM"
	Amount string `json:"amount"` // string encoing of decimal value, eg. "12.3456"
}

type Coins []Coin

//---------- Env ---------

// Env defines the state of the blockchain environment this contract is
// running in. This must contain only trusted data - nothing from the Tx itself
// that has not been verfied (like Signer).
//
// Env are json encoded to a byte slice before passing to the wasm contract.
type Env struct {
	Block       BlockInfo        `json:"block"`
	Transaction *TransactionInfo `json:"transaction"`
	Contract    ContractInfo     `json:"contract"`
}

type BlockInfo struct {
	// block height this transaction is executed
	Height uint64 `json:"height"`
	// time in nanoseconds since unix epoch. Uses string to ensure JavaScript compatibility.
	Time    uint64 `json:"time,string"`
	ChainID string `json:"chain_id"`
}

type ContractInfo struct {
	// Bech32 encoded sdk.AccAddress of the contract, to be used when sending messages
	Address HumanAddress `json:"address"`
}

type TransactionInfo struct {
	// Position of this transaction in the block.
	// The first transaction has index 0
	//
	// Along with BlockInfo.Height, this allows you to get a unique
	// transaction identifier for the chain for future queries
	Index uint32 `json:"index"`
}

type MessageInfo struct {
	// Bech32 encoded sdk.AccAddress executing the contract
	Sender HumanAddress `json:"sender"`
	// Amount of funds send to the contract along with this message
	Funds Coins `json:"funds"`
}
