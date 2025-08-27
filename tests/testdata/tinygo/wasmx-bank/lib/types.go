package lib

import (
	"encoding/json"
	"strconv"

	sdkmath "cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	utils "github.com/loredanacirstea/wasmx-utils"
)

const MODULE_NAME = "bank"

// Type alias for CoinMap
type CoinMap map[string]sdkmath.Int

// DenomUnit_ represents a denomination unit with a value
type DenomUnit_ struct {
	Denom string      `json:"denom"`
	Value sdkmath.Int `json:"value"` // 10^exponent
}

// DenomInfo represents denomination information
type DenomInfo struct {
	Denom    string             `json:"denom"`
	Value    sdkmath.Int        `json:"value"` // 10^exponent
	Contract wasmx.Bech32String `json:"contract"`
}

// Input represents transaction input
type Input struct {
	Address string       `json:"address"`
	Coins   []wasmx.Coin `json:"coins"`
}

// Output represents transaction output
type Output struct {
	Address string       `json:"address"`
	Coins   []wasmx.Coin `json:"coins"`
}

// MsgSend represents a send message
type MsgSend struct {
	FromAddress wasmx.Bech32String `json:"from_address"`
	ToAddress   wasmx.Bech32String `json:"to_address"`
	Amount      []wasmx.Coin       `json:"amount"`
}

type MsgSendResponse struct{}

// MsgMultiSend represents a multi-send message
type MsgMultiSend struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

type MsgMultiSendResponse struct{}

// MsgUpdateParams represents a parameter update message
type MsgUpdateParams struct {
	Authority string `json:"authority"`
	Params    Params `json:"params"`
}

type MsgUpdateParamsResponse struct{}

// MsgSetSendEnabled represents a send enabled update message
type MsgSetSendEnabled struct {
	Authority     string        `json:"authority"`
	SendEnabled   []SendEnabled `json:"send_enabled"`
	UseDefaultFor []string      `json:"use_default_for"`
}

type MsgSetSendEnabledResponse struct{}

// SendEnabled represents send enabled configuration
type SendEnabled struct {
	Denom   string `json:"denom"`
	Enabled bool   `json:"enabled"`
}

// Balance represents account balance
type Balance struct {
	Address wasmx.Bech32String `json:"address"`
	Coins   []wasmx.Coin       `json:"coins"`
}

// Params represents bank module parameters
type Params struct {
	DefaultSendEnabled bool          `json:"default_send_enabled"`
	SendEnabled        []SendEnabled `json:"send_enabled"`
}

// DenomUnit represents a denomination unit
type DenomUnit struct {
	Denom    string   `json:"denom"`
	Exponent uint32   `json:"exponent"`
	Aliases  []string `json:"aliases"`
}

// Metadata represents token metadata
type Metadata struct {
	Description string      `json:"description"`
	DenomUnits  []DenomUnit `json:"denom_units"`
	Base        string      `json:"base"`
	Display     string      `json:"display"`
	Name        string      `json:"name"`
	Symbol      string      `json:"symbol"`
	URI         string      `json:"uri"`
	URIHash     string      `json:"uri_hash"`
}

// GenesisState represents the bank module's genesis state
type GenesisState struct {
	Params      Params                `json:"params"`
	Balances    []Balance             `json:"balances"`
	Supply      []wasmx.Coin          `json:"supply"`
	DenomInfo   []DenomDeploymentInfo `json:"denom_info"`
	SendEnabled []SendEnabled         `json:"send_enabled"`
}

// DenomDeploymentInfoExternal is used for JSON serialization with string fields
type DenomDeploymentInfoExternal struct {
	Metadata  Metadata           `json:"metadata"`
	CodeID    string             `json:"code_id"`
	Admins    []string           `json:"admins"`
	Minters   []string           `json:"minters"`
	Contract  wasmx.Bech32String `json:"contract"`
	BaseDenom string             `json:"base_denom"`
}

// DenomDeploymentInfo represents denomination deployment information
type DenomDeploymentInfo struct {
	Metadata  Metadata           `json:"metadata"`
	CodeID    uint64             `json:"code_id"`
	Admins    []string           `json:"admins"`
	Minters   []string           `json:"minters"`
	Contract  wasmx.Bech32String `json:"contract"`
	BaseDenom string             `json:"base_denom"`
}

