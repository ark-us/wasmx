package types

import (
	storetypes "cosmossdk.io/store/types"
)

// GasRegister abstract source for gas costs
type GasRegister interface {
	// NewContractInstanceCosts costs to create a new contract instance from code
	NewContractInstanceCosts(pinned bool, msgLen int) storetypes.Gas
	// CompileCosts costs to persist and "compile" a new wasm contract
	CompileCosts(byteLength int) storetypes.Gas
	// UncompressCosts costs to unpack a new wasm contract
	UncompressCosts(byteLength int) storetypes.Gas
	// InstantiateContractCosts costs when interacting with a wasm contract
	InstantiateContractCosts(pinned bool, msgLen int) storetypes.Gas
	// ReplyCosts costs to to handle a message reply
	ReplyCosts(pinned bool, reply Reply) storetypes.Gas
	// EventCosts costs to persist an event
	EventCosts(attrs []EventAttribute, events Events) storetypes.Gas
	// ToWasmVMGas converts from sdk gas to wasmvm gas
	ToWasmVMGas(source storetypes.Gas) uint64
	// FromWasmVMGas converts from wasmvm gas to sdk gas
	FromWasmVMGas(source uint64) storetypes.Gas
}
