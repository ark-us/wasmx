package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testdata "mythos/v1/x/wasmx/keeper/testdata/classic"
	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/precompiles"
	vmtypes "mythos/v1/x/wasmx/vm/types"
)

var (
	//go:embed testdata/wasmx/simple_storage.wasm
	wasmxSimpleStorage []byte
)

type SysContract struct {
	Benchmark *BenchmarkRequest `json:"benchmark"`
}

type BenchmarkRequest struct {
	Request   vmtypes.CallRequest `json:"request"`
	Magnitude int32               `json:"magnitude"`
}

func (suite *KeeperTestSuite) TestWasmxBenchmark() {
	wasmbin := precompiles.GetPrecompileByLabel("sys_proxy")
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	sysAddressBz, err := hex.DecodeString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	s.Require().NoError(err)
	sysAddress := sdk.AccAddress(sysAddressBz)

	// deploy an evm contract
	evmcode, err := hex.DecodeString(testdata.SimpleStorage)
	s.Require().NoError(err)
	initvalue := "0000000000000000000000000000000000000000000000000000000000000009"
	initvaluebz, err := hex.DecodeString(initvalue)
	s.Require().NoError(err)
	codeId2, contractAddress2 := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage")
	codeInfo := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId2)
	s.Require().NotNil(codeInfo)

	getHex := `6d4ce63c`

	// an EOA can make a system call by query
	req := &SysContract{
		Benchmark: &BenchmarkRequest{
			Magnitude: 3,
			Request: vmtypes.CallRequest{
				To:       contractAddress2,
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
	suite.Require().True(elapsed.Cmp(big.NewInt(5)) == 0 || elapsed.Cmp(big.NewInt(6)) == 0)

	// an EOA cannot make a system call by tx
	res := appA.ExecuteContractNoCheck(sender, sysAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 1000000, nil)
	suite.Require().True(res.IsErr())

	// a contract cannot make a system call
	evmcode, err = hex.DecodeString(testdata.Call)
	s.Require().NoError(err)
	_, callAddress := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: []byte{}}, nil, "callwasm")
	msg := types.WasmxExecutionMessage{Data: append(sysAddress.Bytes(), data...)}
	res = appA.ExecuteContractNoCheck(sender, callAddress, msg, nil, nil, 1000000, nil)
	suite.Require().True(res.IsErr())

	// cannot deploy a system contract
	codeId := appA.StoreCode(sender, wasmbin, nil)
	msgbz, err := json.Marshal(types.WasmxExecutionMessage{Data: []byte{}})
	s.Require().NoError(err)
	instantiateContractMsg := &types.MsgInstantiateContract{
		Sender: sender.Address.String(),
		CodeId: codeId,
		Label:  "label",
		Msg:    msgbz,
		Funds:  nil,
	}
	res = appA.DeliverTxWithOpts(sender, instantiateContractMsg, 5000000, nil)
	suite.Require().True(res.IsErr(), res.GetLog())
	suite.Require().Contains(res.GetLog(), "invalid address for system contracts")
}

func (suite *KeeperTestSuite) TestWasmxSimpleStorage() {
	wasmbin := wasmxSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	logCount := strings.Count(res.GetLog(), `{"key":"type","value":"wasmx"}`)
	dataCount := strings.Count(res.GetLog(), `{"key":"data","value":"0x"}`)
	topicCount := strings.Count(res.GetLog(), `{"key":"topic","value":"0x68656c6c6f000000000000000000000000000000000000000000000000000000"}`)
	s.Require().Equal(1, logCount, res.GetLog())
	s.Require().Equal(1, dataCount, res.GetLog())
	s.Require().Equal(1, topicCount, res.GetLog())

	initvalue := "sammy"
	keybz := []byte("hello")
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, string(queryres))

	data = []byte(`{"get":{"key":"hello"}}`)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(string(qres), "sammy")
}
