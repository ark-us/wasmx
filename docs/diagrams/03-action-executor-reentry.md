# ActionExecutor Reentry

```mermaid
flowchart TD
    External["External or system event"] --> Source{"Event source"}
    Source --> Timer["Timer timeout"]
    Source --> P2P["Incoming P2P message"]
    Source --> Reentry["Explicit reentry request"]
    Source --> Other["HTTP / email / system hook"]

    Timer --> Goroutine["Start goroutine under parent context"]
    P2P --> Goroutine
    Reentry --> Goroutine
    Other --> Goroutine

    Goroutine --> CancelCheck{"Parent context closed?"}
    CancelCheck -->|Yes| Drop["Do not execute"]
    CancelCheck -->|No| Executor["ActionExecutor.Execute"]

    Executor --> Lock["Acquire executor mutex"]
    Lock --> Header["Create SDK context from block header"]
    Header --> Begin["BaseApp.BeginTransaction"]
    Begin --> Callback["Run callback"]
    Callback --> Contract["Execute contract entry point"]
    Contract --> End["BaseApp.EndTransaction"]
    End --> Result{"Callback error or query mode?"}
    Result -->|Error| NoCommit["Return error without commit"]
    Result -->|Query| NoCommit
    Result -->|Finalize success| Commit["Commit cache context"]
    Commit --> Unlock["Release executor mutex"]
    NoCommit --> Unlock
```

WasmX allows asynchronous events to trigger contract execution, but those executions cannot mutate shared state arbitrarily. They are funneled through one serialized executor per app/chain, which gives every reentry the same transaction hooks and commit rules as deterministic transaction execution.

Source anchors:
- `wasmx/x/network/keeper/executor.go`: `ActionExecutor`
- `wasmx/x/network/keeper/keeper_timeout.go`: timer-triggered reentry
- `wasmx/x/network/keeper/keeper_p2p.go`: P2P-triggered reentry
- `wasmx/x/network/keeper/keeper_reentry_gor.go`: goroutine-backed reentry
