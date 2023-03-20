package keeper_test

import (
	_ "embed"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmeth "wasmx/x/wasmx/ewasm"
	"wasmx/x/wasmx/types"
)

var (
	//go:embed testdata/classic/Curve384Test.wasm
	curve384testbin []byte
)

var (
	precomputeGenHex = "85d3cf13"
	precomputePubHex = "059548ef"

	PkxHi_1 = "00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769"
	PkxLo_1 = "b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"
	PkyHi_1 = "000000000000000000000000000000000d0d9d4f899b00456516b647c5e9b7ed"
	PkyLo_1 = "02c538d7878e63e8da0603396b4cbd9494d42f691141f9e2e5927cf88aac0c63"

	PkxHi_2 = "000000000000000000000000000000007db2259aeae1a60c09b5ab79ea623093"
	PkxLo_2 = "2eea94ccab529e7df1d1eef8505d1b0c5ed6e81a2d0fb77302866dd9d039432c"
	PkyHi_2 = "000000000000000000000000000000004a6b9bdad287d3c05acbb6107abdeea9"
	PkyLo_2 = "e745066f63b91c449790a6de0fd2d1fa71bee691a0f76d6c37836e43ad9e5009"
)

func (suite *KeeperTestSuite) TestEwasmPrecompileCurve384Direct() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	deps := []string{"0x0000000000000000000000000000000000000005"}
	addressCurve := sdk.AccAddress(appA.Hex2bz("0000000000000000000000000000000000000020"))

	// fadd
	fsig := "6422c13f"
	xhi := "00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769"
	xlo := "b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"
	yhi := "000000000000000000000000000000000d0d9d4f899b00456516b647c5e9b7ed"
	ylo := "02c538d7878e63e8da0603396b4cbd9494d42f691141f9e2e5927cf88aac0c63"
	qres := appA.EwasmQuery(sender, addressCurve, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("0x%s%s%s%s%s", fsig, xhi, xlo, yhi, ylo))}, nil, nil)
	hi := "00000000000000000000000000000000d5580bbe4b82f354c197e5336a0aaf56"
	lo := "ba52704a88c4d94eb0ca5ad6871ee0708b22d6cd75b50e65e2c52317488e7095"
	s.Require().Equal(hi+lo, qres)

	// fsub
	fsig = "b387fd8f"
	xhi = "0000000000000000000000000000000003e501df64c8d7065d58eac499351e2a"
	xlo = "fcdc74fda6bd4980919ca5dcf51075e51e36e9442aba748d8d9931e0f1332bd6"
	yhi = "0000000000000000000000000000000049451a30e75e7a6a7f48519b72a60e4f"
	ylo = "f737d5a207bc2e493b8455c10652357e19a1044de6e3c1d680f328cb7015f4ee"
	qres = appA.EwasmQuery(sender, addressCurve, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("0x%s%s%s%s%s", fsig, xhi, xlo, yhi, ylo))}, nil, nil)
	hi = "00000000000000000000000000000000ba9fe7ae7d6a5c9bde109929268f0fdb"
	lo = "05a49f5b9f011b375618501beebe40660495e4f543d6b2b70ca60916811d36e7"
	s.Require().Equal(hi+lo, qres)

	// fmul
	fsig = "970a1fe1"
	xhi = "00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769"
	xlo = "b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"
	yhi = "000000000000000000000000000000000d0d9d4f899b00456516b647c5e9b7ed"
	ylo = "02c538d7878e63e8da0603396b4cbd9494d42f691141f9e2e5927cf88aac0c63"
	qres = appA.EwasmQuery(sender, addressCurve, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("0x%s%s%s%s%s", fsig, xhi, xlo, yhi, ylo))}, nil, deps)
	hi = "000000000000000000000000000000005de8b2b22ecdf6790f0c7de8ea01bdd6"
	lo = "fb8446353273f6053dd29c5ef32974403861d4b388cefccf2e01f63f53b6ffe0"
	s.Require().Equal(hi+lo, qres)

	// fmul 2
	fsig = "970a1fe1"
	xhi = "0000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d"
	xlo = "26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c98"
	yhi = "00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769"
	ylo = "b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432"
	qres = appA.EwasmQuery(sender, addressCurve, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("0x%s%s%s%s%s", fsig, xhi, xlo, yhi, ylo))}, nil, deps)
	hi = "00000000000000000000000000000000858564b53562cbd97f41a5389d7e6673"
	lo = "41d0469bbe77677a1ec703fcfcf7fe3f1d0c7b85bf517be09e3b5d480678f3be"
	s.Require().Equal(hi+lo, qres)

	// finv
	fsig = "17fac034"
	xhi = "0000000000000000000000000000000003e501df64c8d7065d58eac499351e2a"
	xlo = "fcdc74fda6bd4980919ca5dcf51075e51e36e9442aba748d8d9931e0f1332bd6"
	qres = appA.EwasmQuery(sender, addressCurve, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("0x%s%s%s", fsig, xhi, xlo))}, nil, deps)
	hi = "00000000000000000000000000000000ba2909a8e60a55d7a0caf129a18c6c6a"
	lo = "a41434c431646bb4a928e76ad732152f35eb59e6df429de7323e5813809f03dc"
	s.Require().Equal(hi+lo, qres)

	// fsqur
	qres = appA.EwasmQuery(sender, addressCurve, types.WasmxExecutionMessage{Data: appA.Hex2bz("0xd11a2e9e00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432")}, nil, deps)
	// hi = "1195240816300000000000000000000000000000004820074110110193231243"
	// lo = "1592129462351643224710518314155115154117101214196871572721034219"
	// s.Require().Equal(hi+lo, qres)
	rhi, ok := sdk.NewIntFromString("172587436146765776595475267476930568742")
	s.Require().True(ok)
	rlo, ok := sdk.NewIntFromString("48520138635265626271663490472705092939691116329213758288807215238505996767551")
	s.Require().True(ok)
	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)

	// oadd
	qres = appA.EwasmQuery(sender, addressCurve, types.WasmxExecutionMessage{Data: appA.Hex2bz("0x30cf0ca30000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c9800000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432")}, nil, nil)
	rhi, ok = sdk.NewIntFromString("44081175095904796235352876975367904678")
	s.Require().True(ok)
	rlo, ok = sdk.NewIntFromString("100506859219136236035519631350093065801404035061881771274013876202950109718359")
	s.Require().True(ok)
	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)

	// osub
	qres = appA.EwasmQuery(sender, addressCurve, types.WasmxExecutionMessage{Data: appA.Hex2bz("0x4a535b7e0000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c9800000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432")}, nil, nil)
	rhi, ok = sdk.NewIntFromString("192181771008421629849363742203568058066")
	s.Require().True(ok)
	rlo, ok = sdk.NewIntFromString("50253429609568118015677628747952974325121736874976132164053486168204567507417")
	s.Require().True(ok)
	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)

	// oinv
	qres = appA.EwasmQuery(sender, addressCurve, types.WasmxExecutionMessage{Data: appA.Hex2bz("0x536efe3a0000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c98")}, nil, deps)
	rhi, ok = sdk.NewIntFromString("243740159862127312284528648991845804952")
	s.Require().True(ok)
	rlo, ok = sdk.NewIntFromString("114212664340466150970443038717607667640015602627992738166447476421376179050893")
	s.Require().True(ok)
	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)

	// omul
	qres = appA.EwasmQuery(sender, addressCurve, types.WasmxExecutionMessage{Data: appA.Hex2bz("0x6d199f760000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c9800000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde26432")}, nil, deps)
	rhi, ok = sdk.NewIntFromString("111601834606564682910788718900069040274")
	s.Require().True(ok)
	rlo, ok = sdk.NewIntFromString("60776024975184475010760057463893313737709808256475322851953635903292101315646")
	s.Require().True(ok)
	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)

	// osqr
	qres = appA.EwasmQuery(sender, addressCurve, types.WasmxExecutionMessage{Data: appA.Hex2bz("0xb7c5948d0000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c98")}, nil, deps)
	rhi, ok = sdk.NewIntFromString("254801097616806841164035593580960886224")
	s.Require().True(ok)
	rlo, ok = sdk.NewIntFromString("21597171338405059423043823392981743091518353191029311488707255659114934298079")
	s.Require().True(ok)
	s.Require().Equal("00000000000000000000000000000000"+rhi.BigInt().Text(16)+rlo.BigInt().Text(16), qres)
}

