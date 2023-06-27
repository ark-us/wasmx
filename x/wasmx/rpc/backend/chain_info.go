package backend

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"

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

func (b *Backend) BaseFee(blockRes *tmrpctypes.ResultBlockResults) (*big.Int, error) {
	return nil, nil
}

// // ChainConfig returns the latest ethereum chain configuration
// func (b *Backend) ChainConfig() *params.ChainConfig {
// 	params, err := b.queryClient.Params(b.ctx, &wasmxtypes.QueryParamsRequest{})
// 	if err != nil {
// 		return nil
// 	}

// 	return params.Params.ChainConfig.EthereumConfig(b.chainID)
// }
