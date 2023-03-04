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

	//go:embed testdata/classic/estid-wallet.wasm
	estidwalletbin []byte

	//go:embed testdata/classic/curve384-test.wasm
	curve384testbin []byte
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

// func (suite *KeeperTestSuite) TestEwasmPrecompileCurve384Direct() {
// 	sender := suite.GetRandomAccount()
// 	initBalance := sdk.NewInt(1000_000_000)

// 	suite.faucet.Fund(suite.ctx, sender.Address, sdk.NewCoin(suite.denom, initBalance))
// 	suite.Commit()

// curve384 := ewasm.GetPrecompileByLabel("curve384")
// 	codeIdCurve := suite.StoreCode(sender, curve384)
// 	addressCurve := suite.InstantiateCode(sender, codeIdCurve, `{"readonly": true, "data": "0x"}`)

// 	// fadd
// 	fsig := "6422c13f"
// 	xhi := "00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769"
// 	xlo := "b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"
// 	yhi := "000000000000000000000000000000000d0d9d4f899b00456516b647c5e9b7ed"
// 	ylo := "02c538d7878e63e8da0603396b4cbd9494d42f691141f9e2e5927cf88aac0c63"
// 	qres := suite.EwasmQuery(sender, addressCurve, []byte(fmt.Sprintf(`{"readonly": true, "data": "0x%s%s%s%s%s"}`, fsig, xhi, xlo, yhi, ylo)), nil)
// 	hi := "00000000000000000000000000000000d5580bbe4b82f354c197e5336a0aaf56"
// 	lo := "ba52704a88c4d94eb0ca5ad6871ee0708b22d6cd75b50e65e2c52317488e7095"
// 	s.Require().Equal(hi+lo, qres)

// 	// fsub
// 	fsig = "b387fd8f"
// 	xhi = "0000000000000000000000000000000003e501df64c8d7065d58eac499351e2a"
// 	xlo = "fcdc74fda6bd4980919ca5dcf51075e51e36e9442aba748d8d9931e0f1332bd6"
// 	yhi = "0000000000000000000000000000000049451a30e75e7a6a7f48519b72a60e4f"
// 	ylo = "f737d5a207bc2e493b8455c10652357e19a1044de6e3c1d680f328cb7015f4ee"
// 	qres = suite.EwasmQuery(sender, addressCurve, []byte(fmt.Sprintf(`{"readonly": true, "data": "0x%s%s%s%s%s"}`, fsig, xhi, xlo, yhi, ylo)), nil)
// 	hi = "00000000000000000000000000000000ba9fe7ae7d6a5c9bde109929268f0fdb"
// 	lo = "05a49f5b9f011b375618501beebe40660495e4f543d6b2b70ca60916811d36e7"
// 	s.Require().Equal(hi+lo, qres)

// 	// fmul
// 	fsig = "970a1fe1"
// 	xhi = "00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769"
// 	xlo = "b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"
// 	yhi = "000000000000000000000000000000000d0d9d4f899b00456516b647c5e9b7ed"
// 	ylo = "02c538d7878e63e8da0603396b4cbd9494d42f691141f9e2e5927cf88aac0c63"
// 	qres = suite.EwasmQuery(sender, addressCurve, []byte(fmt.Sprintf(`{"readonly": true, "data": "0x%s%s%s%s%s"}`, fsig, xhi, xlo, yhi, ylo)), nil)
// 	hi = "000000000000000000000000000000005de8b2b22ecdf6790f0c7de8ea01bdd6"
// 	lo = "fb8446353273f6053dd29c5ef32974403861d4b388cefccf2e01f63f53b6ffe0"
// 	s.Require().Equal(hi+lo, qres)

// 	// fmul 2
// 	fsig = "970a1fe1"
// 	xhi = "0000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d"
// 	xlo = "26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c98"
// 	yhi = "00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769"
// 	ylo = "b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"
// 	qres = suite.EwasmQuery(sender, addressCurve, []byte(fmt.Sprintf(`{"readonly": true, "data": "0x%s%s%s%s%s"}`, fsig, xhi, xlo, yhi, ylo)), nil)
// 	hi = "00000000000000000000000000000000858564b53562cbd97f41a5389d7e6673"
// 	lo = "41d0469bbe77677a1ec703fcfcf7fe3f1d0c7b85bf517be09e3b5d480678f3be"
// 	s.Require().Equal(hi+lo, qres)

