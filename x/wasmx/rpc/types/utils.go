package types

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/evmos/ethermint/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	wasmxtypes "mythos/v1/x/wasmx/types"
)

// ExceedBlockGasLimitError defines the error message when tx execution exceeds the block gas limit.
// The tx fee is deducted in ante handler, so it shouldn't be ignored in JSON-RPC API.
const ExceedBlockGasLimitError = "out of gas in location: block gas meter; gasWanted:"

// // RawTxToEthTx returns a evm MsgEthereum transaction from raw tx bytes.
// func RawTxToEthTx(clientCtx client.Context, txBz tmtypes.Tx) ([]*evmtypes.MsgEthereumTx, error) {
// 	tx, err := clientCtx.TxConfig.TxDecoder()(txBz)
// 	if err != nil {
// 		return nil, errorsmod.Wrap(errortypes.ErrJSONUnmarshal, err.Error())
// 	}

// 	ethTxs := make([]*evmtypes.MsgEthereumTx, len(tx.GetMsgs()))
// 	for i, msg := range tx.GetMsgs() {
// 		ethTx, ok := msg.(*evmtypes.MsgEthereumTx)
// 		if !ok {
// 			return nil, fmt.Errorf("invalid message type %T, expected %T", msg, &evmtypes.MsgEthereumTx{})
// 		}
// 		ethTx.Hash = ethTx.AsTransaction().Hash().Hex()
// 		ethTxs[i] = ethTx
// 	}
// 	return ethTxs, nil
// }

// EthHeaderFromTendermint is an util function that returns an Ethereum Header
// from a tendermint Header.
func EthHeaderFromTendermint(header tmtypes.Header, bloom ethtypes.Bloom, baseFee *big.Int) *ethtypes.Header {
	txHash := ethtypes.EmptyRootHash
	if len(header.DataHash) == 0 {
		txHash = common.BytesToHash(header.DataHash)
	}

	time := uint64(header.Time.UTC().Unix()) // #nosec G701
	return &ethtypes.Header{
		ParentHash:  common.BytesToHash(header.LastBlockID.Hash.Bytes()),
		UncleHash:   ethtypes.EmptyUncleHash,
		Coinbase:    common.BytesToAddress(header.ProposerAddress),
		Root:        common.BytesToHash(header.AppHash),
		TxHash:      txHash,
		ReceiptHash: ethtypes.EmptyRootHash,
		Bloom:       bloom,
		Difficulty:  big.NewInt(0),
		Number:      big.NewInt(header.Height),
		GasLimit:    0,
		GasUsed:     0,
		Time:        time,
		Extra:       []byte{},
		MixDigest:   common.Hash{},
		Nonce:       ethtypes.BlockNonce{},
		BaseFee:     baseFee,
	}
}

// BlockMaxGasFromConsensusParams returns the gas limit for the current block from the chain consensus params.
func BlockMaxGasFromConsensusParams(goCtx context.Context, clientCtx client.Context, blockHeight int64) (int64, error) {
	resConsParams, err := clientCtx.Client.ConsensusParams(goCtx, &blockHeight)
	defaultGasLimit := int64(^uint32(0)) // #nosec G701
	if err != nil {
		return defaultGasLimit, err
	}

	gasLimit := resConsParams.ConsensusParams.Block.MaxGas
	if gasLimit == -1 {
		// Sets gas limit to max uint32 to not error with javascript dev tooling
		// This -1 value indicating no block gas limit is set to max uint64 with geth hexutils
		// which errors certain javascript dev tooling which only supports up to 53 bits
		gasLimit = defaultGasLimit
	}

	return gasLimit, nil
}

