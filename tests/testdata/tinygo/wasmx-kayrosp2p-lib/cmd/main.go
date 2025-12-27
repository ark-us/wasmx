package main

import (
	"encoding/json"
	"fmt"
	"os"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	fsm "github.com/loredanacirstea/wasmx-fsm/lib"
	lib "github.com/loredanacirstea/wasmx-kayrosp2p-lib/lib"
	raft "github.com/loredanacirstea/wasmx-raft-lib/lib"
	raftp2p "github.com/loredanacirstea/wasmx-raftp2p-lib/lib"
)

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module consensus
//export wasmx_consensus_json_i64_1
func Wasmx_consensus_json_i64_1() {}

//go:wasm-module wasmxcore
//export wasmx_env_core_i64_1
func Wasmx_env_core_i64_1() {}

//go:wasm-module crosschain
//export wasmx_crosschain_json_i64_1
func Wasmx_crosschain_json_i64_1() {}

//go:wasm-module multichain
//export wasmx_multichain_json_i64_1
func Wasmx_multichain_json_i64_1() {}

//go:wasm-module p2p
//export wasmx_p2p_json_i64_1
func Wasmx_p2p_json_i64_1() {}

//go:wasm-module wasmxcore
//export wasmx_nondeterministic_1
func Wasmx_nondeterministic_1() {}

//go:wasm-module wasmx-raftp2p
//export instantiate
func Instantiate() {}

