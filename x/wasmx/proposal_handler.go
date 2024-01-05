package wasmx

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"mythos/v1/x/wasmx/keeper"
	"mythos/v1/x/wasmx/types"
)

// NewWasmxProposalHandler creates a governance handler to manage new proposal types.
func NewWasmxProposalHandler(k *keeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *types.RegisterRoleProposal:
			return handleRegisterRoleProposal(ctx, k, c)
		case *types.DeregisterRoleProposal:
			return handleDeregisterRoleProposal(ctx, k, c)
		default:
			return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}

// handleRegisterRoleProposal handles the registration proposal for a webserver route
func handleRegisterRoleProposal(
	ctx sdk.Context,
	k *keeper.Keeper,
	p *types.RegisterRoleProposal,
) error {
	k.RegisterRoleHandler(ctx, p.Role, p.Label, p.ContractAddress)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeRegisterRole,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyContractAddr, p.ContractAddress),
		sdk.NewAttribute(types.AttributeKeyRole, p.Role),
		sdk.NewAttribute(types.AttributeKeyRoleLabel, p.Label),
	))

	return nil
}

// handleDeregisterRoleProposal handles the deregistration proposal for a route
func handleDeregisterRoleProposal(
	ctx sdk.Context,
	k *keeper.Keeper,
	p *types.DeregisterRoleProposal,
) error {
	k.DeregisterRoleHandler(ctx, p.ContractAddress)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDeregisterRole,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyContractAddr, p.ContractAddress),
	))

	return nil
}
