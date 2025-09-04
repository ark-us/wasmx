module github.com/loredanacirstea/wasmx-fsm

go 1.24

toolchain go1.24.4

require github.com/stretchr/testify v1.10.0

require github.com/loredanacirstea/wasmx-env v0.0.0

require github.com/loredanacirstea/wasmx-utils v0.0.0

require (
	cosmossdk.io/math v1.5.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/loredanacirstea/wasmx-env-utils v0.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require github.com/loredanacirstea/wasmx-env-core v0.0.0

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-utils v0.0.0 => ../wasmx-env-utils

replace github.com/loredanacirstea/wasmx-utils v0.0.0 => ../wasmx-utils

replace github.com/loredanacirstea/wasmx-env-core v0.0.0 => ../wasmx-env-core
