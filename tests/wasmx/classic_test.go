package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	testdata "github.com/loredanacirstea/mythos-tests/testdata/classic"
	mcfg "github.com/loredanacirstea/wasmx/config"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	wasmxkeeper "github.com/loredanacirstea/wasmx/x/wasmx/keeper"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestSendingCoinsToNewAccount() {
	sender := suite.GetRandomAccount()
	newacc := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_ = appA.ExecuteContract(sender, appA.BytesToAccAddressPrefixed(newacc.Address), types.WasmxExecutionMessage{Data: []byte{}}, sdk.Coins{sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(1_000_000))}, nil)

	realBalance := appA.App.WasmxKeeper.GetBankKeeper().GetBalance(appA.Context(), newacc.Address, appA.Chain.Config.BaseDenom)
	s.Require().Equal(sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(1_000_000)), realBalance)
}

func (suite *KeeperTestSuite) TestEwasmOpcodes() {
	suite.SetCurrentChain(mcfg.MYTHOS_CHAIN_ID_TEST)

	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	evmcode, err := hex.DecodeString(testdata.OpcodesAll)
	s.Require().NoError(err)

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
	// "bf29425c": "sar_(uint256,uint256)",
	sarhex := "2ea9b94b"
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
	numberhex := "9cb9a1ab"
	// "287c71e8": "origin_()",
	// "24b60399": "timestamp_()",
	timestamphex := "24b60399"

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

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "allopcodes", nil)
	contractAddressHex := common.BytesToAddress(contractAddress.Bytes()).Hex()

	_, codeInfo1, _, err := appA.App.WasmxKeeper.ContractInstance(appA.Context(), contractAddress)
	s.Require().NoError(err)
	s.Require().Greater(len(codeInfo1.InterpretedBytecodeDeployment), 0)
	s.Require().Greater(len(codeInfo1.InterpretedBytecodeRuntime), 0)

	appA.Faucet.Fund(appA.Context(), contractAddress, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(addresshex)}, nil, nil)
	s.Require().Equal("000000000000000000000000"+strings.ToLower(contractAddressHex[2:]), qres)

	calld := andhex + "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", qres)

	calld = addhex + "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", qres)

	calld = subhex + "00000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

	calld = mulhex + "00000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", qres)

	calld = lthex + "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

	calld = gthex + "00000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

	calld = modhex + "00000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", qres)

	calld = nothex + "0000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd", qres)

	calld = addmodhex + "000000000000000000000000000000000000000000000000000000000000000500000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", qres)

	calld = mulmodhex + "000000000000000000000000000000000000000000000000000000000000000500000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000004"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", qres)

	calld = shrhex + "0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", qres)

	calld = shrhex + "0000000000000000000000000000000000000000000000000000000000000000aa0000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("aa0000000000000000000000000000000000000000000000000000000000000c", qres)

	calld = shrhex + "0000000000000000000000000000000000000000000000000000000000000110c84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", qres)

	calld = shrhex + "0000000000000000000000000000000000000000000000000000000000000080c84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("00000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769", qres)

	calld = shrhex + "0000000000000000000000000000000000000000000000000000000000000100c84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", qres)

	calld = shrhex + "0000000000000000000000000000000000000000000000000000000000000040c84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000c84a6e6ec1e7f30f5c812eeba420f7690000000000000000", qres)

	calld = shrhex + "0000000000000000000000000000000000000000000000000000000000000081c84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("000000000000000000000000000000006425373760f3f987ae409775d2107bb4", qres)

	calld = shrhex + "00000000000000000000000000000000000000000000000000000000000000fb983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000013", qres)

	calld = shrhex + "00000000000000000000000000000000000000000000000000000000000000aa983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("000000000000000000000000000000000000000000260f6fa86bffb7bb894f56", qres)

	calld = shlhex + "0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000030", qres)

	calld = shlhex + "0000000000000000000000000000000000000000000000000000000000000000aa0000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("aa0000000000000000000000000000000000000000000000000000000000000c", qres)

	calld = shlhex + "000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("c84a6e6ec1e7f30f5c812eeba420f76900000000000000000000000000000000", qres)

	calld = shlhex + "000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", qres)

	calld = shlhex + "000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000c84a6e6ec1e7f30f5c812eeba420f769"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000c84a6e6ec1e7f30f5c812eeba420f7690000000000000000", qres)

	calld = shlhex + "00000000000000000000000000000000000000000000000000000000000000fb983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("a800000000000000000000000000000000000000000000000000000000000000", qres)

	calld = shlhex + "00000000000000000000000000000000000000000000000000000000000000aa983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("ce5eb700bda45880968354000000000000000000000000000000000000000000", qres)

	calld = shlhex + "0000000000000000000000000000000000000000000000000000000000000080983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("119058ba35f397adc02f69162025a0d500000000000000000000000000000000", qres)

	calld = sarhex + "0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", qres)

	calld = sarhex + "0000000000000000000000000000000000000000000000000000000000000000aa0000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("aa0000000000000000000000000000000000000000000000000000000000000c", qres)

	calld = sarhex + "0000000000000000000000000000000000000000000000000000000000000002fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd", qres)

	calld = sarhex + "00000000000000000000000000000000000000000000000000000000000000fb983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3", qres)

	calld = sarhex + "00000000000000000000000000000000000000000000000000000000000000aa983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("ffffffffffffffffffffffffffffffffffffffffffe60f6fa86bffb7bb894f56", qres)

	calld = sarhex + "0000000000000000000000000000000000000000000000000000000000000080983dbea1affedeee253d5921804d11ce119058ba35f397adc02f69162025a0d5"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("ffffffffffffffffffffffffffffffff983dbea1affedeee253d5921804d11ce", qres)

	calld = calldataloadhex + "0000000000000000000000000000000000000000000000000000000000000024123456789abcdef111111111111111111111111111111111111fffffffffffff"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("123456789abcdef111111111111111111111111111111111111fffffffffffff", qres)

	calld = calldatasizehex + "112233"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000007", qres)

	calld = callerhex
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("000000000000000000000000"+strings.ToLower(common.BytesToAddress(sender.Address.Bytes()).Hex()[2:]), qres)

	calld = chainidhex
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000001b59", qres)

	calld = gaslimithex
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000005f5e100", qres)

	calld = callvalue
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, sdk.Coins{sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(99999999))}, nil)
	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000005f5e0ff")

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(callvalue)}, sdk.Coins{sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(99999999))}, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000005f5e0ff")

	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(numberhex)}, nil, nil)
	blockno := new(big.Int)
	qresbz, err := hex.DecodeString(qres)
	s.Require().NoError(err)
	blockno.SetBytes(qresbz)
	s.Require().Equal(appA.App.LastBlockHeight(), blockno.Int64())

	// TODO redo this test with correct time; the header time from the consensus contract
	// currentTime := s.Chain().CurrentHeader.Time.Unix()
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(timestamphex)}, nil, nil)
	timestamp := new(big.Int)
	qresbz, err = hex.DecodeString(qres)
	s.Require().NoError(err)
	timestamp.SetBytes(qresbz)
	// s.Require().Equal(currentTime, timestamp.Int64())

	calld = coinbasehex
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	// TODO reenable coinbase
	// s.Require().Equal("000000000000000000000000"+hex.EncodeToString(appA.Context().BlockHeader().ProposerAddress), qres)

	realBalance := appA.App.WasmxKeeper.GetBankKeeper().GetBalancePrefixed(appA.Context(), contractAddress, appA.Chain.Config.BaseDenom)

	calld = balancehex + "00000000000000000000000039B1BF12E9e21D78F0c76d192c26d47fa710Ec98"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Contains(qres, "0000000000000000000000000000000000000000000000000000000000000000")

	calld = balancehex + "000000000000000000000000" + contractAddressHex[2:]
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Contains(qres, "00"+hex.EncodeToString(realBalance.Amount.BigInt().Bytes()))

	calld = selfbalancehex
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Contains(qres, "00"+hex.EncodeToString(realBalance.Amount.BigInt().Bytes()))

	calld = extcodehashhex + "000000000000000000000000" + strings.ToLower(contractAddressHex[2:])
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	codeInfo, err := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().NoError(err)
	s.Require().Equal(qres, hex.EncodeToString(codeInfo.CodeHash))

	calld = gashex
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("00000000000000000000000000000000000000000000000000000000000262d8", qres)

	calld = codesizehex
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000001f41", qres)

	calld = blockhashhex + "0000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal(string(qres), "0000000000000000000000000000000000000000000000000000000000000000")

	calld = gaspricehex
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", string(qres))

	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(basefeehex)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", qres)
}

