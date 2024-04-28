package sample

import (
	address "cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
)

// AccAddress returns a sample account address
func AccAddress(addrCodec address.Codec) string {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	addrstr, err := addrCodec.BytesToString(addr)
	if err != nil {
		panic(err)
	}
	return addrstr
}
