package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"wasmx/x/websrv/client/cli"
)

var (
	RegisterRouteProposalHandler   = govclient.NewProposalHandler(cli.NewRegisterRouteProposalCmd)
	DeregisterRouteProposalHandler = govclient.NewProposalHandler(cli.NewDeregisterRouteProposalCmd)
)
