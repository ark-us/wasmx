module github.com/loredanacirstea/wasmx-blocks

go 1.24

require github.com/loredanacirstea/wasmx-env v0.0.0

require (
	cosmossdk.io/math v1.3.0 // indirect
	github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
)

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-utils v0.0.0 => ../wasmx-env-utils
