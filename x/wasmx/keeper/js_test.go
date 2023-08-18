package keeper_test

import (
	_ "embed"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

var (
	//go:embed testdata/js/javy/simple_storage.wasm
	jsSimpleStorage []byte

	//go:embed testdata/js/simple_storage_interpret.js
	simpleStorageJsInterpret []byte
)

func (suite *KeeperTestSuite) TestJavyJsSimpleStorage() {
	wasmbin := jsSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "jsSimpleStorage", nil)

	data := []byte{}
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
}

func (suite *KeeperTestSuite) TestInterpreterJsSimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	deps := []string{types.INTERPRETER_JS}
	codeId := appA.StoreCode(sender, simpleStorageJsInterpret, deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`"hello"`)}, "simpleStorageJsInterpret", nil)

	key := []byte("jsstore")
	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("hello"), value)

	data := []byte(`{"store":["goodbye"]}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	value = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("goodbye"), value)

	data = []byte(`{"load":[]}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("goodbye"), resp)

	data = []byte(`{"store":["I say goodbye, you say hello"]}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	value = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("I say goodbye, you say hello"), value)

	data = []byte(`{"store":["you say goodbye, I say hello"]}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	value = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("you say goodbye, I say hello"), value)
}
