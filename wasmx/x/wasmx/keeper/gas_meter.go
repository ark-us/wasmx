package keeper

import (
	sdkerr "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"wasmx/v1/x/wasmx/types"
)

const (
	// DefaultGasMultiplier is how many wasm gas points = 1 Cosmos SDK gas point. This strategy is taken from CosmWasm
	//
	// CosmWasm gas strategy is documented in https://github.com/CosmWasm/cosmwasm/blob/v1.0.0-beta/docs/GAS.md.
	// Cosmos SDK reference costs can be found here: https://github.com/cosmos/cosmos-sdk/blob/v0.42.10/store/types/gas.go#L198-L209.
	//
	// The original multiplier of 100 up to CosmWasm 0.16 was based on
	//     "A write at ~3000 gas and ~200us = 10 gas per us (microsecond) cpu/io
	//     Rough timing have 88k gas at 90us, which is equal to 1k sdk gas... (one read)"
	// as well as manual Wasmer benchmarks from 2019. This was then multiplied by 150_000
	// in the 0.16 -> 1.0 upgrade (https://github.com/CosmWasm/cosmwasm/pull/1120).
	//
	// The multiplier deserves more reproducible benchmarking and a strategy that allows easy adjustments.
	// This is tracked in https://github.com/CosmWasm/wasmd/issues/566 and https://github.com/CosmWasm/wasmd/issues/631.
	// Gas adjustments are consensus breaking but may happen in any release marked as consensus breaking.
	// Do not make assumptions on how much gas an operation will consume in places that are hard to adjust,
	// such as hardcoding them in contracts.
	//
	// Please note that all gas prices returned to wasmvm should have this multiplied.
	// Benchmarks and numbers were discussed in: https://github.com/CosmWasm/wasmd/pull/634#issuecomment-938055852
	DefaultGasMultiplier uint64 = 140_000_000
	// DefaultInstanceCost is how much SDK gas we charge each time we load a WASM instance.
	// Creating a new instance is costly, and this helps put a recursion limit to contracts calling contracts.
	// Benchmarks and numbers were discussed in: https://github.com/CosmWasm/wasmd/pull/634#issuecomment-938056803
	DefaultInstanceCost uint64 = 60_000
	// DefaultCompileCost is how much SDK gas is charged *per byte* for compiling WASM code.
	// Benchmarks and numbers were discussed in: https://github.com/CosmWasm/wasmd/pull/634#issuecomment-938056803
	DefaultCompileCost uint64 = 3
	// DefaultEventAttributeDataCost is how much SDK gas is charged *per byte* for attribute data in events.
	// This is used with len(key) + len(value)
	DefaultEventAttributeDataCost uint64 = 1
	// DefaultContractMessageDataCost is how much SDK gas is charged *per byte* of the message that goes to the contract
	// This is used with len(msg). Note that the message is deserialized in the receiving contract and this is charged
	// with wasm gas already. The derserialization of results is also charged in wasmvm. I am unsure if we need to add
	// additional costs here.
	// Note: also used for error fields on reply, and data on reply. Maybe these should be pulled out to a different (non-zero) field
	DefaultContractMessageDataCost uint64 = 0
	// DefaultPerAttributeCost is how much SDK gas we charge per attribute count.
	DefaultPerAttributeCost uint64 = 10
	// DefaultPerCustomEventCost is how much SDK gas we charge per event count.
	DefaultPerCustomEventCost uint64 = 20
	// DefaultEventAttributeDataFreeTier number of bytes of total attribute data we do not charge.
	DefaultEventAttributeDataFreeTier = 100
	// CW has 0.15 gas instead, but we divide by 6
	DefaultPerByteUncompressCost = 1
	UncompressCostDivisor        = 6
)

// GasRegister abstract source for gas costs
type GasRegister interface {
	// NewContractInstanceCosts costs to crate a new contract instance from code
	NewContractInstanceCosts(pinned bool, msgLen int) storetypes.Gas
	// CompileCosts costs to persist and "compile" a new wasm contract
	CompileCosts(byteLength int) storetypes.Gas
	// UncompressCosts costs to unpack a new wasm contract
	UncompressCosts(byteLength int) storetypes.Gas
	// InstantiateContractCosts costs when interacting with a wasm contract
	InstantiateContractCosts(pinned bool, msgLen int) storetypes.Gas
	// EventCosts costs to persist an event
	EventCosts(attrs []types.EventAttribute, events types.Events) storetypes.Gas
	// ToWasmVMGas converts from sdk gas to wasmvm gas
	ToWasmVMGas(source storetypes.Gas) uint64
	// FromWasmVMGas converts from wasmvm gas to sdk gas
	FromWasmVMGas(source uint64) storetypes.Gas
}

// WasmGasRegisterConfig config type
type WasmGasRegisterConfig struct {
	// InstanceCost costs when interacting with a wasm contract
	InstanceCost storetypes.Gas
	// CompileCosts costs to persist and "compile" a new wasm contract
	CompileCost storetypes.Gas
	// UncompressCost costs per byte to unpack a contract
	UncompressCost storetypes.Gas
	// GasMultiplier is how many wasm gas points = 1 sdk gas point
	GasMultiplier storetypes.Gas
	// EventPerAttributeCost is how much SDK gas is charged *per byte* for attribute data in events.
	// This is used with len(key) + len(value)
	EventPerAttributeCost storetypes.Gas
	// EventAttributeDataCost is how much SDK gas is charged *per byte* for attribute data in events.
	// This is used with len(key) + len(value)
	EventAttributeDataCost storetypes.Gas
	// EventAttributeDataFreeTier number of bytes of total attribute data that is free of charge
	EventAttributeDataFreeTier uint64
	// ContractMessageDataCost SDK gas charged *per byte* of the message that goes to the contract
	// This is used with len(msg)
	ContractMessageDataCost storetypes.Gas
	// CustomEventCost cost per custom event
	CustomEventCost uint64
}

