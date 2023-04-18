package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"mythos/v1/x/websrv/client/cli"
)

var (
	RegisterRouteProposalHandler   = govclient.NewProposalHandler(cli.NewRegisterRouteProposalCmd)
	DeregisterRouteProposalHandler = govclient.NewProposalHandler(cli.NewDeregisterRouteProposalCmd)
)
