package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"

	wasmeth "wasmx/x/wasmx/ewasm"
	wasmxkeeper "wasmx/x/wasmx/keeper"
	"wasmx/x/wasmx/types"
)

var (
	//go:embed testdata/classic/opcodes_all.wasm
	opcodeswasm []byte

	//go:embed testdata/classic/call.wasm
	callwasm []byte

	//go:embed testdata/classic/fibonacci.wasm
	fibonacciwasm []byte

	//go:embed testdata/classic/simple_storage.wasm
	simpleStorage []byte

	//go:embed testdata/classic/call_revert.wasm
	callrevertbin []byte

	//go:embed testdata/classic/call_simple.wasm
	callsimplewasm []byte

	//go:embed testdata/classic/call_nested.wasm
	callnestedwasm []byte

	//go:embed testdata/classic/call_nested_deep.wasm
	callnesteddeepwasm []byte

	//go:embed testdata/classic/call_general.wasm
	callgeneralwasm []byte

	//go:embed testdata/classic/call_static.wasm
	callstaticwasm []byte

	//go:embed testdata/classic/call_static_inner.wasm
	callstaticinnerwasm []byte

	//go:embed testdata/classic/call_delegate.wasm
	delegatecallwasm []byte

	//go:embed testdata/classic/call_delegate_lib.wasm
	delegatecalllibwasm []byte

	//go:embed testdata/classic/origin.wasm
	originwasm []byte

	//go:embed testdata/classic/create.wasm
	createwasm []byte

	//go:embed testdata/classic/create2.wasm
	create2wasm []byte

	//go:embed testdata/classic/logs.wasm
	logswasm []byte

	//go:embed testdata/classic/erc20.wasm
	erc20wasm []byte

	//go:embed testdata/classic/switch.wasm
	switchbin []byte
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
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "opcodetest", nil)
	contractAddressHex := common.BytesToAddress(contractAddress.Bytes()).Hex()

	appA.faucet.Fund(appA.Context(), contractAddress, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(addresshex)}, nil, nil)
	s.Require().Equal("000000000000000000000000"+strings.ToLower(contractAddressHex[2:]), qres)

	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(basefeehex)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000000", qres)

	calld := andhex + "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", qres)

	calld = addhex + "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", qres)

	calld = subhex + "00000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

	calld = mulhex + "00000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", qres)

	calld = lthex + "00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

	calld = gthex + "00000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000001", qres)

	calld = modhex + "00000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", qres)

	calld = nothex + "0000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd", qres)

	calld = addmodhex + "000000000000000000000000000000000000000000000000000000000000000500000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000003"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000002", qres)

	calld = mulmodhex + "000000000000000000000000000000000000000000000000000000000000000500000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000004"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", qres)

	calld = shrhex + "0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", qres)

	calld = shlhex + "0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000030", qres)

	calld = sarhex + "0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000003", qres)

	calld = sarhex + "0000000000000000000000000000000000000000000000000000000000000002fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff4"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd", qres)

	calld = calldataloadhex + "0000000000000000000000000000000000000000000000000000000000000024123456789abcdef111111111111111111111111111111111111fffffffffffff"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("123456789abcdef111111111111111111111111111111111111fffffffffffff", qres)

	calld = calldatasizehex + "112233"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000007", qres)

	calld = callerhex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("000000000000000000000000"+strings.ToLower(common.BytesToAddress(sender.Address.Bytes()).Hex()[2:]), qres)

	calld = chainidhex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000001b59", qres)

	calld = gaslimithex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000989680", qres)

	calld = callvalue
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, sdk.Coins{sdk.NewCoin(appA.denom, sdk.NewInt(99999999))}, nil)
	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000005f5e0ff")

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(callvalue)}, sdk.Coins{sdk.NewCoin(appA.denom, sdk.NewInt(99999999))}, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000005f5e0ff")

	calld = coinbasehex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("000000000000000000000000"+hex.EncodeToString(appA.Context().BlockHeader().ProposerAddress), qres)

	realBalance, err := appA.app.BankKeeper.Balance(appA.Context(), &banktypes.QueryBalanceRequest{Address: contractAddress.String(), Denom: appA.denom})
	s.Require().NoError(err)

	calld = selfbalancehex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Contains(qres, "00"+hex.EncodeToString(realBalance.Balance.Amount.BigInt().Bytes()))

	calld = balancehex + "00000000000000000000000039B1BF12E9e21D78F0c76d192c26d47fa710Ec98"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Contains(qres, "0000000000000000000000000000000000000000000000000000000000000000")

	calld = balancehex + "000000000000000000000000" + contractAddressHex[2:]
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Contains(qres, "00"+hex.EncodeToString(realBalance.Balance.Amount.BigInt().Bytes()))

	calld = extcodehashhex + "000000000000000000000000" + strings.ToLower(contractAddressHex[2:])
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	codeInfo := appA.app.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().Equal(qres, hex.EncodeToString(codeInfo.CodeHash))

	calld = gashex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("000000000000000000000000000000000000000000000000000007f615420f00", qres)

	calld = codesizehex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000001e15", qres)

	calld = blockhashhex + "0000000000000000000000000000000000000000000000000000000000000002"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal(string(qres), "0000000000000000000000000000000000000000000000000000000000000000")

	calld = gaspricehex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: s.hex2bz(calld)}, nil, nil)
	s.Require().Equal(string(qres), "0000000000000000000000000000000000000000000000000000000000000000")
}

