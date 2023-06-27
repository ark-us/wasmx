package backend

import (
	"github.com/ethereum/go-ethereum/common/hexutil"

	wasmxtypes "mythos/v1/x/wasmx/types"
)

// ChainID is the EIP-155 replay-protection chain id for the current ethereum chain config.
func (b *Backend) ChainID() (*hexutil.Big, error) {
	eip155ChainID, err := wasmxtypes.ParseChainID(b.clientCtx.ChainID)
	if err != nil {
		panic(err)
	}
	return (*hexutil.Big)(eip155ChainID), nil
}

// // ChainConfig returns the latest ethereum chain configuration
// func (b *Backend) ChainConfig() *params.ChainConfig {
// 	params, err := b.queryClient.Params(b.ctx, &wasmxtypes.QueryParamsRequest{})
// 	if err != nil {
// 		return nil
// 	}

// 	return params.Params.ChainConfig.EthereumConfig(b.chainID)
// }
