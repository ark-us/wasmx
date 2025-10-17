package lib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	p2p "github.com/loredanacirstea/wasmx-env-p2p/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	fsm "github.com/loredanacirstea/wasmx-fsm/lib"
	raftlib "github.com/loredanacirstea/wasmx-raft-lib/lib"
)

// SetupNode initializes local state from InitChainSetup and peers (AS parity)
func SetupNode(_ []fsm.ActionParam, event fsm.EventObject) error {
	// read base64-encoded InitChainSetup from event params (key: "data")
	dataB64 := ""
	for _, p := range event.Params {
		if p.Key == "data" {
			dataB64 = p.Value
			break
		}
	}
	if dataB64 == "" {
		return fmt.Errorf("no initChainSetup found")
	}
	raw, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil {
		return err
	}
	// TODO remove validator private key from logs in initChainSetup
	LoggerDebug("setupNode", []string{"initChainSetup", string(raw)})

	// Reset commit/applied
	if err := raftlib.SetLastApplied(raftlib.LOG_START); err != nil {
		return err
	}

	// Decode into raftlib.InitChainSetup
	var init raftlib.InitChainSetup
	if err := json.Unmarshal(raw, &init); err != nil {
		return err
	}

	// Set current node ID (for single node, this is always 0)
	if err := raftlib.SetCurrentNodeId(0); err != nil {
		return err
	}

	// For single-node consensus, create minimal node info with just ourselves
	// This is needed because raft-lib functions expect node info to be populated
	selfNode := []p2p.NodeInfo{
		{
			Address:   wasmx.Bech32String(init.ValidatorAddress),
			Node:      p2p.NetworkNode{IP: "127.0.0.1"}, // Minimal placeholder
			OutOfSync: false,
		},
	}
	if err := raftlib.SetNodeIPs(selfNode); err != nil {
		return err
	}

	// Initialize chain state and store consensus params for next height
	if err := raftlib.InitChain(init); err != nil {
		return err
	}

	// Initialize Next/Match index arrays with length 1 (single node)
	return raftlib.InitializeIndexArrays(1)
}

func StartNode() error {
	return nil
}

// NewBlock builds and finalizes a block immediately for single-node on-demand consensus
func NewBlock() error {
	// Step 1: Build the block proposal using the raft-lib ProposeBlock
	// This will batch transactions from mempool and append to log
	if err := raftlib.ProposeBlock(nil, fsm.EventObject{}); err != nil {
		return fmt.Errorf("failed to propose block: %v", err)
	}

	// Step 2: Get the newly created log entry (should be at last log index)
	lastLogIndex, err := raftlib.GetLastLogIndex()
	if err != nil {
		return fmt.Errorf("failed to get last log index: %v", err)
	}

	// Step 3: Retrieve the block entry from the log
	entry, err := raftlib.GetLogEntryAggregate(lastLogIndex)
	if err != nil {
		return fmt.Errorf("failed to get log entry at index %d: %v", lastLogIndex, err)
	}
	if entry == nil {
		return fmt.Errorf("log entry is nil at index %d", lastLogIndex)
	}

	LoggerInfo("finalizing block immediately", []string{
		"height", raftlib.Int64ToString(entry.Index),
		"termId", raftlib.Int32ToString(entry.TermID),
	})

	_, err = raftlib.StartBlockFinalizationInternal(entry, false)
	if err != nil {
		return fmt.Errorf("failed to finalize block at index %d: %v", lastLogIndex, err)
	}
	return nil
}