func (suite *KeeperTestSuite) TestEwasmSimpleStorage() {
	wasmbin := simpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)
	getHex := `6d4ce63c`
	setHex := `60fe47b1`
	getHex1 := `054c1a75`
	getHex2 := `d2178b08`

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	initvalue := "0000000000000000000000000000000000000000000000000000000000000005"
	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	queryres := appA.app.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, hex.EncodeToString(queryres))

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(getHex)}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), initvalue)

	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(setHex + "0000000000000000000000000000000000000000000000000000000000000006")}, nil, nil)

	queryres = appA.app.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", hex.EncodeToString(queryres))

	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(getHex)}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000006")

	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(getHex)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", qres)

	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(getHex1)}, nil, nil)
	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000000007")

	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(getHex2)}, nil, nil)
	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000000008")
}

func (suite *KeeperTestSuite) TestCallFibonacci() {
	wasmbin := callwasm
	fibowasm := fibonacciwasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)
	fibhex := "c6c2ea17"
	fibstorehex := "cf837088"

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "callwasm", nil)

	codeIdFibo := appA.StoreCode(sender, fibowasm)
	contractAddressFibo := appA.InstantiateCode(sender, codeIdFibo, types.WasmxExecutionMessage{Data: []byte{}}, "fibonacci", nil)

	value := "0000000000000000000000000000000000000000000000000000000000000005"
	result := "0000000000000000000000000000000000000000000000000000000000000005"
	paddedAddr := append(make([]byte, 12), contractAddressFibo.Bytes()...)
	msgFib := types.WasmxExecutionMessage{Data: append(
		append(paddedAddr, suite.hex2bz(fibhex)...),
		suite.hex2bz(value)...,
	)}
	msgFibStore := types.WasmxExecutionMessage{Data: append(
		append(paddedAddr, suite.hex2bz(fibstorehex)...),
		suite.hex2bz(value)...,
	)}

	// call fibonaci contract directly
	res := appA.ExecuteContract(sender, contractAddressFibo, types.WasmxExecutionMessage{Data: append(
		suite.hex2bz(fibhex),
		suite.hex2bz(value)...,
	)}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), result)

	// call fibonacci contract through the callwasm contract
	deps := []string{wasmeth.EvmAddressFromAcc(contractAddressFibo).Hex()}
	res = appA.ExecuteContract(sender, contractAddress, msgFib, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), result)

	// query fibonacci through the callwasm contract
	qres := appA.EwasmQuery(sender, contractAddress, msgFib, nil, deps)
	s.Require().Equal(result, qres)

	// query fibonacci through the callwasm contract - with storage
	qres = appA.EwasmQuery(sender, contractAddress, msgFibStore, nil, deps)
	s.Require().Equal(result, qres)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	queryres := appA.app.WasmxKeeper.QueryRaw(appA.Context(), contractAddressFibo, keybz)
	suite.Require().Equal("", hex.EncodeToString(queryres))

	res = appA.ExecuteContract(sender, contractAddressFibo, types.WasmxExecutionMessage{Data: append(
		suite.hex2bz(fibstorehex),
		suite.hex2bz(value)...,
	)}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), result)
	queryres = appA.app.WasmxKeeper.QueryRaw(appA.Context(), contractAddressFibo, keybz)
	suite.Require().Equal(result, hex.EncodeToString(queryres))
}

