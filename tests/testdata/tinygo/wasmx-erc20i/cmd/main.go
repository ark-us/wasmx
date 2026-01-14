package main

import (
	"encoding/json"
	"os"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	"github.com/loredanacirstea/wasmx/tests/testdata/tinygo/wasmx-erc20i/lib"
)

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

func main() {
	// Handle internal entry points
	entrypoint := os.Getenv("ENTRY_POINT")
	switch entrypoint {
	case "instantiate":
		lib.InstantiateToken()
		return
	}

	databz := wasmx.GetCallData()

	// Handle empty data
	if len(databz) == 0 {
		wasmx.Finish([]byte(`{"success": true}`))
		return
	}

	var calldata lib.CallData
	if err := json.Unmarshal(databz, &calldata); err != nil {
		lib.Revert("invalid function call data: " + string(databz))
		return
	}

	var result []byte

	switch {
	case calldata.Name != nil:
		result = lib.GetName()
	case calldata.Symbol != nil:
		result = lib.GetSymbol()
	case calldata.Decimals != nil:
		result = lib.GetDecimals()
	case calldata.TotalSupply != nil:
		result = lib.TotalSupply()
	case calldata.BalanceOf != nil:
		result = lib.BalanceOf(calldata.BalanceOf)
	case calldata.Transfer != nil:
		result = lib.Transfer(calldata.Transfer)
	case calldata.TransferFrom != nil:
		result = lib.TransferFrom(calldata.TransferFrom)
	case calldata.Approve != nil:
		result = lib.Approve(calldata.Approve)
	case calldata.ApproveFrom != nil:
		result = lib.ApproveFrom(calldata.ApproveFrom)
	case calldata.Allowance != nil:
		result = lib.Allowance(calldata.Allowance)
	case calldata.Mint != nil:
		result = lib.Mint(calldata.Mint)
	case calldata.Burn != nil:
		result = lib.Burn(calldata.Burn)
	default:
		lib.Revert("invalid function call data: " + string(databz))
		return
	}

	wasmx.Finish(result)
}
