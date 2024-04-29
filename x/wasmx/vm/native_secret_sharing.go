package vm

import (
	"encoding/binary"
	"encoding/json"
	"strings"

	sdkerr "cosmossdk.io/errors"

	aabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/second-state/WasmEdge-go/wasmedge"

	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/precompiles"
	"mythos/v1/x/wasmx/vm/wasmutils"
)

const LENGTH_SIZE = 4

type InputShamirSplit struct {
	Secret    string `json:"secret"`
	Count     uint32 `json:"count"`
	Threshold uint32 `json:"threshold"`
}

type InputShamirRecover struct {
	Shares []string `json:"shares"`
}

type ResultShares struct {
	Shares []string `json:"shares"`
}

type ResultSecret struct {
	Secret string `json:"secret"`
}

var SecretSharingAbi aabi.ABI

func init() {
	var err error
	SecretSharingAbi, err = aabi.JSON(strings.NewReader(SecretSharingAbiStr))
	if err != nil {
		panic(err)
	}
}

func SecretSharing(context *Context, input []byte) ([]byte, error) {
	wasmbin := precompiles.GetPrecompileByLabel(context.CosmosHandler.AddressCodec(), "secret_sharing")
	wasmedge.SetLogErrorLevel()
	conf := wasmedge.NewConfigure(wasmedge.WASI)
	vm := wasmedge.NewVMWithConfig(conf)
	err := wasmutils.InstantiateWasm(vm, "", wasmbin)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret sharing: invalid wasm")
	}

	fabi, err := SecretSharingAbi.MethodById(input[0:4])
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret sharing: abi method not found")
	}

	switch fabi.RawName {
	case "shamirSplit": // "shamirSplit(string,uint32,uint32)"
		unpacked, err := fabi.Inputs.Unpack(input[4:])
		if err != nil {
			return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret sharing: cannot unpack input")
		}

		var args InputShamirSplit
		err = fabi.Inputs.Copy(&args, unpacked)
		if err != nil {
			return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret sharing: cannot unpack input")
		}

		shares, err := ShamirSplit(vm, []byte(args.Secret), args.Count, args.Threshold)
		if err != nil {
			return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret sharing: cannot split")
		}
		result, err := fabi.Outputs.Pack(shares.Shares)
		if err != nil {
			return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret sharing: cannot pack output")
		}
		return result, nil
	case "shamirRecover": // "shamirRecover(string[])"
		unpacked, err := fabi.Inputs.Unpack(input[4:])
		if err != nil {
			return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret recover: cannot unpack input")
		}

		var args InputShamirRecover
		err = fabi.Inputs.Copy(&args, unpacked)
		if err != nil {
			return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret recover: cannot unpack input")
		}
		bz, err := json.Marshal(args)
		if err != nil {
			return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret recover: marshaling failed")
		}

		secret, err := ShamirRecover(vm, bz)
		if err != nil {
			// TODO return error? or just empty result
			return nil, nil
		}
		result, err := fabi.Outputs.Pack(secret.Secret)
		if err != nil {
			return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret recover: cannot unpack output")
		}
		return result, nil
	}

	vm.Release()
	conf.Release()

	return nil, nil
}

func ShamirSplit(vm *wasmedge.VM, secret []byte, count uint32, threshold uint32) (*ResultShares, error) {
	inputLen := len(secret)
	inputPointer, err := allocateInput(vm, secret)
	if err != nil {
		return nil, err
	}

	// Run the function. Given the pointer to the subject.
	result, err := vm.Execute("ShamirSplit", inputPointer, int32(inputLen), int32(count), int32(threshold))
	if err != nil {
		return nil, err
	}

	memData, err := getResult(vm, result)
	if err != nil {
		return nil, err
	}

	shares := &ResultShares{}
	err = json.Unmarshal(memData, shares)
	if err != nil {
		return nil, err
	}

	// Deallocate the subject, and the output.
	vm.Execute("free", inputPointer)
	return shares, nil
}

func ShamirRecover(vm *wasmedge.VM, input []byte) (*ResultSecret, error) {
	inputLen := len(input)
	inputPointer, err := allocateInput(vm, input)
	if err != nil {
		return nil, err
	}

	// Run the function. Given the pointer to the subject.
	result, err := vm.Execute("ShamirRecover", inputPointer, int32(inputLen))
	if err != nil {
		return nil, err
	}

	memData, err := getResult(vm, result)
	if err != nil {
		return nil, err
	}

	data := &ResultSecret{}
	err = json.Unmarshal(memData, data)
	if err != nil {
		return nil, err
	}

	// Deallocate the subject, and the output.
	vm.Execute("free", inputPointer)
	return data, nil
}

func getResult(vm *wasmedge.VM, result []interface{}) ([]byte, error) {
	mod := vm.GetActiveModule()
	mem := mod.FindMemory("memory")

	outputPointer := result[0].(int32)
	memData, err := mem.GetData(uint(outputPointer), LENGTH_SIZE)
	if err != nil {
		return nil, err
	}
	resultLength := binary.BigEndian.Uint32(memData)
	outputPointer = outputPointer + LENGTH_SIZE

	// Read the result
	memData, err = mem.GetData(uint(outputPointer), uint(resultLength))
	if err != nil {
		return nil, err
	}
	return memData, nil
}

func allocateInput(vm *wasmedge.VM, input []byte) (int32, error) {
	inputLen := len(input)

	// Allocate memory for the input, and get a pointer to it.
	// Include a byte for the NULL terminator we add below.
	allocateResult, err := vm.Execute(types.MEMORY_EXPORT_MALLOC, int32(inputLen+1))
	if err != nil {
		return 0, err
	}
	inputPointer := allocateResult[0].(int32)

	// Write the subject into the memory.
	mod := vm.GetActiveModule()
	mem := mod.FindMemory("memory")
	memData, err := mem.GetData(uint(inputPointer), uint(inputLen+1))
	if err != nil {
		return 0, err
	}
	copy(memData, input)

	// C-string terminates by NULL.
	memData[inputLen] = 0

	return inputPointer, nil
}

const SecretSharingAbiStr = `[{"inputs":[{"internalType":"string[]","name":"shares","type":"string[]"}],"name":"shamirRecover","outputs":[{"internalType":"string","name":"secret","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"secret","type":"string"},{"internalType":"uint32","name":"count","type":"uint32"},{"internalType":"uint32","name":"threshold","type":"uint32"}],"name":"shamirSplit","outputs":[{"internalType":"string[]","name":"shares","type":"string[]"}],"stateMutability":"view","type":"function"}]`
