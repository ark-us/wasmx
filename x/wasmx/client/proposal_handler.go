package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"mythos/v1/x/wasmx/client/cli"
)

var (
	RegisterRoleProposalHandler   = govclient.NewProposalHandler(cli.NewRegisterRoleProposalCmd)
	DeregisterRoleProposalHandler = govclient.NewProposalHandler(cli.NewDeregisterRoleProposalCmd)
)