func (suite *KeeperTestSuite) TestEwasmSimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	getHex := `6d4ce63c`
	setHex := `60fe47b1`

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
	s.Require().NoError(err)

	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	initvaluebz, err := hex.DecodeString(initvalue)
	s.Require().NoError(err)
	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage", nil)

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
}

func (suite *KeeperTestSuite) TestCallFibonacci() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	fibhex := "c6c2ea17"
	fibstorehex := "cf837088"
	fibInternal := "b1960274"
	evmcode, err := hex.DecodeString(testdata.Call)
	s.Require().NoError(err)
	fiboevm, err := hex.DecodeString(testdata.Fibonacci)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAddressFibo := appA.DeployEvm(sender, fiboevm, types.WasmxExecutionMessage{Data: []byte{}}, nil, "fibonacci", nil)

	value := "000000000000000000000000000000000000000000000000000000000000000f"
	result := "0000000000000000000000000000000000000000000000000000000000000262"

	start := time.Now()
	// call fibonaci contract directly
	res := appA.ExecuteContractWithGas(sender, contractAddressFibo, types.WasmxExecutionMessage{Data: append(
		appA.Hex2bz(fibInternal),
		appA.Hex2bz(value)...,
	)}, nil, nil, 300_000_000, nil)

	fmt.Println("-fibo-elapsed", time.Since(start))
	s.Require().Contains(hex.EncodeToString(res.Data), result)

	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callwasm", nil)

	paddedAddr := append(make([]byte, 12), contractAddressFibo.Bytes()...)
	msgFibInternal := types.WasmxExecutionMessage{Data: append(
		append(paddedAddr, appA.Hex2bz(fibInternal)...),
		appA.Hex2bz(value)...,
	)}
	msgFib := types.WasmxExecutionMessage{Data: append(
		append(paddedAddr, appA.Hex2bz(fibhex)...),
		appA.Hex2bz(value)...,
	)}
	msgFibStore := types.WasmxExecutionMessage{Data: append(
		append(paddedAddr, appA.Hex2bz(fibstorehex)...),
		appA.Hex2bz(value)...,
	)}

	// call fibonacci contract through the callwasm contract
	deps := []string{types.EvmAddressFromAcc(contractAddressFibo.Bytes()).Hex()}

	// query fibonacci internal through the callwasm contract
	qres := appA.WasmxQuery(sender, contractAddress, msgFibInternal, nil, deps)
	s.Require().Equal(result, qres)

	res = appA.ExecuteContractWithGas(sender, contractAddress, msgFib, nil, deps, uint64(1_000_000_000), nil)
	s.Require().Contains(hex.EncodeToString(res.Data), result)

	// query fibonacci through the callwasm contract
	qres = appA.WasmxQuery(sender, contractAddress, msgFib, nil, deps)
	s.Require().Equal(result, qres)

	// query fibonacci through the callwasm contract - with storage
	qres = appA.WasmxQuery(sender, contractAddress, msgFibStore, nil, deps)
	s.Require().Equal(result, qres)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddressFibo, keybz)
	suite.Require().Equal("", hex.EncodeToString(queryres))

	res = appA.ExecuteContractWithGas(sender, contractAddressFibo, types.WasmxExecutionMessage{Data: append(
		appA.Hex2bz(fibstorehex),
		appA.Hex2bz(value)...,
	)}, nil, nil, 1_000_000_000, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), result)
	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddressFibo, keybz)
	suite.Require().Equal(result, hex.EncodeToString(queryres))
}

