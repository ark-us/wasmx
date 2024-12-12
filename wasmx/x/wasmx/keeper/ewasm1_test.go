package keeper_test

import (
	_ "embed"
	"encoding/hex"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/v1/x/wasmx/types"
)

var (
	//go:embed testdata/classic/simple_storage.wasm
	simpleStorage []byte

	//go:embed testdata/classic/simple_storage_wc.wasm
	simpleStorageWC []byte

	//go:embed testdata/classic/constructor_test.wasm
	constructortestbin []byte
)

func (suite *KeeperTestSuite) TestEwasm1SimpleStorage() {
	wasmbin := simpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	getHex := `6d4ce63c`
	setHex := `60fe47b1`
	getHex1 := `054c1a75`
	getHex2 := `d2178b08`

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	initvalue := "0000000000000000000000000000000000000000000000000000000000000005"
	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex)}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), initvalue)

	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(setHex + "0000000000000000000000000000000000000000000000000000000000000006")}, nil, nil)

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", hex.EncodeToString(queryres))

	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex)}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000006")

	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", qres)

	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex1)}, nil, nil)
	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000000007")

	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex2)}, nil, nil)
	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000000008")
}

func (suite *KeeperTestSuite) TestEwasm1SimpleStorageConstructor() {
	wasmbin := simpleStorageWC
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	getHex := `6d4ce63c`

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	initvalue := "0000000000000000000000000000000000000000000000000000000000000005"
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: appA.Hex2bz(initvalue)}, "simpleStorage", nil)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getHex)}, nil, nil)
	s.Require().Equal(initvalue, qres)
}

func (suite *KeeperTestSuite) TestEwasmCannotExecuteInternal() {
	wasmbin := simpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	setHex := `60fe47b1`

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	executeMsg := []byte(`{"data": "0x` + setHex + `0000000000000000000000000000000000000000000000000000000000000006"}`)

	executeCodeMsg := &types.MsgExecuteWithOriginContract{
		Origin:   appA.MustAccAddressToString(sender.Address),
		Sender:   appA.MustAccAddressToString(sender.Address),
		Contract: contractAddress.String(),
		Msg:      executeMsg,
		Funds:    sdk.Coins{},
	}
	res, err := appA.DeliverTxWithOpts(sender, executeCodeMsg, 5500000, nil)
	s.Require().NoError(err)
	s.Require().False(res.IsOK(), res.GetLog())
	suite.Commit()
}

func (suite *KeeperTestSuite) TestConstructorTestBin() {
	wasmbin := constructortestbin
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	fsig := "c1b4625e"
	fsig2 := "4a53d41e"
	strmap := "e71a136a"

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	calld := "000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000597364666173666761736b736b646d6664736b6e766b6d6c2c76642c2e6777656c2e72336c742c6b34336f702c65726c3b2c2e663b3b2e6673643b6c2c666c6b6d6766646b6e736b6a61646e6b6c6d73646c76642c6c3b732c6600000000000000"
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, "callwasm", nil)
	appA.Faucet.Fund(appA.Context(), contractAddress, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	res := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fsig)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000002163636861727365743d5554462d383e3c2f703e3c2f626f64793e3c2f68746d6c3e00000000000000000000000000000000000000000000000000000000000000", res)

	res = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fsig2)}, nil, nil)
	s.Require().Equal("000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000d93c21444f43545950452068746d6c3e3c68746d6c3e3c686561643e3c6d65746120636861727365743d225554462d38223e3c6d65746120687474702d65717569763d22582d55412d436f6d70617469626c652220636f6e74656e743d2249453d65646765223e3c6d657461206e616d653d2276696577706f72742220636f6e74656e743d2277696474683d6465766963652d77696474682c20696e697469616c2d7363616c653d312e30223e3c7469746c653e466972737420446f63756d656e743c2f7469746c653e3c2f686561643e3c626f64793e3c703e00000000000000", res)

	res = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(strmap + "0000000000000000000000000000000000000000000000000000000000000000")}, nil, nil)
	s.Require().Equal(calld, res)
}
