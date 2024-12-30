# The System

## dev

```
go work init ./wasmx ./wasmx-wasmedge ./wasmx-wazero ./mythos
go work use ./newproj
```

## actions

Test locally with https://github.com/nektos/act

```
act -j build --secret-file .secrets --container-architecture linux/amd64 -P ubuntu-latest=nektos/act-environments-ubuntu:18.04

act -j build --secret-file .secrets --container-architecture linux/amd64 -P ubuntu-latest=nektos/act-environments-ubuntu:18.04 --env GOOS=darwin --env GOARCH=arm64

act -j build --env GOOS=linux --env GOARCH=amd64
act -j build --env GOOS=darwin --env GOARCH=arm64
act -j build --env GOOS=windows --env GOARCH=amd64


```