func (suite *KeeperTestSuite) TestEwasmCallRevert() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	receiver := common.HexToAddress("0x0000000000000000000000000000000000001111")
	receiverAcc := types.AccAddressFromEvm(receiver)
	evmcode, err := hex.DecodeString(testdata.CallRevert)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callrevertbin", nil)

	balance := appA.App.BankKeeper.GetBalance(appA.Context(), receiverAcc, appA.Chain.Config.BaseDenom)
	s.Require().Equal(balance.Amount, sdkmath.NewInt(0))

	// contract does not have funds, so the inner call fails but tx returns
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte{}}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000004")

	balance = appA.App.BankKeeper.GetBalance(appA.Context(), receiverAcc, appA.Chain.Config.BaseDenom)
	s.Require().Equal(balance.Amount, sdkmath.NewInt(0))

	// contract has funds, so the inner call succeeds, but tx fails
	// the inner call sends funds to the 0x1111 address, that does not yet exist
	appA = s.AppContext()
	appA.Faucet.Fund(appA.Context(), contractAddress, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	res, err = appA.ExecuteContractNoCheck(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte{}}, nil, nil, 10_000_000, nil)
	s.Require().NoError(err)
	s.Require().True(res.IsErr(), res.GetLog())
	s.Require().Contains(res.GetLog(), "failed to execute message", res.GetLog())

	balance = appA.App.BankKeeper.GetBalance(appA.Context(), receiverAcc, appA.Chain.Config.BaseDenom)
	s.Require().Equal(balance.Amount, sdkmath.NewInt(0))
}