func (suite *KeeperTestSuite) TestEwasmCallRevert() {
	wasmbin := callrevertbin
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "callrevertbin", nil)

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte{}}, nil, nil)

	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000004")
}

func (suite *KeeperTestSuite) TestEwasmNestedGeneralCall() {
	wasmbin := callgeneralwasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId1 := appA.StoreCode(sender, wasmbin)
	contractAccount1 := appA.InstantiateCode(sender, codeId1, types.WasmxExecutionMessage{Data: []byte{}}, "callgeneralwasm1", nil)

	// Contract 2
	codeId2 := appA.StoreCode(sender, wasmbin)
	contractAccount2 := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "callgeneralwasm2", nil)

	// Contract 3
	codeId3 := appA.StoreCode(sender, wasmbin)
	contractAccount3 := appA.InstantiateCode(sender, codeId3, types.WasmxExecutionMessage{Data: []byte{}}, "callgeneralwasm3", nil)

	// Execute nested calls
	value := `0000000000000000000000000000000000000000000000000000000000000009`
	data := `0000000000000000000000000000000000000000000000000000000000000002` + `000000000000000000000000` + hex.EncodeToString(contractAccount2.Bytes()) + `000000000000000000000000` + hex.EncodeToString(contractAccount3.Bytes()) + value
	deps := []string{wasmeth.EvmAddressFromAcc(contractAccount1).Hex(), wasmeth.EvmAddressFromAcc(contractAccount2).Hex(), wasmeth.EvmAddressFromAcc(contractAccount3).Hex()}
	res := appA.ExecuteContract(sender, contractAccount1, types.WasmxExecutionMessage{Data: suite.hex2bz(data)}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000003", string(res.Data))

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	queryres := appA.app.WasmxKeeper.QueryRaw(appA.Context(), contractAccount1, keybz)
	suite.Require().Equal(value, hex.EncodeToString(queryres))

	queryres = appA.app.WasmxKeeper.QueryRaw(appA.Context(), contractAccount2, keybz)
	suite.Require().Equal(value, hex.EncodeToString(queryres))

	queryres = appA.app.WasmxKeeper.QueryRaw(appA.Context(), contractAccount3, keybz)
	suite.Require().Equal(value, hex.EncodeToString(queryres))
}

func (suite *KeeperTestSuite) TestEwasmCallNested() {
	wasmbin_inner := callsimplewasm
	wasmbin := callnestedwasm
	wasmbin_deep := callnesteddeepwasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(10_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	// Deploy first contract
	codeId1 := appA.StoreCode(sender, wasmbin_inner)
	contractAddress1 := appA.InstantiateCode(sender, codeId1, types.WasmxExecutionMessage{Data: []byte{}}, "callsimplewasm", nil)
	contractHex1 := hex.EncodeToString(contractAddress1.Bytes())

	// Deploy deep contract
	codeId2 := appA.StoreCode(sender, wasmbin_deep)
	contractAddress2 := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "callsimplewasm", nil)
	contractHex2 := hex.EncodeToString(contractAddress2.Bytes())

	// Deploy second contract
	codeId := appA.StoreCode(sender, wasmbin)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: s.hex2bz(contractHex1)}, "callnestedwasm", sdk.NewCoins(sdk.NewCoin(appA.denom, sdk.NewInt(100_000))))

	deps := []string{wasmeth.EvmAddressFromAcc(contractAddress1).Hex(), wasmeth.EvmAddressFromAcc(contractAddress2).Hex()}
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("000000000000000000000000" + contractHex2 + "000000000000000000000000" + contractHex1)}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "00000000000000000000000000000000000000000000000000000000000000740000000000000000000000000000000000000000000000000000000000000011", string(res.Data))
}

