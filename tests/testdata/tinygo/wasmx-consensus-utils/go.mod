module github.com/loredanacirstea/wasmx-consensus-utils

go 1.24

require (
	github.com/loredanacirstea/wasmx-env v0.0.0
	github.com/loredanacirstea/wasmx-env-consensus v0.0.0
	github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect
	github.com/loredanacirstea/wasmx-staking v0.0.0
)

require cosmossdk.io/math v1.5.3

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-utils v0.0.0 => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-env-consensus v0.0.0 => ../wasmx-env-consensus

replace github.com/loredanacirstea/wasmx-staking v0.0.0 => ../wasmx-staking
