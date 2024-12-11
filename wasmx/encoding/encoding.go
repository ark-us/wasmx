package encoding

import (
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"

	mcodec "mythos/v1/codec"
)

type ChainConfig struct {
	Bech32PrefixAccAddr  string
	Bech32PrefixAccPub   string
	Bech32PrefixValAddr  string
	Bech32PrefixValPub   string
	Bech32PrefixConsAddr string
	Bech32PrefixConsPub  string
	Name                 string
	HumanCoinUnit        string
	BaseDenom            string
	DenomUnit            string
	BaseDenomUnit        uint32
	BondBaseDenom        string
	BondDenom            string
}

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func MakeEncodingConfig(cfg *ChainConfig, customSigners []signing.CustomGetSigner) EncodingConfig {
	var err error
	signingOptions := signing.Options{
		AddressCodec:          mcodec.NewAccBech32Codec(cfg.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc),
		ValidatorAddressCodec: mcodec.NewValBech32Codec(cfg.Bech32PrefixValAddr, mcodec.NewAddressPrefixedFromVal),
	}
	for _, customSigner := range customSigners {
		signingOptions.DefineCustomGetSigners(customSigner.MsgType, customSigner.Fn)
	}

	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	amino := codec.NewLegacyAmino()

	configOptions := tx.ConfigOptions{
		EnabledSignModes: tx.DefaultSignModes,
	}
	configOptions.SigningOptions = &signingOptions
	configOptions.SigningContext, err = signing.NewContext(*configOptions.SigningOptions)
	if err != nil {
		panic(err)
	}

	txCfg, err := tx.NewTxConfigWithOptions(appCodec, configOptions)
	if err != nil {
		panic(err)
	}

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         appCodec,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}
