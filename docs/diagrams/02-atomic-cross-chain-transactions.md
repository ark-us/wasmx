# Atomic Cross-Chain Transactions

```mermaid
sequenceDiagram
    autonumber
    participant Leader as "Leader chain"
    participant Atomic as "ExecuteAtomicTxInternal"
    participant MC as "MultiChainContext"
    participant ChainA as "Local subchain app"
    participant ChainB as "Remote subchain app"
    participant Meta as "ExecutionMetaInfo"

    Leader->>Atomic: Submit atomic tx batch
    Atomic->>MC: Set current tx hash and authorized chain IDs
    Atomic->>MC: Create result and cross-chain call channels
    loop Each subtransaction in batch
        Atomic->>MC: Set current subtx index and call index
        alt Subtx belongs to current chain
            Atomic->>ChainA: DeliverTx(txbz)
            ChainA-->>Atomic: ABCI result and events
            Atomic->>MC: Publish partial result on result channel
        else Node has target chain state
            Atomic->>MC: Wait for target chain result channel
            MC-->>Atomic: Target chain partial result
        else Node lacks target chain state
            Atomic->>Meta: Read proposer precomputed cross-chain calls
            Atomic->>ChainA: Re-execute local relevant calls
            Atomic->>Atomic: Compare actual result with metainfo
        end
    end
    Atomic->>MC: Clear current atomic tx hash and chain IDs
    Atomic-->>Leader: Combined atomic tx response
```


The transaction batch behaves like an ordered event log across multiple state machines. Nodes that have all affected chains execute directly; nodes with partial state consume deterministic metainfo produced earlier and verify local replay against expected results.

Each subchain behaves like an ordered partition, and the atomic transaction coordinator enforces ordering, correlation IDs, timeouts, and replay validation across partitions.

Source anchors:
- `ARCHITECTURE.md`: optimistic execution and atomic cross-chain notes
- `wasmx/x/network/keeper/multichain.go`: `ExecuteAtomicTxInternal`, `ExecuteCrossChainTxInternal`
- `wasmx/x/network/types/ctx_metainfo_crosschain.go`: deterministic cross-chain call metainfo
