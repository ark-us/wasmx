# Examples

## levels subchains and cross-chain transactions

### decentralized subchains

* 4 validator mythos setup
* create 2 levels, each with 2 validators

#### create 2 level1 chains

```bash

mythosd testnet init-files --network.initial-chains=level0 --output-dir=$(pwd)/testnet --v=4 --keyring-backend=test --minimum-gas-prices="1000amyt" --same-machine=true --nocors --libp2p --min-level-validators=2 --enable-eid=false

mythosd start --home=./testnet/node0/mythosd --same-machine-node-index=0
mythosd start --home=./testnet/node1/mythosd --same-machine-node-index=1
mythosd start --home=./testnet/node2/mythosd --same-machine-node-index=2
mythosd start --home=./testnet/node3/mythosd --same-machine-node-index=3

# create gentx
mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl.json --chain-id="level0_1000-1" --from node0 --keyring-backend test --home ./testnet/node0/mythosd --fees 200000000000alvl --gas 90000000 --yes --log_level trace --trace

mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl2.json --chain-id="level0_1000-1" --from node1 --keyring-backend test --home ./testnet/node1/mythosd --fees 200000000000alvl --gas 90000000 --yes --log_level trace --trace --node tcp://localhost:26660

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
mythosd tx wasmx store ./x/network/keeper/testdata/wasmx/simple_storage.wasm --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl1 --gas=9000000 --node tcp://localhost:26671 --yes

# store_code, code_id

mythosd tx wasmx instantiate 53 '{"crosschain_contract":"metaregistry"}' --label "simple_storage" --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl1 --gas=9000000 --node tcp://localhost:26671 --yes

# instantiate
# level11t67sqkunyp5etk8khw3yhdxj7ur3dzf2l8dade

# chain level2
mythosd tx wasmx store ./x/network/keeper/testdata/wasmx/crosschain.wasm --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node1 --keyring-backend=test --home=./testnet/node1/mythosd --fees=90000000000alvl2 --gas=9000000 --node tcp://localhost:26772 --yes

# store_code, code_id
# {"crosschain_contract":"metaregistry"}

mythosd tx wasmx instantiate 53 '{"crosschain_contract":"metaregistry"}' --label "crosschain" --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node1 --keyring-backend=test --home=./testnet/node1/mythosd --fees=90000000000alvl2 --gas=9000000 --node tcp://localhost:26772 --yes

# instantiate
# level21umhm4l9nxfhahqshmyd8rlmzwkt2xz36yhycjq

# chain level1
mythosd tx wasmx execute level11t67sqkunyp5etk8khw3yhdxj7ur3dzf2l8dade '{"set":{"key":"hello","value":"brian"}}' --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl1 --gas=9000000 --node tcp://localhost:26671 --yes

mythosd query multichain call level11t67sqkunyp5etk8khw3yhdxj7ur3dzf2l8dade '{"get":{"key":"hello"}}' --from node0 --keyring-backend test --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --home=./testnet/node0/mythosd --node tcp://localhost:26671

# atomic tx sent to chain level2

mythosd tx multichain atomic "/Users/user/dev/blockchain/wasmx-tests/atomictx.json" level2_2_1002-1,level1_1_1002-1 --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl2 --gas=9000000 --node tcp://localhost:26672 --yes


mythosd tx multichain atomic "/Users/user/dev/blockchain/wasmx-tests/atomictx.json" level2_2_1002-1,level1_1_1002-1 --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node1 --keyring-backend=test --home=./testnet/node1/mythosd --fees=90000000000alvl2 --gas=9000000 --node tcp://localhost:26772 --yes

```

* atomictx.json
```json
[{"msg_json":"{\"CrossChain\":{\"sender\":\"\",\"from\":\"\",\"to\":\"level11t67sqkunyp5etk8khw3yhdxj7ur3dzf2l8dade\",\"msg\":\"eyJkYXRhIjoiZXlKelpYUWlPbnNpYTJWNUlqb2lhR1ZzYkc4aUxDSjJZV3gxWlNJNkluTmhiVzE1SW4xOSJ9\",\"funds\":[],\"dependencies\":[],\"from_chain_id\":\"\",\"to_chain_id\":\"level1_1_1002-1\",\"is_query\":false}}","contract": "level21umhm4l9nxfhahqshmyd8rlmzwkt2xz36yhycjq", "multi_chain_id":"level2_2_1002-1"}]

```

