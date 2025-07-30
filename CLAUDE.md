
wasmX is a WASMX-based blockchain engine: multiVM, language agnostic, extensible with various host APIs. It's core contracts are also WASM, from bank, governance, to the actual consensus protocol.

`./wasmx/x/wasmx` is the core engine. with the core VM in `.wasmx/x/wasmx/vm`, where the main host APIs live, along with memory abstractions and handlers for various ways of WASM modules interacting with the host.

There are host API extensions in `./wasmx/x/` but also `./wasmx-vmimap` and `./wasmx-vmsmtp`.

Even the WASM runtime is pluggable and it can be ran with both wazero and wasmedge, though wazero is default.

Most tests are in `./tests`

wasmX was inspired by cosmos-sdk, but has converted most modules into WASM contracts and now it only uses the BaseApp to process transactions.

in `./tests/testdata/` there are examples of contracts in various languages, used in tests. Some, also contain wasmx sdks, like tinygo.
Most core contracts are in AssemblyScript, in the `wasmx-as-contracts` repo.
