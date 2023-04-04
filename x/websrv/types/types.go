package types

import (
	"encoding/hex"
	"math/big"
)

func OauthClientIdToString(id uint64) string {
	bz := big.NewInt(int64(id)).FillBytes(make([]byte, 32))
	return hex.EncodeToString(bz)
}

func OauthClientIdFromString(idstr string) (uint64, error) {
	bz, err := hex.DecodeString(idstr)
	if err != nil {
		return 0, err
	}
	id := new(big.Int)
	id.SetBytes(bz)
	return id.Uint64(), nil
}
