package lib

import (
	fsm "github.com/loredanacirstea/wasmx-fsm/lib"
)

func sGet(key string) string      { return string(fsm.GetContextValueInternal(key)) }
func sSet(key string, val string) { fsm.SetContextValue(key, val) }
