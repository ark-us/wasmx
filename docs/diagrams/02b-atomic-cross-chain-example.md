# Atomic Cross-Chain Example: Two Transactions With Nested Call

```mermaid
sequenceDiagram
    autonumber
    participant Client as "Client / proposer"
    participant Executor as "ActionExecutor locked lane"
    participant Atomic as "ExecuteAtomicTxInternal"
    participant MC as "MultiChainContext channels"
    participant Meta as "ExecutionMetaInfo"
    participant ChainA as "Chain A app"
    participant A1 as "Contract A1"
    participant A2 as "Contract A2"
    participant ChainB as "Chain B app"
    participant B1 as "Contract B1"
    participant B1N as "Contract B1Nested"
    participant B2 as "Contract B2"

    Client->>Atomic: Atomic batch [Tx1, Tx2]
    Atomic->>MC: Set atomic tx hash and allowed chains [A, B]
    Atomic->>MC: Create per-chain result and cross-chain call channels
    Atomic->>Executor: Enter serialized execution lane
    Note over Executor: One ActionExecutor per app/chain. The mutex makes non-deterministic and reentry execution successive, not concurrent.

    rect rgb(238, 246, 255)
        Note over Atomic,ChainB: Tx1 targets Chain A and contains two contract calls
        Atomic->>Executor: Execute ordered subtransaction index 0
        Executor->>ChainA: DeliverTx(Tx1)
        ChainA->>A1: Call 1: execute local contract A1
        A1-->>ChainA: Local state transition and events
        ChainA->>A1: Call 2: execute contract A1 logic that needs Chain B
        A1->>Atomic: executeCrossChainTx(to Chain B, Contract B1)
        Atomic->>MC: Send cross-chain call request {subtx=0, call=0}
        MC->>ChainB: Route request to Chain B execution goroutine
        ChainB->>Executor: Reenter serialized execution lane for cross-chain call
        Executor->>B1: Execute Contract B1
        B1->>B1N: Nested contract call
        B1N-->>B1: Nested result
        B1-->>Executor: Cross-chain result
        Executor-->>ChainB: Commit or rollback by transaction hooks
        ChainB->>MC: Send cross-chain response {subtx=0, call=0}
        MC-->>Atomic: Receive B1 response
        Atomic-->>A1: Return deterministic cross-chain response
        A1-->>ChainA: Finish Tx1
        ChainA-->>Executor: Tx1 ABCI result and emitted events
        Executor-->>Atomic: Ordered result for subtx 0
        Atomic->>Meta: Record request/response for replay
        Atomic->>MC: Publish Tx1 partial result on Chain A result channel
    end

    rect rgb(242, 250, 242)
        Note over Atomic,ChainA: Tx2 targets Chain B and also contains two contract calls
        Atomic->>Executor: Execute ordered subtransaction index 1
        Executor->>ChainB: DeliverTx(Tx2)
        ChainB->>B2: Call 3: execute local contract B2
        B2-->>ChainB: Local state transition and events
        ChainB->>B2: Call 4: execute contract B2 logic that needs Chain A
        B2->>Atomic: executeCrossChainTx(to Chain A, Contract A2)
        Atomic->>MC: Send cross-chain call request {subtx=1, call=0}
        MC->>ChainA: Route request to Chain A execution goroutine
        ChainA->>Executor: Reenter serialized execution lane for cross-chain call
        Executor->>A2: Execute Contract A2
        A2-->>Executor: Cross-chain result
        Executor-->>ChainA: Commit or rollback by transaction hooks
        ChainA->>MC: Send cross-chain response {subtx=1, call=0}
        MC-->>Atomic: Receive A2 response
        Atomic-->>B2: Return deterministic cross-chain response
        B2-->>ChainB: Finish Tx2
        ChainB-->>Executor: Tx2 ABCI result and emitted events
        Executor-->>Atomic: Ordered result for subtx 1
        Atomic->>Meta: Record request/response for replay
        Atomic->>MC: Publish Tx2 partial result on Chain B result channel
    end

    Atomic->>Executor: Exit serialized execution lane
    Atomic->>Atomic: Merge ordered results [Tx1 result, Tx2 result]
    Atomic->>MC: Clear current atomic tx hash and chain set
    Atomic-->>Client: MsgExecuteAtomicTxResponse

    Note over Meta: Validators that cannot execute every chain use metainfo to replay/verify the same ordered cross-chain calls.
```

## What This Shows

The atomic batch is the outer transaction boundary. Inside it, each subtransaction is executed in a deterministic order through the `ActionExecutor` serialized lane. Cross-chain reentries also pass back through that same execution discipline, so they are successive and transaction-hooked rather than free-running goroutines.

A cross-chain contract call is treated as an indexed internal call: `(atomic tx hash, subtx index, cross-chain call index)`.

The nested call under `Contract B1` is still a normal contract subcall inside Chain B execution. It is not a new atomic subtransaction; it is part of the call tree for `Tx1`'s cross-chain call.

Source anchors:
- `wasmx/x/network/keeper/executor.go`: locked `ActionExecutor` and transaction hooks
- `wasmx/x/network/keeper/multichain.go`: atomic batch loop, channels, cross-chain call timeout
- `wasmx/x/network/vmcrosschain/ops.go`: contract-facing cross-chain host API
- `wasmx/x/network/types/ctx_metainfo_crosschain.go`: cross-chain request/response metainfo
- `wasmx/x/wasmx/vm/ops_common.go`: nested contract call handling
