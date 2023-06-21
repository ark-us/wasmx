package keeper_test

import (
	_ "embed"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

var (
	//go:embed testdata/wasmx/simple_storage.wasm
	wasmxSimpleStorage []byte
)

func (suite *KeeperTestSuite) TestWasmxSimpleStorage() {
	wasmbin := wasmxSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	logCount := strings.Count(res.GetLog(), `{"key":"type","value":"wasmx"}`)
	dataCount := strings.Count(res.GetLog(), `{"key":"data","value":"0x"}`)
	topicCount := strings.Count(res.GetLog(), `{"key":"topic","value":"0x68656c6c6f000000000000000000000000000000000000000000000000000000"}`)
	s.Require().Equal(1, logCount, res.GetLog())
	s.Require().Equal(1, dataCount, res.GetLog())
	s.Require().Equal(1, topicCount, res.GetLog())

	initvalue := "sammy"
	keybz := []byte("hello")
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, string(queryres))

	data = []byte(`{"get":{"key":"hello"}}`)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(string(qres), "sammy")
}
