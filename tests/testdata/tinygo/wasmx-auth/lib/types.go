package lib

import (
	"encoding/json"
	"strconv"

	wasmx "github.com/loredanacirstea/wasmx-env"
	"github.com/loredanacirstea/wasmx-utils"
)

const MODULE_NAME = "auth"
const ModuleAccountTypeName = "ModuleAccount"
const BaseAccountTypeName = "BaseAccount"
const TypeUrl_BaseAccount = "/mythos.cosmosmod.v1.BaseAccount"
const TypeUrl_ModuleAccount = "/mythos.cosmosmod.v1.ModuleAccount"

// GenesisState represents the auth module's genesis state
type GenesisState struct {
	Params                 Params         `json:"params"`
	Accounts               []wasmx.AnyWrap `json:"accounts"`
	BaseAccountTypeurl     string         `json:"base_account_typeurl"`
	ModuleAccountTypeurl   string         `json:"module_account_typeurl"`
}

// BaseAccountExternal is used for JSON serialization with string fields
type BaseAccountExternal struct {
	Address       wasmx.Bech32String  `json:"address"`
	PubKey        *wasmx.PublicKey    `json:"pub_key"`
	AccountNumber string              `json:"account_number"`
	Sequence      string              `json:"sequence"`
}

// BaseAccount represents a basic account
type BaseAccount struct {
	Address       wasmx.Bech32String  `json:"address"`
	PubKey        *wasmx.PublicKey    `json:"pub_key"`
	AccountNumber uint64              `json:"account_number"`
	Sequence      uint64              `json:"sequence"`
}

// NewBaseAccount creates a new BaseAccount
func NewBaseAccount(addr wasmx.Bech32String) BaseAccount {
	return BaseAccount{
		Address:       addr,
		PubKey:        nil,
		AccountNumber: 0,
		Sequence:      0,
	}
}

// MarshalJSON implements custom JSON marshaling for BaseAccount
func (ba BaseAccount) MarshalJSON() ([]byte, error) {
	external := BaseAccountExternal{
		Address:       ba.Address,
		PubKey:        ba.PubKey,
		AccountNumber: utils.U64toa(ba.AccountNumber),
		Sequence:      utils.U64toa(ba.Sequence),
	}
	return json.Marshal(external)
}

// UnmarshalJSON implements custom JSON unmarshaling for BaseAccount
func (ba *BaseAccount) UnmarshalJSON(data []byte) error {
	var external BaseAccountExternal
	if err := json.Unmarshal(data, &external); err != nil {
		return err
	}
	
	accountNo, err := strconv.ParseUint(external.AccountNumber, 10, 64)
	if err != nil {
		return err
	}
	
	seq, err := strconv.ParseUint(external.Sequence, 10, 64)
	if err != nil {
		return err
	}
	
	ba.Address = external.Address
	ba.PubKey = external.PubKey
	ba.AccountNumber = accountNo
	ba.Sequence = seq
	return nil
}

// ModuleAccount represents a module account
type ModuleAccount struct {
	BaseAccount BaseAccount `json:"base_account"`
	Name        string      `json:"name"`
	Permissions []string    `json:"permissions"`
}

// ModuleCredential represents module credentials
type ModuleCredential struct {
	ModuleName     string   `json:"module_name"`
	DerivationKeys []string `json:"derivation_keys"`
}

// ParamsExternal is used for JSON serialization with string fields
type ParamsExternal struct {
	MaxMemoCharacters     string `json:"max_memo_characters"`
	TxSigLimit            string `json:"tx_sig_limit"`
	TxSizeCostPerByte     string `json:"tx_size_cost_per_byte"`
	SigVerifyCostEd25519  string `json:"sig_verify_cost_ed25519"`
	SigVerifyCostSecp256k1 string `json:"sig_verify_cost_secp256k1"`
}

// Params represents the auth module parameters
type Params struct {
	MaxMemoCharacters      uint64 `json:"max_memo_characters"`
	TxSigLimit             uint64 `json:"tx_sig_limit"`
	TxSizeCostPerByte      uint64 `json:"tx_size_cost_per_byte"`
	SigVerifyCostEd25519   uint64 `json:"sig_verify_cost_ed25519"`
	SigVerifyCostSecp256k1 uint64 `json:"sig_verify_cost_secp256k1"`
}

