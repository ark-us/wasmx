# FSM-Driven Consensus Contracts

```mermaid
sequenceDiagram
    autonumber
    participant Event as "Consensus event"
    participant Executor as "ActionExecutor / tx execution"
    participant FSM as "FSM interpreter WASM contract"
    participant Def as "FSM definition and context"
    participant Algo as "Consensus algorithm black box"
    participant Actions as "Consensus action library WASM"
    participant Host as "Go host APIs"
    participant BaseApp as "BaseApp transaction engine"

    Event->>Executor: newTransaction / timeout / P2P message / StartNode
    Executor->>FSM: Execute interpreter entry point
    FSM->>Def: Load current state, event, guards, actions
    FSM->>Algo: Interpret next transition
    Note over Algo: The algorithm is data-driven FSM logic, not hard-coded Go control flow.
    Algo-->>FSM: Selected transition and ordered actions
    loop Each action emitted by the transition
        FSM->>Actions: External action call
        alt Timer / delayed state transition
            Actions->>Host: StartTimeout(id, contract, delay, args)
            Host-->>Executor: Later reentry into FSM timed entry point
        else New transaction accepted
            Actions->>Actions: Add transaction to mempool / consensus context
        else Block production action
            Actions->>Host: FinalizeBlock(request, metainfo)
            Host->>BaseApp: FinalizeBlockSimple
            BaseApp->>BaseApp: Execute ordered transactions
            BaseApp-->>Host: tx results, events, app hash inputs
            Host-->>Actions: FinalizeBlock response
        else Network action
            Actions->>Host: Send P2P / state sync / peer message
        end
    end
    FSM->>Def: Persist new FSM state and context
    FSM-->>Executor: Interpreter result
    Executor->>Executor: EndTransaction and commit on success
```

Consensus behavior is expressed as FSM data plus WASM contracts. Go provides host APIs, networking, timers, and transaction boundaries, but the consensus flow is interpreted by the FSM interpreter WASM contract.

A consensus event does not directly run hard-coded Go logic. It enters a WASM interpreter, the interpreter evaluates the current FSM state and transition table, then the selected actions can start concrete transaction execution through host APIs such as `FinalizeBlock`.

Source anchors:
- `ARCHITECTURE.md`: transaction cycles, block end, and reentry model
- `tests/testdata/tinygo/wasmx-fsm/cmd/main.go`: FSM interpreter entry points
- `tests/testdata/tinygo/wasmx-fsm/lib/machine.go`: FSM transition execution and timer integration
- `tests/testdata/tinygo/wasmx-raft-lib/lib/actions.go`: consensus actions that finalize blocks
- `wasmx/x/network/keeper/keeper_timeout.go`: timed reentry
- `wasmx/x/wasmx/vm/ops_wasmx_consensus.go`: host `FinalizeBlock` API
