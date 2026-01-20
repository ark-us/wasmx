module github.com/loredanacirstea/wasmx-kayros-verifier

go 1.24

toolchain go1.24.4

require github.com/loredanacirstea/wasmx-env v0.0.0

require github.com/loredanacirstea/wasmx-env-httpclient v0.0.0

require github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect

require cosmossdk.io/math v1.5.3 // indirect

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-httpclient v0.0.0 => ../wasmx-env-httpclient

replace github.com/loredanacirstea/wasmx-fsm v0.0.0 => ../wasmx-fsm

replace github.com/loredanacirstea/wasmx-env-utils v0.0.0 => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-utils v0.0.0 => ../wasmx-utils
