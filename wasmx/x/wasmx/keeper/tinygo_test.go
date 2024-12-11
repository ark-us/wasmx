package keeper_test

import (
	_ "embed"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"wasmx/v1/x/wasmx/types"
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
	initBalance := sdkmath.NewInt(1000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "tinygoAdd", nil)

	data := []byte{}
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
}

func (suite *KeeperTestSuite) TestWasiTinygoSimpleStorage() {
	wasmbin := tinygoSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
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

func (suite *KeeperTestSuite) TestWasiTinygoSimpleStorageCall() {
	wasmbin := tinygoSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	depsPy := []string{types.INTERPRETER_PYTHON}

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, simpleStoragePy, depsPy)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`"123"`)}, "simpleContractPy", nil)

	codeId2 := appA.StoreCode(sender, wasmbin, nil)
	contractAddressWrap := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte("hello")}, "tinygoSimpleStorage", nil)

	key := []byte("pystore")
	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("123"), value)

	data := []byte(fmt.Sprintf(`{"wrapStore":["%s", "goodbye"]}`, contractAddress.String()))
	appA.ExecuteContract(sender, contractAddressWrap, types.WasmxExecutionMessage{Data: data}, nil, nil)

	value = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte(`goodbye`), value)

	resp := appA.WasmxQueryRaw(sender, contractAddressWrap, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"wrapLoad":["%s"]}`, contractAddress.String()))}, nil, nil)
	s.Require().Equal([]byte("goodbye23"), resp)
}