// 	// finv
// 	fsig = "17fac034"
// 	xhi = "0000000000000000000000000000000003e501df64c8d7065d58eac499351e2a"
// 	xlo = "fcdc74fda6bd4980919ca5dcf51075e51e36e9442aba748d8d9931e0f1332bd6"
// 	qres = suite.EwasmQuery(sender, addressCurve, []byte(fmt.Sprintf(`{"readonly": true, "data": "0x%s%s%s"}`, fsig, xhi, xlo)), nil)
// 	hi = "00000000000000000000000000000000ba2909a8e60a55d7a0caf129a18c6c6a"
// 	lo = "a41434c431646bb4a928e76ad732152f35eb59e6df429de7323e5813809f03dc"
// 	s.Require().Equal(hi+lo, qres)

// 	// fsqur
// 	qres = suite.EwasmQuery(sender, addressCurve, []byte(`{"readonly": true, "data": "0xd11a2e9e00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"}`), nil)
// 	// hi = "1195240816300000000000000000000000000000004820074110110193231243"
// 	// lo = "1592129462351643224710518314155115154117101214196871572721034219"
// 	// s.Require().Equal(hi+lo, qres)
// 	rhi, ok := sdk.NewIntFromString("172587436146765776595475267476930568742")
// 	s.Require().True(ok)
// 	rlo, ok := sdk.NewIntFromString("48520138635265626271663490472705092939691116329213758288807215238505996767551")
// 	s.Require().True(ok)
// 	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)

// 	// oadd
// 	qres = suite.EwasmQuery(sender, addressCurve, []byte(`{"readonly": true, "data": "0x30cf0ca30000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c9800000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"}`), nil)
// 	rhi, ok = sdk.NewIntFromString("44081175095904796235352876975367904678")
// 	s.Require().True(ok)
// 	rlo, ok = sdk.NewIntFromString("100506859219136236035519631350093065801404035061881771274013876202950109718359")
// 	s.Require().True(ok)
// 	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)

// 	// osub
// 	qres = suite.EwasmQuery(sender, addressCurve, []byte(`{"readonly": true, "data": "0x4a535b7e0000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c9800000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"}`), nil)
// 	rhi, ok = sdk.NewIntFromString("192181771008421629849363742203568058066")
// 	s.Require().True(ok)
// 	rlo, ok = sdk.NewIntFromString("50253429609568118015677628747952974325121736874976132164053486168204567507417")
// 	s.Require().True(ok)
// 	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)

// 	// oinv
// 	qres = suite.EwasmQuery(sender, addressCurve, []byte(`{"readonly": true, "data": "0x536efe3a0000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c98"}`), nil)
// 	rhi, ok = sdk.NewIntFromString("243740159862127312284528648991845804952")
// 	s.Require().True(ok)
// 	rlo, ok = sdk.NewIntFromString("114212664340466150970443038717607667640015602627992738166447476421376179050893")
// 	s.Require().True(ok)
// 	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)

// 	// omul
// 	qres = suite.EwasmQuery(sender, addressCurve, []byte(`{"readonly": true, "data": "0x6d199f760000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c9800000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"}`), nil)
// 	rhi, ok = sdk.NewIntFromString("111601834606564682910788718900069040274")
// 	s.Require().True(ok)
// 	rlo, ok = sdk.NewIntFromString("60776024975184475010760057463893313737709808256475322851953635903292101315646")
// 	s.Require().True(ok)
// 	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)

// 	// osqr
// 	qres = suite.EwasmQuery(sender, addressCurve, []byte(`{"readonly": true, "data": "0xb7c5948d0000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c98"}`), nil)
// 	rhi, ok = sdk.NewIntFromString("254801097616806841164035593580960886224")
// 	s.Require().True(ok)
// 	rlo, ok = sdk.NewIntFromString("21597171338405059423043823392981743091518353191029311488707255659114934298079")
// 	s.Require().True(ok)
// 	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)
// }

