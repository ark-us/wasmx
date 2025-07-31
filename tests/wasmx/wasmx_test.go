package keeper_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm/types"

	testdata "github.com/loredanacirstea/mythos-tests/testdata/classic"
	wasmxtest "github.com/loredanacirstea/mythos-tests/testdata/wasmx"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

type SysContract struct {
	Benchmark *BenchmarkRequest `json:"benchmark"`
}

type BenchmarkRequest struct {
	Request   vmtypes.CallRequest `json:"request"`
	Magnitude int32               `json:"magnitude"`
}

func (suite *KeeperTestSuite) TestWasmxBenchmark() {
	SkipFixmeTests(suite.T(), "TestWasmxBenchmark")
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	wasmbin := precompiles.GetPrecompileByLabel(appA.AddressCodec(), "sys_proxy")

	sysAddressBz, err := hex.DecodeString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	s.Require().NoError(err)
	sysAddress := appA.BytesToAccAddressPrefixed(sdk.AccAddress(sysAddressBz))

	// deploy an evm contract
	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
	s.Require().NoError(err)
	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	initvaluebz, err := hex.DecodeString(initvalue)
	s.Require().NoError(err)
	codeId2, contractAddress2 := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage", nil)
	codeInfo, err := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId2)
	s.Require().NoError(err)
	s.Require().NotNil(codeInfo)

	getHex := `6d4ce63c`

	// an EOA can make a system call by query
	req := &SysContract{
		Benchmark: &BenchmarkRequest{
			Magnitude: 3,
			Request: vmtypes.CallRequest{
				To:       contractAddress2.Bytes(),
				From:     sender.Address,
				Value:    big.NewInt(0),
				GasLimit: big.NewInt(1000000),
				Calldata: appA.Hex2bz(getHex),
				Bytecode: codeInfo.InterpretedBytecodeRuntime,
				CodeHash: codeInfo.CodeHash,
			},
		},
	}
	data, err := json.Marshal(req)
	s.Require().NoError(err)

	qres := appA.WasmxQuery(sender, sysAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	elapsed := big.NewInt(0).SetBytes(appA.Hex2bz(qres))
	suite.Require().True(elapsed.Cmp(big.NewInt(4)) == 1, fmt.Sprintf("elapsed: %d", elapsed.Uint64()))

	// an EOA cannot make a system call by tx
	res, err := appA.ExecuteContractNoCheck(sender, sysAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 1000000, nil)
	s.Require().NoError(err)
	suite.Require().True(res.IsErr())

	// a contract cannot make a system call
	evmcode, err = hex.DecodeString(testdata.Call)
	s.Require().NoError(err)
	_, callAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callwasm", nil)
	msg := types.WasmxExecutionMessage{Data: append(sysAddress.Bytes(), data...)}
	res, err = appA.ExecuteContractNoCheck(sender, callAddress, msg, nil, nil, 1000000, nil)
	s.Require().NoError(err)
	suite.Require().True(res.IsErr())

	// cannot deploy a system contract
	codeId := appA.StoreCode(sender, wasmbin, nil)
	msgbz, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	s.Require().NoError(err)
	instantiateContractMsg := &types.MsgInstantiateContract{
		Sender: appA.MustAccAddressToString(sender.Address),
		CodeId: codeId,
		Label:  "label",
		Msg:    msgbz,
		Funds:  nil,
	}
	res, err = appA.DeliverTxWithOpts(sender, instantiateContractMsg, "", 15000000, nil)
	s.Require().NoError(err)
	suite.Require().True(res.IsErr(), res.GetLog())
	suite.Require().Contains(res.GetLog(), "invalid address for system contracts")
}

