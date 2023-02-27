package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"strings"
	"wasmx/x/wasmx/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

var (

	//go:embed testdata/classic/opcodes_all.wasm
	opcodeswasm []byte
)

func (suite *KeeperTestSuite) TestEwasmOpcodes() {
	wasmbin := opcodeswasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	// "758aa8ad": "address_()",
	addresshex := "758aa8ad"
	// "faab7fd7": "balance_(uint256)",
	balancehex := "faab7fd7"
	// "94ef052c": "selfbalance_()",
	selfbalancehex := "94ef052c"
	// "3deeb600": "basefee_()",
	basefeehex := "3deeb600"
	// "8491293f": "and_(uint256,uint256)",
	andhex := "8491293f"
	// "d1cbf1c5": "add_(uint256,uint256)",
	addhex := "d1cbf1c5"
	// "f090359a": "sub_(uint256,uint256)",
	subhex := "f090359a"
	// "3fc4f3f5": "lt_(uint256,uint256)",
	lthex := "3fc4f3f5"
	// "5b1db25f": "gt_(uint256,uint256)",
	gthex := "5b1db25f"
	// "5a7ea41c": "mod_(uint256,uint256)",
	modhex := "5a7ea41c"
	// "c42e9208": "mul_(uint256,uint256)",
	mulhex := "c42e9208"
	// "d6eddb18": "not_(uint256)",
	nothex := "d6eddb18"
	// "d957a807": "addmod_(uint256,uint256,uint256)",
	addmodhex := "d957a807"
	// "ab3300c3": "mulmod_(uint256,uint256,uint256)",
	mulmodhex := "ab3300c3"

	// "d92b22cc": "byte_(uint256,uint256)",
	// "70ea8194": "div_(uint256,uint256)",
	// "d8a1aad7": "eq_(uint256,uint256)",
	// "630834b5": "exp_(uint256,uint256)",
	// "3f8d6558": "or_(uint256,uint256)",
	// "bf29425c": "sar_(int256,int256)",
	sarhex := "bf29425c"
	// "74f6c5bb": "sdiv_(uint256,uint256)",
	// "e7a77a56": "sgt_(uint256,uint256)",

	// "0f58c996": "slt_(uint256,uint256)",
	// "d44aeb8a": "smod_(uint256,uint256)",
	// "27401a41": "xor_(uint256,uint256)"

	// "45a90766": "shl_(uint256,uint256)",
	shlhex := "45a90766"
	// "38619a92": "shr_(uint256,uint256)",
	shrhex := "38619a92"

	// "bb1c8ed4": "signextend_(uint256,uint256)",
	// "44febe2f": "iszero_(uint256)",
	// "b4c4b7ff": "sha3_(bytes)",

	// "4b00ea37": "calldataload_(uint256)",
	calldataloadhex := "4b00ea37"
	// "584a4504": "calldatasize_()",
	calldatasizehex := "584a4504"
	// "c2f490e9": "caller_()",
	callerhex := "c2f490e9"
	// "df48621b": "callvalue_()",
	callvalue := "df48621b"
	// "414d3fbe": "chainid_()",
	chainidhex := "414d3fbe"
	// "fcca1ca2": "codesize_()",
	codesizehex := "fcca1ca2"
	// "9c51d0ba": "coinbase_()",
	coinbasehex := "9c51d0ba"
	// "91a8c7a1": "blockhash_(uint256)",
	blockhashhex := "91a8c7a1"
	// "57296d07": "gas_()",
	gashex := "57296d07"
	// "0dfe3b3d": "gaslimit_()",
	gaslimithex := "0dfe3b3d"
	// "dd5d9040": "gasprice_()",
	gaspricehex := "dd5d9040"

	// "9cb9a1ab": "number_()",
	// "287c71e8": "origin_()",
	// "24b60399": "timestamp_()",

	// "b7af15de": "calldatacopy_(uint256,uint256,uint256)",
	// "7445bcc5": "codecopy_(uint256,uint256,uint256)",

	// "ede5f84d": "log0_(bytes)",
	// "be6891d4": "log1_(bytes,uint256)",
	// "19ded3ba": "log2_(bytes,uint256,uint256)",
	// "b4395768": "log3_(bytes,uint256,uint256,uint256)",
	// "b2627408": "log4_(bytes,uint256,uint256,uint256,uint256)",

	// "4d3874b2": "msize_()",
	// "cbccd294": "return_(bytes)",
	// "b3cbaf3e": "returndatasize_()",
	// "fbd101fb": "revert_(bytes)",
	// "d801346f": "sload_(uint256)",
	// "509665f9": "sstore_(uint256,uint256)",
	// "8abd861a": "create2_(uint256,uint256,uint256,uint256)",
	// "b27ad395": "create_(uint256,uint256,uint256)",
	// "cce14e6b": "call_(uint256,address,uint256,uint256,uint256,uint256,uint256)",
	// "ba27c2d7": "callcode_(uint256,address,uint256,uint256,uint256,uint256,uint256)",
	// "baab795d": "delegatecall_(uint256,address,uint256,uint256,uint256,uint256)",
	// "79625029": "staticcall_(uint256,address,uint256,uint256,uint256,uint256)"
	// "86f7b7d0": "extcodecopy_(address,uint256,uint256,uint256)",
	// "f49fa982": "extcodehash_(uint256)",
	extcodehashhex := "f49fa982"
	// "f78d2c3c": "extcodesize_(address)",
	// "6e35da83": "extcodesize_(uint256)",

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}})
	contractAddressHex := common.BytesToAddress(contractAddress.Bytes()).Hex()

	appA.faucet.Fund(appA.Context(), contractAddress, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(addresshex)}, nil)
	s.Require().Equal("000000000000000000000000"+strings.ToLower(contractAddressHex[2:]), qres)

	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(basefeehex)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", qres)

	calld := andhex + "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", qres)

	calld = addhex + "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", qres)

	calld = subhex + "00000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

	calld = mulhex + "00000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", qres)

	calld = lthex + "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

	calld = gthex + "00000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

	calld = modhex + "00000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", qres)

	calld = nothex + "0000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd", qres)

	calld = addmodhex + "000000000000000000000000000000000000000000000000000000000000000500000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", qres)

	calld = mulmodhex + "000000000000000000000000000000000000000000000000000000000000000500000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000004"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", qres)

	calld = shrhex + "0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", qres)

	calld = shlhex + "0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000030", qres)

	calld = sarhex + "0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", qres)

	calld = sarhex + "0000000000000000000000000000000000000000000000000000000000000002fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd", qres)

	calld = calldataloadhex + "0000000000000000000000000000000000000000000000000000000000000024123456789abcdef111111111111111111111111111111111111fffffffffffff"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("123456789abcdef111111111111111111111111111111111111fffffffffffff", qres)

	calld = calldatasizehex + "112233"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000007", qres)

	calld = callerhex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("000000000000000000000000"+strings.ToLower(common.BytesToAddress(sender.Address.Bytes()).Hex()[2:]), qres)

	calld = chainidhex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000001b58", qres)

	calld = codesizehex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000001a5d", qres)

	calld = gaslimithex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000989680", qres)

	calld = coinbasehex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	// s.Require().Equal("000000000000000000000000"+strings.ToLower(hex.EncodeToString(suite.ProposerAddress)), qres)
	s.Require().Equal(1, 2)

	calld = balancehex + "000000000000000000000000" + contractAddressHex[2:]
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Contains(qres, "00"+hex.EncodeToString(initBalance.BigInt().Bytes()))

	calld = selfbalancehex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Contains(qres, "00"+hex.EncodeToString(initBalance.BigInt().Bytes()))

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(callvalue)}, sdk.Coins{sdk.NewCoin(appA.denom, sdk.NewInt(99999999))})
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000005f5e0ff")

	calld = callvalue
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, sdk.Coins{sdk.NewCoin(appA.denom, sdk.NewInt(99999999))})
	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000005f5e0ff")

	calld = gashex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal("000000000000000000000000000000000000000000000000000175e626ad9975", qres)

	calld = extcodehashhex + "000000000000000000000000" + strings.ToLower(contractAddressHex[2:])
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	codeInfo := appA.app.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().Equal(qres, hex.EncodeToString(codeInfo.CodeHash))

	// queryData := tmservice.GetBlockByHeightRequest{Height: 3}
	// queryDataBz, err := queryData.Marshal()
	// s.Require().NoError(err)
	// resbz := s.Query(sender, contractAddress, queryDataBz, "/cosmos.base.tendermint.v1beta1.Service/GetBlockByHeight")
	// var resp tmservice.GetBlockByHeightResponse
	// err = resp.Unmarshal(resbz.Value)
	// s.Require().NoError(err)

	calld = blockhashhex + "0000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal(string(qres), "0000000000000000000000000000000000000000000000000000000000000000")

	// tmservice.ServiceClient
	// _, err := s.queryClient.GetBlockByHeight(context.Background(), &tmservice.GetBlockByHeightRequest{Height: 1})

	calld = gaspricehex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil)
	s.Require().Equal(string(qres), "0000000000000000000000000000000000000000000000000000000000000000")
}