#### reverse order cross-chain tx

```bash

# chain level1
mythosd tx wasmx store ./x/network/keeper/testdata/wasmx/crosschain.wasm --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl1 --gas=9000000 --node tcp://localhost:26671 --yes

# store_code, code_id

mythosd tx wasmx instantiate 53 '{"crosschain_contract":"metaregistry"}' --label "crosschain" --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --from=node0 --keyring-backend=test --home=./testnet/node0/mythosd --fees=90000000000alvl1 --gas=9000000 --node tcp://localhost:26671 --yes

# instantiate
# level11ts5qjpfvtfh8cer2xqqz343t8wsp5rqzkkzjv3

# chain level2
mythosd tx wasmx store ./x/network/keeper/testdata/wasmx/simple_storage.wasm --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node1 --keyring-backend=test --home=./testnet/node1/mythosd --fees=90000000000alvl2 --gas=9000000 --node tcp://localhost:26772 --yes

# store_code, code_id

mythosd tx wasmx instantiate 53 '{"data":"{}"}' --label "simple_storage" --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node1 --keyring-backend=test --home=./testnet/node1/mythosd --fees=90000000000alvl2 --gas=9000000 --node tcp://localhost:26772 --yes

# instantiate
# level219hp2wzvx6yctd2aexp8hghpnll0fhsnewqqp7m

# chain level2
mythosd tx wasmx execute level219hp2wzvx6yctd2aexp8hghpnll0fhsnewqqp7m '{"set":{"key":"hello","value":"brian"}}' --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node1 --keyring-backend=test --home=./testnet/node1/mythosd --fees=90000000000alvl2 --gas=9000000 --node tcp://localhost:26772 --yes

mythosd query wasmx call level219hp2wzvx6yctd2aexp8hghpnll0fhsnewqqp7m '{"get":{"key":"hello"}}' --from node1 --keyring-backend test --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --home=./testnet/node1/mythosd --node tcp://localhost:26772

# atomic tx sent to chain level2

mythosd tx multichain atomic "/Users/user/dev/blockchain/wasmx-tests/atomictx.json" level2_2_1002-1,level1_1_1002-1 --chain-id=level2_2_1002-1 --registry-chain-id=level0_1000-1 --from=node1 --keyring-backend=test --home=./testnet/node1/mythosd --fees=90000000000alvl1 --gas=9000000 --node tcp://localhost:26772 --yes


```

* atomictx.json
```json
[{"msg_json":"{\"CrossChain\":{\"sender\":\"\",\"from\":\"\",\"to\":\"level219hp2wzvx6yctd2aexp8hghpnll0fhsnewqqp7m\",\"msg\":\"eyJkYXRhIjoiZXlKelpYUWlPbnNpYTJWNUlqb2lhR1ZzYkc4aUxDSjJZV3gxWlNJNkluTmhiVzE1SW4xOSJ9\",\"funds\":[],\"dependencies\":[],\"from_chain_id\":\"\",\"to_chain_id\":\"level2_2_1002-1\",\"is_query\":false}}","contract": "level11ts5qjpfvtfh8cer2xqqz343t8wsp5rqzkkzjv3", "multi_chain_id":"level1_1_1002-1"}]

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

## statesync mythos & create validator

```bash

mythosd testnet init-files --network.initial-chains=mythos --output-dir=$(pwd)/testnet --chain-id=mythos_7000-14 --v=1 --keyring-backend=test --minimum-gas-prices="1000amyt" --same-machine=true --nocors --libp2p --min-level-validators=2 --enable-eid=false

HOMEMAIN=./testnet/node0/mythosd
sed -i.bak -E "s|^(snapshot-interval[[:space:]]+=[[:space:]]+).*$|\110|" $HOMEMAIN/config/app.toml

