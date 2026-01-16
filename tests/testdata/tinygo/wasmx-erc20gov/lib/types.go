package lib

import (
	sdkmath "cosmossdk.io/math"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	erc20i "github.com/loredanacirstea/wasmx/tests/testdata/tinygo/wasmx-erc20i/lib"
)

const MODULE_NAME = "erc20gov"

const ZERO_ADDRESS = erc20i.ZERO_ADDRESS

// Storage keys (reuse from erc20i where possible)
const (
	BALANCE_KEY            = erc20i.BALANCE_KEY
	TOTAL_SUPPLY_KEY       = erc20i.TOTAL_SUPPLY_KEY
	ADMINS_KEY             = erc20i.ADMINS_KEY
	INFO_KEY               = erc20i.INFO_KEY
	MEMBER_COUNT_KEY       = "memberCount"
	GOVERNANCE_ADDRESS_KEY = "governance_address"
)

// Reuse types from erc20i
type TokenInfo = erc20i.TokenInfo
type MsgName = erc20i.MsgName
type MsgNameResponse = erc20i.MsgNameResponse
type MsgSymbol = erc20i.MsgSymbol
type MsgSymbolResponse = erc20i.MsgSymbolResponse
type MsgDecimals = erc20i.MsgDecimals
type MsgDecimalsResponse = erc20i.MsgDecimalsResponse
type MsgTotalSupply = erc20i.MsgTotalSupply
type MsgTotalSupplyResponse = erc20i.MsgTotalSupplyResponse
type MsgBalanceOf = erc20i.MsgBalanceOf
type MsgBalanceOfResponse = erc20i.MsgBalanceOfResponse
type MsgMint = erc20i.MsgMint
type MsgBurn = erc20i.MsgBurn
type MsgTransferFromResponse = erc20i.MsgTransferFromResponse

type InitialBalance struct {
	UserID string `json:"user_id"`
	Amount string `json:"amount"`
}

// CallDataInstantiate is the initialization message (different from erc20i)
type CallDataInstantiate struct {
	Admins            []string         `json:"admins"`
	Name              string           `json:"name"`
	Symbol            string           `json:"symbol"`
	Decimals          int32            `json:"decimals"`
	GovernanceAddress string           `json:"governance_address"`
	InitialBalances   []InitialBalance `json:"initial_balances,omitempty"`
}

// MsgMemberCount - unique to erc20gov
type MsgMemberCount struct{}

type MsgMemberCountResponse struct {
	Count uint64 `json:"count"`
}

// MsgTransfer - only governance can call this
type MsgTransfer struct {
	From  string      `json:"from"`
	To    string      `json:"to"`
	Value sdkmath.Int `json:"value"`
}

type MsgTransferResponse struct{}

// MsgTransferFrom - only admins can call this
type MsgTransferFrom struct {
	From  string      `json:"from"`
	To    string      `json:"to"`
	Value sdkmath.Int `json:"value"`
}

// MsgGetGovernanceAddress - query the governance address
type MsgGetGovernanceAddress struct{}

type MsgGetGovernanceAddressResponse struct {
	GovernanceAddress string `json:"governance_address"`
}

// MsgSetGovernanceAddress - only old governance can call this
type MsgSetGovernanceAddress struct {
	NewGovernanceAddress string `json:"new_governance_address"`
}

type MsgSetGovernanceAddressResponse struct {
	Success bool `json:"success"`
}

// MsgGetAdmins - query the admin list
type MsgGetAdmins struct{}

type MsgGetAdminsResponse struct {
	Admins []string `json:"admins"`
}

// MsgSetAdmins - only governance can call this
type MsgSetAdmins struct {
	Admins []string `json:"admins"`
}

type MsgSetAdminsResponse struct {
	Success bool `json:"success"`
}

// CallData structure for routing
type CallData struct {
	// Queries
	Name                 *MsgName                 `json:"name,omitempty"`
	Symbol               *MsgSymbol               `json:"symbol,omitempty"`
	Decimals             *MsgDecimals             `json:"decimals,omitempty"`
	TotalSupply          *MsgTotalSupply          `json:"totalSupply,omitempty"`
	MemberCount          *MsgMemberCount          `json:"memberCount,omitempty"`
	BalanceOf            *MsgBalanceOf            `json:"balanceOf,omitempty"`
	GetGovernanceAddress *MsgGetGovernanceAddress `json:"getGovernanceAddress,omitempty"`
	GetAdmins            *MsgGetAdmins            `json:"getAdmins,omitempty"`

	// Governance-controlled actions
	Transfer *MsgTransfer `json:"transfer,omitempty"`
	Mint     *MsgMint     `json:"mint,omitempty"`
	Burn     *MsgBurn     `json:"burn,omitempty"`

	// Admin-controlled actions
	TransferFrom *MsgTransferFrom `json:"transferFrom,omitempty"`

	// Governance-only admin functions
	SetGovernanceAddress *MsgSetGovernanceAddress `json:"setGovernanceAddress,omitempty"`
	SetAdmins            *MsgSetAdmins            `json:"setAdmins,omitempty"`

	// Hooks
	RoleChanged *wasmx.RolesChangedHook `json:"RoleChanged,omitempty"`
}
