package lib

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// InstantiateToken initializes the token contract
func InstantiateToken() []byte {
	calldraw := wasmx.GetCallData()
	LoggerDebug("instantiate token", []string{"args", string(calldraw)})

	var calld CallDataInstantiate
	if err := json.Unmarshal(calldraw, &calld); err != nil {
		Revert("failed to parse instantiate data: " + err.Error())
	}

	SetAdmins(calld.Admins)

	minters := calld.Minters
	if len(minters) == 0 {
		minters = []wasmx.Bech32String{wasmx.GetCaller()}
	}
	SetMinters(minters)

	SetInfo(TokenInfo{
		Name:     calld.Name,
		Symbol:   calld.Symbol,
		Decimals: calld.Decimals,
	})

	return []byte("{}")
}

// GetName returns the token name
func GetName() []byte {
	value := GetInfo()
	resp := MsgNameResponse{Name: value.Name}
	data, _ := json.Marshal(resp)
	return data
}

// GetSymbol returns the token symbol
func GetSymbol() []byte {
	value := GetInfo()
	resp := MsgSymbolResponse{Symbol: value.Symbol}
	data, _ := json.Marshal(resp)
	return data
}

// GetDecimals returns the token decimals
func GetDecimals() []byte {
	value := GetInfo()
	resp := MsgDecimalsResponse{Decimals: value.Decimals}
	data, _ := json.Marshal(resp)
	return data
}

// TotalSupply returns the total token supply
func TotalSupply() []byte {
	value := GetTotalSupply()
	info := GetInfo()
	resp := MsgTotalSupplyResponse{
		Supply: wasmx.Coin{Denom: info.Symbol, Amount: value},
	}
	data, _ := json.Marshal(resp)
	return data
}

// BalanceOf returns the balance of an address
func BalanceOf(req *MsgBalanceOf) []byte {
	value := GetBalance(string(req.Owner))
	info := GetInfo()
	resp := MsgBalanceOfResponse{
		Balance: wasmx.Coin{Denom: info.Symbol, Amount: value},
	}
	data, _ := json.Marshal(resp)
	return data
}

// Transfer transfers tokens to another address
func Transfer(req *MsgTransfer) []byte {
	from := wasmx.GetCaller()
	LoggerDebug("transfer", []string{"from", string(from), "to", string(req.To), "value", req.Value.String()})
	Move(from, req.To, req.Value)
	resp := MsgTransferResponse{}
	data, _ := json.Marshal(resp)
	return data
}

// TransferFrom transfers tokens on behalf of another address
func TransferFrom(req *MsgTransferFrom) []byte {
	spender := wasmx.GetCaller()
	admins := GetAdmins()
	authorized := wasmx.IsAuthorized(spender, admins)

	LoggerDebug("transferFrom", []string{
		"from", string(req.From),
		"to", string(req.To),
		"value", req.Value.String(),
		"caller", string(spender),
		"authorized", fmt.Sprintf("%t", authorized),
	})

	if authorized {
		Move(req.From, req.To, req.Value)
	} else {
		allow := GetAllowance(string(req.From), string(spender))
		if allow.GTE(req.Value) {
			Move(req.From, req.To, req.Value)
			allow = allow.Sub(req.Value)
			SetAllowance(string(req.From), string(spender), allow)
		} else {
			Revert(fmt.Sprintf("insufficient allowance: %s < %s", allow.String(), req.Value.String()))
		}
	}

	resp := MsgTransferFromResponse{}
	data, _ := json.Marshal(resp)
	return data
}

// Approve approves an address to spend tokens
func Approve(req *MsgApprove) []byte {
	owner := wasmx.GetCaller()
	SetAllowance(string(owner), string(req.Spender), req.Value)
	LogApproval(owner, req.Spender, req.Value)
	return []byte("{}")
}

// ApproveFrom allows admins (like identity contract) to set allowance on behalf of owner
func ApproveFrom(req *MsgApproveFrom) []byte {
	caller := wasmx.GetCaller()
	admins := GetAdmins()
	authorized := wasmx.IsAuthorized(caller, admins)

	if !authorized {
		Revert("caller not admin for approveFrom: " + string(caller))
	}

	LoggerDebug("approveFrom", []string{
		"owner", string(req.Owner),
		"spender", string(req.Spender),
		"value", req.Value.String(),
		"caller", string(caller),
	})

	SetAllowance(string(req.Owner), string(req.Spender), req.Value)
	LogApproval(req.Owner, req.Spender, req.Value)
	return []byte("{}")
}

