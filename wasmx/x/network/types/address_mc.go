package types

import (
	mcodec "wasmx/v1/codec"
)

func CrossChainAddress(bech32addr string, newprefix string) (string, error) {
	return mcodec.Bech32ifyAddressPrefixedBytes(newprefix, []byte(bech32addr))
}
