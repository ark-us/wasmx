module github.com/loredanacirstea/wasmx-raft-lib

go 1.24

toolchain go1.24.4

require github.com/loredanacirstea/wasmx-env v0.0.0

require github.com/loredanacirstea/wasmx-fsm v0.0.0

require github.com/loredanacirstea/wasmx-env-core v0.0.0 // indirect
require github.com/loredanacirstea/wasmx-env-consensus v0.0.0 // indirect
require github.com/loredanacirstea/wasmx-env-crosschain v0.0.0 // indirect
require github.com/loredanacirstea/wasmx-staking v0.0.0 // indirect
require github.com/loredanacirstea/wasmx-consensus-utils v0.0.0 // indirect
require github.com/loredanacirstea/wasmx-blocks v0.0.0 // indirect

require (
	cosmossdk.io/math v1.5.3 // indirect
	github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect
	github.com/loredanacirstea/wasmx-utils v0.0.0 // indirect
)

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-utils v0.0.0 => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-utils v0.0.0 => ../wasmx-utils

replace github.com/loredanacirstea/wasmx-fsm v0.0.0 => ../wasmx-fsm

replace github.com/loredanacirstea/wasmx-env-core v0.0.0 => ../wasmx-env-core
replace github.com/loredanacirstea/wasmx-env-crosschain v0.0.0 => ../wasmx-env-crosschain
replace github.com/loredanacirstea/wasmx-env-consensus v0.0.0 => ../wasmx-env-consensus
replace github.com/loredanacirstea/wasmx-staking v0.0.0 => ../wasmx-staking
replace github.com/loredanacirstea/wasmx-consensus-utils v0.0.0 => ../wasmx-consensus-utils
replace github.com/loredanacirstea/wasmx-blocks v0.0.0 => ../wasmx-blocks
