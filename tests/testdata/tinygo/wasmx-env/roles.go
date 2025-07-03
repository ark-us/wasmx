package wasmx

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
