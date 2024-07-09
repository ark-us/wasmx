# Examples

## levels subchains and cross-chain transactions

### decentralized subchains

* 4 validator mythos setup
* create 2 levels, each with 2 validators

#### create 2 level1 chains

```bash

mythosd testnet init-files --network.initial-chains=level0 --output-dir=$(pwd)/testnet --v=4 --keyring-backend=test --minimum-gas-prices="1000amyt" --same-machine=true --nocors --libp2p --min-level-validators=2 --enable-eid=false

# create gentx
mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl.json --chain-id="level0_1000-1" --from node0 --keyring-backend test --home ./testnet/node0/mythosd --fees 200000000000alvl --gas 90000000 --yes --log_level trace --trace

mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl2.json --chain-id="level0_1000-1" --from node1 --keyring-backend test --home ./testnet/node1/mythosd --fees 200000000000alvl --gas 90000000 --yes --log_level trace --trace

mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl3.json --chain-id="level0_1000-1" --from node2 --keyring-backend test --home ./testnet/node2/mythosd --fees 200000000000alvl --gas 90000000 --yes --log_level trace --trace

mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl4.json --chain-id="level0_1000-1" --from node3 --keyring-backend test --home ./testnet/node3/mythosd --fees 200000000000alvl --gas 90000000 --yes --log_level trace --trace

```

#### create one level 2 chain

* may need to change what validators join the level2 chain
* choose 2:

```bash
mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl.json --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --from node0 --keyring-backend test --home ./testnet/node0/mythosd --fees 200000000000alvl1 --gas 10000000 --yes --log_level trace --trace

mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl2.json --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --from node1 --keyring-backend test --home ./testnet/node1/mythosd --fees 200000000000alvl1 --gas 10000000 --yes --log_level trace --trace

mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl3.json --chain-id=level1_1_1003-1 --registry-chain-id=level0_1000-1 --from node2 --keyring-backend test --home ./testnet/node2/mythosd --fees 200000000000alvl1 --gas 10000000 --yes --log_level trace --trace

mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl4.json --chain-id=level1_1_1003-1 --registry-chain-id=level0_1000-1 --from node3 --keyring-backend test --home ./testnet/node3/mythosd --fees 200000000000alvl1 --gas 10000000 --yes --log_level trace --trace
```

#### cross-chain tx

```bash

# chain level1
mythosd tx wasmx store ./x/network/keeper/testdata/wasmx/simple_storage.wasm --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl1 --gas=9000000 --yes

# store_code, code_id

mythosd tx wasmx instantiate 53 '{"data":"{}"}' --label "simple_storage" --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl1 --gas=9000000 --yes

# instantiate
# level11m77kmufz9a6lzs6q7sw8fywkxc57c6p95vs7sh

# chain level2
mythosd tx wasmx store ./x/network/keeper/testdata/wasmx/crosschain.wasm --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl2 --gas=9000000 --yes

# store_code, code_id
# {"crosschain_contract":"metaregistry"}

mythosd tx wasmx instantiate 53 '{"crosschain_contract":"metaregistry"}' --label "crosschain" --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl2 --gas=9000000 --yes

# instantiate
# level21k6cmx9qqssmua806sagehx4a2dx3h9zfgn8nxc

# chain level1
mythosd tx wasmx execute level11m77kmufz9a6lzs6q7sw8fywkxc57c6p95vs7sh '{"set":{"key":"hello","value":"brian"}}' --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl1 --gas=9000000 --yes

mythosd query multichain call level11m77kmufz9a6lzs6q7sw8fywkxc57c6p95vs7sh '{"get":{"key":"hello"}}' --from node0 --keyring-backend test --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --home=./testnet/node0/mythosd

# atomic tx sent to chain level2

mythosd tx multichain atomic "/Users/user/dev/blockchain/wasmx-tests/atomictx.json" level2_2_1002-1,level1_1_1002-1 --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl2 --gas=9000000 --yes


mythosd tx multichain atomic "/Users/user/dev/blockchain/wasmx-tests/atomictx.json" level2_2_1002-1,level1_1_1002-1 --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node1 --keyring-backend=test --home=./testnet/node1/mythosd --fees=90000000000alvl2 --gas=9000000 --yes

```