func (suite *KeeperTestSuite) TestEwasmNestedGeneralCall() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	evmcode, err := hex.DecodeString(testdata.CallGeneral)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAccount1 := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callgeneralwasm1", nil)

	_, contractAccount2 := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callgeneralwasm2", nil)

	_, contractAccount3 := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callgeneralwasm3", nil)

	// Execute nested calls
	value := `0000000000000000000000000000000000000000000000000000000000000009`
	data := `0000000000000000000000000000000000000000000000000000000000000002` + `000000000000000000000000` + hex.EncodeToString(contractAccount2.Bytes()) + `000000000000000000000000` + hex.EncodeToString(contractAccount3.Bytes()) + value
	deps := []string{types.EvmAddressFromAcc(contractAccount1.Bytes()).Hex(), types.EvmAddressFromAcc(contractAccount2.Bytes()).Hex(), types.EvmAddressFromAcc(contractAccount3.Bytes()).Hex()}
	res := appA.ExecuteContract(sender, contractAccount1, types.WasmxExecutionMessage{Data: appA.Hex2bz(data)}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000003", string(res.Data))

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAccount1, keybz)
	suite.Require().Equal(value, hex.EncodeToString(queryres))

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAccount2, keybz)
	suite.Require().Equal(value, hex.EncodeToString(queryres))

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAccount3, keybz)
	suite.Require().Equal(value, hex.EncodeToString(queryres))
}

func (suite *KeeperTestSuite) TestEwasmCallNested() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	evm_inner, err := hex.DecodeString(testdata.CallNestedSimple)
	s.Require().NoError(err)
	evmbin, err := hex.DecodeString(testdata.CallNested)
	s.Require().NoError(err)
	evmbin_deep, err := hex.DecodeString(testdata.CallNestedDeep)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	// Deploy first contract
	_, contractAddress1 := appA.DeployEvm(sender, evm_inner, types.WasmxExecutionMessage{Data: []byte{}}, nil, "nested_inner", nil)
	contractHex1 := hex.EncodeToString(contractAddress1.Bytes())

	// Deploy deep contract
	_, contractAddress2 := appA.DeployEvm(sender, evmbin_deep, types.WasmxExecutionMessage{Data: []byte{}}, nil, "nested_deep", nil)
	contractHex2 := hex.EncodeToString(contractAddress2.Bytes())

	// Deploy second contract
	_, contractAddress := appA.DeployEvm(sender, evmbin, types.WasmxExecutionMessage{Data: appA.Hex2bz(contractHex1)}, sdk.NewCoins(sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(100_000))), "callnestedwasm", nil)

	deps := []string{types.EvmAddressFromAcc(contractAddress1.Bytes()).Hex(), types.EvmAddressFromAcc(contractAddress2.Bytes()).Hex()}
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("000000000000000000000000" + contractHex2 + "000000000000000000000000" + contractHex1)}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "00000000000000000000000000000000000000000000000000000000000000740000000000000000000000000000000000000000000000000000000000000011", string(res.Data))
}

func (suite *KeeperTestSuite) TestEwasmStaticCall() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	evmcode, err := hex.DecodeString(testdata.CallStatic)
	s.Require().NoError(err)
	evmcode_inner, err := hex.DecodeString(testdata.CallStaticInner)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, innerContractAddress := appA.DeployEvm(sender, evmcode_inner, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callstaticinnerwasm", nil)
	innerHex1 := hex.EncodeToString(innerContractAddress.Bytes())

	_, scContractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callstaticwasm", nil)

	deps := []string{types.EvmAddressFromAcc(innerContractAddress.Bytes()).Hex()}
	res := appA.ExecuteContract(sender, scContractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("000000000000000000000000" + innerHex1 + "0000000000000000000000000000000000000000000000000000000000000003")}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000004")

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), innerContractAddress, keybz)
	suite.Require().Equal("", hex.EncodeToString(queryres))

	res = appA.ExecuteContract(sender, scContractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("000000000000000000000000" + innerHex1 + "00000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000001")}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000006")

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), innerContractAddress, keybz)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", hex.EncodeToString(queryres))
}

func (suite *KeeperTestSuite) TestEwasmDelegateCall() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	// lib reads from storage key 0 and returns the value
	evmcodelib, err := hex.DecodeString(testdata.CallDelegateLib)
	s.Require().NoError(err)
	// contract stores 9 at key value 0 and returns the delegatecall return value
	evmcode, err := hex.DecodeString(testdata.CallDelegate)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	// Deploy library code
	_, contractAddressAccLib := appA.DeployEvm(sender, evmcodelib, types.WasmxExecutionMessage{Data: []byte{}}, nil, "delegatecalllibwasm", nil)
	contractHex1 := hex.EncodeToString(contractAddressAccLib.Bytes())

	// Deploy second contract
	_, contractAddressAcc := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: appA.Hex2bz(contractHex1)}, sdk.NewCoins(sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(100000))), "delegatecallwasm", nil)

	deps := []string{types.EvmAddressFromAcc(contractAddressAccLib.Bytes()).Hex()}
	res := appA.ExecuteContract(sender, contractAddressAcc, types.WasmxExecutionMessage{Data: appA.Hex2bz("000000000000000000000000" + contractHex1)}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000009")
}

