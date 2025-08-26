package wasmx

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
)

// Type aliases
type HexString string
type Base64String string
type ConsensusAddressString string
type ValidatorAddressString string

type Bech32String string

type RoleChangeRequest struct {
	Role            string                `json:"role"`
	Label           string                `json:"label"`
	ContractAddress string                `json:"contract_address"`
	ActionType      RoleChangedActionType `json:"action_type"`
}

type ContractStorageType int

const (
	CoreConsensus ContractStorageType = iota
	MetaConsensus
	SingleConsensus
	Memory
	Transient
)

const (
	StorageCoreConsensus   = "CoreConsensus"
	StorageMetaConsensus   = "MetaConsensus"
	StorageSingleConsensus = "SingleConsensus"
	StorageMemory          = "Memory"
	StorageTransient       = "Transient"
)

var ContractStorageTypeByString = map[string]ContractStorageType{
	StorageCoreConsensus:   CoreConsensus,
	StorageMetaConsensus:   MetaConsensus,
	StorageSingleConsensus: SingleConsensus,
	StorageMemory:          Memory,
	StorageTransient:       Transient,
}

var ContractStorageTypeByEnum = map[ContractStorageType]string{
	CoreConsensus:   StorageCoreConsensus,
	MetaConsensus:   StorageMetaConsensus,
	SingleConsensus: StorageSingleConsensus,
	Memory:          StorageMemory,
	Transient:       StorageTransient,
}

type RoleChangedActionType int

const (
	ActionReplace RoleChangedActionType = iota
	ActionAdd     RoleChangedActionType = 1
	ActionRemove  RoleChangedActionType = 2
	ActionNoOp    RoleChangedActionType = 3
)

const (
	RoleChangedAction_Replace = "Replace"
	RoleChangedAction_Add     = "Add"
	RoleChangedAction_Remove  = "Remove"
	RoleChangedAction_NoOp    = "NoOp"
)

type RoleChanged struct {
	Role            string                `json:"role"`
	Label           string                `json:"label"`
	ContractAddress string                `json:"contract_address"`
	ActionType      RoleChangedActionType `json:"action_type"`
	PreviousAddress string                `json:"previous_address"`
}

type Role struct {
	Role        string              `json:"role"`
	StorageType ContractStorageType `json:"storage_type"`
	Primary     int                 `json:"primary"`
	Multiple    bool                `json:"multiple"`
	Labels      []string            `json:"labels"`
	Addresses   []string            `json:"addresses"`
}

type RolesChangedHook struct {
	Role        *Role        `json:"role,omitempty"`
	RoleChanged *RoleChanged `json:"role_changed,omitempty"`
}

type RolesGenesis struct {
	Roles               []Role   `json:"roles"`
	IndividualMigration []string `json:"individual_migration"`
}

type SimpleCallRequestRaw struct {
	To       Bech32String `json:"to"`
	Value    *sdkmath.Int `json:"value"`
	GasLimit *big.Int     `json:"gasLimit"`
	Calldata []byte       `json:"calldata"`
	IsQuery  bool         `json:"isQuery"`
}

type CallRequest struct {
	To       string   `json:"to"`
	Calldata string   `json:"calldata"`
	Value    *big.Int `json:"value"`
	GasLimit int64    `json:"gasLimit"`
	IsQuery  bool     `json:"isQuery"`
}

type CallResponse struct {
	Success int    `json:"success"`
	Data    string `json:"data"`
}

type CallResult struct {
	Success int    `json:"success"`
	Data    []byte `json:"data"`
}

type LoggerLog struct {
	Msg   string   `json:"msg"`
	Parts []string `json:"parts"`
}

// Common SDK types mirrored from AS

type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"` // base64 string
	Index bool   `json:"index"`
}

type Event struct {
	Type       string           `json:"type"`
	Attributes []EventAttribute `json:"attributes"`
}

type StorageRangeReq struct {
	StartKey string `json:"start_key"`
	EndKey   string `json:"end_key"`
	Reverse  bool   `json:"reverse"`
}

type StorageDeleteRange struct {
	StartKey string `json:"start_key"`
	EndKey   string `json:"end_key"`
}

type StorageRange struct {
	StartKey string `json:"start_key"` // base64
	EndKey   string `json:"end_key"`   // base64
	Reverse  bool   `json:"reverse"`
}

type StoragePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type StoragePairs struct {
	Values []StoragePair `json:"values"`
}

