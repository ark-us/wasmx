package multichain

import (
	"cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "mythos/v1/codec"
	mcfg "mythos/v1/config"
	appencoding "mythos/v1/encoding"
	cosmosmodtypes "mythos/v1/x/cosmosmod/types"
	networktypes "mythos/v1/x/network/types"
)

func MultiChainCtx(ac address.Codec, clientCtx client.Context) (client.Context, mcodec.AccBech32Codec, mcodec.AccBech32Codec, error) {
	chainId := clientCtx.ChainID
	mcfg.SetGlobalChainConfig(chainId)
	config, err := mcfg.GetChainConfig(chainId)
	if err != nil {
		return clientCtx, mcodec.AccBech32Codec{}, mcodec.AccBech32Codec{}, err
	}
	addrCodec := mcodec.MustUnwrapAccBech32Codec(ac)
	customEncoding := appencoding.MakeEncodingConfig(config)

	customCdc := mcodec.NewAccBech32Codec(config.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)
	customAddrCodec := mcodec.MustUnwrapAccBech32Codec(customCdc)

	clientCtx = clientCtx.WithAccountRetriever(cosmosmodtypes.NewAccountRetriever(chainId, customAddrCodec))

	clientCtx = clientCtx.
		WithCodec(customEncoding.Marshaler).
		WithInterfaceRegistry(customEncoding.InterfaceRegistry).
		WithTxConfig(customEncoding.TxConfig).
		WithLegacyAmino(customEncoding.Amino)

	return clientCtx, addrCodec, customAddrCodec, nil
}

func MultiChainCustomCtx(clientCtx client.Context) (client.Context, mcodec.AccBech32Codec, error) {
	chainId := clientCtx.ChainID
	mcfg.SetGlobalChainConfig(chainId)
	config, err := mcfg.GetChainConfig(chainId)
	if err != nil {
		return clientCtx, mcodec.AccBech32Codec{}, err
	}
	customEncoding := appencoding.MakeEncodingConfig(config)

	customCdc := mcodec.NewAccBech32Codec(config.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc)
	customAddrCodec := mcodec.MustUnwrapAccBech32Codec(customCdc)

	clientCtx = clientCtx.WithAccountRetriever(cosmosmodtypes.NewAccountRetriever(chainId, customAddrCodec))

	clientCtx = clientCtx.
		WithCodec(customEncoding.Marshaler).
		WithInterfaceRegistry(customEncoding.InterfaceRegistry).
		WithTxConfig(customEncoding.TxConfig).
		WithLegacyAmino(customEncoding.Amino)

	return clientCtx, customAddrCodec, nil
}

func MultiChainWrap(clientCtx client.Context, msg sdk.Msg, fromAddr mcodec.AccAddressPrefixed) (*networktypes.MsgMultiChainWrap, error) {
	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}
	msgMultiChain := &networktypes.MsgMultiChainWrap{
		MultiChainId: clientCtx.ChainID,
		Sender:       fromAddr.String(),
		Data:         msgAny,
	}
	return msgMultiChain, nil
}
