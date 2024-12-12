package types

import (
	mcodec "github.com/loredanacirstea/wasmx/codec"
)

func CrossChainAddress(bech32addr string, newprefix string) (string, error) {
	return mcodec.Bech32ifyAddressPrefixedBytes(newprefix, []byte(bech32addr))
}
