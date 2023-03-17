package ewasm

import (
	_ "embed"
)

var (
	//go:embed contracts/precompiles/01.ecrecover.e.wasm
	ecrecover []byte

	//go:embed contracts/precompiles/02.sha2-256.e.wasm
	sha2_256 []byte

	//go:embed contracts/precompiles/03.ripmd160.e.wasm
	ripmd160 []byte

	//go:embed contracts/precompiles/04.identity.e.wasm
	identity []byte

	//go:embed contracts/precompiles/05.modexp.e.wasm
	modexp []byte

	//go:embed contracts/precompiles/06.ecadd.e.wasm
	ecadd []byte

	//go:embed contracts/precompiles/07.ecmul.e.wasm
	ecmul []byte

	//go:embed contracts/precompiles/08.ecpairings.e.wasm
	ecpairings []byte

	//go:embed contracts/precompiles/09.blake2f.e.wasm
	blake2f []byte

	//go:embed contracts/precompiles/20.secp384r1.wasm
	secp384r1 []byte

	//go:embed contracts/precompiles/21.secp384r1_registry.wasm
	secp384r1_registry []byte
)

func GetPrecompileByLabel(label string) []byte {
	var wasmbin []byte
	switch label {
	case "ecrecover":
		wasmbin = ecrecover
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
	}
	return wasmbin
}
