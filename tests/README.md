# wasmx tests


## Run tests

```bash
go test -v ./...

go test --count=1 -short -v ./...

go test --count=1 -short -v ./x/wasmx/keeper

go test --count=1 -timeout 300s -v -run KeeperTestSuite/TestEwasmFibonacci ./x/wasmx/keeper

```
* for wasmedge
```bash
CGO_LDFLAGS='-Wl,-rpath,/Users/user/.wasmedge/lib' go test --count=1 -timeout 3000s -short -v ./...
```
