module github.com/loredanacirstea/wasmx-oauth2-server

go 1.24

toolchain go1.24.4

replace github.com/loredanacirstea/wasmx-env => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-core => ../wasmx-env-core

replace github.com/loredanacirstea/wasmx-env-utils => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-env-httpserver => ../wasmx-env-httpserver

replace github.com/loredanacirstea/wasmx-auth => ../wasmx-auth

replace github.com/loredanacirstea/wasmx-utils => ../wasmx-utils

require github.com/loredanacirstea/wasmx-env v0.0.0

require github.com/loredanacirstea/wasmx-auth v0.0.0

require github.com/loredanacirstea/wasmx-utils v0.0.0 // indirect

require (
	github.com/loredanacirstea/wasmx-env-core v0.0.0-00010101000000-000000000000
	github.com/loredanacirstea/wasmx-env-httpserver v0.0.0
)

require (
	cosmossdk.io/math v1.5.3 // indirect
	github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect
)