func (suite *KeeperTestSuite) TestEwasmStaticCall() {
	wasmbin_inner := callstaticinnerwasm
	wasmbin := callstaticwasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin_inner)
	innerContractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "callstaticinnerwasm", nil)
	innerHex1 := hex.EncodeToString(innerContractAddress.Bytes())

	codeId2 := appA.StoreCode(sender, wasmbin)
	scContractAddress := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "callstaticwasm", nil)

	deps := []string{wasmeth.EvmAddressFromAcc(innerContractAddress).Hex()}
	res := appA.ExecuteContract(sender, scContractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("000000000000000000000000" + innerHex1 + "0000000000000000000000000000000000000000000000000000000000000003")}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000004")

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	queryres := appA.app.WasmxKeeper.QueryRaw(appA.Context(), innerContractAddress, keybz)
	suite.Require().Equal("", hex.EncodeToString(queryres))

	res = appA.ExecuteContract(sender, scContractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("000000000000000000000000" + innerHex1 + "00000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000001")}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000006")

	queryres = appA.app.WasmxKeeper.QueryRaw(appA.Context(), innerContractAddress, keybz)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000006", hex.EncodeToString(queryres))
}

func (suite *KeeperTestSuite) TestEwasmDelegateCall() {
	// lib reads from storage key 0 and returns the value
	wasmlib := delegatecalllibwasm
	// contract stores 9 at key value 0 and returns the delegatecall return value
	wasmbin := delegatecallwasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(10_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	// Deploy library code
	codeIdLib := appA.StoreCode(sender, wasmlib)
	contractAddressAccLib := appA.InstantiateCode(sender, codeIdLib, types.WasmxExecutionMessage{Data: []byte{}}, "delegatecalllibwasm", nil)
	contractHex1 := hex.EncodeToString(contractAddressAccLib.Bytes())

	// Deploy second contract
	codeId := appA.StoreCode(sender, wasmbin)
	contractAddressAcc := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: suite.hex2bz(contractHex1)}, "delegatecallwasm", sdk.NewCoins(sdk.NewCoin(appA.denom, sdk.NewInt(100000))))

	deps := []string{wasmeth.EvmAddressFromAcc(contractAddressAccLib).Hex()}
	res := appA.ExecuteContract(sender, contractAddressAcc, types.WasmxExecutionMessage{Data: suite.hex2bz("000000000000000000000000" + contractHex1)}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000009")
}

func (suite *KeeperTestSuite) TestCallOutOfGas() {
	wasmbin := fibonacciwasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)
	fibstorehex := "cf837088"

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "fibonacci", nil)

	value := "0000000000000000000000000000000000000000000000000000000000000005"
	msgFibStore := types.WasmxExecutionMessage{Data: append(suite.hex2bz(fibstorehex), suite.hex2bz(value)...)}

	res := appA.ExecuteContractNoCheck(sender, contractAddress, msgFibStore, nil, nil, 140_000, nil)
	s.Require().False(res.IsOK(), res.GetLog())
	s.Require().Contains(res.GetLog(), "out of gas", res.GetLog())
	s.Commit()
}

