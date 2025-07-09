package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env"
)

func StoreDkimKey(opts SignOptions) error {
	optsbz, err := json.Marshal(opts)
	if err != nil {
		return err
	}
	wasmx.StorageStore([]byte(`dkim_keys`), optsbz)
	return nil
}

func LoadDkimKey() *SignOptions {
	v := &SignOptions{}
	bz := wasmx.StorageLoad([]byte(`dkim_keys`))
	err := json.Unmarshal(bz, v)
	if err != nil {
		return nil
	}
	return v
}
