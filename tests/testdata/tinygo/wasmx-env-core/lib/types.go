package wasmxcore

import (
	sdkmath "cosmossdk.io/math"

	"github.com/cometbft/cometbft/libs/bytes"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "wasmx-core-env"

type StartBackgroundProcessRequest struct {
	Contract string `json:"contract"`
	Args     []byte `json:"args"`
}

type StartBackgroundProcessResponse struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

type WriteToBackgroundProcessRequest struct {
	Contract string `json:"contract"` // role or address
	Data     []byte `json:"data"`
	PtrFunc  string `json:"ptrFunc"`
}

type WriteToBackgroundProcessResponse struct {
	Error string `json:"error"`
}

type ReadFromBackgroundProcessRequest struct {
	Contract wasmx.Bech32String `json:"contract"` // role or address
	PtrFunc  string             `json:"ptrFunc"`
	LenFunc  string             `json:"lenFunc"`
}

type ReadFromBackgroundProcessResponse struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

type GrpcResponse struct {
	Data  string `json:"data"` // base64
	Error string `json:"error"`
}

type StartTimeoutRequest struct {
	ID       string             `json:"id"`
	Contract wasmx.Bech32String `json:"contract"`
	Delay    int64              `json:"delay"`
	Args     []byte             `json:"args"`
}

type CancelTimeoutRequest struct {
	ID string `json:"id"`
}

type MigrateContractStateByStorageRequest struct {
	ContractAddress   wasmx.Bech32String `json:"contract_address"`
	SourceStorageType string             `json:"source_storage_type"`
	TargetStorageType string             `json:"target_storage_type"`
}

type MigrateContractStateByAddressRequest struct {
	SourceContractAddress wasmx.Bech32String `json:"source_contract_address"`
	TargetContractAddress wasmx.Bech32String `json:"target_contract_address"`
	SourceStorageType     string             `json:"source_storage_type"`
	TargetStorageType     string             `json:"target_storage_type"`
}

type GlobalStorageStoreRequest struct {
	StoreKey string `json:"store_key"`
	Key      []byte `json:"key"`
	Value    []byte `json:"value"`
}

type GlobalStorageLoadRequest struct {
	StoreKey string `json:"store_key"`
	Key      []byte `json:"key"`
}

type GlobalStorageResetRequest struct {
	StoreKey string `json:"store_key"`
}

type GlobalStorageResetResponse struct {
	Error string `json:"error"`
}

type UpdateSystemCacheRequest struct {
	RoleAddress              wasmx.Bech32String  `json:"role_address"`
	RoleRegistryId           uint64              `json:"role_registry_id"`
	RoleRegistryCodeInfo     *wasmx.CodeInfo     `json:"role_registry_code_info"`
	RoleRegistryContractInfo *wasmx.ContractInfo `json:"role_registry_contract_info"`
	CodeRegistryAddress      wasmx.Bech32String  `json:"code_registry_address"`
	CodeRegistryID           uint64              `json:"code_registry_id"`
	CodeRegistryCodeInfo     *wasmx.CodeInfo     `json:"code_registry_code_info"`
	CodeRegistryContractInfo *wasmx.ContractInfo `json:"code_registry_contract_info"`
}

type UpdateSystemCacheResponse struct {
	Error string `json:"error"`
}

type CoreCallRequest struct {
	To              wasmx.Bech32String `json:"to"`
	From            wasmx.Bech32String `json:"from"`
	Value           sdkmath.Int        `json:"value"`
	GasLimit        sdkmath.Int        `json:"gasLimit"`
	Calldata        []byte             `json:"calldata"`
	Bytecode        []byte             `json:"bytecode"`
	CodeHash        []byte             `json:"codeHash"`
	CodeFilePath    string             `json:"codeFilePath"`
	AotFilePath     string             `json:"aotFilePath"`
	CodeId          uint64             `json:"codeId"`
	SystemDeps      []string           `json:"systemDeps"`
	Pinned          bool               `json:"pinned"`
	MeteringOff     bool               `json:"metering_off"`
	IsQuery         bool               `json:"isQuery"`
	StorageType     string             `json:"storage_type"`
	StorageAddress  wasmx.Bech32String `json:"storage_address"`
	ContractAddress wasmx.Bech32String `json:"contract_address"`
}

type CoreCallResponse struct {
	Success uint8  `json:"success"`
	Data    []byte `json:"data"`
}

type ExecuteCliCommandRequest struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	WorkDir string            `json:"work_dir,omitempty"`
	Env     map[string]string `json:"env,omitempty"` // Environment variables to set for the command
}

type ExecuteCliCommandResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error"`
}

type ResultBroadcastTx struct {
	Code      uint32         `json:"code"`
	Data      bytes.HexBytes `json:"data"`
	Log       string         `json:"log"`
	Codespace string         `json:"codespace"`

	Hash bytes.HexBytes `json:"hash"`
}

type BroadcastTxAsyncResponse struct {
	Error    string             `json:"error"`
	Response *ResultBroadcastTx `json:"response"`
}

type PrepareTxRequest struct {
	FromAddress  string `json:"from_address"`
	ToAddress    string `json:"to_address"`
	Data         []byte `json:"data"`
	GasLimit     uint64 `json:"gas_limit"`
	PrivateKeyHex string `json:"private_key_hex"`
}

type PrepareTxResponse struct {
	Error string `json:"error"`
	TxBytes []byte `json:"tx_bytes"`
}
