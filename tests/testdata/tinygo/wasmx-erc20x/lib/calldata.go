package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// GetCallDataWrap parses the call data and routes to the appropriate function
func GetCallDataWrap() []byte {
	calldraw := wasmx.GetCallData()

	var callData CallData
	if err := json.Unmarshal(calldraw, &callData); err != nil {
		Revert("invalid function call data: " + string(calldraw))
		return nil
	}

	var result []byte

	if callData.Name != nil {
		result = GetName()
	} else if callData.Symbol != nil {
		result = GetSymbol()
	} else if callData.Decimals != nil {
		result = GetDecimals()
	} else if callData.TotalSupply != nil {
		result = TotalSupply()
	} else if callData.BalanceOf != nil {
		result = BalanceOf(callData.BalanceOf)
	} else if callData.SignedBalanceOf != nil {
		result = SignedBalanceOf(callData.SignedBalanceOf)
	} else if callData.Transfer != nil {
		result = Transfer(callData.Transfer)
	} else if callData.TransferFrom != nil {
		result = TransferFrom(callData.TransferFrom)
	} else if callData.Approve != nil {
		result = Approve(callData.Approve)
	} else if callData.Allowance != nil {
		result = Allowance(callData.Allowance)
	} else if callData.Mint != nil {
		result = Mint(callData.Mint)
	} else if callData.Burn != nil {
		result = Burn(callData.Burn)
	} else {
		Revert("invalid function call data: " + string(calldraw))
		return nil
	}

	return result
}
