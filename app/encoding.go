package app

import (
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"

	"mythos/v1/app/params"
	mcodec "mythos/v1/codec"
	"mythos/v1/config"
)

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func MakeEncodingConfig(cfg *config.ChainConfig) params.EncodingConfig {
	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          mcodec.NewAccBech32Codec(cfg.Bech32PrefixAccAddr, mcodec.NewAddressPrefixedFromAcc),
			ValidatorAddressCodec: mcodec.NewValBech32Codec(cfg.Bech32PrefixValAddr, mcodec.NewAddressPrefixedFromVal),
		},
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	amino := codec.NewLegacyAmino()
	txCfg := tx.NewTxConfig(appCodec, tx.DefaultSignModes)

	return params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         appCodec,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}
