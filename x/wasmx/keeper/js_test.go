package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testdata "mythos/v1/x/wasmx/keeper/testdata/classic"
	"mythos/v1/x/wasmx/types"
)

var (
	//go:embed testdata/js/javy/simple_storage.wasm
	jsSimpleStorage []byte

	//go:embed testdata/js/simple_storage.js
	simpleStorageJsInterpret []byte

	//go:embed testdata/js/call.js
	callSimpleStorageJsInterpret []byte

	//go:embed testdata/js/call_evm.js
	callEvmSimpleStorageJsInterpret []byte
)

func (suite *KeeperTestSuite) TestWasiJavyJsSimpleStorage() {
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

func (suite *KeeperTestSuite) TestWasiInterpreterJsSimpleStorage() {
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

func (suite *KeeperTestSuite) TestWasiInterpreterJsCallSimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	deps := []string{types.INTERPRETER_JS}
	codeId := appA.StoreCode(sender, simpleStorageJsInterpret, deps)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`"123"`)}, "simpleContractJs", nil)

	codeId2 := appA.StoreCode(sender, callSimpleStorageJsInterpret, deps)
	contractAddressCall := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "CallSimpleContractJs", nil)

	key := []byte("jsstore")
	data := []byte(fmt.Sprintf(`{"store":["%s", "str1"]}`, contractAddress.String()))
	appA.ExecuteContract(sender, contractAddressCall, types.WasmxExecutionMessage{Data: data}, nil, nil)

	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("str1"), value)

	data = []byte(fmt.Sprintf(`{"load":["%s"]}`, contractAddress.String()))
	resp := appA.WasmxQueryRaw(sender, contractAddressCall, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("str1"), resp)
}

func (suite *KeeperTestSuite) TestWasiInterpreterJsCallPySimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	depsJs := []string{types.INTERPRETER_JS}
	depsPy := []string{types.INTERPRETER_PYTHON}
	codeId := appA.StoreCode(sender, simpleStoragePy, depsPy)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`"123"`)}, "simpleContractPy", nil)

	codeId2 := appA.StoreCode(sender, callSimpleStorageJsInterpret, depsJs)
	contractAddressCall := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "CallSimpleContractJs", nil)

	key := []byte("pystore")
	data := []byte(fmt.Sprintf(`{"store":["%s", "str12"]}`, contractAddress.String()))
	appA.ExecuteContract(sender, contractAddressCall, types.WasmxExecutionMessage{Data: data}, nil, nil)

	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("str12"), value)

	data = []byte(fmt.Sprintf(`{"load":["%s"]}`, contractAddress.String()))
	resp := appA.WasmxQueryRaw(sender, contractAddressCall, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("str12"), resp)
}

func (suite *KeeperTestSuite) TestWasiInterpreterJsCallEvmSimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)
	depsJs := []string{types.INTERPRETER_JS}

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
	s.Require().NoError(err)
	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	initvaluebz, err := hex.DecodeString(initvalue)
	s.Require().NoError(err)
	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage")

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

	codeId2 := appA.StoreCode(sender, callEvmSimpleStorageJsInterpret, depsJs)
	contractAddressCall := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "CallSimpleContractJs", nil)

	data := []byte(fmt.Sprintf(`{"store":["%s", "str12"]}`, contractAddress.String()))
	appA.ExecuteContract(sender, contractAddressCall, types.WasmxExecutionMessage{Data: data}, nil, nil)

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000007", hex.EncodeToString(queryres))

	data = []byte(fmt.Sprintf(`{"load":["%s"]}`, contractAddress.String()))
	resp := appA.WasmxQueryRaw(sender, contractAddressCall, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000007", hex.EncodeToString(resp))
}
