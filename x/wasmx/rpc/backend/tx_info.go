package backend

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	etherminttypes "github.com/evmos/ethermint/types"

	rpctypes "mythos/v1/x/wasmx/rpc/types"
	"mythos/v1/x/wasmx/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
	// wasmxvm "mythos/v1/x/wasmx/vm"
)

// GetTransactionByHash returns the Ethereum format transaction identified by Ethereum transaction hash
func (b *Backend) GetTransactionByHash(txHash common.Hash) (*rpctypes.RPCTransaction, error) {
	res, err := b.GetTxByEthHash(txHash)

	// if err != nil {
	// 	return b.getTransactionByHashPending(txHash)
	// }

	if err != nil {
		return nil, errors.New("tx does not exist")
	}

	block, err := b.TendermintBlockByNumber(rpctypes.BlockNumber(res.Height))
	if err != nil {
		return nil, err
	}

	tx, err := b.clientCtx.TxConfig.TxDecoder()(block.Block.Txs[res.TxIndex])
	if err != nil {
		return nil, err
	}

	// the `res.MsgIndex` is inferred from tx index, should be within the bound.
	msg, ok := tx.GetMsgs()[res.MsgIndex].(*wasmxtypes.MsgExecuteEth)
	if !ok {
		return nil, errors.New("invalid ethereum tx")
	}

	blockRes, err := b.TendermintBlockResultByNumber(&block.Block.Height)
	if err != nil {
		b.logger.Debug("block result not found", "height", block.Block.Height, "error", err.Error())
		return nil, nil
	}

	// if res.EthTxIndex == -1 {
	// 	// Fallback to find tx index by iterating all valid eth transactions
	// 	msgs := b.EthMsgsFromTendermintBlock(block, blockRes)
	// 	for i := range msgs {
	// 		if msgs[i].Hash == hexTx {
	// 			res.EthTxIndex = int32(i)
	// 			break
	// 		}
	// 	}
	// }
	// if we still unable to find the eth tx index, return error, shouldn't happen.
	if res.EthTxIndex == -1 {
		return nil, errors.New("can't find index of ethereum tx")
	}

	baseFee, err := b.BaseFee(blockRes)
	if err != nil {
		// handle the error for pruned node.
		b.logger.Error("failed to fetch Base Fee from prunned block. Check node prunning configuration", "height", blockRes.Height, "error", err)
	}

	return rpctypes.NewTransactionFromMsg(
		txHash,
		msg,
		common.BytesToHash(block.BlockID.Hash.Bytes()),
		uint64(res.Height),
		uint64(res.EthTxIndex),
		baseFee,
		b.chainID,
	)
}

