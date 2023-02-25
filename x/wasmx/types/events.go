package types

const (
	// WasmModuleEventType is stored with any contract TX that returns non empty EventAttributes
	WasmModuleEventType = "wasmx"
	// CustomContractEventPrefix contracts can create custom events. To not mix them with other system events they got the `wasm-` prefix.
	CustomContractEventPrefix = "wasmx-"

	EventTypeStoreCode   = "store_code"
	EventTypeInstantiate = "instantiate"
	EventTypeExecute     = "execute"
	EventTypeMigrate     = "migrate"
	EventTypePinCode     = "pin_code"
	EventTypeUnpinCode   = "unpin_code"
)

// event attributes returned from contract execution
const (
	AttributeReservedPrefix = "_"

	AttributeKeyContractAddr       = "contract_address"
	AttributeKeyCodeID             = "code_id"
	AttributeKeyChecksum           = "code_checksum"
	AttributeKeyResultDataHex      = "result"
	AttributeKeyRequiredCapability = "required_capability"
)
