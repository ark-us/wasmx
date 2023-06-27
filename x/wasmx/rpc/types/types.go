package types

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/tx"

	wasmxtypes "mythos/v1/x/wasmx/types"
)

// QueryClient defines a gRPC Client used for:
//   - Transaction simulation
//   - EVM module queries
//   - Fee market module queries
type QueryClient struct {
	tx.ServiceClient
	wasmxtypes.QueryClient
}

// NewQueryClient creates a new gRPC query client
func NewQueryClient(clientCtx client.Context) *QueryClient {
	return &QueryClient{
		ServiceClient: tx.NewServiceClient(clientCtx),
		QueryClient:   wasmxtypes.NewQueryClient(clientCtx),
	}
}
