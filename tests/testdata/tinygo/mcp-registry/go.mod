module github.com/loredanacirstea/wasmx-mcp-registry

go 1.24

toolchain go1.24.4

require (
	cosmossdk.io/math v1.5.3
	github.com/loredanacirstea/wasmx-env v0.0.0
	github.com/loredanacirstea/wasmx-env-httpserver v0.0.0
)

require github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-utils v0.0.0 => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-utils v0.0.0 => ../wasmx-utils

replace github.com/loredanacirstea/wasmx-env-httpserver v0.0.0 => ../wasmx-env-httpserver

replace github.com/loredanacirstea/wasmx-env-oauth2server v0.0.0 => ../wasmx-env-oauth2server

replace github.com/loredanacirstea/wasmx-env-httpclient v0.0.0 => ../wasmx-env-httpclient

replace github.com/loredanacirstea/wasmx-env-postgresql v0.0.0 => ../wasmx-env-postgresql