func (suite *KeeperTestSuite) TestEwasmPrecompileCurve384Test() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, curve384testbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "curve384testbin", nil)

	// test_cadd
	calldata := "38e3a7eb0000000000000000000000000000000058df4b4c45b7d92e15838cc2ec62e63d26a7a65903a36031844d06d753766895e2ebf62f2d593d88f797f25a39a72c9800000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769b78d377301367565d6c4579d1bd222dbf64ea76464731482fd32a61ebde2643200000000000000000000000000000000aa87ca22be8b05378eb1c71ef320ad746e1d3b628ba79b9859f741e082542a385502f25dbf55296c3a545e3872760ab7000000000000000000000000000000003617de4a96262c6f5d9e98bf9292dc29f8f41dbd289a147ce9da3113b5f0b8c00a60b1ce1d7e819d7a431d7c90ea0e5f"
	deps := []string{"0x0000000000000000000000000000000000000005"}
	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, deps)
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
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, deps)
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
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, deps)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001", qres)

	// test_cmul2
	calldata = "0dcdcb38000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calldata)}, nil, deps)
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
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("0492770c")}, nil, deps)

	// test_finv
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("eb69ead0")}, nil, deps)

	// test_fmul
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("9ae6c915")}, nil, deps)

	// test_fmul2
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("cbd1c8d6")}, nil, deps)

	// test_fsub
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("5fb43bb4")}, nil, deps)

	// test_cadd()
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("6c96097e")}, nil, deps)

	// test_cdbl
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("b3baeed1")}, nil, deps)
}

