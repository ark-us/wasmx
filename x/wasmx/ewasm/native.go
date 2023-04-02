package ewasm

import (
	"bytes"
	// "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	// "github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var EMPTY_ADDRESS = bytes.Repeat([]byte{0}, 20)
var NativeMap = map[string]func([]byte) []byte{
	"0x0000000000000000000000000000000000000001": Secp256k1Recover,
}

func Secp256k1Recover(msg []byte) []byte {
	return nil
}