func (suite *KeeperTestSuite) TestCallOutOfGas() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	fibstorehex := "cf837088"
	evmcode, err := hex.DecodeString(testdata.Fibonacci)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "fibonacci", nil)

	value := "0000000000000000000000000000000000000000000000000000000000000005"
	msgFibStore := types.WasmxExecutionMessage{Data: append(appA.Hex2bz(fibstorehex), appA.Hex2bz(value)...)}
	// gas limit is chosen to be > than needed to run the antehandler but < than needed to execute the actual transaction
	res, err := appA.ExecuteContractNoCheck(sender, contractAddress, msgFibStore, nil, nil, 2_000_000, nil)
	// 1093970 - is executed just for the antehandler
	// 2235235 successful execution
	s.Require().NoError(err)
	s.Require().False(res.IsOK(), res.GetLog())
	s.Require().True(res.IsErr(), res.GetLog())
	s.Require().Contains(res.GetLog(), "out of gas", res.GetLog())
}

func (suite *KeeperTestSuite) TestInvalidTransaction() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	fibstorehex := "cf837088"
	evmcode, err := hex.DecodeString(testdata.Fibonacci)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "fibonacci", nil)

	value := "0000000000000000000000000000000000000000000000000000000000000005"
	msgFibStore := types.WasmxExecutionMessage{Data: append(appA.Hex2bz(fibstorehex), appA.Hex2bz(value)...)}

	// create an invalid transaction, make sure it gets rejected from the mempool
	msgbz, err := json.Marshal(msgFibStore)
	s.Require().NoError(err)
	senderstr, err := appA.AddressCodec().BytesToString(sender.Address)
	s.Require().NoError(err)
	executeContractMsg := &types.MsgExecuteContract{
		Sender:       senderstr,
		Contract:     contractAddress.String(),
		Msg:          msgbz,
		Funds:        nil,
		Dependencies: nil,
	}

	bz := appA.PrepareCosmosTx(sender, []sdk.Msg{executeContractMsg}, nil, nil, "")
	// change a byte, so encoding is off
	// this will make the consensus contract revert the "newTransaction" transaction
	bz[0] = 8
	_, err = suite.AddToMempoolFSM([][]byte{bz})
	s.Require().Error(err, "tx should not have been included in the mempool")
}

func (suite *KeeperTestSuite) TestInvalidMessageTransaction() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	fibstorehex := "cf837088"
	evmcode, err := hex.DecodeString(testdata.Fibonacci)
	s.Require().NoError(err)

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	sender2Prefixed := appA.BytesToAccAddressPrefixed(sender2.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	// fund with minimal balance, just to create the account
	appA.Faucet.Fund(appA.Context(), sender2Prefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(1)))
	suite.Commit()

	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "fibonacci", nil)

	value := "0000000000000000000000000000000000000000000000000000000000000005"
	msgFibStore := types.WasmxExecutionMessage{Data: append(appA.Hex2bz(fibstorehex), appA.Hex2bz(value)...)}

	// create an invalid transaction, make sure it gets rejected in baseapp.CheckTx
	msgbz, err := json.Marshal(msgFibStore)
	s.Require().NoError(err)
	executeContractMsg := &types.MsgExecuteContract{
		Sender:       senderPrefixed.String(),
		Contract:     contractAddress.String(),
		Msg:          msgbz,
		Funds:        nil,
		Dependencies: nil,
	}

	bz := appA.PrepareCosmosTx(sender2, []sdk.Msg{executeContractMsg}, nil, nil, "")
	_, err = suite.AddToMempoolFSM([][]byte{bz})
	s.Require().Error(err, "tx should not have been included in the mempool")
}

func (suite *KeeperTestSuite) TestEwasmFibonacci() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	fibhex := "c6c2ea17"
	fibstorehex := "cf837088"
	evmcode, err := hex.DecodeString(testdata.Fibonacci)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "fibonacciwasm", nil)

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fibhex + "0000000000000000000000000000000000000000000000000000000000000005")}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000005")
	s.Require().Equal(0, len(appA.GetEwasmEvents(res.GetEvents())))

	queryMsg := fibhex + "0000000000000000000000000000000000000000000000000000000000000005"
	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", qres)

	queryMsg = fibstorehex + "0000000000000000000000000000000000000000000000000000000000000005"
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", qres)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("", hex.EncodeToString(queryres))

	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fibstorehex + "0000000000000000000000000000000000000000000000000000000000000005")}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000005")

	logs, err := appA.GetEwasmLogs(appA.AddressCodec(), res.GetEvents())
	s.Require().NoError(err)
	s.Require().Equal(7, len(logs))
	s.Require().Equal(7, len(appA.GetEventsByAttribute(res.GetEvents(), "topic", "0x5566666666666666666666666666666666666666666666666666666666666677")))
	logd := appA.GetEventsByAttribute(res.GetEvents(), "topic", "0x0000000000000000000000000000000000000000000000000000000000000005")
	s.Require().Equal(1, len(logd))

	queryres = appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", hex.EncodeToString(queryres))
}