// FormatBlock creates an ethereum block from a tendermint header and ethereum-formatted
// transactions.
func FormatBlock(
	header tmtypes.Header, size int, gasLimit int64,
	gasUsed *big.Int, transactions []interface{}, bloom ethtypes.Bloom,
	validatorAddr common.Address, baseFee *big.Int,
) map[string]interface{} {
	var transactionsRoot common.Hash
	if len(transactions) == 0 {
		transactionsRoot = ethtypes.EmptyRootHash
	} else {
		transactionsRoot = common.BytesToHash(header.DataHash)
	}

	result := map[string]interface{}{
		"number":           hexutil.Uint64(header.Height),
		"hash":             hexutil.Bytes(header.Hash()),
		"parentHash":       common.BytesToHash(header.LastBlockID.Hash.Bytes()),
		"nonce":            ethtypes.BlockNonce{},   // PoW specific
		"sha3Uncles":       ethtypes.EmptyUncleHash, // No uncles in Tendermint
		"logsBloom":        bloom,
		"stateRoot":        hexutil.Bytes(header.AppHash),
		"miner":            validatorAddr,
		"mixHash":          common.Hash{},
		"difficulty":       (*hexutil.Big)(big.NewInt(0)),
		"extraData":        "0x",
		"size":             hexutil.Uint64(size),
		"gasLimit":         hexutil.Uint64(gasLimit), // Static gas limit
		"gasUsed":          (*hexutil.Big)(gasUsed),
		"timestamp":        hexutil.Uint64(header.Time.Unix()),
		"transactionsRoot": transactionsRoot,
		"receiptsRoot":     ethtypes.EmptyRootHash,

		"uncles":          []common.Hash{},
		"transactions":    transactions,
		"totalDifficulty": (*hexutil.Big)(big.NewInt(0)),
	}

	if baseFee != nil {
		result["baseFeePerGas"] = (*hexutil.Big)(baseFee)
	}

	return result
}

// NewTransactionFromMsg returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func NewTransactionFromMsg(
	txHash common.Hash,
	msg *wasmxtypes.MsgExecuteEth,
	blockHash common.Hash,
	blockNumber, index uint64,
	baseFee *big.Int,
	chainID *big.Int,
) (*RPCTransaction, error) {
	tx := msg.AsTransaction()
	return NewRPCTransaction(txHash, tx, blockHash, blockNumber, index, baseFee, chainID)
}

// // NewTransactionFromData returns a transaction that will serialize to the RPC
// // representation, with the given location metadata set (if available).
// func NewRPCTransaction(
// 	tx *ethtypes.Transaction, blockHash common.Hash, blockNumber, index uint64, baseFee *big.Int,
// 	chainID *big.Int,
// ) (*RPCTransaction, error) {
// 	// Determine the signer. For replay-protected transactions, use the most permissive
// 	// signer, because we assume that signers are backwards-compatible with old
// 	// transactions. For non-protected transactions, the homestead signer signer is used
// 	// because the return value of ChainId is zero for those transactions.
// 	var signer ethtypes.Signer
// 	if tx.Protected() {
// 		signer = ethtypes.LatestSignerForChainID(tx.ChainId())
// 	} else {
// 		signer = ethtypes.HomesteadSigner{}
// 	}
// 	from, _ := ethtypes.Sender(signer, tx) // #nosec G703
// 	v, r, s := tx.RawSignatureValues()
// 	result := &RPCTransaction{
// 		Type:     hexutil.Uint64(tx.Type()),
// 		From:     from,
// 		Gas:      hexutil.Uint64(tx.Gas()),
// 		GasPrice: (*hexutil.Big)(tx.GasPrice()),
// 		Hash:     tx.Hash(),
// 		Input:    hexutil.Bytes(tx.Data()),
// 		Nonce:    hexutil.Uint64(tx.Nonce()),
// 		To:       tx.To(),
// 		Value:    (*hexutil.Big)(tx.Value()),
// 		V:        (*hexutil.Big)(v),
// 		R:        (*hexutil.Big)(r),
// 		S:        (*hexutil.Big)(s),
// 		ChainID:  (*hexutil.Big)(chainID),
// 	}
// 	if blockHash != (common.Hash{}) {
// 		result.BlockHash = &blockHash
// 		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
// 		result.TransactionIndex = (*hexutil.Uint64)(&index)
// 	}
// 	switch tx.Type() {
// 	case ethtypes.AccessListTxType:
// 		al := tx.AccessList()
// 		result.Accesses = &al
// 		result.ChainID = (*hexutil.Big)(tx.ChainId())
// 	case ethtypes.DynamicFeeTxType:
// 		al := tx.AccessList()
// 		result.Accesses = &al
// 		result.ChainID = (*hexutil.Big)(tx.ChainId())
// 		result.GasFeeCap = (*hexutil.Big)(tx.GasFeeCap())
// 		result.GasTipCap = (*hexutil.Big)(tx.GasTipCap())
// 		// if the transaction has been mined, compute the effective gas price
// 		if baseFee != nil && blockHash != (common.Hash{}) {
// 			// price = min(tip, gasFeeCap - baseFee) + baseFee
// 			price := math.BigMin(new(big.Int).Add(tx.GasTipCap(), baseFee), tx.GasFeeCap())
// 			result.GasPrice = (*hexutil.Big)(price)
// 		} else {
// 			result.GasPrice = (*hexutil.Big)(tx.GasFeeCap())
// 		}
// 	}
// 	return result, nil
// }