* atomictx.json
```json
[{"msg_json":"{\"CrossChain\":{\"sender\":\"\",\"from\":\"\",\"to\":\"level11m77kmufz9a6lzs6q7sw8fywkxc57c6p95vs7sh\",\"msg\":\"eyJkYXRhIjoiZXlKelpYUWlPbnNpYTJWNUlqb2lhR1ZzYkc4aUxDSjJZV3gxWlNJNkluTmhiVzE1SW4xOSJ9\",\"funds\":[],\"dependencies\":[],\"from_chain_id\":\"\",\"to_chain_id\":\"level1_1_1002-1\",\"is_query\":false}}","contract": "level21k6cmx9qqssmua806sagehx4a2dx3h9zfgn8nxc", "multi_chain_id":"level2_2_1002-1"}]

```




## multiregistry multi-chain, cross-chain transaction

```bash

# 3 validator mythos setup
mythosd testnet init-files --chain-id=mythos_7000-14 --output-dir=$(pwd)/testnet --v=3 --keyring-backend=test --minimum-gas-prices="1000amyt" --same-machine=true --nocors --libp2p --min-level-validators=1 --enable-eid=false

# create 2 levels, each with 1 validator

mythosd testnet create-levels 2 1 --chain-id="mythos_7000-14" --keyring-backend test --home ./testnet

# chain level1
mythosd tx wasmx store ./x/network/keeper/testdata/wasmx/simple_storage.wasm --chain-id=chain0_1_1001-1 --registry-chain-id=mythos_7000-14 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl0 --gas=9000000 --yes

mythosd tx wasmx instantiate 50 '{"data":"{}"}' --label "simple_storage" --chain-id=chain0_1_1001-1 --registry-chain-id=mythos_7000-14 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl0 --gas=9000000 --yes

# chain016tsljek8g3av2rp8wnztga65xkn2dns8vdh4rl

# chain level2
mythosd tx wasmx store ./x/network/keeper/testdata/wasmx/crosschain.wasm --chain-id=leveln_2_1002-1 --registry-chain-id=mythos_7000-14 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl2 --gas=9000000 --yes

mythosd tx wasmx instantiate 50 '{"data":"{}"}' --label "crosschain" --chain-id=leveln_2_1002-1 --registry-chain-id=mythos_7000-14 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl2 --gas=9000000 --yes


# leveln1dpfdf0r42qttzgg6qnkkc7tyscx4t6r44fdmxf

# chain level1
mythosd tx wasmx execute chain016tsljek8g3av2rp8wnztga65xkn2dns8vdh4rl '{"set":{"key":"hello","value":"brian"}}' --chain-id=chain0_1_1001-1 --registry-chain-id=mythos_7000-14 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl0 --gas=9000000 --yes

mythosd query multichain call chain016tsljek8g3av2rp8wnztga65xkn2dns8vdh4rl '{"get":{"key":"hello"}}' --from node0 --keyring-backend test --chain-id=chain0_1_1001-1 --registry-chain-id=mythos_7000-14 --home=./testnet/node0/mythosd

# atomic tx sent to chain level2

mythosd tx multichain atomic "/Users/user/dev/blockchain/wasmx-tests/atomictx.json" leveln_2_1002-1,chain0_1_1001-1 --chain-id=leveln_2_1002-1 --registry-chain-id=mythos_7000-14 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl2 --gas=9000000 --yes

```

* atomictx.json
```json
[{"msg_json":"{\"CrossChain\":{\"sender\":\"\",\"from\":\"\",\"to\":\"chain016tsljek8g3av2rp8wnztga65xkn2dns8vdh4rl\",\"msg\":\"eyJkYXRhIjoiZXlKelpYUWlPbnNpYTJWNUlqb2lhR1ZzYkc4aUxDSjJZV3gxWlNJNkluTmhiVzE1SW4xOSJ9\",\"funds\":[],\"dependencies\":[],\"from_chain_id\":\"\",\"to_chain_id\":\"chain0_1_1001-1\",\"is_query\":false}}","contract": "leveln1dpfdf0r42qttzgg6qnkkc7tyscx4t6r44fdmxf", "multi_chain_id":"leveln_2_1002-1"}]

```
