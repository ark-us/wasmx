# Deterministic SQL Transactions

```mermaid
sequenceDiagram
    autonumber
    participant Tx as "Blockchain transaction"
    participant Contract as "WASM smart contract"
    participant SQLHost as "vmsql host API"
    participant SQLModule as "vmsql AppModule hooks"
    participant DB as "SQLite / SQL database"

    Tx->>Contract: Execute contract entry point
    Contract->>SQLHost: Connect / Execute / Query
    SQLHost->>SQLHost: Resolve connection by chain + contract + id
    alt No open SQL transaction
        SQLHost->>DB: BEGIN
        SQLHost->>DB: SAVEPOINT sp0
    end
    SQLHost->>DB: Execute parameterized SQL
    DB-->>SQLHost: Rows / rows affected / error
    SQLHost-->>Contract: JSON response in WASM memory
    Contract-->>Tx: Success or revert
    Tx->>SQLModule: EndTransaction(txerr)
    alt Finalize mode and no tx error
        SQLModule->>DB: RELEASE sp0
        SQLModule->>DB: COMMIT
    else Finalize mode with tx error
        SQLModule->>DB: ROLLBACK TO sp0
        SQLModule->>DB: COMMIT cleanup
    else Query / non-finalize mode
        SQLModule->>DB: ROLLBACK
    end
    SQLModule->>SQLModule: Clear open tx and savepoint map
```

Source anchors:
- `wasmx/x/vmsql/ops.go`: `Connect`, `Execute`, `BatchAtomic`, `Query`, `beginDbTx`
- `wasmx/x/vmsql/module.go`: `EndTransaction`
- `tests/vmsql/sqlite_test.go`: rollback and nested rollback coverage
