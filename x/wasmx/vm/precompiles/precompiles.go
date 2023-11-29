package precompiles

import (
	_ "embed"

	"mythos/v1/x/wasmx/types"
)

var (
	//go:embed 01.ecrecover.e.wasm
	ecrecovereth []byte

	//go:embed 02.sha2-256.e.wasm
	sha2_256 []byte

	//go:embed 03.ripmd160.e.wasm
	ripmd160 []byte

	//go:embed 04.identity.e.wasm
	identity []byte

	//go:embed 05.modexp.e.wasm
	modexp []byte

	//go:embed 06.ecadd.e.wasm
	ecadd []byte

	//go:embed 07.ecmul.e.wasm
	ecmul []byte

	//go:embed 08.ecpairings.e.wasm
	ecpairings []byte

	//go:embed 09.blake2f.e.wasm
	blake2f []byte

	//go:embed 20.secp384r1.wasm
	secp384r1 []byte

	//go:embed 21.secp384r1_registry.wasm
	secp384r1_registry []byte

	//go:embed 22.secret_sharing.wasm
	secret_sharing []byte

	//go:embed 23.evm_shanghai.wasm
	interpreter_evm_shanghai []byte

	//go:embed 24.alias_eth.wasm
	alias_eth []byte

	//go:embed 26.rustpython.wasm
	rustpython []byte

	//go:embed 27.quickjs.wasm
	quickjs []byte

	//go:embed 28.finite_state_machine.wasm
	state_machine []byte

	//go:embed ff.sys_proxy.wasm
	sys_proxy []byte
)

func GetPrecompileByLabel(label string) []byte {
	var wasmbin []byte
	switch label {
	case "ecrecovereth":
		wasmbin = ecrecovereth
	case "sha2-256":
		wasmbin = sha2_256
	case "ripmd160":
		wasmbin = ripmd160
	case "identity":
		wasmbin = identity
	case "modexp":
		wasmbin = modexp
	case "ecadd":
		wasmbin = ecadd
	case "ecmul":
		wasmbin = ecmul
	case "ecpairings":
		wasmbin = ecpairings
	case "blake2f":
		wasmbin = blake2f
	case "secp384r1":
		wasmbin = secp384r1
	case "secp384r1_registry":
		wasmbin = secp384r1_registry
	case "secret_sharing":
		wasmbin = secret_sharing
	case types.INTERPRETER_EVM_SHANGHAI:
		wasmbin = interpreter_evm_shanghai
	case "alias_eth":
		wasmbin = alias_eth
	case types.INTERPRETER_PYTHON:
		wasmbin = rustpython
	case types.INTERPRETER_JS:
		wasmbin = quickjs
	case types.INTERPRETER_FSM:
		wasmbin = state_machine
	case "sys_proxy":
		wasmbin = sys_proxy
	}
	return wasmbin
}