func (suite *KeeperTestSuite) TestEwasmSwitchJump() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	evmcode, err := hex.DecodeString(testdata.Switch)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "switchbin", nil)

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("0000000000000000000000000000000000000000000000000000000000000001")}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000001")

	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("0000000000000000000000000000000000000000000000000000000000000000")}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000008")
}

func (suite *KeeperTestSuite) TestEwasmLogs() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	evmcode, err := hex.DecodeString(testdata.Logs)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "logswasm", nil)

	data := "8888888888888888888888888888888888888888888888888888888888888888"
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(data)}, nil, nil)

	ewasmlogs := appA.GetEwasmEvents(res.GetEvents())
	s.Require().NoError(err)
	s.Require().Equal(5, len(ewasmlogs))

	evs := appA.GetEventsByAttribute(res.GetEvents(), "data", "0x8888888888888888888888888888888888888888888888888888888888888888")
	s.Require().Equal(5, len(evs))
}

func (suite *KeeperTestSuite) TestEwasmCreate1() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	evmcode, err := hex.DecodeString(testdata.Create)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	createHex := "780900dc"
	// create2Hex := "cb858002"

	// Deploy factory
	_, factoryAccount := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "create", nil)

	creationFunds := sdk.Coins{sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(10000))}
	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"

	res := appA.ExecuteContract(sender, factoryAccount, types.WasmxExecutionMessage{Data: appA.Hex2bz(createHex + initvalue)}, creationFunds, nil)

	// contract creation logs
	createdContractAddressStr := appA.GetContractAddressFromEvents(res.GetEvents())
	createdContractAddress, err := appA.AddressStringToAccAddressPrefixed(createdContractAddressStr)
	s.Require().NoError(err)

	wrappedCtx := appA.Context()
	createdContractFunds := appA.App.BankKeeper.GetAllBalancesPrefixed(wrappedCtx, createdContractAddress)
	s.Require().Equal(creationFunds, createdContractFunds)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), createdContractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

	contractInfo, err := appA.App.WasmxKeeper.GetContractInfo(appA.Context(), createdContractAddress)
	s.Require().NoError(err)
	s.Require().NotNil(contractInfo)
	s.Require().Equal(factoryAccount.String(), contractInfo.Provenance)

	_factoryAccount, err := appA.App.AccountKeeper.GetAccountPrefixed(appA.Context(), factoryAccount)
	s.Require().NoError(err)
	_nonce := _factoryAccount.GetSequence() - 1
	_createdContractAddress := wasmxkeeper.EwasmBuildContractAddressClassic(factoryAccount.Bytes(), _nonce)
	s.Require().Equal(createdContractAddress.String(), appA.MustAccAddressToString(_createdContractAddress))

	// create second contract
	res = appA.ExecuteContract(sender, factoryAccount, types.WasmxExecutionMessage{Data: appA.Hex2bz(createHex + initvalue)}, creationFunds, nil)

	// contract creation logs
	createdContractAddressStr = appA.GetContractAddressFromEvents(res.GetEvents())
	createdContractAddress, err = appA.AddressStringToAccAddressPrefixed(createdContractAddressStr)
	s.Require().NoError(err)

	_factoryAccount, err = appA.App.AccountKeeper.GetAccountPrefixed(appA.Context(), factoryAccount)
	s.Require().NoError(err)
	_nonce = _factoryAccount.GetSequence() - 1
	_createdContractAddress = wasmxkeeper.EwasmBuildContractAddressClassic(factoryAccount.Bytes(), _nonce)
	s.Require().Equal(createdContractAddress.String(), appA.MustAccAddressToString(_createdContractAddress))
}