func (suite *KeeperTestSuite) TestEwasmFibonacci() {
	wasmbin := fibonacciwasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)
	fibhex := "c6c2ea17"
	fibstorehex := "cf837088"

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "fibonacciwasm", nil)

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(fibhex + "0000000000000000000000000000000000000000000000000000000000000005")}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000005")

	queryMsg := fibhex + "0000000000000000000000000000000000000000000000000000000000000005"
	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", qres)

	queryMsg = fibstorehex + "0000000000000000000000000000000000000000000000000000000000000005"
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", qres)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	queryres := appA.app.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("", hex.EncodeToString(queryres))

	suite.Require().NotContains(res.GetLog(), "topic_0")
	suite.Require().NotContains(res.GetLog(), "wasmx-ewasm_log")

	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(fibstorehex + "0000000000000000000000000000000000000000000000000000000000000005")}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000005")

	queryres = appA.app.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", hex.EncodeToString(queryres))

	suite.Require().Contains(res.GetLog(), `{"key":"data","value":"0x"},{"key":"topic_0","value":"0x5566666666666666666666666666666666666666666666666666666666666677"},{"key":"topic_1","value":"0x0000000000000000000000000000000000000000000000000000000000000005"}`)
}

func (suite *KeeperTestSuite) TestEwasmCreate2() {
	wasmToCreate := callstaticinnerwasm
	wasmbin := create2wasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)
	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmToCreate)

	// Deploy factory
	codeId2 := appA.StoreCode(sender, wasmbin)
	factoryAccount := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "create2wasm", nil)

	codeIdHex := fmt.Sprintf("%064s", strconv.FormatUint(codeId, 16))
	innerExecuteMsg := "0000000000000000000000000000000000000000000000000000000000000008"
	salt := "0000000000000000000000000000000000000000000000000000000000000001"
	creationFunds := sdk.Coins{sdk.NewCoin(appA.denom, sdk.NewInt(10000))}
	res := appA.ExecuteContract(sender, factoryAccount, types.WasmxExecutionMessage{Data: suite.hex2bz(salt + codeIdHex + innerExecuteMsg)}, creationFunds, nil)

	saltb, _ := hex.DecodeString(salt)
	innerExecuteMsgb, _ := hex.DecodeString(innerExecuteMsg)
	codeInfo := appA.app.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().NotNil(codeInfo)

	_createdContractAddress := wasmxkeeper.EwasmPredictableAddressGenerator(factoryAccount, saltb, innerExecuteMsgb, false)(appA.Context(), codeId2, codeInfo.CodeHash)

	_contractInfo := appA.app.WasmxKeeper.GetContractInfo(appA.Context(), _createdContractAddress)
	s.Require().NotNil(_contractInfo)
	s.Require().Equal(codeId, _contractInfo.CodeId)
	s.Require().Equal(factoryAccount.String(), _contractInfo.Creator)

	s.Require().Contains(hex.EncodeToString(res.Data), "000000000000000000000000"+hex.EncodeToString(_createdContractAddress.Bytes()))

	wrappedCtx := sdk.WrapSDKContext(appA.Context())
	createdContractFunds, err := appA.app.BankKeeper.AllBalances(wrappedCtx, &banktypes.QueryAllBalancesRequest{Address: _createdContractAddress.String()})
	s.Require().NoError(err)
	s.Require().Equal(creationFunds, createdContractFunds.Balances)

	// contract creation logs
	createdContractAddressStr := suite.GetContractAddressFromLog(res.GetLog())
	_, err = sdk.AccAddressFromBech32(createdContractAddressStr)
	s.Require().NoError(err)

	res = appA.ExecuteContract(sender, _createdContractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(innerExecuteMsg)}, nil, nil)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	queryres := appA.app.WasmxKeeper.QueryRaw(appA.Context(), _createdContractAddress, keybz)
	suite.Require().Equal(innerExecuteMsg, hex.EncodeToString(queryres))
}

func (suite *KeeperTestSuite) TestEwasmSwitchJump() {
	wasmbin := switchbin
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "switchbin", nil)

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("0000000000000000000000000000000000000000000000000000000000000001")}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000001")

	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("0000000000000000000000000000000000000000000000000000000000000000")}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000008")
}

func (suite *KeeperTestSuite) TestEwasmAddressGeneration() {
	// TODO
	// addressGenerator := appA.app.WasmxKeeper.Keeper.EwasmClassicAddressGenerator(sender.Address)
}

