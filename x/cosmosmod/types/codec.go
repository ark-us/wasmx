package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// gov
	cdc.RegisterConcrete(govtypes.MsgSubmitProposal{}, "/cosmos.gov.v1.MsgSubmitProposal", nil)
	cdc.RegisterConcrete(govtypes.MsgVote{}, "/cosmos.gov.v1.MsgVote", nil)
	cdc.RegisterConcrete(govtypes.MsgVoteWeighted{}, "/cosmos.gov.v1.MsgVoteWeighted", nil)
	cdc.RegisterConcrete(govtypes.MsgDeposit{}, "/cosmos.gov.v1.MsgDeposit", nil)

	// auth
	cdc.RegisterInterface((*sdk.ModuleAccountI)(nil), nil)
	cdc.RegisterInterface((*authtypes.GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*sdk.AccountI)(nil), nil)
	cdc.RegisterConcrete(&authtypes.BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	cdc.RegisterConcrete(&authtypes.ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
	cdc.RegisterConcrete(authtypes.Params{}, "cosmos-sdk/x/auth/Params", nil)
	cdc.RegisterConcrete(&authtypes.ModuleCredential{}, "cosmos-sdk/GroupAccountCredential", nil)

	legacy.RegisterAminoMsg(cdc, &authtypes.MsgUpdateParams{}, "cosmos-sdk/x/auth/MsgUpdateParams")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// gov
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&govtypes.MsgSubmitProposal{},
		&govtypes.MsgVote{},
		&govtypes.MsgVoteWeighted{},
		&govtypes.MsgDeposit{},
	)

	// auth
	registry.RegisterInterface(
		"cosmos.auth.v1beta1.AccountI",
		(*authtypes.AccountI)(nil),
		&authtypes.BaseAccount{},
		&authtypes.ModuleAccount{},
	)

	registry.RegisterInterface(
		"cosmos.auth.v1beta1.AccountI",
		(*sdk.AccountI)(nil),
		&authtypes.BaseAccount{},
		&authtypes.ModuleAccount{},
	)

	registry.RegisterInterface(
		"cosmos.auth.v1beta1.GenesisAccount",
		(*authtypes.GenesisAccount)(nil),
		&authtypes.BaseAccount{},
		&authtypes.ModuleAccount{},
	)

	registry.RegisterInterface(
		"cosmos.auth.v1.ModuleCredential",
		(*cryptotypes.PubKey)(nil),
		&authtypes.ModuleCredential{},
	)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&authtypes.MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_MsgStaking_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_MsgBank_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_MsgGov_serviceDesc)
	msgservice.RegisterMsgServiceDesc(registry, &_MsgAuth_serviceDesc)
}
