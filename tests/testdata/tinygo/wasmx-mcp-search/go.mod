module github.com/loredanacirstea/wasmx-mcp-search

go 1.24

toolchain go1.24.4

require (
	github.com/loredanacirstea/wasmx-env v0.0.0
	github.com/loredanacirstea/wasmx-env-core v0.0.0
	github.com/loredanacirstea/wasmx-env-postgresql v0.0.0
)

require (
	cosmossdk.io/math v1.5.3 // indirect
	github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect
)

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-utils v0.0.0 => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-env-core v0.0.0 => ../wasmx-env-core

replace github.com/loredanacirstea/wasmx-env-postgresql v0.0.0 => ../wasmx-env-postgresql

replace github.com/loredanacirstea/wasmx-utils v0.0.0 => ../wasmx-utils
