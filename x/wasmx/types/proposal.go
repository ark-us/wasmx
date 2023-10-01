package types

import (
	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	v1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// constants
const (
	ProposalTypeRegisterRole   string = "RegisterRole"
	ProposalTypeDeregisterRole string = "DeregisterRole"
)

// Implements Proposal Interface
var (
	_ v1beta1.Content = &RegisterRoleProposal{}
	_ v1beta1.Content = &DeregisterRoleProposal{}
)

func init() {
	v1beta1.RegisterProposalType(ProposalTypeRegisterRole)
	v1beta1.RegisterProposalType(ProposalTypeDeregisterRole)
	v1beta1.ModuleCdc.Amino.RegisterConcrete(&RegisterRoleProposal{}, "wasmx/RegisterRoleProposal", nil)
	v1beta1.ModuleCdc.Amino.RegisterConcrete(&DeregisterRoleProposal{}, "wasmx/DeregisterRoleProposal", nil)
}

// NewRegisterRoleProposal returns new instance of RegisterRoleProposal
func NewRegisterRoleProposal(title, description string, role string, label string, contractAddress string) v1beta1.Content {
	return &RegisterRoleProposal{
		Title:           title,
		Description:     description,
		Role:            role,
		Label:           label,
		ContractAddress: contractAddress,
	}
}

// ProposalRoute returns router key for this proposal
func (*RegisterRoleProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*RegisterRoleProposal) ProposalType() string {
	return ProposalTypeRegisterRole
}

// ValidateBasic performs a stateless check of the proposal fields
func (p *RegisterRoleProposal) ValidateBasic() error {
	if p.Role == "" {
		return sdkerr.Wrapf(sdkerrors.ErrInvalidRequest, "empty role")
	}
	if p.Label == "" {
		return sdkerr.Wrapf(sdkerrors.ErrInvalidRequest, "empty label")
	}

	if _, err := sdk.AccAddressFromBech32(p.ContractAddress); err != nil {
		return err
	}

	return v1beta1.ValidateAbstract(p)
}

// NewDeregisterRoleProposal returns new instance of DeregisterRoleProposal
func NewDeregisterRoleProposal(title, description string, contractAddress string) v1beta1.Content {
	return &DeregisterRoleProposal{
		Title:           title,
		Description:     description,
		ContractAddress: contractAddress,
	}
}

// ProposalRoute returns router key for this proposal
func (*DeregisterRoleProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*DeregisterRoleProposal) ProposalType() string {
	return ProposalTypeDeregisterRole
}

// ValidateBasic performs a stateless check of the proposal fields
func (p *DeregisterRoleProposal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(p.ContractAddress); err != nil {
		return err
	}

	return v1beta1.ValidateAbstract(p)
}