// MarshalJSON implements custom JSON marshaling for Params
func (p Params) MarshalJSON() ([]byte, error) {
	external := ParamsExternal{
		MaxMemoCharacters:      utils.U64toa(p.MaxMemoCharacters),
		TxSigLimit:             utils.U64toa(p.TxSigLimit),
		TxSizeCostPerByte:      utils.U64toa(p.TxSizeCostPerByte),
		SigVerifyCostEd25519:   utils.U64toa(p.SigVerifyCostEd25519),
		SigVerifyCostSecp256k1: utils.U64toa(p.SigVerifyCostSecp256k1),
	}
	return json.Marshal(external)
}

// UnmarshalJSON implements custom JSON unmarshaling for Params
func (p *Params) UnmarshalJSON(data []byte) error {
	var external ParamsExternal
	if err := json.Unmarshal(data, &external); err != nil {
		return err
	}
	
	mmc, err := strconv.ParseUint(external.MaxMemoCharacters, 10, 64)
	if err != nil {
		return err
	}
	
	tsl, err := strconv.ParseUint(external.TxSigLimit, 10, 64)
	if err != nil {
		return err
	}
	
	tscb, err := strconv.ParseUint(external.TxSizeCostPerByte, 10, 64)
	if err != nil {
		return err
	}
	
	svced, err := strconv.ParseUint(external.SigVerifyCostEd25519, 10, 64)
	if err != nil {
		return err
	}
	
	svcsecp, err := strconv.ParseUint(external.SigVerifyCostSecp256k1, 10, 64)
	if err != nil {
		return err
	}
	
	p.MaxMemoCharacters = mmc
	p.TxSigLimit = tsl
	p.TxSizeCostPerByte = tscb
	p.SigVerifyCostEd25519 = svced
	p.SigVerifyCostSecp256k1 = svcsecp
	return nil
}

// Message types
type MsgUpdateParams struct {
	Authority string `json:"authority"`
	Params    Params `json:"params"`
}

type MsgUpdateParamsResponse struct{}

type MsgSetAccount struct {
	Account wasmx.AnyWrap `json:"account"`
}

type MsgNewBaseAccount struct {
	Address wasmx.Bech32String `json:"address"`
}

type MsgNewModuleAccount struct {
	Address     wasmx.Bech32String `json:"address"`
	Name        string             `json:"name"`
	Permissions []string           `json:"permissions"`
}

// Query types
type QueryAccountsRequest struct {
	Pagination wasmx.PageRequest `json:"pagination"`
}

type QueryAccountsResponse struct {
	Accounts   []wasmx.AnyWrap    `json:"accounts"`
	Pagination wasmx.PageResponse `json:"pagination"`
}

type QueryAccountRequest struct {
	Address string `json:"address"`
}

type QueryAccountResponse struct {
	Account *wasmx.AnyWrap `json:"account"`
}

type QueryHasAccountResponse struct {
	Found bool `json:"found"`
}

type QueryAccountAddressByIDRequest struct {
	ID        int64  `json:"id"`
	AccountID uint64 `json:"account_id"`
}

type QueryAccountAddressByIDResponse struct {
	AccountAddress string `json:"account_address"`
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QueryModuleAccountsRequest struct{}

type QueryModuleAccountsResponse struct {
	Accounts []wasmx.AnyWrap `json:"accounts"`
}

type QueryModuleAccountByNameRequest struct {
	Name string `json:"name"`
}

type QueryModuleAccountByNameResponse struct {
	Account *wasmx.AnyWrap `json:"account"`
}

type Bech32PrefixRequest struct{}

type Bech32PrefixResponse struct {
	Bech32Prefix string `json:"bech32_prefix"`
}

type AddressBytesToStringRequest struct {
	AddressBytes string `json:"address_bytes"` // base64 string
}

type AddressBytesToStringResponse struct {
	AddressString string `json:"address_string"`
}

type AddressStringToBytesRequest struct {
	AddressString string `json:"address_string"`
}

type AddressStringToBytesResponse struct {
	AddressBytes string `json:"address_bytes"` // base64 string
}

type QueryAccountInfoRequest struct {
	Address string `json:"address"`
}

type QueryAccountInfoResponse struct {
	Info BaseAccount `json:"info"`
}

// Utility function to create a new BaseAccount wrapped in AnyWrap
func NewBaseAccountAnyWrap(typeUrl string, addr wasmx.Bech32String) wasmx.AnyWrap {
	data := NewBaseAccount(addr)
	jsonData, _ := json.Marshal(data)
	return wasmx.NewAnyWrap(typeUrl, string(jsonData))
}

// Calldata structure
type CallData struct {
	GetParams *MsgGetParams `json:"GetParams"`

	InitGenesis *MsgInitGenesis `json:"InitGenesis"`
}

type MsgInitGenesis struct{}

type MsgGetParams struct{}
