package vm

import (
	"encoding/binary"
	"encoding/json"
	"strings"

	sdkerr "cosmossdk.io/errors"

	aabi "github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"
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

	vm := context.newIVmFn(context.Ctx)
	defer func() {
		vm.Cleanup()
	}()
	err := vm.InitWasi([]string{}, []string{}, []string{})
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret sharing: cannot initialize WASI")
	}
	err = vm.InstantiateWasm("", wasmbin)
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret sharing: invalid wasm")
	}
	mem, err := vm.GetMemory()
	if err != nil {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "secret sharing: missing memory")
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

		shares, err := ShamirSplit(vm, mem, []byte(args.Secret), args.Count, args.Threshold)
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

		secret, err := ShamirRecover(vm, mem, bz)
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
	return nil, nil
}

func ShamirSplit(vm memc.IVm, mem memc.IMemory, secret []byte, count uint32, threshold uint32) (*ResultShares, error) {
	inputLen := len(secret)
	inputPointer, err := allocateInput(vm, mem, secret)
	if err != nil {
		return nil, err
	}

	// Run the function. Given the pointer to the subject.
	result, err := vm.Call("ShamirSplit", []interface{}{inputPointer, int32(inputLen), int32(count), int32(threshold)})
	if err != nil {
		return nil, err
	}

	memData, err := getResult(mem, result)
	if err != nil {
		return nil, err
	}

	shares := &ResultShares{}
	err = json.Unmarshal(memData, shares)
	if err != nil {
		return nil, err
	}

	// Deallocate the subject, and the output.
	vm.Call("free", []interface{}{inputPointer})
	return shares, nil
}

func ShamirRecover(vm memc.IVm, mem memc.IMemory, input []byte) (*ResultSecret, error) {
	inputLen := len(input)
	inputPointer, err := allocateInput(vm, mem, input)
	if err != nil {
		return nil, err
	}

	// Run the function. Given the pointer to the subject.
	result, err := vm.Call("ShamirRecover", []interface{}{inputPointer, int32(inputLen)})
	if err != nil {
		return nil, err
	}

	memData, err := getResult(mem, result)
	if err != nil {
		return nil, err
	}

	data := &ResultSecret{}
	err = json.Unmarshal(memData, data)
	if err != nil {
		return nil, err
	}

	// Deallocate the subject, and the output.
	vm.Call("free", []interface{}{inputPointer})
	return data, nil
}

func getResult(mem memc.IMemory, result []int32) ([]byte, error) {
	outputPointer := result[0]
	memData, err := mem.Read(outputPointer, LENGTH_SIZE)
	if err != nil {
		return nil, err
	}
	resultLength := binary.BigEndian.Uint32(memData)
	outputPointer = outputPointer + LENGTH_SIZE

	// Read the result
	memData, err = mem.Read(outputPointer, int32(resultLength))
	if err != nil {
		return nil, err
	}
	return memData, nil
}

func allocateInput(vm memc.IVm, mem memc.IMemory, input []byte) (int32, error) {
	inputLen := len(input)

	// Allocate memory for the input, and get a pointer to it.
	// Include a byte for the NULL terminator we add below.
	allocateResult, err := vm.Call(types.MEMORY_EXPORT_MALLOC, []interface{}{int32(inputLen + 1)})
	if err != nil {
		return 0, err
	}
	inputPointer := allocateResult[0]

	// Write the subject into the memory.
	memData, err := mem.Read(inputPointer, int32(inputLen)+1)
	if err != nil {
		return 0, err
	}
	copy(memData, input)

	// C-string terminates by NULL.
	memData[inputLen] = 0

	return inputPointer, nil
}

const SecretSharingAbiStr = `[{"inputs":[{"internalType":"string[]","name":"shares","type":"string[]"}],"name":"shamirRecover","outputs":[{"internalType":"string","name":"secret","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"secret","type":"string"},{"internalType":"uint32","name":"count","type":"uint32"},{"internalType":"uint32","name":"threshold","type":"uint32"}],"name":"shamirSplit","outputs":[{"internalType":"string[]","name":"shares","type":"string[]"}],"stateMutability":"view","type":"function"}]`
