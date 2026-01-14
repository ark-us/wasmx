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
		callerUserID := GetCallerUserID()
		if callerUserID == "" {
			Revert("caller must have a registered identity")
		}
		minters = []string{callerUserID}
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
	value := GetBalance(req.Owner)
	info := GetInfo()
	resp := MsgBalanceOfResponse{
		Balance: wasmx.Coin{Denom: info.Symbol, Amount: value},
	}
	data, _ := json.Marshal(resp)
	return data
}

// Transfer transfers tokens to another address
func Transfer(req *MsgTransfer) []byte {
	from := GetCallerUserID()
	if from == "" {
		Revert("caller must have a registered identity")
	}
	LoggerDebug("transfer", []string{"from", from, "to", req.To, "value", req.Value.String()})
	Move(from, req.To, req.Value)
	resp := MsgTransferResponse{}
	data, _ := json.Marshal(resp)
	return data
}

// TransferFrom transfers tokens on behalf of another address
func TransferFrom(req *MsgTransferFrom) []byte {
	spender := GetCallerUserID()
	if spender == "" {
		Revert("caller must have a registered identity")
	}
	admins := GetAdmins()
	authorized := StringInList(spender, admins)

	LoggerDebug("transferFrom", []string{
		"from", req.From,
		"to", req.To,
		"value", req.Value.String(),
		"caller", spender,
		"authorized", fmt.Sprintf("%t", authorized),
	})

	if authorized {
		Move(req.From, req.To, req.Value)
	} else {
		allow := GetAllowance(req.From, spender)
		if allow.GTE(req.Value) {
			Move(req.From, req.To, req.Value)
			allow = allow.Sub(req.Value)
			SetAllowance(req.From, spender, allow)
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
	owner := GetCallerUserID()
	if owner == "" {
		Revert("caller must have a registered identity")
	}
	SetAllowance(owner, req.Spender, req.Value)
	LogApproval(owner, req.Spender, req.Value)
	return []byte("{}")
}

// ApproveFrom allows admins (like system contracts) to set allowance on behalf of owner
func ApproveFrom(req *MsgApproveFrom) []byte {
	caller := GetCallerUserID()
	if caller == "" {
		Revert("caller must have a registered identity")
	}
	admins := GetAdmins()
	authorized := StringInList(caller, admins)

	if !authorized {
		Revert("caller not admin for approveFrom: " + caller)
	}

	LoggerDebug("approveFrom", []string{
		"owner", req.Owner,
		"spender", req.Spender,
		"value", req.Value.String(),
		"caller", caller,
	})

	SetAllowance(req.Owner, req.Spender, req.Value)
	LogApproval(req.Owner, req.Spender, req.Value)
	return []byte("{}")
}

// Allowance returns the allowance for an owner/spender pair
func Allowance(req *MsgAllowance) []byte {
	value := GetAllowance(req.Owner, req.Spender)
	resp := MsgAllowanceResponse{Remaining: value}
	data, _ := json.Marshal(resp)
	return data
}

// Mint mints new tokens
func Mint(req *MsgMint) []byte {
	caller := GetCallerUserID()
	if caller == "" {
		Revert("caller must have a registered identity")
	}
	minters := GetMinters()
	authorized := StringInList(caller, minters)

	if !authorized {
		Revert("caller cannot mint: " + caller)
	}

	LoggerDebug("mint", []string{
		"to", req.To,
		"value", req.Value.String(),
		"authorized", fmt.Sprintf("%t", authorized),
	})

	balance := GetBalance(req.To)
	balance = balance.Add(req.Value)
	SetBalance(req.To, balance)

	supply := GetTotalSupply()
	supply = supply.Add(req.Value)
	SetTotalSupply(supply)

	LogTransfer(ZERO_ADDRESS, req.To, req.Value)

	return []byte("{}")
}

// Burn burns tokens
func Burn(req *MsgBurn) []byte {
	caller := GetCallerUserID()
	if caller == "" {
		Revert("caller must have a registered identity")
	}
	minters := GetMinters()
	authorized := StringInList(caller, minters)

	if !authorized {
		Revert("caller cannot burn: " + caller)
	}

	LoggerDebug("burn", []string{
		"from", req.From,
		"value", req.Value.String(),
		"authorized", fmt.Sprintf("%t", authorized),
	})

	balance := GetBalance(req.From)
	if balance.LT(req.Value) {
		Revert(fmt.Sprintf("balance not enough for burning: %s; burning %s", balance.String(), req.Value.String()))
	}

	balance = balance.Sub(req.Value)
	SetBalance(req.From, balance)

	supply := GetTotalSupply()
	supply = supply.Sub(req.Value)
	SetTotalSupply(supply)

	LogTransfer(req.From, ZERO_ADDRESS, req.Value)

	return []byte("{}")
}

// Move transfers tokens between users
func Move(from, to string, amount sdkmath.Int) {
	balanceFrom := GetBalance(from)
	if balanceFrom.LT(amount) {
		Revert(fmt.Sprintf("cannot move coins from %s to %s; amount: %s; balance: %s",
			from, to, amount.String(), balanceFrom.String()))
	}

	balanceTo := GetBalance(to)
	balanceFrom = balanceFrom.Sub(amount)
	balanceTo = balanceTo.Add(amount)

	SetBalance(from, balanceFrom)
	LogTransfer(from, to, amount)
	SetBalance(to, balanceTo)
}

// LogTransfer emits a Transfer event
func LogTransfer(from, to string, amount sdkmath.Int) {
	topic0str := "Transfer(address,address,uint256)"
	topic1 := userIDToTopic(from)
	topic2 := userIDToTopic(to)
	topic3 := hexToUint8Array32(amount.BigInt().Text(16))

	wasmx.LogWithMsgTopic(topic0str, []byte{}, [][32]byte{topic1, topic2, topic3})
}

// LogApproval emits an Approval event
func LogApproval(owner, spender string, amount sdkmath.Int) {
	topic0str := "Approval(address,address,uint256)"
	topic1 := userIDToTopic(owner)
	topic2 := userIDToTopic(spender)
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
