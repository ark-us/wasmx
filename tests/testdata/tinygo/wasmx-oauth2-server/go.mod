module github.com/loredanacirstea/wasmx-oauth2-server

go 1.24

toolchain go1.24.4

replace github.com/loredanacirstea/wasmx-env => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-core => ../wasmx-env-core

replace github.com/loredanacirstea/wasmx-env-utils => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-env-httpserver => ../wasmx-env-httpserver

require github.com/loredanacirstea/wasmx-env v0.0.0

require (
	github.com/loredanacirstea/wasmx-env-core v0.0.0-00010101000000-000000000000
	github.com/loredanacirstea/wasmx-env-httpserver v0.0.0
)

require (
	cosmossdk.io/math v1.3.0 // indirect
	github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
)