func main() {
	entrypoint := os.Getenv("ENTRY_POINT")
	lib.LoggerDebugExtended("fsm: ENTRY_POINT: "+entrypoint, nil)
	// these entrypoints are internal by default
	switch entrypoint {
	case "instantiate":
		databz := wasmx.GetCallData()
		var msg lib.InstantiateMsg
		if err := json.Unmarshal(databz, &msg); err != nil {
			wasmx.Revert([]byte("invalid instantiate message: " + err.Error()))
			return
		}

		// Initialize Kayros client with the provided API URL
		config := lib.KayrosConfig{
			ApiBaseUrl: msg.KayrosApiUrl,
		}
		if err := lib.InitializeKayrosClient(config); err != nil {
			wasmx.Revert([]byte("failed to initialize Kayros client: " + err.Error()))
			return
		}

		// Store consensus configuration
		consensusConfig := lib.ConsensusConfig{
			ThresholdCommit:   msg.ThresholdCommit,
			ThresholdFinalize: msg.ThresholdFinalize,
			GenesisUUID:       msg.GenesisUUID,
		}
		// Set defaults if not provided
		if consensusConfig.ThresholdCommit == 0 {
			consensusConfig.ThresholdCommit = 51
		}
		if consensusConfig.ThresholdFinalize == 0 {
			consensusConfig.ThresholdFinalize = 75
		}
		if err := lib.SetConsensusConfig(consensusConfig); err != nil {
			wasmx.Revert([]byte("failed to set consensus config: " + err.Error()))
			return
		}

		// Set genesis UUID as initial last UUID if provided
		if msg.GenesisUUID != "" {
			lib.SetLastKayrosUUID(msg.GenesisUUID)
		}

		lib.LoggerInfo("Kayros client initialized", []string{
			"api_url", msg.KayrosApiUrl,
			"genesis_uuid", msg.GenesisUUID,
			"threshold_commit", fmt.Sprintf("%d", consensusConfig.ThresholdCommit),
			"threshold_finalize", fmt.Sprintf("%d", consensusConfig.ThresholdFinalize),
		})
		wasmx.Finish([]byte{})
		return
	}

	// Only internal
	wasmx.OnlyInternal(raftp2p.MODULE_NAME, "")

	databz := wasmx.GetCallData()
	var calld fsm.ExternalActionCallData
	if err := json.Unmarshal(databz, &calld); err != nil {
		raftp2p.Revert("invalid call data: " + err.Error() + ": " + string(databz))
		return
	}

	// Helper to read params from event
	get := func(key string) string {
		for _, p := range calld.Event.Params {
			if p.Key == key {
				return p.Value
			}
		}
		for _, p := range calld.Params {
			if p.Key == key {
				return p.Value
			}
		}
		return ""
	}

	// Route methods
	switch calld.Method {
	case "ifNodeIsValidator":
		ok, err := raftp2p.IfNodeIsValidator(calld.Params, calld.Event)
		if err != nil {
			raftp2p.Revert(err.Error())
			return
		}
		res := raft.WrapGuard(ok)
		wasmx.Finish(res)
		return
	case "ifNewTransaction":
		ok, err := raftp2p.IfNewTransaction(calld.Params, calld.Event)
		if err != nil {
			raftp2p.Revert(err.Error())
			return
		}
		res := raft.WrapGuard(ok)
		wasmx.Finish(res)
		return
	case "setupNode":
		if err := raftp2p.SetupNode(calld.Params, calld.Event); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "connectPeers":
		st, err := raft.GetCurrentState()
		if err != nil {
			raftp2p.Revert(err.Error())
			return
		}
		if err := raftp2p.ConnectPeersInternal(raftp2p.GetProtocolIdFromState(st)); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "connectRooms":
		if err := raftp2p.ConnectRooms(); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "registerWithKayros":
		err := lib.RegisterWithKayros(calld.Params, calld.Event)
		if err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "getKayrosTxs":
		err := lib.GetKayrosTxs(calld.Params, calld.Event)
		if err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "forwardMsgToChat":
		if err := raftp2p.ForwardMsgToChat(calld.Params, calld.Event); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "requestBlockSync":
		if err := raftp2p.RequestBlockSync(); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "receiveStateSyncRequest":
		entry := get("entry")
		sig := get("signature")
		sender := get("sender")
		if err := raftp2p.ReceiveStateSyncRequest(entry, sig, wasmx.Bech32String(sender)); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "receiveStateSyncResponse":
		entry := get("entry")
		sender := get("sender")
		if err := raftp2p.ReceiveStateSyncResponse(entry, sender); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "receiveUpdateNodeResponse":
		entry := get("entry")
		sig := get("signature")
		sender := get("sender")
		if err := raftp2p.ReceiveUpdateNodeResponse(entry, sig, wasmx.Bech32String(sender)); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "receiveCommit":
		if err := lib.ReceiveCommit(calld.Params, calld.Event); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "receiveMissingTransactions":
		if err := lib.ReceiveMissingTransactions(calld.Params, calld.Event); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "receiveMissingTransactionsRequest":
		if err := lib.ReceiveMissingTransactionsRequest(calld.Params, calld.Event); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "ifAllTransactions":
		ok := lib.IfAllTransactions(nil, calld.Event)
		res := raft.WrapGuard(ok)
		wasmx.Finish(res)
		return
	case "sendNewTransactionResponse":
		if err := raft.SendNewTransactionResponse(nil, fsm.EventObject{}); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "addToMempool":
		if err := raftp2p.AddToMempool(calld.Params, calld.Event); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "updateNodeAndReturn":
		if err := raftp2p.UpdateNodeAndReturn(calld.Params, calld.Event); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "registerValidatorWithNetwork":
		if err := raftp2p.RegisterValidatorWithNetwork(nil, fsm.EventObject{}); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "setup":
		if err := raft.Setup(calld.Params, calld.Event); err != nil {
			raftp2p.Revert(err.Error())
			return
		}
	case "bootstrapAfterStateSync":
		if err := raft.BootstrapAfterStateSync(calld.Params, calld.Event); err != nil {
			raftp2p.Revert("bootstrapAfterStateSync failed: " + err.Error())
			return
		}
	case "commitAfterStateSync":
		if err := raft.CommitAfterStateSync(calld.Params, calld.Event); err != nil {
			raftp2p.Revert("commitAfterStateSync failed: " + err.Error())
			return
		}
	case "VerifyCommitLight":
		if err := lib.VerifyCommitLight(calld.Params, calld.Event); err != nil {
			raftp2p.Revert("VerifyCommitLight failed: " + err.Error())
			return
		}
	case "rollback":
		if err := raft.Rollback(calld.Params, calld.Event); err != nil {
			raftp2p.Revert("Rollback failed: " + err.Error())
			return
		}
	default:
		wasmx.Revert(append([]byte("invalid function call data: "), databz...))
		return
	}
	wasmx.Finish(wasmx.GetFinishData())
}
