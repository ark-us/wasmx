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

// Role name constants mirrored from AssemblyScript
const (
    ROLE_EID_REGISTRY            = "eid_registry"
    ROLE_STORAGE                 = "storage"
    ROLE_STORAGE_CONTRACTS       = "storage_contracts"
    ROLE_STAKING                 = "staking"
    ROLE_BANK                    = "bank"
    ROLE_DENOM                   = "denom"
    ROLE_HOOKS                   = "hooks"
    ROLE_HOOKS_NONC              = "hooks_nonconsensus"
    ROLE_GOVERNANCE              = "gov"
    ROLE_AUTH                    = "auth"
    ROLE_ROLES                   = "roles"
    ROLE_SLASHING                = "slashing"
    ROLE_DISTRIBUTION            = "distribution"
    ROLE_INTERPRETER             = "interpreter"
    ROLE_PRECOMPILE              = "precompile"
    ROLE_ALIAS                   = "alias"
    ROLE_CONSENSUS               = "consensus"
    ROLE_INTERPRETER_PYTHON      = "interpreter_python"
    ROLE_INTERPRETER_JS          = "interpreter_javascript"
    ROLE_INTERPRETER_FSM         = "interpreter_state_machine"
    ROLE_LIBRARY                 = "deplibrary"
    ROLE_CHAT                    = "chat"
    ROLE_TIME                    = "time"
    ROLE_LEVEL0                  = "level0"
    ROLE_LEVELN                  = "leveln"
    ROLE_MULTICHAIN_REGISTRY     = "multichain_registry"
    ROLE_MULTICHAIN_REGISTRY_LOCAL = "multichain_registry_local"
    ROLE_SECRET_SHARING          = "secret_sharing"
    ROLE_LOBBY                   = "lobby"
    ROLE_METAREGISTRY            = "metaregistry"
    ROLE_FEE_COLLECTOR           = "fee_collector"
    ROLE_DTYPE                   = "dtype"
)