// MarshalJSON implements custom JSON marshaling for DenomDeploymentInfo
func (ddi DenomDeploymentInfo) MarshalJSON() ([]byte, error) {
	external := DenomDeploymentInfoExternal{
		Metadata:  ddi.Metadata,
		CodeID:    utils.U64toa(ddi.CodeID),
		Admins:    ddi.Admins,
		Minters:   ddi.Minters,
		Contract:  ddi.Contract,
		BaseDenom: ddi.BaseDenom,
	}
	return json.Marshal(external)
}

// UnmarshalJSON implements custom JSON unmarshaling for DenomDeploymentInfo
func (ddi *DenomDeploymentInfo) UnmarshalJSON(data []byte) error {
	var external DenomDeploymentInfoExternal
	if err := json.Unmarshal(data, &external); err != nil {
		return err
	}

	codeID, err := strconv.ParseUint(external.CodeID, 10, 64)
	if err != nil {
		return err
	}

	ddi.Metadata = external.Metadata
	ddi.CodeID = codeID
	ddi.Admins = external.Admins
	ddi.Minters = external.Minters
	ddi.Contract = external.Contract
	ddi.BaseDenom = external.BaseDenom
	return nil
}

// MsgRegisterDenom represents a denomination registration message
type MsgRegisterDenom struct {
	Contract wasmx.Bech32String `json:"contract"`
	Metadata Metadata           `json:"metadata"`
}

// MsgMintCoins represents a coin minting message
type MsgMintCoins struct {
	Address wasmx.Bech32String `json:"address"`
	Coins   []wasmx.Coin       `json:"coins"`
}

// CallDataInstantiate represents instantiate call data
type CallDataInstantiate struct {
	Authorities []string `json:"authorities"`
}

// Query types
type QueryBalanceRequest struct {
	Address string `json:"address"`
	Denom   string `json:"denom"`
}

type QueryBalanceResponse struct {
	Balance wasmx.Coin `json:"balance"`
}

type QueryAllBalancesRequest struct {
	Address      string            `json:"address"`
	Pagination   wasmx.PageRequest `json:"pagination"`
	ResolveDenom bool              `json:"resolve_denom"`
}

type QueryAllBalancesResponse struct {
	Balances   []wasmx.Coin       `json:"balances"`
	Pagination wasmx.PageResponse `json:"pagination"`
}

type QuerySpendableBalancesRequest struct{}

type QuerySpendableBalancesResponse struct{}

type QuerySpendableBalanceByDenomRequest struct{}

type QuerySpendableBalanceByDenomResponse struct{}

type QueryTotalSupplyRequest struct {
	Pagination wasmx.PageRequest `json:"pagination"`
}

type QueryTotalSupplyResponse struct {
	Supply     []wasmx.Coin       `json:"supply"`
	Pagination wasmx.PageResponse `json:"pagination"`
}

type QuerySupplyOfRequest struct {
	Denom string `json:"denom"`
}

type QuerySupplyOfResponse struct {
	Amount wasmx.Coin `json:"amount"`
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct{}

type QueryDenomMetadataRequest struct{}

type QueryDenomMetadataResponse struct{}

type QueryDenomMetadataByQueryStringRequest struct{}

type QueryDenomMetadataByQueryStringResponse struct{}

type QueryDenomsMetadataRequest struct{}

type QueryDenomsMetadataResponse struct{}

type QueryDenomOwnersRequest struct{}

type QueryDenomOwnersResponse struct{}

type QuerySendEnabledRequest struct{}

type QuerySendEnabledResponse struct{}

type QueryAddressByDenom struct {
	Denom string `json:"denom"`
}

type QueryAddressByDenomResponse struct {
	Address wasmx.Bech32String `json:"address"`
}

// Calldata structure
type CallData struct {
	GetParams *MsgGetParams `json:"GetParams"`

	InitGenesis *MsgInitGenesis `json:"InitGenesis"`
}

type MsgInitGenesis struct{}

type MsgGetParams struct{}
