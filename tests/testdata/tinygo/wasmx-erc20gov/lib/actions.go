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

	if calld.GovernanceAddress == "" {
		Revert("governance_address is required")
	}

	SetGovernanceAddress(calld.GovernanceAddress)

	admins := calld.Admins
	if len(admins) == 0 {
		callerUserID := GetCallerUserID()
		if callerUserID == "" {
			Revert("caller must have a registered identity")
		}
		admins = []string{callerUserID}
	}
	SetAdmins(admins)

	SetInfo(TokenInfo{
		Name:     calld.Name,
		Symbol:   calld.Symbol,
		Decimals: calld.Decimals,
	})

	totalSupply := GetTotalSupply()
	for _, bal := range calld.InitialBalances {
		if bal.UserID == "" {
			continue
		}
		amt, ok := sdkmath.NewIntFromString(bal.Amount)
		if !ok {
			Revert("invalid initial balance amount: " + bal.Amount)
		}
		if amt.IsNegative() {
			Revert("initial balance must be non-negative")
		}
		if amt.IsZero() {
			continue
		}
		SetBalance(bal.UserID, amt)
		totalSupply = totalSupply.Add(amt)
	}
	SetTotalSupply(totalSupply)

	return []byte("{}")
}

// ============================================================================
// Query Functions
// ============================================================================

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

// MemberCount returns the count of members with balance > 0
func MemberCount() []byte {
	count := GetMemberCount()
	resp := MsgMemberCountResponse{Count: count}
	data, _ := json.Marshal(resp)
	return data
}

// BalanceOf returns the balance of an address
func BalanceOf(req *MsgBalanceOf) []byte {
	value := GetBalance(req.Owner)
	if value.IsZero() {
		if userID, _ := QueryUserByAddress(req.Owner); userID != "" && userID != req.Owner {
			value = GetBalance(userID)
		}
	}
	info := GetInfo()
	resp := MsgBalanceOfResponse{
		Balance: wasmx.Coin{Denom: info.Symbol, Amount: value},
	}
	data, _ := json.Marshal(resp)
	return data
}

// QueryGetGovernanceAddress returns the governance contract address
func QueryGetGovernanceAddress() []byte {
	address := GetGovernanceAddress()
	resp := MsgGetGovernanceAddressResponse{GovernanceAddress: address}
	data, _ := json.Marshal(resp)
	return data
}

// QueryGetAdmins returns the admin list
func QueryGetAdmins() []byte {
	admins := GetAdmins()
	resp := MsgGetAdminsResponse{Admins: admins}
	data, _ := json.Marshal(resp)
	return data
}

// ============================================================================
// Governance-controlled Functions
// ============================================================================

// Transfer transfers tokens between addresses - ONLY governance can call this
func Transfer(req *MsgTransfer) []byte {
	RequireGovernance()

	LoggerDebug("transfer (gov)", []string{
		"from", req.From,
		"to", req.To,
		"value", req.Value.String(),
	})

	from := ResolveUserIDOrAddress(req.From)
	to := ResolveUserIDOrAddress(req.To)
	Move(from, to, req.Value)
	resp := MsgTransferResponse{}
	data, _ := json.Marshal(resp)
	return data
}

// Mint mints new tokens - ONLY governance can call this
func Mint(req *MsgMint) []byte {
	RequireGovernance()

	LoggerDebug("mint (gov)", []string{
		"to", req.To,
		"value", req.Value.String(),
	})

	to := ResolveUserIDOrAddress(req.To)
	balance := GetBalance(to)
	balance = balance.Add(req.Value)
	SetBalance(to, balance)

	supply := GetTotalSupply()
	supply = supply.Add(req.Value)
	SetTotalSupply(supply)

	LogTransfer(ZERO_ADDRESS, to, req.Value)

	return []byte("{}")
}

// Burn burns tokens - ONLY governance can call this
func Burn(req *MsgBurn) []byte {
	RequireGovernance()

	LoggerDebug("burn (gov)", []string{
		"from", req.From,
		"value", req.Value.String(),
	})

	from := ResolveUserIDOrAddress(req.From)
	balance := GetBalance(from)
	if balance.LT(req.Value) {
		Revert(fmt.Sprintf("balance not enough for burning: %s; burning %s", balance.String(), req.Value.String()))
	}

	balance = balance.Sub(req.Value)
	SetBalance(from, balance)

	supply := GetTotalSupply()
	supply = supply.Sub(req.Value)
	SetTotalSupply(supply)

	LogTransfer(from, ZERO_ADDRESS, req.Value)

	return []byte("{}")
}

// DoSetGovernanceAddress updates the governance address - ONLY old governance can call this
func DoSetGovernanceAddress(req *MsgSetGovernanceAddress) []byte {
	RequireGovernance()

	if req.NewGovernanceAddress == "" {
		Revert("new governance address cannot be empty")
	}

	oldAddress := GetGovernanceAddress()
	SetGovernanceAddress(req.NewGovernanceAddress)

	LoggerInfo("governance address updated", []string{
		"old", oldAddress,
		"new", req.NewGovernanceAddress,
	})

	resp := MsgSetGovernanceAddressResponse{Success: true}
	data, _ := json.Marshal(resp)
	return data
}

// DoSetAdmins updates the admin list - ONLY governance can call this
func DoSetAdmins(req *MsgSetAdmins) []byte {
	RequireGovernance()

	SetAdmins(req.Admins)

	LoggerInfo("admins updated", []string{
		"count", fmt.Sprintf("%d", len(req.Admins)),
	})

	resp := MsgSetAdminsResponse{Success: true}
	data, _ := json.Marshal(resp)
	return data
}

// ============================================================================
// Admin-controlled Functions
// ============================================================================

// TransferFrom transfers tokens on behalf of another address - ONLY admins can call this
func TransferFrom(req *MsgTransferFrom) []byte {
	RequireAdminOrGovernance()

	callerUserID := GetCallerUserID()

	LoggerDebug("transferFrom", []string{
		"from", req.From,
		"to", req.To,
		"value", req.Value.String(),
		"caller", callerUserID,
	})

	from := ResolveUserIDOrAddress(req.From)
	to := ResolveUserIDOrAddress(req.To)
	Move(from, to, req.Value)

	resp := MsgTransferFromResponse{}
	data, _ := json.Marshal(resp)
	return data
}

// ============================================================================
// Internal Functions
// ============================================================================

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
