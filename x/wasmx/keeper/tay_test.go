package keeper_test

import (
	_ "embed"
	"encoding/hex"
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

func (suite *KeeperTestSuite) TestInterpreterTayOpcodes() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	senderP := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderP, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.OpcodesTay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{}}`)}, "OpcodesTay", nil)

	data := []byte(`{"getChainId":{}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("level0_1000-1", string(resp))

	data = []byte(`{"base64dec":{"value":"aGVsbG9vYQ=="}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("hellooa", string(resp))

	data = []byte(`{"base64dec":{"value":"aGVsbG8="}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("hello", string(resp))

	data = []byte(`{"base64enc":{"value":"hellooa"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("aGVsbG9vYQ==", string(resp))

	data = []byte(`{"base64enc":{"value":"hello"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("aGVsbG8=", string(resp))

	data = []byte(`{"sha256":{"value":"aGVsbG8="}}`) // hello
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8", hex.EncodeToString(resp))

	data = []byte(`{"bufferstringify":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(`{"somebuf":"AQIDBAU="}`, string(resp))

	// data = []byte(`{"uintstringify":{}}`)
	// resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	// s.Require().Equal(`{"someu256":"AQIDBAU="}`, string(resp))

	data = []byte(`{"getCaller":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(senderP.String(), string(resp))

	data = []byte(`{"getAddress":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(contractAddress.String(), string(resp))

	data = []byte(`{"storageStore":{"key":"hello","value":"sammy"}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"storageLoad":{"key":"hello"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("sammy", string(resp))

	data = []byte(`{"storageStore":{"key":"hella","value":"bart"}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"storageLoadRange":{"startkey":"hella","endkey":"hello"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("sammy", string(resp))

	data = []byte(`{"storageLoadRangePairs":{"startkey":"hella","endkey":"hello"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("sammy", string(resp))
}

func (suite *KeeperTestSuite) TestInterpreterTayU256() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	senderP := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderP, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.OpcodesTay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{}}`)}, "OpcodesTay", nil)

	data := []byte(`{"add":{"a":"0x00000000000000000000000000000000000000000000000000000000000003e8","b": "0x00000000000000000000000000000000000000000000000000000000000003e8"}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("00000000000000000000000000000000000000000000000000000000000007d0", hex.EncodeToString(resp))

	data = []byte(`{"sub":{"a":"0x0000000000000000000000000000000000000000000000000000000000000003","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", hex.EncodeToString(resp))

	data = []byte(`{"mul":{"a":"0x0000000000000000000000000000000000000000000000000000000000000003","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", hex.EncodeToString(resp))

	data = []byte(`{"mul":{"a":"0x0000000000000000000000000ff0000000000000000000000000000000000003","b": "0x00000000000000000000000000000000000000000000000000000ff000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("00000000000000fe010000002fd00000000000000000000000002fd000000009", hex.EncodeToString(resp))

	data = []byte(`{"div":{"a":"0x0000000000000000000000000000000000000000000000000000000000000009","b": "0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", hex.EncodeToString(resp))

	data = []byte(`{"div":{"a":"0x000000000000000000000000000000000000000000000000fffffffffffffff4","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000007ffffffffffffffa", hex.EncodeToString(resp))

	data = []byte(`{"lt":{"a":"0x0000000000000000000000000000000000000000000000000000000000000003","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("false", string(resp))

	data = []byte(`{"gt":{"a":"0x0000000000000000000000000000000000000000000000000000000000000003","b": "0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("true", string(resp))

	data = []byte(`{"eq":{"a":"0x0000000000000000000000000022222222222222222222222222222222222222","b": "0x0000000000000000000000000022222222222222222222222222222222222222"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("true", string(resp))

	data = []byte(`{"mod":{"a":"0x0000000000000000000000000000000000000000000000000000000000000008","b": "0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", hex.EncodeToString(resp))

	data = []byte(`{"not":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd", hex.EncodeToString(resp))

	data = []byte(`{"addmod":{"a":"0x0000000000000000000000000000000000000000000000000000000000000005","b": "0x0000000000000000000000000000000000000000000000000000000000000003","c":"0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", hex.EncodeToString(resp))

	data = []byte(`{"mulmod":{"a":"0x0000000000000000000000000000000000000000000000000000000000000005","b": "0x0000000000000000000000000000000000000000000000000000000000000003","c":"0x0000000000000000000000000000000000000000000000000000000000000004"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", hex.EncodeToString(resp))

	data = []byte(`{"and":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0x000000000000000000000000000000000000000000000000000000000000000c"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x0000000000000000000000000000000000000000000000000000000000000000","b": "0xaa0000000000000000000000000000000000000000000000000000000000000c"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("aa0000000000000000000000000000000000000000000000000000000000000c", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x0000000000000000000000000000000000000000000000000000000000000110","b": "0xc84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x0000000000000000000000000000000000000000000000000000000000000080","b": "0xc84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x0000000000000000000000000000000000000000000000000000000000000100","b": "0xc84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(resp))

	data = []byte(`{"shr":{"a":"0x00000000000000000000000000000000000000000000000000000000000000aa","b": "0x983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("000000000000000000000000000000000000000000260f6fa86bffb7bb894f56", hex.EncodeToString(resp))

	data = []byte(`{"shl":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0x000000000000000000000000000000000000000000000000000000000000000c"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000030", hex.EncodeToString(resp))

	data = []byte(`{"shl":{"a":"0x00000000000000000000000000000000000000000000000000000000000000aa","b": "0x983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("ce5eb700bda45880968354000000000000000000000000000000000000000000", hex.EncodeToString(resp))

	data = []byte(`{"sar":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0x000000000000000000000000000000000000000000000000000000000000000c"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", hex.EncodeToString(resp))

	data = []byte(`{"sar":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd", hex.EncodeToString(resp))

	data = []byte(`{"exp":{"a":"0x0000000000000000000000000000000000000000000000000000000000000002","b": "0x0000000000000000000000000000000000000000000000000000000000000003"}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000008", hex.EncodeToString(resp))
}

// TODO byte
