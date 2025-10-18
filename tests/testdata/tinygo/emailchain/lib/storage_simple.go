package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
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

func StoreServerPassword(password string) {
	wasmx.StorageStore([]byte(`server_password`), []byte(password))
}

func LoadServerPassword() string {
	bz := wasmx.StorageLoad([]byte(`server_password`))
	return string(bz)
}
