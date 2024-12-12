package backend

import (
	"context"
	"encoding/json"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	config "github.com/loredanacirstea/wasmx/config"
	rpctypes "github.com/loredanacirstea/wasmx/x/wasmx/rpc/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// GasPrice returns the current gas price
func (b *Backend) GasPrice() (*hexutil.Big, error) {
	result := big.NewInt(1000000)
	return (*hexutil.Big)(result), nil
}

// DoCall performs a simulated call operation through the evmtypes. It returns the
// estimated gas used on the operation or an error if fails.
func (b *Backend) DoCall(
	args rpctypes.TransactionArgs, blockNr rpctypes.BlockNumber,
) ([]byte, error) {
	to := wasmxtypes.AccAddressFromEvm(*args.To)
	var from string
	var err error
	if args.From != nil {
		from, err = b.addressCodec.BytesToString(wasmxtypes.AccAddressFromEvm(*args.From))
		if err != nil {
			return nil, err
		}
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
		funds = sdk.NewCoins(sdk.NewCoin(config.BaseDenom, sdkmath.NewIntFromBigInt((*big.Int)(args.Value))))
	}

	tostr, err := b.addressCodec.BytesToString(to)
	if err != nil {
		return nil, err
	}

	req := wasmxtypes.QuerySmartContractCallRequest{
		Sender:    from,
		Address:   tostr,
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
	senderstr, err := b.addressCodec.BytesToString(sender)
	if err != nil {
		return common.Hash{}, err
	}
	ethereumTx.Sender = senderstr

	if err := ethereumTx.ValidateBasic(); err != nil {
		b.logger.Debug("tx failed basic validation", "error", err.Error())
		return common.Hash{}, err
	}

	// TODO denom from smart contract
	cosmosTx, err := ethereumTx.BuildTx(b.clientCtx.TxConfig.NewTxBuilder(), config.BaseDenom)
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

// SimulateTransaction send a raw Ethereum transaction.
func (b *Backend) simulateTransaction(ctx context.Context, args rpctypes.TransactionArgs) (hexutil.Uint64, error) {
	var err error
	nonce := uint64(0)
	if args.Nonce != nil {
		nonce, err = hexutil.DecodeUint64(args.Nonce.String())
		if err != nil {
			b.logger.Error("nonce decoding failed", "error", err.Error())
			return hexutil.Uint64(0), err
		}
	}
	gasLimit := wasmxtypes.DefaultWasmConfig().SmartQueryGasLimit
	if args.Gas != nil {
		gasLimit, err = hexutil.DecodeUint64(args.Gas.String())
		if err != nil {
			b.logger.Error("gasLimit decoding failed", "error", err.Error())
			return hexutil.Uint64(0), err
		}
	}
	value := big.NewInt(0)
	if args.Value != nil {
		value = args.Value.ToInt()
	}
	minGasPrice := b.cfg.GetMinGasPrices()
	gasPrice := minGasPrice.AmountOf(config.BaseDenom).BigInt()
	// gasPrice := big.NewInt(10000)
	if args.GasPrice != nil {
		gasPrice = args.GasPrice.ToInt()
	}

	tx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    nonce,
		To:       args.To,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     *args.Data,
	})

	data, err := tx.MarshalBinary()
	if err != nil {
		b.logger.Error("tx marshaling failed", "error", err.Error())
		return hexutil.Uint64(0), err
	}

	// extract sender address from signature
	ethereumTx := &wasmxtypes.MsgExecuteEth{Data: data}
	senderstr, err := b.addressCodec.BytesToString(wasmxtypes.AccAddressFromEvm(*args.From))
	if err != nil {
		return hexutil.Uint64(0), err
	}
	ethereumTx.Sender = senderstr

	// TODO denom from smart contract
	cosmosTx, err := ethereumTx.BuildTx(b.clientCtx.TxConfig.NewTxBuilder(), config.BaseDenom)
	if err != nil {
		b.logger.Error("failed to build cosmos tx", "error", err.Error())
		return hexutil.Uint64(0), err
	}

	txBytes, err := b.clientCtx.TxConfig.TxEncoder()(cosmosTx)
	if err != nil {
		b.logger.Error("failed to encode eth tx using default encoder", "error", err.Error())
		return hexutil.Uint64(0), err
	}

	simreq := txtypes.SimulateRequest{
		TxBytes: txBytes,
	}
	simResponse, err := b.queryClient.Simulate(ctx, &simreq)
	if err != nil {
		return hexutil.Uint64(0), err
	}

	gasInfo := simResponse.GasInfo
	return hexutil.Uint64(gasInfo.GasUsed + 31000), nil
}

// EstimateGas returns an estimate of gas usage for the given smart contract call.
func (b *Backend) EstimateGas(args rpctypes.TransactionArgs, blockNrOptional *rpctypes.BlockNumber) (hexutil.Uint64, error) {
	blockNr := rpctypes.EthPendingBlockNumber
	if blockNrOptional != nil {
		blockNr = *blockNrOptional
	}

	// From ContextWithHeight: if the provided height is 0,
	// it will return an empty context and the gRPC query will use
	// the latest block height for querying.
	ctx := rpctypes.ContextWithHeight(blockNr.Int64())
	if args.From == nil {
		from := common.BigToAddress(big.NewInt(0))
		args.From = &from
	}
	return b.simulateTransaction(ctx, args)
}
