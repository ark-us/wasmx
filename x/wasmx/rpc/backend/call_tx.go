package backend

import (
	"context"
	"encoding/json"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"mythos/v1/app"
	rpctypes "mythos/v1/x/wasmx/rpc/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
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

	to := wasmxtypes.AccAddressFromEvm(*args.To)
	var from string
	if args.From != nil {
		from = wasmxtypes.AccAddressFromEvm(*args.From).String()
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

	funds := sdk.Coins{}
	if args.Value != nil {
		funds = sdk.NewCoins(sdk.NewCoin(app.BondDenom, sdk.NewIntFromBigInt((*big.Int)(args.Value))))
	}

	req := wasmxtypes.QuerySmartContractCallRequest{
		Sender:    from,
		Address:   to.String(),
		QueryData: bz,
		Funds:     funds,
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
	if err != nil {
		return nil, err
	}

	return data.Data, nil
}

// SendRawTransaction send a raw Ethereum transaction.
func (b *Backend) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	// RLP decode raw transaction bytes
	tx := &ethtypes.Transaction{}
	if err := tx.UnmarshalBinary(data); err != nil {
		b.logger.Error("transaction decoding failed", "error", err.Error())
		return common.Hash{}, err
	}

	// check the local node config in case unprotected txs are disabled
	if !b.UnprotectedAllowed() && !tx.Protected() {
		// Ensure only eip155 signed transactions are submitted if EIP155Required is set.
		return common.Hash{}, errors.New("only replay-protected (EIP-155) transactions allowed over RPC")
	}

	// extract sender address from signature
	ethereumTx := &wasmxtypes.MsgExecuteEth{Data: data}
	ethSigner := ethtypes.LatestSignerForChainID(b.chainID)
	if !tx.Protected() {
		ethSigner = ethtypes.HomesteadSigner{}
	}
	sender, err := ethereumTx.GetSignerFromSignature(ethSigner)
	if err != nil {
		return common.Hash{}, err
	}
	ethereumTx.Sender = sender.String()

	if err := ethereumTx.ValidateBasic(); err != nil {
		b.logger.Debug("tx failed basic validation", "error", err.Error())
		return common.Hash{}, err
	}

	// TODO denom from smart contract
	cosmosTx, err := ethereumTx.BuildTx(b.clientCtx.TxConfig.NewTxBuilder(), app.BondDenom)
	if err != nil {
		b.logger.Error("failed to build cosmos tx", "error", err.Error())
		return common.Hash{}, err
	}

	txBytes, err := b.clientCtx.TxConfig.TxEncoder()(cosmosTx)
	if err != nil {
		b.logger.Error("failed to encode eth tx using default encoder", "error", err.Error())
		return common.Hash{}, err
	}

	syncCtx := b.clientCtx.WithBroadcastMode(flags.BroadcastSync)
	rsp, err := syncCtx.BroadcastTx(txBytes)
	if rsp != nil && rsp.Code != 0 {
		err = errorsmod.ABCIError(rsp.Codespace, rsp.Code, rsp.RawLog)
	}

	// we use Cosmos hashes
	txHash := common.HexToHash(rsp.TxHash)
	if err != nil {
		b.logger.Error("failed to broadcast tx", "error", err.Error())
		return txHash, err
	}

	return txHash, nil
}

// EstimateGas returns an estimate of gas usage for the given smart contract call.
func (b *Backend) EstimateGas(args rpctypes.TransactionArgs, blockNrOptional *rpctypes.BlockNumber) (hexutil.Uint64, error) {
	return hexutil.Uint64(500000), nil

	// blockNr := rpctypes.EthPendingBlockNumber
	// if blockNrOptional != nil {
	// 	blockNr = *blockNrOptional
	// }

	// bz, err := json.Marshal(&args)
	// if err != nil {
	// 	return 0, err
	// }

	// header, err := b.TendermintBlockByNumber(blockNr)
	// if err != nil {
	// 	// the error message imitates geth behavior
	// 	return 0, errors.New("header not found")
	// }

	// req := evmtypes.EthCallRequest{
	// 	Args:            bz,
	// 	GasCap:          b.RPCGasCap(),
	// 	ProposerAddress: sdk.ConsAddress(header.Block.ProposerAddress),
	// 	ChainId:         b.chainID.Int64(),
	// }

	// // From ContextWithHeight: if the provided height is 0,
	// // it will return an empty context and the gRPC query will use
	// // the latest block height for querying.
	// res, err := b.queryClient.EstimateGas(rpctypes.ContextWithHeight(blockNr.Int64()), &req)
	// if err != nil {
	// 	return 0, err
	// }
	// return hexutil.Uint64(res.Gas), nil
}