// Allowance returns the allowance for an owner/spender pair
func Allowance(req *MsgAllowance) []byte {
	value := GetAllowance(string(req.Owner), string(req.Spender))
	resp := MsgAllowanceResponse{Remaining: value}
	data, _ := json.Marshal(resp)
	return data
}

// Mint mints new tokens
func Mint(req *MsgMint) []byte {
	caller := wasmx.GetCaller()
	minters := GetMinters()
	authorized := wasmx.IsAuthorized(caller, minters)

	if !authorized {
		Revert("caller cannot mint: " + string(caller))
	}

	LoggerDebug("mint", []string{
		"to", string(req.To),
		"value", req.Value.String(),
		"authorized", fmt.Sprintf("%t", authorized),
	})

	balance := GetBalance(string(req.To))
	balance = balance.Add(req.Value)
	SetBalance(string(req.To), balance)

	supply := GetTotalSupply()
	supply = supply.Add(req.Value)
	SetTotalSupply(supply)

	LogTransfer(wasmx.Bech32String(ZERO_ADDRESS), req.To, req.Value)

	return []byte("{}")
}

// Burn burns tokens
func Burn(req *MsgBurn) []byte {
	caller := wasmx.GetCaller()
	minters := GetMinters()
	authorized := wasmx.IsAuthorized(caller, minters)

	if !authorized {
		Revert("caller cannot burn: " + string(caller))
	}

	LoggerDebug("burn", []string{
		"from", string(req.From),
		"value", req.Value.String(),
		"authorized", fmt.Sprintf("%t", authorized),
	})

	balance := GetBalance(string(req.From))
	if balance.LT(req.Value) {
		Revert(fmt.Sprintf("balance not enough for burning: %s; burning %s", balance.String(), req.Value.String()))
	}

	balance = balance.Sub(req.Value)
	SetBalance(string(req.From), balance)

	supply := GetTotalSupply()
	supply = supply.Sub(req.Value)
	SetTotalSupply(supply)

	LogTransfer(req.From, wasmx.Bech32String(ZERO_ADDRESS), req.Value)

	return []byte("{}")
}

// Move transfers tokens between addresses
func Move(from, to wasmx.Bech32String, amount sdkmath.Int) {
	balanceFrom := GetBalance(string(from))
	if balanceFrom.LT(amount) {
		Revert(fmt.Sprintf("cannot move coins from %s to %s; amount: %s; balance: %s",
			from, to, amount.String(), balanceFrom.String()))
	}

	balanceTo := GetBalance(string(to))
	balanceFrom = balanceFrom.Sub(amount)
	balanceTo = balanceTo.Add(amount)

	SetBalance(string(from), balanceFrom)
	LogTransfer(from, to, amount)
	SetBalance(string(to), balanceTo)
}

// LogTransfer emits a Transfer event
func LogTransfer(from, to wasmx.Bech32String, amount sdkmath.Int) {
	topic0str := "Transfer(address,address,uint256)"
	topic1 := hexToUint8Array32(string(from))
	topic2 := hexToUint8Array32(string(to))
	topic3 := hexToUint8Array32(amount.BigInt().Text(16))

	wasmx.LogWithMsgTopic(topic0str, []byte{}, [][32]byte{topic1, topic2, topic3})
}

// LogApproval emits an Approval event
func LogApproval(owner, spender wasmx.Bech32String, amount sdkmath.Int) {
	topic0str := "Approval(address,address,uint256)"
	topic1 := hexToUint8Array32(string(owner))
	topic2 := hexToUint8Array32(string(spender))
	topic3 := hexToUint8Array32(amount.BigInt().Text(16))

	wasmx.LogWithMsgTopic(topic0str, []byte{}, [][32]byte{topic1, topic2, topic3})
}

// hexToUint8Array32 converts a hex string to a 32-byte array
func hexToUint8Array32(hexStr string) [32]byte {
	var result [32]byte

	// Remove 0x prefix if present
	if len(hexStr) > 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	// Pad to 64 characters (32 bytes)
	if len(hexStr) < 64 {
		hexStr = fmt.Sprintf("%064s", hexStr)
	}

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		// Return zero-filled array on error
		return result
	}

	// Copy to fixed-size array (right-aligned for numbers)
	if len(bytes) < 32 {
		copy(result[32-len(bytes):], bytes)
	} else {
		copy(result[:], bytes[:32])
	}

	return result
}
