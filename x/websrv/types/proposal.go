package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	v1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// constants
const (
	ProposalTypeRegisterRoute   string = "RegisterRoute"
	ProposalTypeDeregisterRoute string = "DeregisterRoute"
)

// Implements Proposal Interface
var (
	_ v1beta1.Content = &RegisterRouteProposal{}
	_ v1beta1.Content = &DeregisterRouteProposal{}
)

func init() {
	v1beta1.RegisterProposalType(ProposalTypeRegisterRoute)
	v1beta1.RegisterProposalType(ProposalTypeDeregisterRoute)
	v1beta1.ModuleCdc.Amino.RegisterConcrete(&RegisterRouteProposal{}, "wasmx/RegisterRouteProposal", nil)
	v1beta1.ModuleCdc.Amino.RegisterConcrete(&DeregisterRouteProposal{}, "wasmx/DeregisterRouteProposal", nil)
}

// NewRegisterRouteProposal returns new instance of RegisterRouteProposal
func NewRegisterRouteProposal(title, description string, path string, contractAddress string) v1beta1.Content {
	return &RegisterRouteProposal{
		Title:           title,
		Description:     description,
		Path:            path,
		ContractAddress: contractAddress,
	}
}

// ProposalRoute returns router key for this proposal
func (*RegisterRouteProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*RegisterRouteProposal) ProposalType() string {
	return ProposalTypeRegisterRoute
}

// ValidateBasic performs a stateless check of the proposal fields
func (p *RegisterRouteProposal) ValidateBasic() error {
	if p.Path == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty route path")
	}

	if string(p.Path[0]) != "/" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "path must start with /")
	}

	if len(p.Path) > 1 && string(p.Path[len(p.Path)-1]) == "/" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "path must not end with /")
	}

	if _, err := sdk.AccAddressFromBech32(p.ContractAddress); err != nil {
		return err
	}

	return v1beta1.ValidateAbstract(p)
}

// NewDeregisterRouteProposal returns new instance of DeregisterRouteProposal
func NewDeregisterRouteProposal(title, description string, path string, contractAddress string) v1beta1.Content {
	return &DeregisterRouteProposal{
		Title:           title,
		Description:     description,
		Path:            path,
		ContractAddress: contractAddress,
	}
}

// ProposalRoute returns router key for this proposal
func (*DeregisterRouteProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns proposal type for this proposal
func (*DeregisterRouteProposal) ProposalType() string {
	return ProposalTypeDeregisterRoute
}

// ValidateBasic performs a stateless check of the proposal fields
func (p *DeregisterRouteProposal) ValidateBasic() error {
	if p.Path == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty route path")
	}

	if string(p.Path[0]) != "/" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "path must start with /")
	}

	if len(p.Path) > 1 && string(p.Path[len(p.Path)-1]) == "/" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "path must not end with /")
	}

	if _, err := sdk.AccAddressFromBech32(p.ContractAddress); err != nil {
		return err
	}

	return v1beta1.ValidateAbstract(p)
}