func (suite *KeeperTestSuite) TestEwasmPrecompileCurve384TestLong() {
	SkipCIExpensiveTests(suite.T(), "TestEwasmPrecompileCurve384TestLong")

	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)
	deps := []string{"0x0000000000000000000000000000000000000005"}

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, curve384testbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "curve384testbin", nil)

	// test_cmul
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("e5582b4d")}, nil, deps, 1_000_000_000_000, nil)

	fmt.Println("--test_cmul--")
	start := time.Now()
	res := appA.ExecuteContractNoCheck(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("e5582b4d")}, nil, deps, 1_000_000_000_000, nil)
	duration := time.Since(start)
	fmt.Println("Elapsed: ", duration)
	s.Require().True(res.IsOK(), res.GetLog())
	s.Require().NotContains(res.GetLog(), "failed to execute message", res.GetLog())
	s.Commit()

	fmt.Println("--test_verify--")
	start = time.Now()
	res = appA.ExecuteContractNoCheck(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("ba33b770")}, nil, deps, 1_000_000_000_000, nil)
	duration = time.Since(start)
	fmt.Println("Elapsed: ", duration)
	s.Require().True(res.IsOK(), res.GetLog())
	s.Require().NotContains(res.GetLog(), "failed to execute message", res.GetLog())
	s.Commit()
	s.Require().True(false)
}

func (suite *KeeperTestSuite) TestEwasmPrecompileCurve384TestLong2() {
	SkipCIExpensiveTests(suite.T(), "TestEwasmPrecompileCurve384TestLong2")
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)
	deps := []string{"0x0000000000000000000000000000000000000005"}

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, curve384testbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "curve384testbin", nil)

	msgHash := "d093b45258f603020e15de2c058029ae30e73c794212b8c10f58180cb5ce0beb"
	rhi := "0000000000000000000000000000000042359a721ee3f60efdb4096fd48c32e8"
	rlo := "6df129d5028be3fa1626b192458daf49d4c7676c08663a62decad8df853340ad"
	shi := "0000000000000000000000000000000000103c7e7fb0c04197a5371923adda8e"
	slo := "ae415624e6419214f98bebac9a3cf9ddc8bf28eb2871142e9d0371a59598f2dd"

	fmt.Println("--precomputeGenHex--")
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(precomputeGenHex)}, nil, deps, 100_000_000_000, nil) // 52_810_317

	fmt.Println("--precomputePubHex--")
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s%s%s%s", precomputePubHex, PkxHi_2, PkxLo_2, PkyHi_2, PkyLo_2))}, nil, deps, 100_000_000_000, nil) // 52_810_448

	fmt.Println("--test_verify_fast--")
	start := time.Now()
	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s%s%s%s%s", "5879e57c", msgHash, rhi, rlo, shi, slo))}, nil, deps) // , 1_000_000_000_000, nil)
	duration := time.Since(start)
	fmt.Println("Elapsed: ", duration)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)
}

