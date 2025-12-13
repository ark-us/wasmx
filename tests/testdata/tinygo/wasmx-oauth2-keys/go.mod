module github.com/loredanacirstea/wasmx/tests/testdata/tinygo/wasmx-oauth2-keys

go 1.24

toolchain go1.24.4

replace github.com/loredanacirstea/wasmx-env => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-utils => ../wasmx-env-utils

require (
	github.com/loredanacirstea/wasmx-env v0.0.0
	golang.org/x/crypto v0.31.0
)

require (
	cosmossdk.io/math v1.3.0 // indirect
	github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
)
