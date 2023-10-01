package wasmutils

import (
	sdkerrors "cosmossdk.io/errors"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

func InstantiateWasm(contractVm *wasmedge.VM, filePath string, wasmbuffer []byte) error {
	var err error
	if wasmbuffer == nil {
		err = contractVm.LoadWasmFile(filePath)
		if err != nil {
			return sdkerrors.Wrapf(err, "load wasm file failed %s", filePath)
		}
	} else {
		err = contractVm.LoadWasmBuffer(wasmbuffer)
		if err != nil {
			return sdkerrors.Wrapf(err, "load wasm file failed from buffer")
		}
	}
	err = contractVm.Validate()
	if err != nil {
		return sdkerrors.Wrapf(err, "wasm validate failed")
	}
	err = contractVm.Instantiate()
	if err != nil {
		return sdkerrors.Wrapf(err, "wasm instantiate failed")
	}
	return nil
}