func (suite *KeeperTestSuite) TestEwasmPrecompileWalletRegistry() {
	SkipCIExpensiveTests(suite.T(), "TestEwasmPrecompileWalletRegistry")
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)
	deps := []string{"0x0000000000000000000000000000000000000005"}

	senderHex := wasmeth.EvmAddressFromAcc(sender.Address).Hex()

	EXPIRATION_DELTA := 31556952 // 1 year in seconds
	// RENEWAL_TIMESTAMP_DELTA := 604800; // 1 week in seconds

	// register(uint256,uint256,uint256,uint256)
	register := "375a7c7f"
	// finishRegistration()
	finishRegistration := "f6aead24"
	// // verifySignature(bytes32,uint256,uint256,uint256,uint256)
	// verifySignature := "dd3ee290"
	// // verifySignatureFast(bytes32,uint256,uint256,uint256,uint256)
	// verifySignatureFast := "b448884d"
	// // verifySignatureByIndex(uint256,bytes32,uint256,uint256,uint256,uint256)
	// verifySignatureByIndex := "3d907d52"
	// // verifySignatureFastByIndex(uint256,bytes32,uint256,uint256,uint256,uint256)
	// verifySignatureFastByIndex := "53e05baf"

	// isActive(address)
	isActive := "9f8a13d7"
	// isRegistered(uint256)
	isRegistered := "579a6988"
	// isRegisteredAddress(address)
	isRegisteredAddress := "db0c7ca8"
	// isExpired(uint256)
	isExpired := "d9548e53"
	// isExpiredAddress(address)
	isExpiredAddress := "9bb59591"
	// "counter()
	counter := "61bc221a"
	// expirations(uint256)
	// expirations := "c251ddf8"

	// registerAddress(uint256,address,uint256,uint256,uint256,uint256)
	// registerAddress := "41b709c8"
	// removeAddress(uint256,address,uint256,uint256,uint256,uint256)
	// removeAddress := "e5b3db52"
	// renewAccount(uint256,uint256,uint256,uint256,uint256,uint256)
	// renewAccount := "52e616f8"
	// replaceAccount(uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256,uint256)
	// replaceAccount := "7b7ceb53"

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	registryAddress := sdk.AccAddress(appA.Hex2bz("0000000000000000000000000000000000000021"))

	fmt.Println("--register--")
	appA.ExecuteContractWithGas(sender, registryAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s%s%s%s", register, PkxHi_2, PkxLo_2, PkyHi_2, PkyLo_2))}, nil, deps, 20_000_000_000_000, nil) // 52_810_317

	fmt.Println("--finishRegistration--")
	appA.ExecuteContractWithGas(sender, registryAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(finishRegistration)}, nil, deps, 20_000_000, nil) // 52_810_448

	registered := appA.EwasmQuery(sender, registryAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s", isRegistered, "0000000000000000000000000000000000000000000000000000000000000001"))}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", registered)

	registered = appA.EwasmQuery(sender, registryAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s", isRegisteredAddress, "000000000000000000000000"+senderHex[2:]))}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", registered)

	expired := appA.EwasmQuery(sender, registryAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s", isExpired, "0000000000000000000000000000000000000000000000000000000000000001"))}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", expired)

	expired = appA.EwasmQuery(sender, registryAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s", isExpiredAddress, "000000000000000000000000"+senderHex[2:]))}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", expired)

	active := appA.EwasmQuery(sender, registryAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s", isActive, "000000000000000000000000"+senderHex[2:]))}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", active)

	count := appA.EwasmQuery(sender, registryAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(counter)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", count)

	// fmt.Println("--verifySignature--")
	// qres := appA.EwasmQuery(sender, registryAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s%s%s%s%s", verifySignature, msgHash, rhi, rlo, shi, slo))}, nil, deps) // , 1_000_000_000_000, nil)
	// s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

	// fmt.Println("--verifySignatureFast--")
	// qres = appA.EwasmQuery(sender, registryAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s%s%s%s%s", verifySignatureFast, msgHash, rhi, rlo, shi, slo))}, nil, deps) // , 1_000_000_000_000, nil)
	// s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

	delta_one_year := uint64(EXPIRATION_DELTA/5 + 10) // 5 sec blocks
	s.CommitNBlocks(s.chainA, delta_one_year)

	expired = appA.EwasmQuery(sender, registryAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s", isExpired, "0000000000000000000000000000000000000000000000000000000000000001"))}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", expired)
}
