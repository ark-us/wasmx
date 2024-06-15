# Mythos

## prerequisites

```
curl -sSf https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install.sh | bash -s -- -v 0.11.2
```

## testnet

```

mythosd testnet init-files --chain-id=mythos_7000-14 --output-dir=$(pwd)/testnet --v=1 --keyring-backend=test --minimum-gas-prices="1000amyt" --nocors --libp2p --min-level-validators 2

# same machine validators
mythosd testnet init-files --chain-id=mythos_7000-14 --output-dir=$(pwd)/testnet --v=2 --keyring-backend=test --minimum-gas-prices="1000amyt" --same-machine=true --nocors  --libp2p --min-level-validators=2 --enable-eid=false

# add new node to existing testnet
mythosd testnet add-node 2 "mythos167eea4stw39as3tjsc5mryqsusyt6hhs62mq07@/ip4/127.0.0.1/tcp/5001/p2p/12D3KooWAcvC67ydPNLzd7jsnKr47yngw6H5rVr86etnySDd9aXP" --chain-id=mythos_7000-14 --output-dir=$(pwd)/testnet --keyring-backend=test --minimum-gas-prices="1000amyt" --same-machine=true --nocors --libp2p

mythosd start --home=./testnet/node0/mythosd

mythosd start --home=./testnet/node1/mythosd

# create chain levels after chain start
mythosd testnet create-levels 1 2 --chain-id="mythos_7000-14" --keyring-backend test --home ./testnet

```


## Run tests

```
go test -v ./...

go test --count=1 -short -v ./...

go test --count=1 -short -v ./x/wasmx/keeper

go test --count=1 -timeout 300s -v -run KeeperTestSuite/TestEwasmFibonacci ./x/wasmx/keeper

```
* for macos 14
```
CGO_LDFLAGS='-Wl,-rpath,/Users/user/.wasmedge/lib' go test --count=1 -short -v ./...
```

## Precompiles

| Precompile         | CodeHash     | address    |
|--------------------|--------------|------------|
| ecrecovereth | f0da10e5c8a458dc24165afcd9b4e9a546b764a29388f382d336a4fcb9cd6263 | 0x000000000000000000000000000000000000000000000000000000000000001f |
| sha2-256     | 33d40ed73a30238f095587397b7f431c2ed0e893c08e759dcd36d82d51cf78a1 | 0x0000000000000000000000000000000000000000000000000000000000000002 |
| ripmd160     | f5446c0aa6756c77b46b9a9400ce0a83907b5ef3bbc855e43ea1e405f5b9fc21 | 0x0000000000000000000000000000000000000000000000000000000000000003 |
| identity     | 8520544489571f0eb2a1db89de33165bc7165572ce7fc075f3cc8bb52948f529 | 0x0000000000000000000000000000000000000000000000000000000000000004 |
| modexp       | 3fd5041e057f85c4d1ab6ffa8d6c95da30496efd043095e80043e81f1739724f | 0x0000000000000000000000000000000000000000000000000000000000000005 |
| ecadd        | 3698fca4441588bb8d651b624480a230c3d70fc096d473c4a5833c3f3c552cc3 | 0x0000000000000000000000000000000000000000000000000000000000000006 |
| ecmul        | 70456af278e41e91045eacd7e1d9136f12676f614c2aac5623e6c7b4fd8d2f47 | 0x0000000000000000000000000000000000000000000000000000000000000007 |
| ecpairings   | 9e3b23a3a2bbabc04683549f669959ae029f894e35e00b0f7cb6b1eb88184859 | 0x0000000000000000000000000000000000000000000000000000000000000008 |
| blake2f      | 72975c3f90e3f327ae9691f912c70629f56af571d82b1a6ec80f1d40f5b93c8c | 0x0000000000000000000000000000000000000000000000000000000000000009 |
| secp384r1    | 564953b2fecd1d407c12fdc5561cba42d943875f5052a9fdae07867f1503e425 | 0x0000000000000000000000000000000000000000000000000000000000000020 |
| secp384r1_registry | 7826207180357cfff71028dfd847688e6379cfaac6f8f7d5624bd801fb99111f | 0x0000000000000000000000000000000000000000000000000000000000000021 |
| evm_shanghai | 8870e4eb2859ccaaa50a06de94cec658d78617df336a8ec29f0a5c9f29bf975a | 0x0000000000000000000000000000000000000000000000000000000000000023 |

## Default Ports

* `1317` API server; REST HTTP server generated from protos (cosmos-sdk)
* `9090` gRPC server generated from protos (cosmos-sdk)
* `26656` consensus p2p incoming connections - p2p.laddr (former cometbft port)
* `26657` consensus RPC - HTTP server + JSON-RPC server - rpc.laddr (former cometbft port)
* `26657/websocket` consensus RPC - websockets (former cometbft port)
* `8090` network module - smart contract GRPC requests
* `8545` JSON-RPC
* `8546` JSON-RPC websockets
* `9999` websrv webserver
* `6060` pprof
* `5001` libp2p port

## Multichain Commands

```

mythosd tx multichain register-subchain logos lyt 18 1 "10000000000" --chain-id="leveln_1000-1" --from node0 --keyring-backend test --home ./testnet/node0/mythosd --fees 200000000000alvl --gas 90000000 --yes --log_level trace --trace

mythosd query multichain subchains --chain-id="leveln_1000-1" --from node0 --keyring-backend test --home ./testnet/node0/mythosd

mythosd query multichain subchain logos_10001-1 --chain-id="leveln_1000-1" --from node0 --keyring-backend test --home ./testnet/node0/mythosd

mythosd tx multichain register-subchain-validator logos_10001-1 /Users/user/dev/blockchain/wasmx-tests/validator_lvl.json --chain-id="leveln_1000-1" --from node0 --keyring-backend test --home ./testnet/node0/mythosd --fees 200000000000alvl --gas 90000000 --yes --log_level trace --trace

mythosd tx multichain init-subchain logos_10001-1 --chain-id="leveln_1000-1" --from node0 --keyring-backend test --home ./testnet/node0/mythosd --fees 200000000000alvl --gas 90000000 --yes --log_level trace --trace

```

## Protos

```
make proto-gen
```

### Note! we need to manually fix the custom proto file for `network`

* `custom.pulsar.go`: rename `networkv1` to `types`
* `custom.pb.go`: comment out `MsgExecuteAtomicTxRequest` definition and methods `Reset`, `String`, `ProtoMessage`, `Descriptor`, `GetTxs`, `GetLeaderChainId`, `GetSender`. Also comment out `QueryAtomicMultiChainRequest` def & methods