func (suite *KeeperTestSuite) TestEwasmCreate2() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	evmcode, err := hex.DecodeString(testdata.Create)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	create2Hex := "cb858002"

	// Deploy factory
	_, factoryAccount := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "create", nil)

	creationFunds := sdk.Coins{sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(10000))}
	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	salt := "0000000000000000000000000000000000000000000000000000000000000001"

	res := appA.ExecuteContract(sender, factoryAccount, types.WasmxExecutionMessage{Data: appA.Hex2bz(create2Hex + salt + initvalue)}, creationFunds, nil)

	// contract creation logs
	createdContractAddressStr := appA.GetContractAddressFromEvents(res.GetEvents())
	createdContractAddress, err := appA.AddressStringToAccAddressPrefixed(createdContractAddressStr)
	s.Require().NoError(err)

	wrappedCtx := appA.Context()
	createdContractFunds := appA.App.BankKeeper.GetAllBalancesPrefixed(wrappedCtx, createdContractAddress)
	s.Require().NoError(err)
	s.Require().Equal(creationFunds, createdContractFunds)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), createdContractAddress, keybz)
	suite.Require().Equal(hex.EncodeToString(queryres), initvalue)

	contractInfo, err := appA.App.WasmxKeeper.GetContractInfo(appA.Context(), createdContractAddress)
	s.Require().NoError(err)
	s.Require().NotNil(contractInfo)
	s.Require().Equal(factoryAccount.String(), contractInfo.Provenance)

	saltb, _ := hex.DecodeString(salt)
	codeInfo, err := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), contractInfo.CodeId)
	s.Require().NoError(err)
	s.Require().NotNil(codeInfo)

	_createdContractAddress := appA.App.WasmxKeeper.EwasmPredictableAddressGenerator(factoryAccount, saltb, []byte{}, false)(appA.Context(), contractInfo.CodeId, codeInfo.CodeHash)
	s.Require().Equal(createdContractAddress.String(), _createdContractAddress.String())

	// second child contract
	salt = "0000000000000000000000000000000000000000000000000000000000000002"
	res = appA.ExecuteContract(sender, factoryAccount, types.WasmxExecutionMessage{Data: appA.Hex2bz(create2Hex + salt + initvalue)}, creationFunds, nil)

	// contract creation logs
	createdContractAddressStr = appA.GetContractAddressFromEvents(res.GetEvents())
	createdContractAddress, err = appA.AddressStringToAccAddressPrefixed(createdContractAddressStr)
	s.Require().NoError(err)

	saltb, _ = hex.DecodeString(salt)
	_createdContractAddress = appA.App.WasmxKeeper.EwasmPredictableAddressGenerator(factoryAccount, saltb, []byte{}, false)(appA.Context(), contractInfo.CodeId, codeInfo.CodeHash)
	s.Require().Equal(createdContractAddress.String(), _createdContractAddress.String())
}

func (suite *KeeperTestSuite) TestEwasmOrigin() {
	sender := suite.GetRandomAccount()
	senderAddressHex := hex.EncodeToString(sender.Address.Bytes())
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	originbz, err := hex.DecodeString(testdata.Origin)
	s.Require().NoError(err)
	evmcode, err := hex.DecodeString(testdata.CallStatic)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, innerContractAddress := appA.DeployEvm(sender, originbz, types.WasmxExecutionMessage{Data: []byte{}}, nil, "originwasm", nil)
	innerHex1 := hex.EncodeToString(innerContractAddress.Bytes())

	// Deploy staticcall contract
	_, scContractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callstaticwasm", nil)

	deps := []string{types.EvmAddressFromAcc(innerContractAddress.Bytes()).Hex()}
	res := appA.ExecuteContract(sender, scContractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz("000000000000000000000000" + innerHex1 + "0000000000000000000000000000000000000000000000000000000000000000")}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), senderAddressHex)
}

