module github.com/loredanacirstea/wasmx/tests/testdata/tinygo/wasmx-oauth2-keys

go 1.24

toolchain go1.24.4

replace github.com/loredanacirstea/wasmx-env => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-utils => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-env-core => ../wasmx-env-core

replace github.com/loredanacirstea/wasmx-env-httpserver => ../wasmx-env-httpserver

replace github.com/loredanacirstea/wasmx-auth v0.0.0 => ../wasmx-auth

replace github.com/loredanacirstea/wasmx-utils => ../wasmx-utils

require (
	cosmossdk.io/math v1.5.3
	github.com/loredanacirstea/wasmx-auth v0.0.0
	github.com/loredanacirstea/wasmx-env v0.0.0
	github.com/loredanacirstea/wasmx-env-core v0.0.0
	github.com/loredanacirstea/wasmx-env-httpserver v0.0.0
	golang.org/x/crypto v0.31.0
)

require (
	github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect
	github.com/loredanacirstea/wasmx-utils v0.0.0 // indirect
)
