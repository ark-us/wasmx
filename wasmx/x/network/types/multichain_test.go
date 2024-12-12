package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	mcodec "github.com/loredanacirstea/wasmx/v1/codec"
)

func TestMultiChainAddresses(t *testing.T) {
	sender := GetRandomAccount()

	prefix1 := "mythos"
	addressCodec := mcodec.NewAccBech32Codec(prefix1, mcodec.NewAddressPrefixedFromAcc)
	accBech32Codec := mcodec.MustUnwrapAccBech32Codec(addressCodec)

	// wrap with newpref
	newchainBech32Str := mcodec.MustBech32ifyAddressPrefixedBytes("newpref", []byte(sender.Address))
	// create mythos compatible address
	crossChainBech32Str := accBech32Codec.BytesToAccAddressPrefixed([]byte(newchainBech32Str))

	// decode mythos compatible prefixed address
	crossChainBech32, err := accBech32Codec.StringToAccAddressPrefixed(crossChainBech32Str.String())
	require.NoError(t, err)
	require.Equal(t, crossChainBech32Str.Bytes(), crossChainBech32.Bytes())

	// from bytes retrieve
	initialAddr, err := accBech32Codec.Bech32Codec.StringToAddressPrefixedUnsafe(string(crossChainBech32.Bytes()))
	require.NoError(t, err)
	require.Equal(t, "newpref", initialAddr.Prefix())
	require.Equal(t, sender.Address.Bytes(), initialAddr.Bytes())
}

func GetRandomAccount() simulation.Account {
	pk := ed25519.GenPrivKey()
	privKey := secp256k1.GenPrivKeyFromSecret(pk.GetKey().Seed())
	pubKey := privKey.PubKey()
	address := sdk.AccAddress(pubKey.Address())
	account := simulation.Account{
		PrivKey: privKey,
		PubKey:  pubKey,
		Address: address,
	}
	return account
}
