package keeper_test

import (
	_ "embed"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	rusttest "github.com/loredanacirstea/mythos-tests/testdata/rust"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

func (suite *KeeperTestSuite) TestWasmxRustSimpleStorage() {
	SkipFixmeTests(suite.T(), "TestWasmxRustSimpleStorage")
	wasmbin := rusttest.RustSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"key":"hello","value":"bill"}`)}, "simpleStorage", nil)

	keybz := []byte("hello")

	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("bill", string(queryres))

	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	wasmlogs := appA.GetWasmxEvents(res.GetEvents())
	topicLogs := appA.GetWasmxEventsByAttribute(wasmlogs, "topic", "0x68656c6c6f000000000000000000000000000000000000000000000000000000")
	dataLogs := appA.GetWasmxEventsByAttribute(topicLogs, "data", "0x")
	s.Require().GreaterOrEqual(len(wasmlogs), 1, res.GetEvents())
	s.Require().Equal(1, len(topicLogs), res.GetEvents())
	s.Require().Equal(1, len(dataLogs))

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("sammy", string(queryres))

	data = []byte(`{"get":{"key":"hello"}}`)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(string(qres), "sammy")
}
