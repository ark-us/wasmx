package websrv

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"wasmx/v1/x/websrv/keeper"
	"wasmx/v1/x/websrv/types"
)

// NewWebsrvProposalHandler creates a governance handler to manage new proposal types.
func NewWebsrvProposalHandler(k *keeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *types.RegisterRouteProposal:
			return handleRegisterRouteProposal(ctx, k, c)
		case *types.DeregisterRouteProposal:
			return handleDeregisterRouteProposal(ctx, k, c)
		default:
			return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}

// handleRegisterRouteProposal handles the registration proposal for a webserver route
func handleRegisterRouteProposal(
	ctx sdk.Context,
	k *keeper.Keeper,
	p *types.RegisterRouteProposal,
) error {
	k.RegisterRouteHandler(ctx, p.Path, p.ContractAddress)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeRegisterRoute,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyRoute, p.Path),
		sdk.NewAttribute(types.AttributeKeyContract, p.ContractAddress),
	))

	return nil
}

// handleDeregisterRouteProposal handles the deregistration proposal for a route
func handleDeregisterRouteProposal(
	ctx sdk.Context,
	k *keeper.Keeper,
	p *types.DeregisterRouteProposal,
) error {
	k.DeregisterRouteHandler(ctx, p.Path, p.ContractAddress)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDeregisterRoute,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(types.AttributeKeyRoute, p.Path),
		sdk.NewAttribute(types.AttributeKeyContract, p.ContractAddress),
	))

	return nil
}
