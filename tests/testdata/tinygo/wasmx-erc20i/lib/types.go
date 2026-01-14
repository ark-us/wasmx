package lib

import (
	sdkmath "cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "erc20i"

const ZERO_ADDRESS = "user_0"

// Storage keys
const (
	ALLOWANCE_KEY    = "allowance_"
	BALANCE_KEY      = "balance_"
	TOTAL_SUPPLY_KEY = "totalSupply"
	ADMINS_KEY       = "admins"
	MINTERS_KEY      = "minters"
	INFO_KEY         = "info"
)

// CallDataInstantiate is the initialization message
type CallDataInstantiate struct {
	Admins    []string `json:"admins"`
	Minters   []string `json:"minters"`
	Name      string   `json:"name"`
	Symbol    string   `json:"symbol"`
	Decimals  int32    `json:"decimals"`
	BaseDenom string   `json:"base_denom"`
}

// TokenInfo stores basic token information
type TokenInfo struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int32  `json:"decimals"`
}

// Message types

type MsgName struct{}

type MsgNameResponse struct {
	Name string `json:"name"`
}

type MsgSymbol struct{}

type MsgSymbolResponse struct {
	Symbol string `json:"symbol"`
}

type MsgDecimals struct{}

type MsgDecimalsResponse struct {
	Decimals int32 `json:"decimals"`
}

type MsgTotalSupply struct{}

type MsgTotalSupplyResponse struct {
	Supply wasmx.Coin `json:"supply"`
}

type MsgBalanceOf struct {
	Owner string `json:"owner"`
}

type MsgBalanceOfResponse struct {
	Balance wasmx.Coin `json:"balance"`
}

type MsgTransfer struct {
	To    string     `json:"to"`
	Value sdkmath.Int `json:"value"`
}

type MsgTransferResponse struct{}

type MsgTransferFrom struct {
	From  string     `json:"from"`
	To    string     `json:"to"`
	Value sdkmath.Int `json:"value"`
}

type MsgTransferFromResponse struct{}

type MsgApprove struct {
	Spender string     `json:"spender"`
	Value   sdkmath.Int `json:"value"`
}

type MsgApproveResponse struct {
	Success bool `json:"success"`
}

// MsgApproveFrom allows admins (like system contracts) to set allowance on behalf of owner
type MsgApproveFrom struct {
	Owner   string     `json:"owner"`
	Spender string     `json:"spender"`
	Value   sdkmath.Int `json:"value"`
}

type MsgApproveFromResponse struct {
	Success bool `json:"success"`
}

type MsgAllowance struct {
	Owner   string `json:"owner"`
	Spender string `json:"spender"`
}

type MsgAllowanceResponse struct {
	Remaining sdkmath.Int `json:"remaining"`
}

type MsgMint struct {
	To    string     `json:"to"`
	Value sdkmath.Int `json:"value"`
}

type MsgBurn struct {
	From  string     `json:"from"`
	Value sdkmath.Int `json:"value"`
}

// CallData structure for routing
type CallData struct {
	Name         *MsgName         `json:"name,omitempty"`
	Symbol       *MsgSymbol       `json:"symbol,omitempty"`
	Decimals     *MsgDecimals     `json:"decimals,omitempty"`
	TotalSupply  *MsgTotalSupply  `json:"totalSupply,omitempty"`
	BalanceOf    *MsgBalanceOf    `json:"balanceOf,omitempty"`
	Transfer     *MsgTransfer     `json:"transfer,omitempty"`
	TransferFrom *MsgTransferFrom `json:"transferFrom,omitempty"`
	Approve      *MsgApprove      `json:"approve,omitempty"`
	ApproveFrom  *MsgApproveFrom  `json:"approveFrom,omitempty"`
	Allowance    *MsgAllowance    `json:"allowance,omitempty"`
	Mint         *MsgMint         `json:"mint,omitempty"`
	Burn         *MsgBurn         `json:"burn,omitempty"`
}
