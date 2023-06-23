# Mythos

## prerequisites

```
curl -sSf https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install.sh | bash -s -- -v 0.11.2
```

## testnet

```

mythosd testnet init-files --chain-id=mythos_7000-14 --output-dir=$(pwd)/testnet --v=1 --keyring-backend=test --minimum-gas-prices="1000amyt"

mythosd start --home=./testnet/node0/mythosd

```

## Get started

```
ignite chain serve

ignite chain build
```

`serve` command installs dependencies, builds, initializes, and starts your blockchain in development.

### Configure

Your blockchain in development can be configured with `config.yml`.

### Web Frontend

```
cd vue
npm install
npm run serve
```

## Release
To release a new version of your blockchain, create and push a new tag with `v` prefix. A new draft release with the configured targets will be created.

```
git tag v0.1
git push origin v0.1
```

After a draft release is created, make your final changes from the release page and publish it.

### Install
To install the latest version of your blockchain node's binary, execute the following command on your machine:

```
curl https://get.ignite.com/username/wasmx@latest! | sudo bash
```
`username/wasmx` should match the `username` and `repo_name` of the Github repository to which the source code was pushed. Learn more about [the install process](https://github.com/allinbits/starport-installer).


## Run tests

```
go test -v ./...

go test --count=1 -short -v ./...

go test --count=1 -short -v ./x/wasmx/keeper

go test --count=1 -v -run KeeperTestSuite/TestEwasmFibonacci ./x/wasmx/keeper

```
