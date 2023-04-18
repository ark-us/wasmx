package keeper_test

import (
	_ "embed"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"wasmx/v1/x/wasmx/types"
)

var (
	//go:embed testdata/interpreter/simple_storage.wasm
	iSimpleStorage []byte

	//go:embed testdata/interpreter/contract.wasm
	iContract []byte
)

func (suite *KeeperTestSuite) TestInterpreterContractTest() {
	wasmbin := iContract
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
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

// 	codeId := appA.StoreCode(sender, wasmbin)
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

// 	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex)}, nil, nil)
// 	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", qres)

// 	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex1)}, nil, nil)
// 	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000000007")

// 	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex2)}, nil, nil)
// 	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000000008")
// }