func (suite *KeeperTestSuite) TestEwasmPrecompileCurve384Direct2() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, curve384testbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "curve384testbin", nil)

	// test_cadd
	calldata := "38e3a7eb0000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c9800000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde2643200000000000000000000000000000000aa87ca22be8b05378eb1c71ef320ad746e1d3b628ba79b9859f741e082542a385502f25dbf55296c3a545e3872760ab7000000000000000000000000000000003617de4a96262c6f5d9e98bf9292dc29f8f41dbd289a147ce9da3113b5f0b8c00a60b1ce1d7e819d7a431d7c90ea0e5f"
	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	xhi, ok := sdk.NewIntFromString("269429570830637862272946976383477993217")
	s.Require().True(ok)
	xlo, ok := sdk.NewIntFromString("33164330947121508193438466374545365582299057006401131236103295757451220504030")
	s.Require().True(ok)
	yhi, ok := sdk.NewIntFromString("6242975231567010735109018283462506476")
	s.Require().True(ok)
	ylo, ok := sdk.NewIntFromString("34117617441610352774892141774972852389619107500421055621444636069156687180366")
	s.Require().True(ok)
	s.Require().Equal("00000000000000000000000000000000"+xhi.BigInt().Text(16)+xlo.BigInt().Text(16)+"000000000000000000000000000000000"+yhi.BigInt().Text(16)+ylo.BigInt().Text(16), qres)

	// test_cdbl
	calldata = "1b2874700000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c9800000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	xhi, ok = sdk.NewIntFromString("217463657038587930709755035849976368433")
	s.Require().True(ok)
	xlo, ok = sdk.NewIntFromString("26937741697835980130682031784621424112724455479537907929997665780497544946843")
	s.Require().True(ok)
	yhi, ok = sdk.NewIntFromString("98383791699341512653244436809366107612")
	s.Require().True(ok)
	ylo, ok = sdk.NewIntFromString("5059714547996504795299887014531209785147688643102472020979042830755696392339")
	s.Require().True(ok)
	s.Require().Equal("00000000000000000000000000000000"+xhi.BigInt().Text(16)+xlo.BigInt().Text(16)+"00000000000000000000000000000000"+yhi.BigInt().Text(16)+"0"+ylo.BigInt().Text(16), qres)

	// test_cmul
	calldata = "0dcdcb38000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001", qres)

	// test_cmul2
	calldata = "0dcdcb38000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(calldata)}, nil, nil)
	xhi, ok = sdk.NewIntFromString("245781082670945953400715156833221871355")
	s.Require().True(ok)
	xlo, ok = sdk.NewIntFromString("63561210578733136403401621863479295999629046857190053662312855096114351736595")
	s.Require().True(ok)
	yhi, ok = sdk.NewIntFromString("180019394029199568620537828433208332692")
	s.Require().True(ok)
	ylo, ok = sdk.NewIntFromString("33725942635937835172842051555502887548933743943411496924409836942121128525645")
	s.Require().True(ok)
	s.Require().Equal("00000000000000000000000000000000"+xhi.BigInt().Text(16)+xlo.BigInt().Text(16)+"00000000000000000000000000000000"+yhi.BigInt().Text(16)+ylo.BigInt().Text(16), qres)

	// test_fadd()
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("0492770c")}, nil, nil)

	// test_finv
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("eb69ead0")}, nil, nil)

	// test_fmul
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("9ae6c915")}, nil, nil)

	// test_fmul2
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("cbd1c8d6")}, nil, nil)

	// test_fsub
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("5fb43bb4")}, nil, nil)

	// test_cadd()
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("6c96097e")}, nil, nil)

	// test_cdbl
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("b3baeed1")}, nil, nil)

	// test_cmul
	// suite.ExecuteContractWithOpts(sender, contractAddress, fmt.Sprintf(`{"readonly": true, "data": "0x%s"}`, "e5582b4d"), nil, 1_000_000_000_000, nil)

	// test_verify()
	// 0xba33b770

	// test_verify_fast()
	// 0xb8ae44bd

	// test_verify_neg
	// 0x2cb2e7f1
}

