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

func StoreKayrosBotNoStore(noStore bool) {
	var val byte
	if noStore {
		val = 1
	}
	wasmx.StorageStore([]byte(`kayrosbot_no_store`), []byte{val})
}

func LoadKayrosBotNoStore() bool {
	bz := wasmx.StorageLoad([]byte(`kayrosbot_no_store`))
	if len(bz) == 0 {
		return false
	}
	return bz[0] == 1
}

func StoreKayrosApiUserKey(userKey string) {
	wasmx.StorageStore([]byte(`kayros_api_user_key`), []byte(userKey))
}

func LoadKayrosApiUserKey() string {
	bz := wasmx.StorageLoad([]byte(`kayros_api_user_key`))
	return string(bz)
}
