package ewasm

import (
	"bytes"

	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	btc_ecdsa "github.com/btcsuite/btcd/btcec/v2/ecdsa"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

var EMPTY_ADDRESS = bytes.Repeat([]byte{0}, 20)
var NativeMap = map[string]func([]byte) []byte{
	"0x0000000000000000000000000000000000000001": Secp256k1Recover,
}

func Secp256k1Recover(msg []byte) []byte {
	msgHash := msg[0:32]
	signature := msg[32:]
	sig := make([]byte, SignatureLength)
	copy(sig, signature[:])

	pubKeyBz, err := Ecrecover(msgHash, sig)
	if err != nil {
		fmt.Println("Secp256k1Recover err", err)
		return EMPTY_ADDRESS
	}

	pubKey := secp256k1.PubKey{Key: pubKeyBz}
	return pubKey.Address()
}

var SignatureLength = 65
var RecoveryIDOffset = 64

// Ecrecover returns the uncompressed public key that created the given signature.
func Ecrecover(hash, sig []byte) ([]byte, error) {
	pub, err := sigToPub(hash, sig)
	if err != nil {
		return nil, err
	}
	// bytes := pub.SerializeUncompressed()
	bytes := pub.SerializeCompressed()
	return bytes, err
}

func sigToPub(hash, sig []byte) (*btcec.PublicKey, error) {
	if len(sig) != SignatureLength {
		return nil, errors.New("invalid signature")
	}
	// Convert to btcec input format with 'recovery id' v at the beginning.
	btcsig := make([]byte, SignatureLength)
	btcsig[0] = sig[RecoveryIDOffset] + 27
	copy(btcsig[1:], sig)

	pub, _, err := btc_ecdsa.RecoverCompact(btcsig, hash)
	return pub, err
}

// SigToPub returns the public key that created the given signature.
func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	pub, err := sigToPub(hash, sig)
	if err != nil {
		return nil, err
	}
	return pub.ToECDSA(), nil
}

// Sign calculates an ECDSA signature.
//
// This function is susceptible to chosen plaintext attacks that can leak
// information about the private key that is used for signing. Callers must
// be aware that the given hash cannot be chosen by an adversary. Common
// solution is to hash any input before calculating the signature.
//
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(hash []byte, prv *ecdsa.PrivateKey) ([]byte, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}
	if prv.Curve != btcec.S256() {
		return nil, fmt.Errorf("private key curve is not secp256k1")
	}
	// ecdsa.PrivateKey -> btcec.PrivateKey
	var priv btcec.PrivateKey
	if overflow := priv.Key.SetByteSlice(prv.D.Bytes()); overflow || priv.Key.IsZero() {
		return nil, fmt.Errorf("invalid private key")
	}
	defer priv.Zero()
	sig, err := btc_ecdsa.SignCompact(&priv, hash, false) // ref uncompressed pubkey
	if err != nil {
		return nil, err
	}

	// Convert to Ethereum signature format with 'recovery id' v at the end.
	v := sig[0] - 27
	copy(sig, sig[1:])
	sig[RecoveryIDOffset] = v

	return sig, nil
}
