package main

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	fsm "github.com/loredanacirstea/wasmx-fsm/lib"
	lib "github.com/loredanacirstea/wasmx-ondemand-single-lib/lib"
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

//go:wasm-module wasmx-ondemand-single
//export instantiate
func Instantiate() {}

func main() {
	// Only internal
	wasmx.OnlyInternal(lib.MODULE_NAME, "")

	databz := wasmx.GetCallData()
	var calld fsm.ExternalActionCallData
	if err := json.Unmarshal(databz, &calld); err != nil {
		lib.Revert("invalid call data: " + err.Error() + ": " + string(databz))
		return
	}

	// Route methods
	switch calld.Method {
	case "ifNewTransaction":
		ok, err := raftp2p.IfNewTransaction(calld.Params, calld.Event)
		if err != nil {
			lib.Revert(err.Error())
			return
		}
		res := raft.WrapGuard(ok)
		wasmx.Finish(res)
		return
	case "addToMempool":
		if err := raftp2p.AddToMempool(calld.Params, calld.Event); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "setupNode":
		if err := lib.SetupNode(calld.Params, calld.Event); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "sendNewTransactionResponse":
		if err := raft.SendNewTransactionResponse(nil, fsm.EventObject{}); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "newBlock":
		if err := lib.NewBlock(); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "setup":
		if err := raft.Setup(calld.Params, calld.Event); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "StartNode":
		// TODO
		if err := lib.StartNode(); err != nil {
			lib.Revert(err.Error())
			return
		}
	case "rollback":
		if err := raft.Rollback(calld.Params, calld.Event); err != nil {
			lib.Revert("Rollback failed: " + err.Error())
			return
		}
	default:
		wasmx.Revert(append([]byte("invalid function call data: "), databz...))
		return
	}
	wasmx.Finish(wasmx.GetFinishData())
}
