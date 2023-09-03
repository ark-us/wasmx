package types

const (
	// WasmModuleEventType is stored with any contract TX that returns non empty EventAttributes
	WasmModuleEventType = "wasmx"
	// CustomContractEventPrefix contracts can create custom events. To not mix them with other system events they got the `wasmx-` prefix.
	CustomContractEventPrefix = "wasmx"

	EventTypeStoreCode   = "store_code"
	EventTypeInstantiate = "instantiate"
	EventTypeDeploy      = "deploy"
	EventTypeExecute     = "execute"
	EventTypeExecuteEth  = "execute-eth"
	EventTypeMigrate     = "migrate"
	EventTypePinCode     = "pin_code"
	EventTypeUnpinCode   = "unpin_code"

	EventTypeRegisterRole   = "register_role"
	EventTypeDeregisterRole = "deregister_role"
)

// event attributes returned from contract execution
const (
	AttributeReservedPrefix = "_"

	AttributeKeyContractAddr       = "contract_address"
	AttributeKeyCodeID             = "code_id"
	AttributeKeyChecksum           = "code_checksum"
	AttributeKeyResultDataHex      = "result"
	AttributeKeyRequiredCapability = "required_capability"

	AttributeKeyDependency = "dependency"

	AttributeKeyRole      = "role"
	AttributeKeyRoleLabel = "role_label"
)

// wasmx
var (
	// this is prefixed with types CustomContractEventPrefix
	EventTypeWasmxLog               = "log"
	AttributeKeyEventType           = "type"
	AttributeKeyIndex               = "index"
	AttributeKeyData                = "data"
	AttributeKeyTopic               = "topic"
	AttributeKeyCallContractAddress = "contract_address_call"
)
