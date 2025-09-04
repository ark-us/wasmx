package wasmxcore

import (
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
	CodeRegistryAddress      wasmx.Bech32String  `json:"code_registry_address"`
	CodeRegistryID           uint64              `json:"code_registry_id"`
	CodeRegistryCodeInfo     *wasmx.CodeInfo     `json:"code_registry_code_info"`
	CodeRegistryContractInfo *wasmx.ContractInfo `json:"code_registry_contract_info"`
}

type UpdateSystemCacheResponse struct {
	Error string `json:"error"`
}