// // BaseFeeFromEvents parses the feemarket basefee from cosmos events
// func BaseFeeFromEvents(events []abci.Event) *big.Int {
// 	for _, event := range events {
// 		if event.Type != feemarkettypes.EventTypeFeeMarket {
// 			continue
// 		}

// 		for _, attr := range event.Attributes {
// 			if bytes.Equal(attr.Key, []byte(feemarkettypes.AttributeKeyBaseFee)) {
// 				result, success := new(big.Int).SetString(string(attr.Value), 10)
// 				if success {
// 					return result
// 				}

// 				return nil
// 			}
// 		}
// 	}
// 	return nil
// }

// CheckTxFee is an internal function used to check whether the fee of
// the given transaction is _reasonable_(under the cap).
func CheckTxFee(gasPrice *big.Int, gas uint64, cap float64) error {
	// Short circuit if there is no cap for transaction fee at all.
	if cap == 0 {
		return nil
	}
	totalfee := new(big.Float).SetInt(new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gas)))
	// 1 token in 10^18 atoken
	oneToken := new(big.Float).SetInt(big.NewInt(params.Ether))
	// quo = rounded(x/y)
	feeEth := new(big.Float).Quo(totalfee, oneToken)
	// no need to check error from parsing
	feeFloat, _ := feeEth.Float64()
	if feeFloat > cap {
		return fmt.Errorf("tx fee (%.2f ether) exceeds the configured cap (%.2f ether)", feeFloat, cap)
	}
	return nil
}

// TxExceedBlockGasLimit returns true if the tx exceeds block gas limit.
func TxExceedBlockGasLimit(res *abci.ResponseDeliverTx) bool {
	return strings.Contains(res.Log, ExceedBlockGasLimitError)
}

// TxSuccessOrExceedsBlockGasLimit returns true if the transaction was successful
// or if it failed with an ExceedBlockGasLimit error
func TxSuccessOrExceedsBlockGasLimit(res *abci.ResponseDeliverTx) bool {
	return res.Code == 0 || TxExceedBlockGasLimit(res)
}

// ParseTxIndexerResult parse tm tx result to a format compatible with Ethereum.
func ParseTxIndexerResult(txResult *tmrpctypes.ResultTx, tx sdk.Tx) (*types.TxResult, error) {
	// txs, err := ParseTxResult(&txResult.TxResult, tx)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to parse tx events: block %d, index %d, %v", txResult.Height, txResult.Index, err)
	// }

	// index := uint32(parsedTx.MsgIndex) // #nosec G701
	return &types.TxResult{
		Height:  txResult.Height,
		TxIndex: txResult.Index,
		// MsgIndex:          index,
		EthTxIndex:        int32(txResult.Index),
		Failed:            txResult.TxResult.IsErr(),
		GasUsed:           uint64(txResult.TxResult.GasUsed),
		CumulativeGasUsed: uint64(txResult.TxResult.GasUsed),
	}, nil
}

// ParsedTx is the tx infos parsed from events.
type ParsedTx struct {
	MsgIndex int

	// the following fields are parsed from events

	Hash common.Hash
	// -1 means uninitialized
	EthTxIndex int32
	GasUsed    uint64
	Failed     bool
}

// NewParsedTx initialize a ParsedTx
func NewParsedTx(msgIndex int) ParsedTx {
	return ParsedTx{MsgIndex: msgIndex, EthTxIndex: -1}
}

// ParseTxResult parse eth tx infos from cosmos-sdk events.
// It supports two event formats, the formats are described in the comments of the format constants.
func ParseTxResult(result *abci.ResponseDeliverTx, tx sdk.Tx) (*ParsedTx, error) {
	var ethtx *ParsedTx

	// for _, event := range result.Events {
	// 	switch event.Type {
	// 	case wasmxtypes.EventTypeExecute:
	// 		// ethtx.
	// 	case wasmxtypes.EventTypeDeploy:
	// 	default:
	// 		continue
	// 	}
	// }

	// // some old versions miss some events, fill it with tx result
	// gasUsed := uint64(result.GasUsed) // #nosec G701
	// if len(p.Txs) == 1 {
	// 	p.Txs[0].GasUsed = gasUsed
	// }

	// this could only happen if tx exceeds block gas limit
	if result.Code != 0 && tx != nil {
		ethtx.Failed = true
		// replace gasUsed with gasLimit because that's what's actually deducted.
		gasLimit := tx.GetMsgs()[0].(*wasmxtypes.MsgExecuteEth).AsTransaction().Gas()
		ethtx.GasUsed = gasLimit

	}
	return ethtx, nil
}
