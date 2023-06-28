package backend

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"

	rpctypes "mythos/v1/x/wasmx/rpc/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
	wasmxvm "mythos/v1/x/wasmx/vm"
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

// DoCall performs a simulated call operation through the evmtypes. It returns the
// estimated gas used on the operation or an error if fails.
func (b *Backend) DoCall(
	args rpctypes.TransactionArgs, blockNr rpctypes.BlockNumber,
) ([]byte, error) {
	// bz, err := json.Marshal(&args)
	// if err != nil {
	// 	return nil, err
	// }
	// header, err := b.TendermintBlockByNumber(blockNr)
	// if err != nil {
	// 	// the error message imitates geth behavior
	// 	return nil, errors.New("header not found")
	// }

	to := wasmxvm.AccAddressFromEvm(*args.To)
	var from string
	if args.From != nil {
		from = wasmxvm.AccAddressFromEvm(*args.From).String()
	} else {
		from = "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqvvnu6d"
	}

	evmrq := wasmxtypes.WasmxExecutionMessage{
		Data: wasmxtypes.RawContractMessage(*args.Data),
	}
	bz, err := json.Marshal(evmrq)
	if err != nil {
		return nil, err
	}

	req := wasmxtypes.QuerySmartContractCallRequest{
		Sender:    from,
		Address:   to.String(),
		QueryData: bz,
		// Funds: sdk.NewCoins(sdk.NewCoin(app.BondDenom, sdk.NewIntFromBigInt((*big.Int).(args.Value)))),
	}

	// From ContextWithHeight: if the provided height is 0,
	// it will return an empty context and the gRPC query will use
	// the latest block height for querying.
	ctx := rpctypes.ContextWithHeight(blockNr.Int64())
	timeout := b.RPCEVMTimeout()

	// Setup context so it may be canceled the call has completed
	// or, in case of unmetered gas, setup a context with a timeout.
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	// Make sure the context is canceled when the call has completed
	// this makes sure resources are cleaned up.
	defer cancel()

	res, err := b.queryClient.SmartContractCall(ctx, &req)
	if err != nil {
		return nil, err
	}

	var data wasmxtypes.WasmxQueryResponse
	err = json.Unmarshal(res.Data, &data)

	return data.Data, nil
}
