package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GasRegister abstract source for gas costs
type GasRegister interface {
	// NewContractInstanceCosts costs to create a new contract instance from code
	NewContractInstanceCosts(pinned bool, msgLen int) sdk.Gas
	// CompileCosts costs to persist and "compile" a new wasm contract
	CompileCosts(byteLength int) sdk.Gas
	// UncompressCosts costs to unpack a new wasm contract
	UncompressCosts(byteLength int) sdk.Gas
	// InstantiateContractCosts costs when interacting with a wasm contract
	InstantiateContractCosts(pinned bool, msgLen int) sdk.Gas
	// ReplyCosts costs to to handle a message reply
	ReplyCosts(pinned bool, reply Reply) sdk.Gas
	// EventCosts costs to persist an event
	EventCosts(attrs []EventAttribute, events Events) sdk.Gas
	// ToWasmVMGas converts from sdk gas to wasmvm gas
	ToWasmVMGas(source sdk.Gas) uint64
	// FromWasmVMGas converts from wasmvm gas to sdk gas
	FromWasmVMGas(source uint64) sdk.Gas
}
