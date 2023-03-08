package keeper_test

import (
	_ "embed"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"wasmx/x/wasmx/ewasm"
	"wasmx/x/wasmx/types"
)

var (
	//go:embed testdata/classic/ecrecover.wasm
	ecrecoverbin []byte
)

func (suite *KeeperTestSuite) TestEwasmPrecompileIdentityDirect() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	contractAddress := ewasm.AccAddressFromHex("0x0000000000000000000000000000000000000004")

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("aa0000000000000000000000000000000000000000000000000000000077")}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "aa0000000000000000000000000000000000000000000000000000000077")

	queryMsg := "aa0000000000000000000000000000000000000000000000000000000077"
	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal("aa0000000000000000000000000000000000000000000000000000000077", qres)
}

func (suite *KeeperTestSuite) TestEwasmPrecompileEcrecoverDirect() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	contractAddress := ewasm.AccAddressFromHex("0x0000000000000000000000000000000000000001")

	inputhex := "38d18acb67d25c8bb9942764b62f18e17054f66a817bd4295423adf9ed98873e000000000000000000000000000000000000000000000000000000000000001b38d18acb67d25c8bb9942764b62f18e17054f66a817bd4295423adf9ed98873e789d1dd423d25f0772d2748d60f7e4b81bb14d086eba8e8e8efb6dcff8a4ae02"

	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(inputhex)}, nil, nil)
	s.Require().Equal("000000000000000000000000ceaccac640adf55b2028469bd36ba501f28b699d", qres)
}

func (suite *KeeperTestSuite) TestEwasmPrecompileEcrecover() {
	wasmbin := ecrecoverbin
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "ecrecoverbin", nil)

	appA.faucet.Fund(appA.Context(), contractAddress, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	inputhex := "38d18acb67d25c8bb9942764b62f18e17054f66a817bd4295423adf9ed98873e000000000000000000000000000000000000000000000000000000000000001b38d18acb67d25c8bb9942764b62f18e17054f66a817bd4295423adf9ed98873e789d1dd423d25f0772d2748d60f7e4b81bb14d086eba8e8e8efb6dcff8a4ae02"
	deps := []string{"0x0000000000000000000000000000000000000001"}
	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(inputhex)}, nil, deps)
	s.Require().Equal("000000000000000000000000ceaccac640adf55b2028469bd36ba501f28b699d", qres)
}

func (suite *KeeperTestSuite) TestEwasmPrecompileModexpDirect() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	contractAddress := ewasm.AccAddressFromHex("0x0000000000000000000000000000000000000005")

	// <length_of_BASE> <length_of_EXPONENT> <length_of_MODULUS> <BASE> <EXPONENT> <MODULUS>
	// https://eips.ethereum.org/EIPS/eip-198
	// https://github.com/ethereumproject/evm-rs/blob/master/precompiled/modexp/src/lib.rs#L133
	// https://github.com/ethereum/tests/blob/develop/GeneralStateTests/stPreCompiledContracts/modexpTests.json

	calldata := "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000002003fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2efffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f"
	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	expected := "0000000000000000000000000000000000000000000000000000000000000001"
	s.Require().Equal(expected, qres)

	calldata = "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000020fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2efffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	expected = "0000000000000000000000000000000000000000000000000000000000000000"
	s.Require().Equal(expected, qres)

	calldata = "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002003ffff800000000000000000000000000000000000000000000000000000000000000007"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	expected = "3b01b01ac41f2d6e917c6d6a221ce793802469026d9ab7578fa2e79e4da6aaab"
	s.Require().Equal(expected, qres)

	calldata = "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000002003ffff80"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	expected = "3b01b01ac41f2d6e917c6d6a221ce793802469026d9ab7578fa2e79e4da6aaab"
	s.Require().Equal(expected, qres)

	calldata = "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd"
	res := appA.ExecuteContractNoCheck(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil, 1500000, nil)
	s.Require().True(res.IsErr(), res.GetLog())

	calldata = "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000004003fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2efffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2ffffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	expected = "fd24265072b6b01f9bc93300ae72996f1eb0ef2cc4a943b140c6bf2215143e51765316da9900a45dc6b1c0f71df37fbf1a15f274353de964b74822bf76b98b19"
	s.Require().Equal(expected, qres)

	// modexp less than 32 bytes
	calldata = "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001c03fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2efffffffffffffffffffffffffffffffffffffffffffffffffffffffe"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	expected = "bfcd5188ac621ad4d2690eaee537a3b7114509341402010dcb8f1c31"
	s.Require().Equal(expected, qres)

	// modexp with modulus 48 bytes
	calldata = "0000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000300a366e771bd3bb98a41a1e5c0748561b41e8dc9c22ebc6a9d39dd8797631915ddf7e7dcc5700e7727d336f7c61b3e28a2654cd1c523fb5f67bb11684dad2da807ca0b2a17c660a0f2639fa5026ab9d2ac8d8116848e534bdf3c0955850d7175601fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffeffffffff0000000000000000ffffffff"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	expected = "5de8b2b22ecdf6790f0c7de8ea01bdd6fb8446353273f6053dd29c5ef32974403861d4b388cefccf2e01f63f53b6ffe0"
	s.Require().Equal(expected, qres)

	// extra
	calldata = "00000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000003003e501df64c8d7065d58eac499351e2afcdc74fda6bd4980919ca5dcf51075e51e36e9442aba748d8d9931e0f1332bd6fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffeffffffff0000000000000000fffffffdfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffeffffffff0000000000000000ffffffff"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	expected = "ba2909a8e60a55d7a0caf129a18c6c6aa41434c431646bb4a928e76ad732152f35eb59e6df429de7323e5813809f03dc"
	s.Require().Equal(expected, qres)
}
