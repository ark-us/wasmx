module github.com/loredanacirstea/wasmx/tests/testdata/tinygo/wasmx-oauth2-keys

go 1.24

toolchain go1.24.4

replace github.com/loredanacirstea/wasmx-env => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-utils => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-env-core => ../wasmx-env-core

replace github.com/loredanacirstea/wasmx-env-httpserver => ../wasmx-env-httpserver

require (
	cosmossdk.io/math v1.3.0
	github.com/loredanacirstea/wasmx-env v0.0.0
	github.com/loredanacirstea/wasmx-env-core v0.0.0
	github.com/loredanacirstea/wasmx-env-httpserver v0.0.0
	golang.org/x/crypto v0.31.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/exp v0.0.0-20240404231335-c0f41cb1a7a0 // indirect
)
