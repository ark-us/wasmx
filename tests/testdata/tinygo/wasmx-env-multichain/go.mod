module github.com/loredanacirstea/wasmx-env-multichain

go 1.24

require cosmossdk.io/math v1.3.0 // indirect

require golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect

require github.com/loredanacirstea/wasmx-env v0.0.0

require github.com/loredanacirstea/wasmx-env-utils v0.0.0

require github.com/loredanacirstea/wasmx-env-consensus v0.0.0

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-utils v0.0.0 => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-env-consensus v0.0.0 => ../wasmx-env-consensus
