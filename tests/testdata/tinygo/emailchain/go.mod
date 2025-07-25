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
	golang.org/x/crypto v0.40.0
)

require (
	cosmossdk.io/math v1.5.3 // indirect
	github.com/loredanacirstea/wasmx-utils v0.0.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/text v0.27.0 // indirect
)

replace github.com/loredanacirstea/wasmx-env-sql => ../wasmx-env-sql

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-imap v0.0.0 => ../wasmx-env-imap

replace github.com/loredanacirstea/wasmx-env-smtp v0.0.0 => ../wasmx-env-smtp

replace github.com/loredanacirstea/wasmx-utils v0.0.0 => ../wasmx-utils

replace github.com/loredanacirstea/wasmx-env-httpclient v0.0.0 => ../wasmx-env-httpclient

// replace github.com/emersion/go-msgauth => ../../../../../go-msgauth

// github.com/loredanacirstea/mailverif@188e4581f4a628b77101bf8708f8bbb99821b23c
require github.com/loredanacirstea/mailverif v0.0.0-20250725161918-188e4581f4a6

// replace github.com/loredanacirstea/mailverif => ../../../../../mailverif
