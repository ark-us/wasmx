package types

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
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
	EventCosts(attrs []EventAttribute, events Events) storetypes.Gas
	// ToWasmVMGas converts from sdk gas to wasmvm gas
	ToWasmVMGas(source storetypes.Gas) uint64
	// FromWasmVMGas converts from wasmvm gas to sdk gas
	FromWasmVMGas(source uint64) storetypes.Gas
}

// MultipliedGasMeter wraps the GasMeter from context and multiplies all reads by out defined multiplier
type MultipliedGasMeter struct {
	originalMeter storetypes.GasMeter
	GasRegister   GasRegister
}

func NewMultipliedGasMeter(originalMeter storetypes.GasMeter, gr GasRegister) MultipliedGasMeter {
	return MultipliedGasMeter{originalMeter: originalMeter, GasRegister: gr}
}

var _ GasMeter = MultipliedGasMeter{}

// gas consumed in wasm VM units
func (m MultipliedGasMeter) GasConsumed() storetypes.Gas {
	return m.GasRegister.ToWasmVMGas(m.originalMeter.GasConsumed())
}

// consume gas in wasm VM units
func (m MultipliedGasMeter) ConsumeGas(gas storetypes.Gas, descriptor string) {
	descriptor = fmt.Sprintf("wasmx: %s", descriptor)
	sdkgas := m.GasRegister.FromWasmVMGas(gas)
	m.originalMeter.ConsumeGas(sdkgas, descriptor)
}

// gas limit in wasm VM units
func (m MultipliedGasMeter) GasLimit() storetypes.Gas {
	return m.GasRegister.ToWasmVMGas(m.originalMeter.Limit())
}

func (m MultipliedGasMeter) GasRemaining() storetypes.Gas {
	return m.GasRegister.ToWasmVMGas(m.originalMeter.GasRemaining())
}
