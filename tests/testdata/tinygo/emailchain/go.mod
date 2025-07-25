module github.com/loredanacirstea/emailchain

go 1.24

toolchain go1.24.4

require (
	github.com/emersion/go-message v0.18.2
	github.com/loredanacirstea/wasmx-env v0.0.0
	github.com/loredanacirstea/wasmx-env-httpclient v0.0.0
	github.com/loredanacirstea/wasmx-env-imap v0.0.0
	github.com/loredanacirstea/wasmx-env-smtp v0.0.0
	github.com/loredanacirstea/wasmx-env-sql v0.0.0
	github.com/stretchr/testify v1.10.0
	golang.org/x/crypto v0.40.0
)

require (
	cosmossdk.io/math v1.5.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/loredanacirstea/wasmx-utils v0.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/loredanacirstea/wasmx-env-sql => ../wasmx-env-sql

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-imap v0.0.0 => ../wasmx-env-imap

replace github.com/loredanacirstea/wasmx-env-smtp v0.0.0 => ../wasmx-env-smtp

replace github.com/loredanacirstea/wasmx-utils v0.0.0 => ../wasmx-utils

replace github.com/loredanacirstea/wasmx-env-httpclient v0.0.0 => ../wasmx-env-httpclient

// replace github.com/emersion/go-msgauth => ../../../../../go-msgauth

// github.com/loredanacirstea/mailverif@08d14b77d989b7965916b86b998b3d08b2bba3fb
// replace github.com/loredanacirstea/mailverif => ../../mailverif
require github.com/loredanacirstea/mailverif v0.0.0-20250724172646-08d14b77d989
