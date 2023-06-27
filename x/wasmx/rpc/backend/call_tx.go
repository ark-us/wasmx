package backend

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// GasPrice returns the current gas price based on Ethermint's gas price oracle.
func (b *Backend) GasPrice() (*hexutil.Big, error) {
	var (
		result *big.Int
		// err    error
	)
	// if head := b.CurrentHeader(); head.BaseFee != nil {
	// 	result, err = b.SuggestGasTipCap(head.BaseFee)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	result = result.Add(result, head.BaseFee)
	// } else {
	// 	result = big.NewInt(b.RPCMinGasPrice())
	// }

	// // return at least GlobalMinGasPrice from FeeMarket module
	// minGasPrice, err := b.GlobalMinGasPrice()
	// if err != nil {
	// 	return nil, err
	// }
	// minGasPriceInt := minGasPrice.TruncateInt().BigInt()
	// if result.Cmp(minGasPriceInt) < 0 {
	// 	result = minGasPriceInt
	// }

	result = big.NewInt(1000000)

	return (*hexutil.Big)(result), nil
}
