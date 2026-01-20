# wasmx tests


## Run tests

```bash
go test -v ./...

go test --count=1 -short -v ./...

go test --count=1 -short -v ./x/wasmx/keeper

go test --count=1 -timeout 300s -v -run KeeperTestSuite/TestEwasmFibonacci ./x/wasmx/keeper

go test --count=1 -timeout 3000s -v -run TestKeeperTestSuite/TestWasmxSimpleStorage ./tests/wasmx -benchmark=true -wasm-runtime=wazero

go test --count=1 -timeout 3000s -v -run TestKeeperTestSuite/TestWasmxSimpleStorage ./tests/wasmx -wasm-runtime=wazero -consensus-label=consensus_kayrosp2p_ondemand_0.0.1 -kayros_user_key=0x0000000000000000000000000000000000000000000000000000000000000001 -kayros_base_url=https://kayros_indexer_url


```
* for wasmedge
```bash
CGO_LDFLAGS='-Wl,-rpath,/Users/user/.wasmedge/lib' go test --count=1 -timeout 3000s -short -v ./...

CGO_LDFLAGS='-Wl,-rpath,/Users/user/.wasmedge/lib' go test --count=1 -timeout 3000s -v -run TestKeeperTestSuite/TestEwasmPrecompileCurve384TestLong ./tests/wasmx -benchmark=true -wasm-runtime=wasmedge
```
