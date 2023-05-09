package keeper_test

import (
	_ "embed"

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

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	data := []byte(`{"method":"set","params":"{\"key\":\"hello\",\"value\":\"sammy\"}"}`)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Contains(res.GetLog(), "wasmx-log_ewasm_0")
	suite.Require().Contains(res.GetLog(), `{"key":"topic_0","value":"0x68656c6c6f000000000000000000000000000000000000000000000000000000"}`)

	initvalue := "sammy"
	keybz := []byte("hello")
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, string(queryres))

	data = []byte(`{"method":"get","params":"{\"key\":\"hello\"}"}`)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(string(qres), "sammy")
}