// DefaultGasRegisterConfig default values
func DefaultGasRegisterConfig() WasmGasRegisterConfig {
	return WasmGasRegisterConfig{
		InstanceCost:               DefaultInstanceCost,
		CompileCost:                DefaultCompileCost,
		GasMultiplier:              DefaultGasMultiplier,
		EventPerAttributeCost:      DefaultPerAttributeCost,
		CustomEventCost:            DefaultPerCustomEventCost,
		EventAttributeDataCost:     DefaultEventAttributeDataCost,
		EventAttributeDataFreeTier: DefaultEventAttributeDataFreeTier,
		ContractMessageDataCost:    DefaultContractMessageDataCost,
		UncompressCost:             DefaultPerByteUncompressCost,
	}
}

// WasmxGasRegister implements GasRegister interface
type WasmxGasRegister struct {
	c WasmGasRegisterConfig
}

// NewDefaultWasmGasRegister creates instance with default values
func NewDefaultWasmGasRegister() WasmxGasRegister {
	return NewWasmGasRegister(DefaultGasRegisterConfig())
}

// NewWasmGasRegister constructor
func NewWasmGasRegister(c WasmGasRegisterConfig) WasmxGasRegister {
	if c.GasMultiplier == 0 {
		panic(sdkerr.Wrap(sdkerrors.ErrLogic, "GasMultiplier can not be 0"))
	}
	return WasmxGasRegister{
		c: c,
	}
}

// NewContractInstanceCosts costs to crate a new contract instance from code
func (g WasmxGasRegister) NewContractInstanceCosts(pinned bool, msgLen int) storetypes.Gas {
	return g.InstantiateContractCosts(pinned, msgLen)
}

// CompileCosts costs to persist and "compile" a new wasm contract
func (g WasmxGasRegister) CompileCosts(byteLength int) storetypes.Gas {
	if byteLength < 0 {
		panic(sdkerr.Wrap(types.ErrInvalid, "negative length"))
	}
	return g.c.CompileCost * uint64(byteLength)
}

// UncompressCosts costs to unpack a new wasm contract
func (g WasmxGasRegister) UncompressCosts(byteLength int) storetypes.Gas {
	if byteLength < 0 {
		panic(sdkerr.Wrap(types.ErrInvalid, "negative length"))
	}
	return g.c.UncompressCost * (uint64(byteLength) / uint64(UncompressCostDivisor))
}

// InstantiateContractCosts costs when interacting with a wasm contract
func (g WasmxGasRegister) InstantiateContractCosts(pinned bool, msgLen int) storetypes.Gas {
	if msgLen < 0 {
		panic(sdkerr.Wrap(types.ErrInvalid, "negative length"))
	}
	dataCosts := storetypes.Gas(msgLen) * g.c.ContractMessageDataCost
	if pinned {
		return dataCosts
	}
	return g.c.InstanceCost + dataCosts
}

// EventCosts costs to persist an event
func (g WasmxGasRegister) EventCosts(attrs []types.EventAttribute, events types.Events) storetypes.Gas {
	gas, remainingFreeTier := g.eventAttributeCosts(attrs, g.c.EventAttributeDataFreeTier)
	for _, e := range events {
		gas += g.c.CustomEventCost
		gas += storetypes.Gas(len(e.Type)) * g.c.EventAttributeDataCost // no free tier with event type
		var attrCost storetypes.Gas
		attrCost, remainingFreeTier = g.eventAttributeCosts(e.Attributes, remainingFreeTier)
		gas += attrCost
	}
	return gas
}

func (g WasmxGasRegister) eventAttributeCosts(attrs []types.EventAttribute, freeTier uint64) (storetypes.Gas, uint64) {
	if len(attrs) == 0 {
		return 0, freeTier
	}
	var storedBytes uint64
	for _, l := range attrs {
		storedBytes += uint64(len(l.Key)) + uint64(len(l.Value))
	}
	storedBytes, freeTier = calcWithFreeTier(storedBytes, freeTier)
	// total Length * costs + attribute count * costs
	r := sdkmath.NewIntFromUint64(g.c.EventAttributeDataCost).Mul(sdkmath.NewIntFromUint64(storedBytes)).
		Add(sdkmath.NewIntFromUint64(g.c.EventPerAttributeCost).Mul(sdkmath.NewIntFromUint64(uint64(len(attrs)))))
	if !r.IsUint64() {
		panic(storetypes.ErrorOutOfGas{Descriptor: "overflow"})
	}
	return r.Uint64(), freeTier
}

// apply free tier
func calcWithFreeTier(storedBytes uint64, freeTier uint64) (uint64, uint64) {
	if storedBytes <= freeTier {
		return 0, freeTier - storedBytes
	}
	storedBytes -= freeTier
	return storedBytes, 0
}

// ToWasmVMGas convert to wasmVM contract runtime gas unit
func (g WasmxGasRegister) ToWasmVMGas(source storetypes.Gas) uint64 {
	x := source * g.c.GasMultiplier
	if x < source {
		panic(storetypes.ErrorOutOfGas{Descriptor: "overflow"})
	}
	return x
}

// FromWasmVMGas converts to SDK gas unit
func (g WasmxGasRegister) FromWasmVMGas(source uint64) storetypes.Gas {
	return source / g.c.GasMultiplier
}
