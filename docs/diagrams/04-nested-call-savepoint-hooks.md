# Nested Contract Call Savepoint Hooks

```mermaid
sequenceDiagram
    autonumber
    participant A as "Contract A"
    participant VM as "WASM VM call router"
    participant Hooks as "App subcall hooks"
    participant SQL as "vmsql module"
    participant DB as "SQL transaction"
    participant B as "Contract B"

    A->>VM: Internal call to Contract B
    VM->>VM: Create nested call context
    VM->>Hooks: BeginSubCall(level, index, isQuery)
    Hooks->>SQL: BeginSubCall
    SQL->>DB: SAVEPOINT sp_action_level_index
    VM->>B: Execute nested contract call
    alt Nested call succeeds and is not query
        B-->>VM: Return data
        VM->>VM: Commit nested cache context and emit events
        VM->>Hooks: EndSubCall(nil)
        Hooks->>SQL: EndSubCall
        SQL->>DB: RELEASE sp_action_level_index
    else Nested call reverts or is query
        B-->>VM: Error or query result
        VM->>Hooks: EndSubCall(error or isQuery)
        Hooks->>SQL: EndSubCall
        SQL->>DB: ROLLBACK TO sp_action_level_index
    end
    VM-->>A: Success flag and return data
```

Implementation detail behind deterministic SQL. The VM already tracks nested contract call level and index. I used that execution structure as instrumentation: each nested call gets a unique SQL savepoint, so inner call failures and query-only paths roll back without corrupting the outer transaction.

Source anchors:
- `wasmx/x/wasmx/vm/ops_common.go`: nested context and `BeginSubCall` / `EndSubCall`
- `wasmx/x/vmsql/module.go`: savepoint naming and rollback/release logic
- `tests/vmsql/sqlite_test.go`: nested call rollback scenarios