func (suite *KeeperTestSuite) TestEwasmErc20() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	getDecimalsHex := `313ce567`
	getNameHex := `06fdde03`
	getSymbolHex := `95d89b41`
	mintHex := `1249c58b`
	balanceOfHex := `70a08231`
	evmcode, err := hex.DecodeString(testdata.ERC20)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	tokenName := "00000000000000000000000000000000000000000000000000000000000000074d79546f6b656e00000000000000000000000000000000000000000000000000"
	tokenSymbol := "0000000000000000000000000000000000000000000000000000000000000003544b4e0000000000000000000000000000000000000000000000000000000000"
	constructorArgs := "00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000080" + tokenName + tokenSymbol

	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: appA.Hex2bz(constructorArgs)}, nil, "erc20wasm", nil)

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(getDecimalsHex)}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000012")

	queryMsg := getDecimalsHex
	qres := appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(queryMsg)}, nil, nil)
	s.Require().Contains(qres, "0000000000000000000000000000000000000000000000000000000000000012")
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000012", qres)

	queryMsg = getNameHex
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(queryMsg)}, nil, nil)
	s.Require().Contains(qres, tokenName)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000020"+tokenName, qres)

	queryMsg = getSymbolHex
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(queryMsg)}, nil, nil)
	s.Require().Contains(qres, tokenSymbol)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000020"+tokenSymbol, qres)

	// Test minting, test callvalue
	queryMsg = balanceOfHex + "000000000000000000000000" + types.EvmAddressFromAcc(sender.Address).Hex()[2:]
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000000000")

	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(mintHex)}, sdk.Coins{sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(10000000))}, nil)

	queryMsg = balanceOfHex + "000000000000000000000000" + types.EvmAddressFromAcc(sender.Address).Hex()[2:]
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000989680")

	// mint through a contract call
	evmcodeCall, err := hex.DecodeString(testdata.Call)
	s.Require().NoError(err)

	_, contractAddressCall := appA.DeployEvm(sender, evmcodeCall, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callwasm", nil)
	contractAddressErc20Hex := types.Evm32AddressFromAcc(contractAddress.Bytes()).Hex()

	deps := []string{types.EvmAddressFromAcc(contractAddress.Bytes()).Hex()}
	res = appA.ExecuteContract(sender, contractAddressCall, types.WasmxExecutionMessage{Data: appA.Hex2bz(contractAddressErc20Hex[2:] + mintHex)}, sdk.Coins{sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(8888))}, deps)

	queryMsg = balanceOfHex + "000000000000000000000000" + types.EvmAddressFromAcc(contractAddressCall.Bytes()).Hex()[2:]
	qres = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal(qres, "00000000000000000000000000000000000000000000000000000000000022b8")
}

func (suite *KeeperTestSuite) TestContractTransfer() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	sendETH := "c664c714"

	receiver := common.HexToAddress("0x89ec06bFA519Ca6182b3ADaFDe0f05Eeb15394A9")
	value := "0000000000000000000000000000000000000000000000000000000000000001"
	evmcode, err := hex.DecodeString(testdata.Transfer)
	s.Require().NoError(err)

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callwasm", nil)

	appA.Faucet.Fund(appA.Context(), contractAddress, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	cbalance := appA.App.BankKeeper.GetBalancePrefixed(appA.Context(), contractAddress, appA.Chain.Config.BaseDenom)
	s.Require().Equal(initBalance, cbalance.Amount)

	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: appA.Hex2bz(fmt.Sprintf("%s%s%s", sendETH, value, receiver.Hex()[2:]))}, sdk.NewCoins(sdk.NewCoin(appA.Chain.Config.BaseDenom, sdkmath.NewInt(1))), nil)
	realBalance := appA.App.BankKeeper.GetBalance(appA.Context(), types.AccAddressFromEvm(receiver), appA.Chain.Config.BaseDenom)
	s.Require().Equal(realBalance.Amount, sdkmath.NewInt(1))
}

func (suite *KeeperTestSuite) TestKeccak256() {
	var input string
	var expected string
	var res string
	var inputbz []byte
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)
	evmcode, err := hex.DecodeString(testdata.Keccak256Test)
	s.Require().NoError(err)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, contractAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "Keccak256Test", nil)

	input = "0000000000000000000000000000000000000000000000000000000000000000"
	expected = "290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563"
	inputbz, err = hex.DecodeString(input)
	s.Require().NoError(err)
	res = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: inputbz}, nil, nil)
	s.Require().Equal(expected, res)

	input = "1122000000000000000000000000000000000000000000000000000000000000"
	expected = "8d68541fa58fc102a5b96a6e237ecb2983f1a85ab7d0775e54abc387c3c8c398"
	inputbz, err = hex.DecodeString(input)
	s.Require().NoError(err)
	res = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: inputbz}, nil, nil)
	s.Require().Equal(expected, res)

	input = "39B1BF12E9e21D78F0c76d192c26d47fa710Ec9839B1BF12E9e21D78F0c76d192c26d47fa710Ec9839B1BF12E9e21D78F0c76d192c26d47fa710Ec9839B1BF12E9e21D78F0c76d192c26d47fa710Ec9839B1BF12E9e21D78F0c76d192c26d47fa710Ec9839B1BF12E9e21D78F0c76d192c26d47fa710Ec9839B1BF12E9e21D78F0c76d192c26d47fa710Ec9839B1BF12E9e21D78F0c76d192c26d47fa710Ec9839B1BF12E9e21D78F0c76d192c26d47fa710Ec9839B1BF12E9e21D78F0c76d192c26d47fa710Ec98"
	expected = "2e064f6f67e2db421e871e20dd66564e18cba7fa7def016cce92643a23da36ec"
	inputbz, err = hex.DecodeString(input)
	s.Require().NoError(err)
	res = appA.WasmxQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: inputbz}, nil, nil)
	s.Require().Equal(expected, res)
}
