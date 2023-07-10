package crypto

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/btcec/v2"
	btc_ecdsa "github.com/btcsuite/btcd/btcec/v2/ecdsa"

	sdksecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

// from cosmos-sdk/crypto/keys/secp256k1

// used to reject malleable signatures
// see:
//   - https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/signature_nocgo.go#L90-L93
//   - https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/crypto.go#L39
var secp256k1halfN = new(big.Int).Rsh(secp256k1.S256().N, 1)

// Read Signature struct from R || S. Caller needs to ensure
// that len(sigStr) == 64.
func SignatureFromBytes(sigStr []byte) *secp256k1.Signature {
	return &secp256k1.Signature{
		R: new(big.Int).SetBytes(sigStr[:32]),
		S: new(big.Int).SetBytes(sigStr[32:64]),
	}
}

// VerifyBytes verifies a signature of the form R || S.
// It rejects signatures which are not in lower-S form.
func VerifySignature(pubKey *sdksecp256k1.PubKey, msgHash []byte, sigStr []byte) bool {
	if len(sigStr) != 64 {
		return false
	}
	if len(msgHash) != 32 {
		return false
	}
	pub, err := secp256k1.ParsePubKey(pubKey.Key, secp256k1.S256())
	if err != nil {
		return false
	}
	// parse the signature:
	signature := SignatureFromBytes(sigStr)
	// Reject malleable signatures. libsecp256k1 does this check but btcec doesn't.
	// see: https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/signature_nocgo.go#L90-L93
	if signature.S.Cmp(secp256k1halfN) > 0 {
		return false
	}
	return signature.Verify(msgHash, pub)
}

var SignatureLength = 65
var RecoveryIDOffset = 64

// Secp256k1Recover returns the uncompressed public key that created the given signature.
func Secp256k1Recover(hash, sig []byte) ([]byte, error) {
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
		return nil, errors.New("invalid signature length")
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