type Account struct {
	Address       Bech32String `json:"address"`
	PubKey        string       `json:"pubKey"`
	AccountNumber int64        `json:"accountNumber"`
	Sequence      int64        `json:"sequence"`
}

type WasmxLog struct {
	Data   []byte
	Topics [][32]byte
}

// Coins
type Coin struct {
	Denom  string      `json:"denom"`
	Amount sdkmath.Int `json:"amount"`
}

func NewCoin(denom string, amount sdkmath.Int) Coin {
	return Coin{Denom: denom, Amount: amount}
}

type DecCoin struct {
	Denom  string       `json:"denom"`
	Amount *sdkmath.Int `json:"amount"`
}
type CreateAccountRequest struct {
	CodeID uint64 `json:"code_id"`
	Msg    string `json:"msg"`
	Funds  []Coin `json:"funds"`
	Label  string `json:"label"`
}
type CreateAccountResponse struct {
	Address Bech32String `json:"address"`
}

type Create2AccountRequest struct {
	CodeID uint64 `json:"code_id"`
	Msg    string `json:"msg"`
	Salt   string `json:"salt"`
	Funds  []Coin `json:"funds"`
	Label  string `json:"label"`
}
type Create2AccountResponse struct {
	Address Bech32String `json:"address"`
}

type BlockInfo struct {
	// block height this transaction is executed
	Height uint64 `json:"height"`
	// time in nanoseconds since unix epoch.
	Timestamp uint64       `json:"timestamp"`
	GasLimit  uint64       `json:"gasLimit"`
	Hash      []byte       `json:"hash"`
	Proposer  Bech32String `json:"proposer"`
}

type WasmxExecutionMessage struct {
	Data Base64String `json:"data"`
}

type MerkleSlices struct {
	Slices []string `json:"slices"` // base64 encoded
}

// SignMode enum
type SignMode int

const (
	SignMode_SIGN_MODE_UNSPECIFIED       SignMode = 0
	SignMode_SIGN_MODE_DIRECT            SignMode = 1
	SignMode_SIGN_MODE_TEXTUAL           SignMode = 2
	SignMode_SIGN_MODE_DIRECT_AUX        SignMode = 3
	SignMode_SIGN_MODE_LEGACY_AMINO_JSON SignMode = 127
	SignMode_SIGN_MODE_EIP_191           SignMode = 191
)

const (
	SIGN_MODE_UNSPECIFIED       = "SIGN_MODE_UNSPECIFIED"
	SIGN_MODE_DIRECT            = "SIGN_MODE_DIRECT"
	SIGN_MODE_TEXTUAL           = "SIGN_MODE_TEXTUAL"
	SIGN_MODE_DIRECT_AUX        = "SIGN_MODE_DIRECT_AUX"
	SIGN_MODE_LEGACY_AMINO_JSON = "SIGN_MODE_LEGACY_AMINO_JSON"
	SIGN_MODE_EIP_191           = "SIGN_MODE_EIP_191"
)

// PublicKey types
type Ed25519PubKey struct {
	Key Base64String `json:"key"`
}

type Secp256k1PubKey struct {
	Key Base64String `json:"key"`
}

type DefaultPubKey struct {
	Key Base64String `json:"key"`
}

type PublicKey struct {
	TypeUrl string       `json:"type_url"`
	Value   Base64String `json:"value"`
}

// ModeInfo types
type ModeInfoSingle struct {
	Mode string `json:"mode"`
}

type ModeInfoMulti struct {
	ModeInfos []ModeInfo `json:"mode_infos"`
}

type ModeInfo struct {
	Single *ModeInfoSingle `json:"single,omitempty"`
	Multi  *ModeInfoMulti  `json:"multi,omitempty"`
}

// Transaction types
type TxMessage struct {
	TypeUrl string       `json:"type_url"`
	Value   Base64String `json:"value"`
}

type TxBody struct {
	Messages                    []TxMessage `json:"messages"`
	Memo                        string      `json:"memo"`
	TimeoutHeight               uint64      `json:"timeout_height"`
	ExtensionOptions            []AnyWrap   `json:"extension_options"`
	NonCriticalExtensionOptions []AnyWrap   `json:"non_critical_extension_options"`
}

type SignerInfo struct {
	PublicKey *PublicKey `json:"public_key"`
	ModeInfo  ModeInfo   `json:"mode_info"`
	Sequence  uint64     `json:"sequence"`
}

