package types

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	wasmxtypes "wasmx/v1/x/wasmx/types"
)

// QueryClient defines a gRPC Client used for:
//   - Transaction simulation
//   - EVM module queries
//   - Fee market module queries
type QueryClient struct {
	tx.ServiceClient
	wasmxtypes.QueryClient
	Bank banktypes.QueryClient
}

// NewQueryClient creates a new gRPC query client
func NewQueryClient(clientCtx client.Context) *QueryClient {
	return &QueryClient{
		ServiceClient: tx.NewServiceClient(clientCtx),
		QueryClient:   wasmxtypes.NewQueryClient(clientCtx),
		Bank:          banktypes.NewQueryClient(clientCtx),
	}
}
