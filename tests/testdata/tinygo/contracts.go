package tinygo

import (
	_ "embed"
)

var (
	//go:embed forward.wasm
	TinyGoForward []byte

	//go:embed add.wasm
	TinyGoAdd []byte

	//go:embed simple_storage.wasm
	TinyGoSimpleStorage []byte

	//go:embed imaptest.wasm
	ImapTestWrapSdk []byte

	//go:embed smtptest.wasm
	SmtpTestWrapSdk []byte

	// //go:embed emailchain.wasm
	// EmailChain []byte
)
