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
	golang.org/x/crypto v0.39.0
)

require (
	cosmossdk.io/math v1.3.0 // indirect
	github.com/loredanacirstea/wasmx-utils v0.0.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/text v0.26.0 // indirect
)

replace github.com/loredanacirstea/wasmx-env-sql => ../wasmx-env-sql

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-imap v0.0.0 => ../wasmx-env-imap

replace github.com/loredanacirstea/wasmx-env-smtp v0.0.0 => ../wasmx-env-smtp

replace github.com/loredanacirstea/wasmx-utils v0.0.0 => ../wasmx-utils

replace github.com/loredanacirstea/wasmx-env-httpclient v0.0.0 => ../wasmx-env-httpclient

// replace github.com/emersion/go-msgauth => ../../../../../go-msgauth

require github.com/loredanacirstea/mailverif v0.0.0-20250702112238-a372606d8d47

// github.com/loredanacirstea/mailverif@a372606d8d4716e5cf2cc71979c378d9aacaa1da
// replace github.com/loredanacirstea/mailverif => ../../mailverif
replace github.com/loredanacirstea/mailverif => github.com/loredanacirstea/mailverif v0.0.0-20250702112238-a372606d8d47
