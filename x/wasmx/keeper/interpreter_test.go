package keeper_test

import (
	_ "embed"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testdata "mythos/v1/x/wasmx/keeper/testdata/classic"
	"mythos/v1/x/wasmx/types"
)

var (
	//go:embed testdata/interpreter/simple_storage.wasm
	iSimpleStorage []byte

	//go:embed testdata/interpreter/contract.wasm
	iContract []byte

	//go:embed testdata/interpreter/evm_interpreter.wasm
	evmInterpreter []byte
)

func (suite *KeeperTestSuite) TestInterpreterEvm() {
	wasmbin := evmInterpreter
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCodeEwasmEnv1(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "evmInterpreter", nil)

	// res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("66eeeeeeeeeeeeee60005260206000f3")}, nil, nil)
	// s.Require().Contains(hex.EncodeToString(res.Data), initvalue)

	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("000000000000000000000000000000000000000000000000000000000000001066eeeeeeeeeeeeee60005260206000f3")}, nil, nil)
	s.Require().Equal("00000000000000000000000000000000000000000000000000eeeeeeeeeeeeee", qres)

	calld := "00000000000000000000000000000000000000000000000000000000000007df" + testdata.Fibonacci[64:] + "0000000000000000000000000000000000000000000000000000000000000024b19602740000000000000000000000000000000000000000000000000000000000000012"
	fmt.Println("---fibo")
	// fmt.Println(calld)
	start := time.Now()
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)

	fmt.Println("-fibo-elapsed", time.Since(start))

	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", qres)
}

func (suite *KeeperTestSuite) TestInterpreterContractTest() {
	wasmbin := iContract
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCodeEwasmEnv1(sender, wasmbin)
	appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "contract", nil)
}

// func (suite *KeeperTestSuite) TestInterpreterSimpleStorage() {
// 	wasmbin := iSimpleStorage
// 	sender := suite.GetRandomAccount()
// 	initBalance := sdk.NewInt(1000_000_000)
// 	getHex := `6d4ce63c`
// 	setHex := `60fe47b1`
// 	getHex1 := `054c1a75`
// 	getHex2 := `d2178b08`

// 	appA := s.GetAppContext(s.chainA)
// 	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
// 	suite.Commit()

// 	codeId := appA.StoreCodeEwasmEnv1(sender, wasmbin)
// 	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

// 	initvalue := "0000000000000000000000000000000000000000000000000000000000000005"
// 	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
// 	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
// 	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

// 	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex)}, nil, nil)
// 	s.Require().Contains(hex.EncodeToString(res.Data), initvalue)

// 	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(setHex + "0000000000000000000000000000000000000000000000000000000000000006")}, nil, nil)

// 	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
// 	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", hex.EncodeToString(queryres))

// 	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex)}, nil, nil)
// 	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000006")

// 	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex)}, nil, nil)
// 	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", qres)

// 	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex1)}, nil, nil)
// 	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000000007")

// 	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex2)}, nil, nil)
// 	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000000008")
// }
