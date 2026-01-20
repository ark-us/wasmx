package lib

import (
	"encoding/base64"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	fsm "github.com/loredanacirstea/wasmx-fsm/lib"
	raftlib "github.com/loredanacirstea/wasmx-raft-lib/lib"
)

// IfMempoolEmpty checks if the mempool has no transactions
func IfMempoolEmpty(params []fsm.ActionParam, event fsm.EventObject) (bool, error) {
	mempool, err := raftlib.GetMempool()
	if err != nil {
		return false, err
	}
	return len(mempool.Map) == 0, nil
}

// IfMempoolNotEmpty checks if the mempool has at least one transaction
func IfMempoolNotEmpty(params []fsm.ActionParam, event fsm.EventObject) (bool, error) {
	isEmpty, err := IfMempoolEmpty(params, event)
	if err != nil {
		return false, err
	}
	return !isEmpty, nil
}

// IfNewTransaction checks if a transaction is new (not already in mempool)
func IfNewTransaction(params []fsm.ActionParam, event fsm.EventObject) (bool, error) {
	// Extract base64 transaction
	txB64 := ""
	if len(event.Params) > 0 {
		for _, p := range event.Params {
			if p.Key == "transaction" {
				txB64 = p.Value
				break
			}
		}
	}
	if txB64 == "" {
		for _, p := range params {
			if p.Key == "transaction" {
				txB64 = p.Value
				break
			}
		}
	}
	if txB64 == "" {
		return false, fmt.Errorf("no transaction found")
	}

	// Decode transaction from base64 and compute hash
	txBytes, err := base64.StdEncoding.DecodeString(txB64)
	if err != nil {
		return false, err
	}
	txhash := base64.StdEncoding.EncodeToString(wasmx.Sha256(txBytes))

	// Get mempool
	mp, err := raftlib.GetMempool()
	if err != nil {
		return false, err
	}

	// Check if transaction has been seen
	existent := mp.HasSeen(txhash)
	if existent {
		LoggerDebug("mempool: transaction already added or seen", []string{"txhash", txhash})
	}

	return !existent, nil
}

// IfOldTransaction checks if a transaction is already known (not new)
func IfOldTransaction(params []fsm.ActionParam, event fsm.EventObject) (bool, error) {
	isNew, err := IfNewTransaction(params, event)
	if err != nil {
		return false, err
	}
	return !isNew, nil
}

// IfMempoolFull checks if the mempool batch is full based on max_gas and max_bytes
func IfMempoolFull(params []fsm.ActionParam, event fsm.EventObject) (bool, error) {
	mempool, err := raftlib.GetMempool()
	if err != nil {
		return false, err
	}

	// Get consensus params to determine max gas and max bytes
	consensusParams, err := raftlib.GetConsensusParams(0)
	if err != nil {
		return false, err
	}

	maxBytes := consensusParams.Block.MaxBytes
	if maxBytes == -1 {
		maxBytes = raftlib.MaxBlockSizeBytes
	}

	// Check if mempool batch is full
	return mempool.IsBatchFull(consensusParams.Block.MaxGas, maxBytes), nil
}