func (suite *KeeperTestSuite) TestWasmxSimpleStorage() {
	wasmbin := wasmxtest.WasmxSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	wasmlogs := appA.GetWasmxEvents(res.GetEvents())
	topicLogs := appA.GetEventsByAttribute(wasmlogs, "topic", "0x68656c6c6f000000000000000000000000000000000000000000000000000000")
	dataLogs := appA.GetEventsByAttribute(topicLogs, "data", "0x")
	s.Require().GreaterOrEqual(len(wasmlogs), 1, res.GetEvents())
	s.Require().Equal(1, len(topicLogs), res.GetEvents())
	s.Require().Equal(1, len(dataLogs))

	initvalue := "sammy"
	keybz := []byte("hello")
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, string(queryres))

	data = []byte(`{"get":{"key":"hello"}}`)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(string(qres), "sammy")
}

func (suite *KeeperTestSuite) TestWasmxSameCode() {
	wasmbin := wasmxtest.WasmxSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	contractInfo, codeInfo, _, err := appA.App.WasmxKeeper.ContractInstance(appA.Context(), contractAddress)
	s.Require().NoError(err)
	s.Require().NotNil(codeInfo)
	s.Require().NotNil(contractInfo)

	codeId2 := appA.StoreCode(sender, wasmbin, nil)
	// TODO we may eventually force same codeid
	// s.Require().Equal(codeId, codeId2)
	s.Require().Equal(codeId+1, codeId2)
}

func (suite *KeeperTestSuite) TestWasmxTime() {
	SkipFixmeTests(suite.T(), "TestWasmxTime")
	SkipCIExpensiveTests(suite.T(), "TestWasmxTime")

	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	wasmbin := precompiles.GetPrecompileByLabel(appA.AddressCodec(), types.TIME_v001)

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "time", nil)

	data := []byte(``)
	msg := &types.WasmxExecutionMessage{Data: data}
	msgbz, err := json.Marshal(msg)
	s.Require().NoError(err)
	_, err = appA.App.WasmxKeeper.ExecuteEntryPoint(appA.Context(), "time", contractAddress, appA.BytesToAccAddressPrefixed(sender.Address), msgbz, nil, false)
	s.Require().NoError(err)

	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	time.Sleep(time.Second * 15)
	// time.Sleep(time.Minute * 10)
}

func (suite *KeeperTestSuite) TestWasmxLevel0() {
	SkipFixmeTests(suite.T(), "TestWasmxLevel0")
	SkipCIExpensiveTests(suite.T(), "TestWasmxLevel0")

	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	timeAddress := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_TIME))
	level0Address := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_LEVEL0))

	// start time chain
	msgexec := types.WasmxExecutionMessage{Data: []byte(`{"StartNode":{}}`)}
	msgbz, err := json.Marshal(&msgexec)
	suite.Require().NoError(err)
	_, err = appA.App.WasmxKeeper.Execute(appA.Context(), timeAddress, appA.BytesToAccAddressPrefixed(sender.Address), msgbz, nil, nil, false)
	suite.Require().NoError(err)

	time.Sleep(time.Second * 10)

	contractAddress := types.AccAddressFromHex(types.ADDR_IDENTITY)
	internalmsg := types.WasmxExecutionMessage{Data: appA.Hex2bz("aa0000000000000000000000000000000000000000000000000000000077")}
	msgbz, err = json.Marshal(internalmsg)
	suite.Require().NoError(err)
	msg := &types.MsgExecuteContract{
		Sender:       appA.MustAccAddressToString(sender.Address),
		Contract:     contractAddress.String(),
		Msg:          msgbz,
		Funds:        nil,
		Dependencies: nil,
	}
	_, err = appA.App.AccountKeeper.GetSequence(appA.Context(), sender.Address)
	suite.Require().NoError(err)
	tx := appA.PrepareCosmosTx(sender, []sdk.Msg{msg}, nil, nil, "")
	txstr := base64.StdEncoding.EncodeToString(tx)

	data := fmt.Sprintf(`{"newTransaction":{"transaction":"%s"}}`, txstr)
	msgexec = types.WasmxExecutionMessage{Data: []byte(data)}
	msgbz, err = json.Marshal(&msgexec)
	suite.Require().NoError(err)
	_, err = appA.App.WasmxKeeper.Execute(appA.Context(), level0Address, appA.BytesToAccAddressPrefixed(sender.Address), msgbz, nil, nil, false)
	suite.Require().NoError(err)
}