type Fee struct {
	Amount   []Coin       `json:"amount"`
	GasLimit uint64       `json:"gas_limit"`
	Payer    Bech32String `json:"payer"`
	Granter  Bech32String `json:"granter"`
}

type Tip struct {
	Amount []Coin       `json:"amount"`
	Tipper Bech32String `json:"tipper"`
}

type AuthInfo struct {
	SignerInfos []SignerInfo `json:"signer_infos"`
	Fee         *Fee         `json:"fee"`
	Tip         *Tip         `json:"tip"`
}

type SignedTransaction struct {
	Body       TxBody         `json:"body"`
	AuthInfo   AuthInfo       `json:"auth_info"`
	Signatures []Base64String `json:"signatures"`
}

// Pagination
type PageRequest struct {
	Key        uint8  `json:"key"`
	Offset     uint64 `json:"offset"`
	Limit      uint64 `json:"limit"`
	CountTotal bool   `json:"count_total"`
	Reverse    bool   `json:"reverse"`
}

type PageResponse struct {
	Total uint64 `json:"total"`
}

// Verification
type VerifyCosmosTxResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error"`
}

// Cross-chain types
type MsgCrossChainCallRequest struct {
	From         string       `json:"from"`
	To           string       `json:"to"`
	Msg          Base64String `json:"msg"`
	Funds        []Coin       `json:"funds"`
	Dependencies []string     `json:"dependencies"`
	FromChainId  string       `json:"from_chain_id"`
	ToChainId    string       `json:"to_chain_id"`
	IsQuery      bool         `json:"is_query"`
	TimeoutMs    uint64       `json:"timeout_ms"`
}

type MsgCrossChainCallResponse struct {
	Error string       `json:"error"`
	Data  Base64String `json:"data"`
}

type MsgIsAtomicTxInExecutionRequest struct {
	SubChainId string       `json:"sub_chain_id"`
	TxHash     Base64String `json:"tx_hash"`
}

type MsgIsAtomicTxInExecutionResponse struct {
	IsInExecution bool `json:"is_in_execution"`
}

// Code and Contract types
type CodeOrigin struct {
	ChainId string       `json:"chain_id"`
	Address Bech32String `json:"address"`
}

type CodeMetadata struct {
	Name       string       `json:"name"`
	Categ      []string     `json:"categ"`
	Icon       string       `json:"icon"`
	Author     string       `json:"author"`
	Site       string       `json:"site"`
	Abi        Base64String `json:"abi"`
	JsonSchema string       `json:"json_schema"`
	Origin     *CodeOrigin  `json:"origin"`
}

type ContractStorage struct {
	Key   HexString    `json:"key"`
	Value Base64String `json:"value"`
}

type CodeInfo struct {
	CodeHash                      Base64String `json:"code_hash"`
	Creator                       Bech32String `json:"creator"`
	Deps                          []string     `json:"deps"`
	Pinned                        bool         `json:"pinned"`
	MeteringOff                   bool         `json:"metering_off"`
	Metadata                      CodeMetadata `json:"metadata"`
	InterpretedBytecodeDeployment Base64String `json:"interpreted_bytecode_deployment"`
	InterpretedBytecodeRuntime    Base64String `json:"interpreted_bytecode_runtime"`
	RuntimeHash                   Base64String `json:"runtime_hash"`
}

type SystemContract struct {
	Address       string            `json:"address"`
	Label         string            `json:"label"`
	StorageType   string            `json:"storage_type"`
	InitMessage   Base64String      `json:"init_message"`
	Pinned        bool              `json:"pinned"`
	MeteringOff   bool              `json:"metering_off"`
	Native        bool              `json:"native"`
	Role          string            `json:"role"`
	Deps          []string          `json:"deps"`
	Metadata      CodeMetadata      `json:"metadata"`
	ContractState []ContractStorage `json:"contract_state"`
}

type ContractInfo struct {
	CodeId      uint64       `json:"code_id"`
	Creator     Bech32String `json:"creator"`
	Label       string       `json:"label"`
	StorageType string       `json:"storage_type"`
	InitMessage Base64String `json:"init_message"`
	Provenance  string       `json:"provenance"`
	IbcPortId   string       `json:"ibc_port_id"`
}

type MsgSetup struct {
	PreviousAddress Bech32String `json:"previous_address"`
}
