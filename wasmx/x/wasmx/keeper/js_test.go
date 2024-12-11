package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	testdata "mythos/v1/x/wasmx/keeper/testdata/classic"
	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm/types"
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

	//go:embed testdata/js/blockchain.js
	blockchainJsInterpret []byte
)

func (suite *KeeperTestSuite) TestWasiJavyJsSimpleStorage() {
	wasmbin := jsSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "jsSimpleStorage", nil)

	data := []byte{}
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
}

func (suite *KeeperTestSuite) TestWasiInterpreterJsSimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
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
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
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
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
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
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	depsJs := []string{types.INTERPRETER_JS}

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
	s.Require().NoError(err)
	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	initvaluebz, err := hex.DecodeString(initvalue)
	s.Require().NoError(err)
	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage", nil)

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

func (suite *KeeperTestSuite) TestWasiInterpreterJsBlockchain() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	deps := []string{types.INTERPRETER_JS}
	codeId := appA.StoreCode(sender, blockchainJsInterpret, deps)

	data := []byte(`["jsstore","hello"]`)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: data}, "blockchainJsInterpret", nil)

	data = []byte(`{"getEnv":[]}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().True(len(resp) > 0)
	// TODO check env
	// var env types.Env
	// err = json.Unmarshal(resp, &env)
	// s.Require().NoError(err)
	// s.Require().Equal(env.Chain.ChainIdFull, appA.Chain.ChainID)
	// s.Require().Equal(appA.MustAccAddressToString(env.CurrentCall.Sender), appA.MustAccAddressToString(sender.Address))
	// s.Require().Equal(appA.MustAccAddressToString(env.Contract.Address), contractAddress.String())

	data = []byte(fmt.Sprintf(`{"getBalance":["%s"]}`, appA.MustAccAddressToString(sender.Address)))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	balance := appA.App.BankKeeper.GetBalance(appA.Context(), sender.Address, appA.Chain.Config.BaseDenom)
	s.Require().Equal(balance.Amount.BigInt().FillBytes(make([]byte, 32)), resp)

	data = []byte(fmt.Sprintf(`{"getAccount":["%s"]}`, contractAddress.String()))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().True(len(resp) > 0)
	// TODO check this
	// var acc types.EnvContractInfo
	// err = json.Unmarshal(resp, &acc)

	data = []byte(`{"keccak256":["somedata"]}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal("fb763c3da6141a6a1464a68583e30d9a77bb999b1f1c491992dcfac7738ecfb4", hex.EncodeToString(resp))

	// TODO propagate the error properly
	// initMsg := types.WasmxExecutionMessage{Data: []byte(`"hello"`)}
	initMsg := types.WasmxExecutionMessage{Data: []byte(`["jsstore","hello"]`)}
	initMsgBz, err := json.Marshal(initMsg)
	s.Require().NoError(err)
	data = []byte(fmt.Sprintf(`{"instantiateAccount":[%d,"%s","%s"]}`, codeId, hex.EncodeToString(initMsgBz), "0000000000000000000000000000000000000000000000000000000000000000"))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(32, len(resp))
	expectedContractAddress := sdk.AccAddress(vmtypes.CleanupAddress(resp))
	contractInfo := appA.App.WasmxKeeper.GetContractInfo(appA.Context(), expectedContractAddress)
	s.Require().Nil(contractInfo)

	// we actually execute the contract creation
	txresp := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	createdContractAddressStr := appA.GetContractAddressFromEvents(txresp.GetEvents())
	createdContractAddress, err := appA.AddressStringToAccAddress(createdContractAddressStr)
	s.Require().NoError(err)
	contractInfo = appA.App.WasmxKeeper.GetContractInfo(appA.Context(), createdContractAddress)
	s.Require().NotNil(contractInfo)

	// instantiate2
	data = []byte(fmt.Sprintf(`{"instantiateAccount2":[%d, "%s", "%s","%s"]}`, codeId, "0000000000000000000000000000000000000000000000000000000000000011", hex.EncodeToString(initMsgBz), "0000000000000000000000000000000000000000000000000000000000000000"))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(32, len(resp))
	expectedContractAddress = sdk.AccAddress(vmtypes.CleanupAddress(resp))
	contractInfo = appA.App.WasmxKeeper.GetContractInfo(appA.Context(), expectedContractAddress)
	s.Require().Nil(contractInfo)

	// we actually execute the contract creation
	txresp = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	createdContractAddressStr = appA.GetContractAddressFromEvents(txresp.GetEvents())
	createdContractAddress, err = appA.AddressStringToAccAddress(createdContractAddressStr)
	s.Require().NoError(err)
	contractInfo = appA.App.WasmxKeeper.GetContractInfo(appA.Context(), createdContractAddress)
	s.Require().NotNil(contractInfo)

	data = []byte(`{"justError":[]}`)
	txresp, err = appA.ExecuteContractNoCheck(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 2000000, nil)
	s.Require().NoError(err)
	s.Require().True(txresp.IsErr(), txresp.GetLog())
	s.Require().Contains(txresp.GetLog(), "failed to execute message", txresp.GetLog())
	s.Require().Contains(txresp.GetLog(), "just error", txresp.GetLog())
	s.Commit()

	// TODO proper getBlockHash
	// data = []byte(`{"getBlockHash":[4]}`)
	// resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	// s.Require().Equal(resp, appA.Context().HeaderHash().Bytes())
}
