module github.com/loredanacirstea/mailsrv

go 1.24

toolchain go1.24.4

require (
	github.com/loredanacirstea/wasmx-env-imap v0.0.0
	github.com/loredanacirstea/wasmx-env-smtp v0.0.0
	github.com/loredanacirstea/wasmx-env-sql v0.0.0
)

require (
	github.com/emersion/go-sasl v0.0.0-20241020182733-b788ff22d5a6 // indirect
	github.com/emersion/go-smtp v0.23.0 // indirect
	github.com/loredanacirstea/wasmx-utils v0.0.0 // indirect
)

replace github.com/loredanacirstea/wasmx-env v0.0.0 => ../wasmx-env

replace github.com/loredanacirstea/wasmx-env-imap v0.0.0 => ../wasmx-env-imap

replace github.com/loredanacirstea/wasmx-env-smtp v0.0.0 => ../wasmx-env-smtp

replace github.com/loredanacirstea/wasmx-utils v0.0.0 => ../wasmx-utils

replace github.com/loredanacirstea/wasmx-env-httpclient v0.0.0 => ../wasmx-env-httpclient

replace github.com/loredanacirstea/wasmx-env-sql v0.0.0 => ../wasmx-env-sql

replace github.com/loredanacirstea/mailverif => ../../../../../mailverif
