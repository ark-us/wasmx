package python

import (
	_ "embed"
)

var (
	//go:embed forward.py
	PyForward []byte

	//go:embed simple_storage.py
	PySimpleStorage []byte

	//go:embed call.py
	PyCallSimpleStorage []byte

	//go:embed blockchain.py
	PyBlockchain []byte

	//go:embed demo1.py
	PyDemo []byte
)
