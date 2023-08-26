package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"mythos/v1/x/wasmx/types"
	vmtypes "mythos/v1/x/wasmx/vm/types"
)

var (
	//go:embed testdata/python/simple_storage.py
	simpleStoragePy []byte

	//go:embed testdata/python/call.py
	callSimpleStoragePy []byte

	//go:embed testdata/python/blockchain.py
	blockchainPyInterpret []byte
)

// func (suite *KeeperTestSuite) TestWasiInterpreterPython() {
// 	sender := suite.GetRandomAccount()
// 	initBalance := sdk.NewInt(1_000_000_000_000_000_000)

// 	appA := s.GetAppContext(s.chainA)
// 	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
// 	suite.Commit()

// 	pyInterpreterAddress := types.AccAddressFromHex("0x0000000000000000000000000000000000000026")

// 	data := []byte(`
// from wasmx import storage_store
// storage_store("pystore", "222")
// `)
// 	appA.ExecuteContract(sender, pyInterpreterAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

// 	key := []byte("pystore")
// 	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), pyInterpreterAddress, key)
// 	s.Require().Equal([]byte("222"), value)
// }

func (suite *KeeperTestSuite) TestWasiInterpreterPythonSimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	deps := []string{types.INTERPRETER_PYTHON}
	codeId := appA.StoreCode(sender, simpleStoragePy, deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`"123"`)}, "SimpleContractPy", nil)

	key := []byte("pystore")
	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("123"), value)

	data := []byte(`{"store":["234"]}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	value = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("234"), value)

	data = []byte(`{"load":[]}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("234"), resp)
}

func (suite *KeeperTestSuite) TestWasiInterpreterPythonCallSimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	deps := []string{types.INTERPRETER_PYTHON}
	codeId := appA.StoreCode(sender, simpleStoragePy, deps)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`"123"`)}, "SimpleContractPy", nil)

	codeId2 := appA.StoreCode(sender, callSimpleStoragePy, deps)
	contractAddressCall := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "callSimpleStoragePy", nil)

	key := []byte("pystore")
	data := []byte(fmt.Sprintf(`{"store":["%s", "str111"]}`, contractAddress.String()))
	appA.ExecuteContract(sender, contractAddressCall, types.WasmxExecutionMessage{Data: data}, nil, nil)

	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("str111"), value)

	data = []byte(fmt.Sprintf(`{"load":["%s"]}`, contractAddress.String()))
	resp := appA.WasmxQueryRaw(sender, contractAddressCall, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("str11123"), resp)
}

func (suite *KeeperTestSuite) TestWasiInterpreterPythonBlockchain() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	deps := []string{types.INTERPRETER_PYTHON}
	codeId := appA.StoreCode(sender, blockchainPyInterpret, deps)

	data := []byte(`["pystore","hello"]`)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: data}, "blockchainPyInterpret", nil)

	data = []byte(`{"getEnv":[]}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().True(len(resp) > 0)
	// TODO check this
	// var env types.Env
	// err = json.Unmarshal(resp, &env)
	// s.Require().NoError(err)
	// s.Require().Equal(env.Chain.ChainIdFull, appA.Chain.ChainID)
	// s.Require().Equal(env.CurrentCall.Sender.String(), sender.Address.String())
	// s.Require().Equal(env.Contract.Address.String(), contractAddress.String())

	data = []byte(fmt.Sprintf(`{"getBalance":["%s"]}`, sender.Address.String()))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	balance, err := appA.App.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: sender.Address.String(), Denom: appA.Denom})
	s.Require().NoError(err)
	s.Require().Equal(balance.GetBalance().Amount.BigInt().FillBytes(make([]byte, 32)), resp)

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
	createdContractAddressStr := appA.GetContractAddressFromLog(txresp.GetLog())
	createdContractAddress := sdk.MustAccAddressFromBech32(createdContractAddressStr)
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
	createdContractAddressStr = appA.GetContractAddressFromLog(txresp.GetLog())
	createdContractAddress = sdk.MustAccAddressFromBech32(createdContractAddressStr)
	contractInfo = appA.App.WasmxKeeper.GetContractInfo(appA.Context(), createdContractAddress)
	s.Require().NotNil(contractInfo)

	data = []byte(`{"justError":[]}`)
	txresp = appA.ExecuteContractNoCheck(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 2000000, nil)
	s.Require().True(txresp.IsErr(), txresp.GetLog())
	s.Require().Contains(txresp.GetLog(), "failed to execute message", txresp.GetLog())
	// TODO
	// s.Require().Contains(txresp.GetLog(), "just error", txresp.GetLog())
	s.Commit()

	// TODO proper getBlockHash
	// data = []byte(`{"getBlockHash":[4]}`)
	// resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	// s.Require().Equal(resp, appA.Context().HeaderHash().Bytes())
}
