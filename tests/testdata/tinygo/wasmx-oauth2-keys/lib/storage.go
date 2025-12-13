package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// LoadKeyPair loads a key pair by public key
func LoadKeyPair(publicKey string) (*EphemeralKeyPair, error) {
	key := []byte(STORAGE_KEY_PREFIX + publicKey)
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
	key := []byte(STORAGE_KEY_PREFIX + keyPair.PublicKey)
	data, err := json.Marshal(keyPair)
	if err != nil {
		return err
	}
	wasmx.StorageStore(key, data)
	return nil
}

// DeleteKeyPair deletes a key pair
func DeleteKeyPair(publicKey string) {
	key := STORAGE_KEY_PREFIX + publicKey
	wasmx.StorageDelete(key)
}

// LoadPublicKeyByToken loads public key by OAuth token
func LoadPublicKeyByToken(oauthToken string) string {
	key := []byte(STORAGE_TOKEN_PREFIX + oauthToken)
	data := wasmx.StorageLoad(key)
	return string(data)
}

// SaveTokenMapping saves OAuth token to public key mapping
func SaveTokenMapping(oauthToken string, publicKey string) {
	key := []byte(STORAGE_TOKEN_PREFIX + oauthToken)
	wasmx.StorageStore(key, []byte(publicKey))
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
