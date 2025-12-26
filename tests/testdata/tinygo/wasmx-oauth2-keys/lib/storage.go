package lib

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// LoadKeyPair loads a key pair by public key
func LoadKeyPair(publicKey []byte) (*EphemeralKeyPair, error) {
	key := []byte(STORAGE_KEY_PREFIX + hex.EncodeToString(publicKey))
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return nil, nil
	}

	var keyPair EphemeralKeyPair
	if err := json.Unmarshal(data, &keyPair); err != nil {
		return nil, err
	}
	return &keyPair, nil
}

// SaveKeyPair saves a key pair
func SaveKeyPair(keyPair *EphemeralKeyPair) error {
	key := []byte(STORAGE_KEY_PREFIX + hex.EncodeToString(keyPair.PublicKey))
	data, err := json.Marshal(keyPair)
	if err != nil {
		return err
	}
	wasmx.StorageStore(key, data)
	return nil
}

// DeleteKeyPair deletes a key pair
func DeleteKeyPair(publicKey []byte) {
	key := STORAGE_KEY_PREFIX + hex.EncodeToString(publicKey)
	wasmx.StorageDelete(key)
}

// LoadPublicKeyByToken loads public key by OAuth token
func LoadPublicKeyByToken(oauthToken string) []byte {
	key := []byte(STORAGE_TOKEN_PREFIX + oauthToken)
	data := wasmx.StorageLoad(key)
	fmt.Println("--wasmx.oauth2keys.LoadPublicKeyByToken--", string(key), hex.EncodeToString(data))
	return data
}

// SaveTokenMapping saves OAuth token to public key mapping
func SaveTokenMapping(oauthToken string, publicKey []byte) {
	key := []byte(STORAGE_TOKEN_PREFIX + oauthToken)
	wasmx.StorageStore(key, publicKey)
	fmt.Println("--wasmx.oauth2keys.SaveTokenMapping--", string(key), hex.EncodeToString(publicKey))
}

// DeleteTokenMapping deletes OAuth token to public key mapping
func DeleteTokenMapping(oauthToken string) {
	key := STORAGE_TOKEN_PREFIX + oauthToken
	wasmx.StorageDelete(key)
}

// LoadServerSecret loads the master secret for key encryption
func LoadServerSecret() string {
	key := []byte(STORAGE_SERVER_SECRET)
	data := wasmx.StorageLoad(key)
	return string(data)
}

// SaveServerSecret saves the master secret
func SaveServerSecret(secret string) {
	key := []byte(STORAGE_SERVER_SECRET)
	wasmx.StorageStore(key, []byte(secret))
}

// LoadFunderPrivateKey loads the funder private key
func LoadFunderPrivateKey() string {
	key := []byte(STORAGE_FUNDER_PRIVKEY)
	data := wasmx.StorageLoad(key)
	return string(data)
}

// SaveFunderPrivateKey saves the funder private key
func SaveFunderPrivateKey(privKey string) {
	key := []byte(STORAGE_FUNDER_PRIVKEY)
	wasmx.StorageStore(key, []byte(privKey))
}

// LoadInitAccountAmount loads the init account amount as a CoinJSON
func LoadInitAccountAmount() *wasmx.Coin {
	key := []byte("init_account_amt")
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return &wasmx.Coin{}
	}

	var coin wasmx.Coin
	if err := json.Unmarshal(data, &coin); err != nil {
		return &wasmx.Coin{}
	}
	return &coin
}

// SaveInitAccountAmount saves the init account amount as a CoinJSON
func SaveInitAccountAmount(coin wasmx.Coin) error {
	key := []byte("init_account_amt")
	data, err := json.Marshal(coin)
	if err != nil {
		return err
	}
	wasmx.StorageStore(key, data)
	return nil
}

// LoadGasPrice loads the gas price
func LoadGasPrice() *wasmx.Coin {
	key := []byte("gas_price")
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return &wasmx.Coin{}
	}

	var coin wasmx.Coin
	if err := json.Unmarshal(data, &coin); err != nil {
		return &wasmx.Coin{}
	}
	return &coin
}

// SaveGasPrice saves the gas price
func SaveGasPrice(coin wasmx.Coin) error {
	key := []byte("gas_price")
	data, err := json.Marshal(coin)
	if err != nil {
		return err
	}
	wasmx.StorageStore(key, data)
	return nil
}

// LoadRoutePrefix loads the HTTP route prefix
func LoadRoutePrefix() string {
	key := []byte("route_prefix")
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return "/auth" // default
	}
	return string(data)
}

// SaveRoutePrefix saves the HTTP route prefix
func SaveRoutePrefix(prefix string) {
	key := []byte("route_prefix")
	wasmx.StorageStore(key, []byte(prefix))
}