// GetTransactionReceipt returns the transaction receipt identified by hash.
func (b *Backend) GetTransactionReceipt(hash common.Hash) (map[string]interface{}, error) {
	hexTx := hash.Hex()
	b.logger.Debug("eth_getTransactionReceipt", "hash", hexTx)

	res, err := b.GetTxByEthHash(hash)
	if err != nil {
		b.logger.Debug("tx not found", "hash", hexTx, "error", err.Error())
		return nil, nil
	}
	resBlock, err := b.TendermintBlockByNumber(rpctypes.BlockNumber(res.Height))
	if err != nil {
		b.logger.Debug("block not found", "height", res.Height, "error", err.Error())
		return nil, nil
	}
	tx, err := b.clientCtx.TxConfig.TxDecoder()(resBlock.Block.Txs[res.TxIndex])
	if err != nil {
		b.logger.Debug("decoding failed", "error", err.Error())
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}
	ethMsg := tx.GetMsgs()[res.MsgIndex].(*wasmxtypes.MsgExecuteEth)
	txData, err := wasmxtypes.UnpackTxData(ethMsg.Data)
	if err != nil {
		b.logger.Error("failed to unpack tx data", "error", err.Error())
		return nil, err
	}

	cumulativeGasUsed := uint64(0)
	blockRes, err := b.TendermintBlockResultByNumber(&res.Height)
	if err != nil {
		b.logger.Debug("failed to retrieve block results", "height", res.Height, "error", err.Error())
		return nil, nil
	}
	for _, txResult := range blockRes.TxsResults[0:res.TxIndex] {
		cumulativeGasUsed += uint64(txResult.GasUsed)
	}
	cumulativeGasUsed += res.CumulativeGasUsed

	var status hexutil.Uint
	if res.Failed {
		status = hexutil.Uint(ethtypes.ReceiptStatusFailed)
	} else {
		status = hexutil.Uint(ethtypes.ReceiptStatusSuccessful)
	}
	sender := types.EvmAddressFromAcc(sdk.MustAccAddressFromBech32(ethMsg.Sender))

	// parse tx logs from events
	// sets contract address, data, topics, log index
	logs, err := TxLogsFromEvents(blockRes.TxsResults[res.TxIndex].Events)
	if err != nil {
		b.logger.Debug("failed to parse logs", "hash", hexTx, "error", err.Error())
	}
	for i := range logs {
		logs[i].BlockNumber = uint64(res.Height)
		logs[i].BlockHash = common.BytesToHash(resBlock.Block.Header.Hash())
		logs[i].TxHash = hash
		logs[i].TxIndex = uint(res.EthTxIndex)
	}

	// if res.EthTxIndex == -1 {
	// 	// Fallback to find tx index by iterating all valid eth transactions
	// 	msgs := b.EthMsgsFromTendermintBlock(resBlock, blockRes)
	// 	for i := range msgs {
	// 		if msgs[i].Hash == hexTx {
	// 			res.EthTxIndex = int32(i) // #nosec G701
	// 			break
	// 		}
	// 	}
	// }
	// // return error if still unable to find the eth tx index
	// if res.EthTxIndex == -1 {
	// 	return nil, errors.New("can't find index of ethereum tx")
	// }

	receipt := map[string]interface{}{
		// Consensus fields: These fields are defined by the Yellow Paper
		"status":            status,
		"cumulativeGasUsed": hexutil.Uint64(cumulativeGasUsed),
		"logsBloom":         ethtypes.BytesToBloom(ethtypes.LogsBloom(logs)),
		"logs":              logs,

		// Implementation fields: These fields are added by geth when processing a transaction.
		// They are stored in the chain database.
		"transactionHash": hash,
		"contractAddress": nil,
		"gasUsed":         hexutil.Uint64(res.GasUsed),

		// Inclusion information: These fields provide information about the inclusion of the
		// transaction corresponding to this receipt.
		"blockHash":        common.BytesToHash(resBlock.Block.Header.Hash()).Hex(),
		"blockNumber":      hexutil.Uint64(res.Height),
		"transactionIndex": hexutil.Uint64(res.EthTxIndex),

		// sender and receiver (contract or EOA) addreses
		"from": sender,
		"to":   txData.To(),
		"type": hexutil.Uint(ethMsg.AsTransaction().Type()),
	}

	if logs == nil {
		receipt["logs"] = [][]*ethtypes.Log{}
	}

	// TODO get contractAddress from EventTypeDeploy, AttributeKeyContractAddr (bech32)

	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	if txData.To() == nil {
		// get the contract address from the logs
		newContractAddress := ContractAddressFromEvents(blockRes.TxsResults[res.TxIndex].Events)
		receipt["contractAddress"] = newContractAddress
	}

	// if dynamicTx, ok := txData.(*evmtypes.DynamicFeeTx); ok {
	// 	baseFee, err := b.BaseFee(blockRes)
	// 	if err != nil {
	// 		// tolerate the error for pruned node.
	// 		b.logger.Error("fetch basefee failed, node is pruned?", "height", res.Height, "error", err)
	// 	} else {
	// 		receipt["effectiveGasPrice"] = hexutil.Big(*dynamicTx.EffectiveGasPrice(baseFee))
	// 	}
	// }

	return receipt, nil
}

// GetTxByEthHash uses `/tx_query` to find transaction by ethereum tx hash
// TODO: Don't need to convert once hashing is fixed on Tendermint
// https://github.com/tendermint/tendermint/issues/6539
func (b *Backend) GetTxByEthHash(hash common.Hash) (*etherminttypes.TxResult, error) {
	txHash := hash.Big().FillBytes(make([]byte, 32))
	txResult, err := b.clientCtx.Client.Tx(b.ctx, txHash, false)
	if err != nil {
		return nil, err
	}
	if txResult == nil {
		return nil, errors.New("ethereum tx not found")
	}
	if !rpctypes.TxSuccessOrExceedsBlockGasLimit(&txResult.TxResult) {
		return nil, errors.New("invalid ethereum tx")
	}

	var tx sdk.Tx
	if txResult.TxResult.Code != 0 {
		// it's only needed when the tx exceeds block gas limit
		tx, err = b.clientCtx.TxConfig.TxDecoder()(txResult.Tx)
		if err != nil {
			return nil, fmt.Errorf("invalid ethereum tx")
		}
	}

	return rpctypes.ParseTxIndexerResult(txResult, tx)
}

// queryTendermintTxIndexer query tx in tendermint tx indexer
func (b *Backend) queryTendermintTxIndexer(query string) (*etherminttypes.TxResult, error) {
	resTxs, err := b.clientCtx.Client.TxSearch(b.ctx, query, false, nil, nil, "")
	if err != nil {
		return nil, err
	}
	if len(resTxs.Txs) == 0 {
		return nil, errors.New("ethereum tx not found")
	}
	txResult := resTxs.Txs[0]
	if !rpctypes.TxSuccessOrExceedsBlockGasLimit(&txResult.TxResult) {
		return nil, errors.New("invalid ethereum tx")
	}

	var tx sdk.Tx
	if txResult.TxResult.Code != 0 {
		// it's only needed when the tx exceeds block gas limit
		tx, err = b.clientCtx.TxConfig.TxDecoder()(txResult.Tx)
		if err != nil {
			return nil, fmt.Errorf("invalid ethereum tx")
		}
	}

	return rpctypes.ParseTxIndexerResult(txResult, tx)
}