func (suite *KeeperTestSuite) TestEwasmLogs() {
	wasmbin := logswasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "logswasm", nil)

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("0000000000000000000000000000000000000000000000000000000000000008")}, nil, nil)
	logCount := strings.Count(res.GetLog(), "wasmx-ewasm_log_")
	s.Require().Equal(5, logCount, res.GetLog())
}

func (suite *KeeperTestSuite) TestEwasmCreate() {
	wasmToCreate := callstaticinnerwasm
	wasmbin := createwasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)
	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmToCreate)

	// Deploy factory
	codeId2 := appA.StoreCode(sender, wasmbin)
	factoryAccount := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	codeIdHex := fmt.Sprintf("%064s", strconv.FormatUint(codeId, 16))
	creationFunds := sdk.Coins{sdk.NewCoin(appA.denom, sdk.NewInt(10000))}
	res := appA.ExecuteContract(sender, factoryAccount, types.WasmxExecutionMessage{Data: suite.hex2bz(codeIdHex)}, creationFunds, nil)

	_factoryAccount := appA.app.AccountKeeper.GetAccount(appA.Context(), factoryAccount)
	_nonce := _factoryAccount.GetSequence()
	// TODO this should be _nonce - 1
	_createdContractAddress := wasmxkeeper.EwasmBuildContractAddressClassic(factoryAccount, _nonce)

	_contractInfo := appA.app.WasmxKeeper.GetContractInfo(appA.Context(), _createdContractAddress)
	s.Require().NotNil(_contractInfo)
	s.Require().Equal(codeId, _contractInfo.CodeId)
	s.Require().Equal(factoryAccount.String(), _contractInfo.Creator)

	s.Require().Contains(hex.EncodeToString(res.Data), "000000000000000000000000"+hex.EncodeToString(_createdContractAddress.Bytes()))

	// contract creation logs
	createdContractAddressStr := suite.GetContractAddressFromLog(res.GetLog())
	_, err := sdk.AccAddressFromBech32(createdContractAddressStr)
	s.Require().NoError(err)

	wrappedCtx := sdk.WrapSDKContext(appA.Context())
	createdContractFunds, err := appA.app.BankKeeper.AllBalances(wrappedCtx, &banktypes.QueryAllBalancesRequest{Address: _createdContractAddress.String()})
	s.Require().NoError(err)
	s.Require().Equal(creationFunds, createdContractFunds.Balances)

	innerExecuteMsg := "0000000000000000000000000000000000000000000000000000000000000008"
	res = appA.ExecuteContract(sender, _createdContractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(innerExecuteMsg)}, nil, nil)

	keybz := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	queryres := appA.app.WasmxKeeper.QueryRaw(appA.Context(), _createdContractAddress, keybz)
	suite.Require().Equal(innerExecuteMsg, hex.EncodeToString(queryres))
}

func (suite *KeeperTestSuite) TestEwasmOrigin() {
	wasmorigin := originwasm
	wasmbin := callstaticwasm
	sender := suite.GetRandomAccount()
	senderAddressHex := hex.EncodeToString(sender.Address.Bytes())
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmorigin)
	innerContractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "originwasm", nil)
	innerHex1 := hex.EncodeToString(innerContractAddress.Bytes())

	// Deploy staticcall contract
	codeId2 := appA.StoreCode(sender, wasmbin)
	scContractAddress := appA.InstantiateCode(sender, codeId2, types.WasmxExecutionMessage{Data: []byte{}}, "callstaticwasm", nil)

	deps := []string{wasmeth.EvmAddressFromAcc(innerContractAddress).Hex()}
	res := appA.ExecuteContract(sender, scContractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz("000000000000000000000000" + innerHex1 + "0000000000000000000000000000000000000000000000000000000000000000")}, nil, deps)
	s.Require().Contains(hex.EncodeToString(res.Data), senderAddressHex)
}

// func (suite *KeeperTestSuite) TestCosmWasm() {
// 	wasmbin := erc20cw
// }

