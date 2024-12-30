# The System

## dev

```
go work init ./wasmx ./wasmx-wasmedge ./wasmx-wazero ./mythos
go work use ./newproj
```

## actions

Test locally with https://github.com/nektos/act

```bash
act -j build --secret-file .secrets --container-architecture linux/amd64 -P ubuntu-latest=nektos/act-environments-ubuntu:18.04

act -j build --secret-file .secrets --container-architecture linux/amd64 -P ubuntu-latest=nektos/act-environments-ubuntu:18.04 --env GOOS=darwin --env GOARCH=arm64

# --env GOOS=linux --env GOARCH=amd64
# --env GOOS=darwin --env GOARCH=arm64
# --env GOOS=windows --env GOARCH=amd64

```

## release

```bash
git tag v0.0.10
git push origin v0.0.10
```