// func (suite *KeeperTestSuite) TestEwasmPrecompileEstoniaIDDirect() {
// 	sender := suite.GetRandomAccount()
// 	initBalance := sdk.NewInt(1_000_000_000_000)
// 	precomputeGenHex := "85d3cf13"
// 	precomputePubHex := "059548ef"
// 	sendETHHex := "5271a63f"
// 	// verifySignatureFast := "b448884d"
// 	value := "0000000000000000000000000000000000000000000000000000000000002710"
// 	receiver := "0x89ec06bFA519Ca6182b3ADaFDe0f05Eeb15394A9"
// 	receiverPadded := "000000000000000000000000" + strings.ToLower(receiver[2:])
// 	// msgHash := "22607b0c4e4dd059e8ee00ea75ffcebd439b27107093226e518df2a08fa3a34c"
// 	PkxHi := "000000000000000000000000000000007db2259aeae1a60c09b5ab79ea623093"
// 	PkxLo := "2eea94ccab529e7df1d1eef8505d1b0c5ed6e81a2d0fb77302866dd9d039432c"
// 	PkyHi := "000000000000000000000000000000004a6b9bdad287d3c05acbb6107abdeea9"
// 	PkyLo := "e745066f63b91c449790a6de0fd2d1fa71bee691a0f76d6c37836e43ad9e5009"
// 	rhi := "00000000000000000000000000000000722a955a00ee534ade4e81b6b2683d8a"
// 	rlo := "4b6d5faced90a48917665cd9d37c37385ec76887325d9535a50d8a3b65ad4b08"
// 	shi := "00000000000000000000000000000000558bc64997b0cb2fce1c232120b058f8"
// 	slo := "514de78a3add4fb60d4ef388f09ac47be8795b5a57b9afb87db123bd8f528f77"

// 	suite.faucet.Fund(suite.ctx, sender.Address, sdk.NewCoin(suite.denom, initBalance))
// 	suite.Commit()

// 	codeId := suite.StoreCode(sender, estidwalletbin)

// 	contractAddress := suite.InstantiateCode(sender, codeId, fmt.Sprintf(`{"readonly": true, "data": "0x%s%s%s%s"}`, PkxHi, PkxLo, PkyHi, PkyLo)) // 2_494_074
// 	suite.faucet.Fund(suite.ctx, sender.Address, sdk.NewCoin(suite.denom, initBalance))
// 	suite.Commit()

// 	fmt.Println("--ExecuteContractWithOpts--")
// 	fmt.Println(fmt.Sprintf(`{"readonly": true, "data": "0x%s"}`, precomputeGenHex))
// 	fmt.Println(fmt.Sprintf(`{"readonly": true, "data": "0x%s%s%s%s%s"}`, precomputePubHex, PkxHi, PkxLo, PkyHi, PkyLo))

// 	suite.ExecuteContractWithOpts(sender, contractAddress, fmt.Sprintf(`{"readonly": true, "data": "0x%s"}`, precomputeGenHex), nil, 100_000_000_000, nil) // 52_810_317

// 	suite.ExecuteContractWithOpts(sender, contractAddress, fmt.Sprintf(`{"readonly": true, "data": "0x%s%s%s%s%s"}`, precomputePubHex, PkxHi, PkxLo, PkyHi, PkyLo), nil, 100_000_000_000, nil) // 52_810_448

// 	// fmt.Println("---verifySignatureFast---")
// 	// fmt.Println(fmt.Sprintf(`{"readonly": true, "data": "0x%s%s%s%s%s%s"}`, verifySignatureFast, msgHash, rhi, rlo, shi, slo))
// 	// verifySignatureFast(bytes32 hash, uint256 rhi, uint256 rlo, uint256 shi, uint256 slo)
// 	// qres := suite.EwasmQuery(sender, contractAddress, []byte(fmt.Sprintf(`{"readonly": true, "data": "0x%s%s%s%s%s%s"}`, verifySignatureFast, msgHash, rhi, rlo, shi, slo)), nil)
// 	// s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

// 	// verifySignature(bytes32 hash, uint256 rhi, uint256 rlo, uint256 shi, uint256 slo)

// 	suite.ExecuteContractWithOpts(sender, contractAddress, fmt.Sprintf(`{"readonly": true, "data": "0x%s%s%s%s%s%s%s"}`, sendETHHex, value, receiverPadded, rhi, rlo, shi, slo), nil, 100_000_000_000, nil) // 21_439_647

// 	wrappedCtx := sdk.WrapSDKContext(suite.ctx)
// 	receiverAddress := ewasm.AccAddressFromHex(receiver)
// 	expectedBalances := sdk.NewCoins(sdk.NewCoin(suite.denom, sdk.NewInt(2466)))
// 	balances, err := suite.app.EwasmKeeper.BankKeeper.AllBalances(wrappedCtx, &banktypes.QueryAllBalancesRequest{Address: receiverAddress.String()})
// 	s.Require().NoError(err)
// 	s.Require().Equal(expectedBalances, balances.Balances)
// }
