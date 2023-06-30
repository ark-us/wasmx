package backend

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (b *Backend) PendingTransactions() ([]*sdk.Tx, error) {
	res, err := b.clientCtx.Client.UnconfirmedTxs(b.ctx, nil)
	if err != nil {
		return nil, err
	}

	result := make([]*sdk.Tx, 0, len(res.Txs))
	for _, txBz := range res.Txs {
		tx, err := b.clientCtx.TxConfig.TxDecoder()(txBz)
		if err != nil {
			return nil, err
		}
		result = append(result, &tx)
	}

	return result, nil
}
