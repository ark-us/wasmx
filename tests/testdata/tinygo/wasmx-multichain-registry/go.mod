module github.com/loredanacirstea/wasmx-multichain-registry

go 1.24

toolchain go1.24.4

require github.com/loredanacirstea/wasmx-env v0.0.0

require github.com/loredanacirstea/wasmx-consensus v0.0.0

require github.com/loredanacirstea/wasmx-utils v0.0.0

require github.com/loredanacirstea/wasmx-auth v0.0.0
require github.com/loredanacirstea/wasmx-bank v0.0.0
require github.com/loredanacirstea/wasmx-staking v0.0.0
require github.com/loredanacirstea/wasmx-slashing v0.0.0
require github.com/loredanacirstea/wasmx-distribution v0.0.0
require github.com/loredanacirstea/wasmx-gov v0.0.0

require (
	cosmossdk.io/math v1.5.3 // indirect
	github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect
)

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-consensus v0.0.0 => ../wasmx-consensus

replace github.com/loredanacirstea/wasmx-env-utils v0.0.0 => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-utils v0.0.0 => ../wasmx-utils

replace github.com/loredanacirstea/wasmx-auth v0.0.0 => ../wasmx-auth
replace github.com/loredanacirstea/wasmx-bank v0.0.0 => ../wasmx-bank
replace github.com/loredanacirstea/wasmx-staking v0.0.0 => ../wasmx-staking
replace github.com/loredanacirstea/wasmx-slashing v0.0.0 => ../wasmx-slashing
replace github.com/loredanacirstea/wasmx-distribution v0.0.0 => ../wasmx-distribution
replace github.com/loredanacirstea/wasmx-gov v0.0.0 => ../wasmx-gov
