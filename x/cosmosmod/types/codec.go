package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(govtypes.MsgSubmitProposal{}, "/cosmos.gov.v1.MsgSubmitProposal", nil)
	cdc.RegisterConcrete(govtypes.MsgVote{}, "/cosmos.gov.v1.MsgVote", nil)
	cdc.RegisterConcrete(govtypes.MsgVoteWeighted{}, "/cosmos.gov.v1.MsgVoteWeighted", nil)
	cdc.RegisterConcrete(govtypes.MsgDeposit{}, "/cosmos.gov.v1.MsgDeposit", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&govtypes.MsgSubmitProposal{},
		&govtypes.MsgVote{},
		&govtypes.MsgVoteWeighted{},
		&govtypes.MsgDeposit{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_MsgStaking_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_MsgBank_serviceDesc)
}
