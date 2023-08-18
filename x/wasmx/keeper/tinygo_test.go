package keeper_test

import (
	_ "embed"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

var (
	//go:embed testdata/tinygo/add.wasm
	tinygoAdd []byte

	//go:embed testdata/tinygo/simple_storage.wasm
	tinygoSimpleStorage []byte
)

func (suite *KeeperTestSuite) TestWasiTinygoAdd() {
	wasmbin := tinygoAdd
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "tinygoAdd", nil)

	data := []byte{}
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
}

func (suite *KeeperTestSuite) TestWasiTinygoSimpleStorage() {
	wasmbin := tinygoSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte("hello")}, "tinygoSimpleStorage", nil)

	key := []byte("storagekey")
	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("hello"), value)

	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(`{"store":["goodbye"]}`)}, nil, nil)

	value = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte(`goodbye`), value)

	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(`{"load":[]}`)}, nil, nil)
	s.Require().Equal([]byte("goodbye"), resp)

	appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(`{"store":["hello"]}`)}, nil, nil)

	value = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("goodbye"), value)
}