mythosd testnet add-node 1 "mythos1y23ljx5p0r0c6f9umttr73uuetly338n7pyz9h@/ip4/127.0.0.1/tcp/5001/p2p/12D3KooWCGtqm7YufRVjfrq591tTg6RnXLTXCZGFuSoQvQ3MuG4N" --network.initial-chains=mythos --chain-id=mythos_7000-14 --output-dir=$(pwd)/testnet --keyring-backend=test --minimum-gas-prices="1000amyt" --same-machine=true --nocors --libp2p

mythosd start --home=./testnet/node0/mythosd --same-machine-node-index=0

mythosd tx cosmosmod bank send node0 mythos1pdya3tu98y6grwr28vsfaw6jd9prdzrza88m5y 120000000000000000000amyt --keyring-backend test --home ./testnet/node0/mythosd --fees 200000000000amyt --gas 900000 --chain-id=mythos_7000-14 --yes

mythosd tendermint unsafe-reset-all --home=./testnet/node1/mythosd

HOMEMAIN=./testnet/node1/mythosd
RPC="http://localhost:26657"
RECENT_HEIGHT=$(curl -s $RPC/block | jq -r .result.block.header.height)
TRUST_HEIGHT=$((RECENT_HEIGHT - 1))
TRUST_HASH=$(curl -s "$RPC/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash)
sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; \
s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"$RPC,$RPC\"| ; \
s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$TRUST_HEIGHT| ; \
s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"|" $HOMEMAIN/config/config.toml


mythosd start --home=./testnet/node1/mythosd --same-machine-node-index=1

# after sync, disable statesync ./config/config.toml
sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1false|" $HOMEMAIN/config/config.toml

# change validator public key in validator.json
mythosd tendermint show-validator --home ./testnet/node1/mythosd

mythosd tx cosmosmod staking create-validator /Users/user/dev/blockchain/wasmx-tests/validator.json --from node1 --chain-id=mythos_7000-14 --keyring-backend=test --home=./testnet/node1/mythosd --fees 200000000000000amyt --gas auto --gas-adjustment 1.4 --memo="mythos1pdya3tu98y6grwr28vsfaw6jd9prdzrza88m5y@/ip4/127.0.0.1/tcp/5003/p2p/12D3KooWN8nCZMKQWQ88BZRYLJTHiaBmCBrCZqHEE9d4U4Mekotv" --node tcp://127.0.0.1:26658 --yes

```


## statesync subchains

* similar to the statesync & create validator example, but for mythos,level0

```sh
mythosd testnet init-files --network.initial-chains=mythos,level0 --output-dir=$(pwd)/testnet --chain-id=mythos_7000-14 --v=1 --keyring-backend=test --minimum-gas-prices="1000amyt" --same-machine=true --nocors --libp2p --min-level-validators=2 --enable-eid=false

HOMEMAIN=./testnet/node0/mythosd
sed -i.bak -E "s|^(snapshot-interval[[:space:]]+=[[:space:]]+).*$|\110|" $HOMEMAIN/config/app.toml

mythosd testnet add-node 1 "mythos19f332uvw38tjyrukhfwwv4kxsmxfpcnscgmqtn@/ip4/127.0.0.1/tcp/5001/p2p/12D3KooWAWZ6M3FM34R3Fkx1za4WxUcRry2gmgxGoiVEE594oZXy" --network.initial-chains=mythos,level0 --chain-id=mythos_7000-14 --output-dir=$(pwd)/testnet --keyring-backend=test --minimum-gas-prices="1000amyt" --same-machine=true --nocors --libp2p

mythosd start --home=./testnet/node0/mythosd --same-machine-node-index=0

mythosd tx cosmosmod bank send node0 mythos1w8p2qrncjjkm3mt7y9dmv4myz73dt2zn4l47tl 120000000000000000000amyt --keyring-backend test --home ./testnet/node0/mythosd --fees 200000000000amyt --gas 900000 --chain-id=mythos_7000-14 --yes

mythosd tendermint unsafe-reset-all --home=./testnet/node1/mythosd


HOMEMAIN=./testnet/node1/mythosd
RPC="http://localhost:26657"
RECENT_HEIGHT=$(curl -s $RPC/block | jq -r .result.block.header.height)
TRUST_HEIGHT=$((RECENT_HEIGHT - 1))
TRUST_HASH=$(curl -s "$RPC/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash)
sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; \
s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"$RPC,$RPC\"| ; \
s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$TRUST_HEIGHT| ; \
s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"|" $HOMEMAIN/config/config.toml

mythosd start --home=./testnet/node1/mythosd --same-machine-node-index=1

# after sync, disable statesync ./config/config.toml
sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1false|" $HOMEMAIN/config/config.toml

# change validator public key in validator.json
mythosd tendermint show-validator --home ./testnet/node1/mythosd

mythosd tx cosmosmod staking create-validator /Users/user/dev/blockchain/wasmx-tests/validator.json --from node1 --chain-id=mythos_7000-14 --keyring-backend=test --home=./testnet/node1/mythosd --fees 200000000000000amyt --gas auto --gas-adjustment 1.4 --memo="mythos1w8p2qrncjjkm3mt7y9dmv4myz73dt2zn4l47tl@/ip4/127.0.0.1/tcp/5003/p2p/12D3KooWHRogAgkV2NZDvMkuGpQqUHp7ZnMz7HjeodsjbyP7HZ4E" --node tcp://127.0.0.1:26659 --yes

```

### create a subchain  level1 with the first 2 nodes

* follow tutorial "create 2 level1 chains" - create gentx

```bash
mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl.json --chain-id="level0_1000-1" --from node0 --keyring-backend test --home ./testnet/node0/mythosd --fees 200000000000alvl --gas 90000000 --yes --log_level trace --trace

mythosd tx multichain register-subchain-gentx /Users/user/dev/blockchain/wasmx-tests/validator_lvl2.json --chain-id="level0_1000-1" --from node1 --keyring-backend test --home ./testnet/node1/mythosd --fees 200000000000alvl --gas 90000000 --yes --log_level trace --trace --node tcp://localhost:26660
```

### add a 3rd node

```bash
mythosd testnet add-node 2 "mythos19f332uvw38tjyrukhfwwv4kxsmxfpcnscgmqtn@/ip4/127.0.0.1/tcp/5001/p2p/12D3KooWAWZ6M3FM34R3Fkx1za4WxUcRry2gmgxGoiVEE594oZXy" --network.initial-chains=mythos,level0 --chain-id=mythos_7000-14 --output-dir=$(pwd)/testnet --keyring-backend=test --minimum-gas-prices="1000amyt" --same-machine=true --nocors --libp2p

mythosd tendermint unsafe-reset-all --home=./testnet/node2/mythosd

HOMEMAIN=./testnet/node2/mythosd
RPC="http://localhost:26657"
RECENT_HEIGHT=$(curl -s $RPC/block | jq -r .result.block.header.height)
TRUST_HEIGHT=$((RECENT_HEIGHT - 2))
TRUST_HASH=$(curl -s "$RPC/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash)
sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; \
s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"$RPC,$RPC\"| ; \
s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$TRUST_HEIGHT| ; \
s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"|" $HOMEMAIN/config/config.toml

mythosd start --home=./testnet/node2/mythosd --same-machine-node-index=2

# after sync, disable statesync ./config/config.toml
sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1false|" $HOMEMAIN/config/config.toml

mythosd tx cosmosmod bank send node0 mythos17nknqm99dmukyhf4e2tyxcjgxjk5cnxh5g7rng 120000000000000000000amyt --keyring-backend test --home ./testnet/node0/mythosd --fees 200000000000amyt --gas 900000 --chain-id=mythos_7000-14 --yes

# change validator public key in validator.json
mythosd tendermint show-validator --home ./testnet/node2/mythosd

mythosd tx cosmosmod staking create-validator /Users/user/dev/blockchain/wasmx-tests/validator.json --from node2 --chain-id=mythos_7000-14 --keyring-backend=test --home=./testnet/node2/mythosd --fees 200000000000000amyt --gas auto --gas-adjustment 1.4 --memo="mythos17nknqm99dmukyhf4e2tyxcjgxjk5cnxh5g7rng@/ip4/127.0.0.1/tcp/5005/p2p/12D3KooWPoLGpkrC9nMUC2j8a7cj7qxHRDVJhRSXSn9NVovgMGUs" --node tcp://127.0.0.1:26661 --yes

```

Now we have 3 validators on Mythos. And a level1 chain. We will sync the level1 chain on our 3rd node.

* replace peer_address with a level1 peer address; we can use the level0 bech32 address, it will be converted.
* rpc is also for level1 peer

```sh

RPC="http://localhost:26771"
RECENT_HEIGHT=$(curl -s $RPC/block | jq -r .result.block.header.height) && echo $RECENT_HEIGHT
TRUST_HEIGHT=$((RECENT_HEIGHT - 1)) && echo $TRUST_HEIGHT
TRUST_HASH=$(curl -s "$RPC/block?height=$TRUST_HEIGHT" | jq -r .result.block_id.hash) && echo $TRUST_HASH

mythosd query multichain call level01qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqztgdv8vl '{"StartStateSync":{"chain_id":"level1_1_1002-1","verification_contract_address":"level01qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzgx7e3zq","peer_address":"level019f332uvw38tjyrukhfwwv4kxsmxfpcnsjatum0@/ip4/127.0.0.1/tcp/5002/p2p/12D3KooWAWZ6M3FM34R3Fkx1za4WxUcRry2gmgxGoiVEE594oZXy","rpc":"tcp://127.0.0.1:26771","chain_config":{"Bech32PrefixAccAddr":"level1","Bech32PrefixAccPub":"level1","Bech32PrefixValAddr":"level1","Bech32PrefixValPub":"level1","Bech32PrefixConsAddr":"level1","Bech32PrefixConsPub":"level1","Name":"level1","HumanCoinUnit":"lvl1","BaseDenom":"alvl1","DenomUnit":"lvl1","BaseDenomUnit":18,"BondBaseDenom":"aslvl1","BondDenom":"slvl1"},"statesync_config":{"rpc_servers":["http://localhost:26771","http://localhost:26771"],"trust_period":36000000,"trust_height":146,"trust_hash":"5B4C7B54C27810114886A58F401FE8DC066D110562C5CA3548C971D2A9FC1298","enable":true,"temp_dir":"","discovery_time":15000,"chunk_request_timeout":10000,"chunk_fetchers":4}}}' --chain-id=level0_1000-1 --from=node2 --keyring-backend=test --home=./testnet/node2/mythosd --node tcp://localhost:26662

```

# mythosd tx multichain reset-subchain-data --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --home=./testnet/node2/mythosd

* create validator on level1 for the 3rd node:

```bash

mythosd tx cosmosmod bank send node0 level017nknqm99dmukyhf4e2tyxcjgxjk5cnxh7awlr5 1200000000000000000alvl1 --keyring-backend test --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --home ./testnet/node0/mythosd --fees 200000000000alvl1 --gas 900000 --node tcp://localhost:26771 --yes

# change validator public key in validator.json
mythosd tendermint show-validator --home ./testnet/node2/mythosd

mythosd tx cosmosmod staking create-validator /Users/user/dev/blockchain/wasmx-tests/validator_create_level1.json --from node2 --chain-id=level1_1_1002-1 --registry-chain-id=level0_1000-1 --keyring-backend=test --home=./testnet/node2/mythosd --fees 200000000000000alvl1 --gas auto --gas-adjustment 1.4 --memo="level117nknqm99dmukyhf4e2tyxcjgxjk5cnxh4re5wl@/ip4/127.0.0.1/tcp/5006/p2p/12D3KooWPoLGpkrC9nMUC2j8a7cj7qxHRDVJhRSXSn9NVovgMGUs" --node tcp://127.0.0.1:27171 --yes

```
