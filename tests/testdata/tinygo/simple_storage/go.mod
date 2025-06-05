module simple_storage

go 1.23.0

toolchain go1.23.2

require github.com/loredanacirstea/wasmx-env v0.0.0

require (
	cosmossdk.io/math v1.3.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
)

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env
