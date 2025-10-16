package tinygo

import (
	_ "embed"
)

var (
	//go:embed forward.wasm
	TinyGoForward []byte

	//go:embed determ.wasm
	TinyGoDeterministic []byte

	//go:embed nondeterm.wasm
	TinyGoNonDeterministic []byte

	//go:embed simple_storage.wasm
	TinyGoSimpleStorage []byte

	//go:embed imaptest.wasm
	ImapTestWrapSdk []byte

	//go:embed smtptest.wasm
	SmtpTestWrapSdk []byte
)
