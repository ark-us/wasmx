module github.com/loredanacirstea/wasmx-bank

go 1.24

toolchain go1.24.4

require github.com/loredanacirstea/wasmx-env v0.0.0

require (
	cosmossdk.io/math v1.5.3
	github.com/loredanacirstea/wasmx-utils v0.0.0
)

require github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-utils v0.0.0 => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-utils v0.0.0 => ../wasmx-utils
