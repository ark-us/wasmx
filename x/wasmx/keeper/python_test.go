package keeper_test

import (
	_ "embed"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

var (
	//go:embed testdata/python/simple_storage.py
	simpleStoragePy []byte
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