func (suite *KeeperTestSuite) TestEwasmCannotExecuteInternal() {
	wasmbin := simpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)
	setHex := `60fe47b1`

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	executeMsg := []byte(`{"readonly": false, "data": "0x` + setHex + `0000000000000000000000000000000000000000000000000000000000000006"}`)

	executeCodeMsg := &types.MsgExecuteWithOriginContract{
		Origin:   sender.Address.String(),
		Sender:   sender.Address.String(),
		Contract: contractAddress.String(),
		Msg:      executeMsg,
		Funds:    sdk.Coins{},
	}
	res := appA.DeliverTxWithOpts(sender, executeCodeMsg, 1500000, nil)
	s.Require().False(res.IsOK(), res.GetLog())
	suite.Commit()
}

func (suite *KeeperTestSuite) TestEwasmErc20() {
	wasmbin := erc20wasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)
	getDecimalsHex := `313ce567`
	getNameHex := `06fdde03`
	getSymbolHex := `95d89b41`
	mintHex := `1249c58b`
	balanceOfHex := `70a08231`

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)

	tokenName := "00000000000000000000000000000000000000000000000000000000000000074d79546f6b656e00000000000000000000000000000000000000000000000000"
	tokenSymbol := "0000000000000000000000000000000000000000000000000000000000000003544b4e0000000000000000000000000000000000000000000000000000000000"
	constructorArgs := "00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000080" + tokenName + tokenSymbol

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: suite.hex2bz(constructorArgs)}, "erc20wasm", nil)

	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(getDecimalsHex)}, nil, nil)
	s.Require().Contains(hex.EncodeToString(res.Data), "0000000000000000000000000000000000000000000000000000000000000012")

	queryMsg := getDecimalsHex
	qres := appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(queryMsg)}, nil, nil)
	s.Require().Contains(qres, "0000000000000000000000000000000000000000000000000000000000000012")
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000012", qres)

	queryMsg = getNameHex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(queryMsg)}, nil, nil)
	s.Require().Contains(qres, tokenName)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000020"+tokenName, qres)

	queryMsg = getSymbolHex
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(queryMsg)}, nil, nil)
	s.Require().Contains(qres, tokenSymbol)
	s.Require().Equal("0000000000000000000000000000000000000000000000000000000000000020"+tokenSymbol, qres)

	// Test minting, test callvalue
	queryMsg = balanceOfHex + "000000000000000000000000" + wasmeth.EvmAddressFromAcc(sender.Address).Hex()[2:]
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000000000")

	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(mintHex)}, sdk.Coins{sdk.NewCoin(appA.denom, sdk.NewInt(10000000))}, nil)

	queryMsg = balanceOfHex + "000000000000000000000000" + wasmeth.EvmAddressFromAcc(sender.Address).Hex()[2:]
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal(qres, "0000000000000000000000000000000000000000000000000000000000989680")

	// mint through a contract call
	wasmbincall := callwasm
	codeIdCall := appA.StoreCode(sender, wasmbincall)
	contractAddressCall := appA.InstantiateCode(sender, codeIdCall, types.WasmxExecutionMessage{Data: []byte{}}, "callwasm", nil)
	contractAddressErc20Hex := wasmeth.Evm32AddressFromAcc(contractAddress).Hex()

	deps := []string{wasmeth.EvmAddressFromAcc(contractAddress).Hex()}
	res = appA.ExecuteContract(sender, contractAddressCall, types.WasmxExecutionMessage{Data: suite.hex2bz(contractAddressErc20Hex[2:] + mintHex)}, sdk.Coins{sdk.NewCoin(appA.denom, sdk.NewInt(8888))}, deps)

	queryMsg = balanceOfHex + "000000000000000000000000" + wasmeth.EvmAddressFromAcc(contractAddressCall).Hex()[2:]
	qres = appA.EwasmQuery(sender, contractAddress, types.WasmxExecutionMessage{Data: suite.hex2bz(queryMsg)}, nil, nil)
	s.Require().Equal(qres, "00000000000000000000000000000000000000000000000000000000000022b8")
}
