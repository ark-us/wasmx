package keeper_test

import (
	_ "embed"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	testdata "mythos/v1/x/wasmx/keeper/testdata/taylor"
	"mythos/v1/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestInterpreterTaySimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.SimpleStorageTay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{}}`)}, "SimpleContractTay", nil)

	key := []byte("hello")
	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("sammy"), value)

	data = []byte(`{"get":{"key":"hello"}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("sammy"), resp)
}

func (suite *KeeperTestSuite) TestInterpreterTayERC20() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.ERC20Tay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{"name":"Token","symbol":"TKN","decimals":8}}`)}, "ERC20Tay", nil)

	data := []byte(`{"getName":{}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte(`{"name":"Token"}`), resp)

	data = []byte(`{"getSymbol":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte(`{"symbol":"TKN"}`), resp)

	data = []byte(`{"getDecimals":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte(`{"decimals":8}`), resp)

	data = []byte(`{"mint":{}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, senderPrefixed.String()))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte(`{"balance":"0x00000000000000000000000000000000000000000000000000000000000003e8"}`), resp)

	data = []byte(`{"totalSupply":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte(`{"supply":"0x00000000000000000000000000000000000000000000000000000000000003e8"}`), resp)

	data = []byte(`{"transfer":{"receiver":"level01w9g4qsddjlfajfl7vehgagvysrzec9flwknee2","value":"0x20"}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"balanceOf":{"owner":"level01w9g4qsddjlfajfl7vehgagvysrzec9flwknee2"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte(`{"balance":"0x0000000000000000000000000000000000000000000000000000000000000020"}`), resp)
}

func (suite *KeeperTestSuite) TestInterpreterTayJson() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.JsonTestTay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{}}`)}, "JsonTestTay", nil)

	data := []byte(`{"add":{"set":{"key":"hello","value":"sammy"}}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`{"additional":{"obj":{"a":8,"b":9},"another":4},"set":{"key":"hello","value":"sammy"}}`, string(resp))

	data = []byte(`{"identity":[1, 2, 3]}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`[1,2,3]`, string(resp))

	data = []byte(`{"identity":{"set":{"key":"hello","value":"sammy", "somearr": [6, "aaa", 5], "somev": true}}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`{"set":{"key":"hello","value":"sammy","somearr":[6,"aaa",5],"somev":true}}`, string(resp))

	data = []byte(`{"identity":"ana"}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`"ana"`, string(resp))

	data = []byte(`{"identity":5}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`5`, string(resp))

	data = []byte(`{"identity":true}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`true`, string(resp))
}
