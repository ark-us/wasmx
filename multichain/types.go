package multichain

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

type AppwithTxConfig interface {
	TxConfig() client.TxConfig
	LegacyAmino() *codec.LegacyAmino
	AppCodec() codec.Codec
	InterfaceRegistry() codectypes.InterfaceRegistry
}
