package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	testdata "mythos/v1/x/wasmx/keeper/testdata/classic"
	"mythos/v1/x/wasmx/types"
	"mythos/v1/x/wasmx/vm/precompiles"
	vmtypes "mythos/v1/x/wasmx/vm/types"
)

var (
	//go:embed testdata/wasmx/simple_storage.wasm
	wasmxSimpleStorage []byte

	//go:embed testdata/wasmx/state_machine.wasm
	wasmxStateMachine []byte
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
	initBalance := sdkmath.NewInt(1000_000_000)

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
	codeId2, contractAddress2 := appA.DeployEvm(sender, evmcode, types.WasmxExecutionMessage{Data: initvaluebz}, nil, "simpleStorage", nil)
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
		Sender: sender.Address.String(),
		CodeId: codeId,
		Label:  "label",
		Msg:    msgbz,
		Funds:  nil,
	}
	res, err = appA.DeliverTxWithOpts(sender, instantiateContractMsg, 5000000, nil)
	s.Require().NoError(err)
	suite.Require().True(res.IsErr(), res.GetLog())
	suite.Require().Contains(res.GetLog(), "invalid address for system contracts")
}

func (suite *KeeperTestSuite) TestWasmxSimpleStorage() {
	wasmbin := wasmxSimpleStorage
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "simpleStorage", nil)

	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	wasmlogs := appA.GetWasmxEvents(res.GetEvents())
	emptyDataLogs := appA.GetEventsByAttribute(wasmlogs, "data", "0x")
	topicLogs := appA.GetEventsByAttribute(wasmlogs, "topic", "0x68656c6c6f000000000000000000000000000000000000000000000000000000")
	s.Require().Equal(1, len(wasmlogs), res.GetEvents())
	s.Require().Equal(1, len(emptyDataLogs), res.GetEvents())
	s.Require().Equal(1, len(topicLogs), res.GetEvents())

	initvalue := "sammy"
	keybz := []byte("hello")
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(initvalue, string(queryres))

	data = []byte(`{"get":{"key":"hello"}}`)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(string(qres), "sammy")
}

func (suite *KeeperTestSuite) TestWasmxStateMachineERC20() {
	wasmbin := wasmxStateMachine
	owner := suite.GetRandomAccount()
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(100_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), owner.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	fmt.Println("---TestWasmxStateMachineERC20---")
	codeId := appA.StoreCode(owner, wasmbin, nil)
	fmt.Println("---TestWasmxStateMachineERC20-InstantiateCode--")
	contractAddress := appA.InstantiateCode(owner, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "stateMachine", nil)

	config := `{"context":[{"key":"admin","value":"010101010101"},{"key":"supply","value":"0"},{"key":"tokenName","value":"Token"},{"key":"tokenSymbol","value":"TKN"}],"id":"ERC20","initial":"uninitialized","states":[{"name":"uninitialized","on":[{"name":"initialize","target":"unlocked","guard":"","actions":[]}],"exit":[],"entry":[],"initial":"","states":[]},{"name":"unlocked","on":[{"name":"lock","target":"locked","guard":"isAdmin","actions":[]}],"exit":[],"entry":[],"initial":"active","states":[{"name":"active","on":[{"name":"transfer","target":"intransfer","guard":"","actions":[{"type":"xstate.raise","event":{"type":"move","params":[{"key":"from","value":"getCaller()"},{"key":"to","value":"to"},{"key":"amount","value":"amount"}]},"params":[]}]},{"name":"mint","target":"active","guard":"isAdmin","actions":[{"type":"mint","params":[]}]},{"name":"approve","target":"active","guard":"","actions":[{"type":"approve","params":[]}]},{"name":"transferFrom","target":"intransfer","guard":"hasEnoughAllowance","actions":[{"type":"xstate.raise","event":{"type":"move","params":[{"key":"from","value":"from"},{"key":"to","value":"to"},{"key":"amount","value":"amount"}]},"params":[]}]}],"exit":[],"entry":[],"initial":"","states":[]},{"name":"intransfer","on":[],"exit":[],"entry":[],"initial":"unmoved","states":[{"name":"unmoved","on":[{"name":"move","target":"moved","guard":"hasEnoughBalance","actions":[{"type":"move","params":[]},{"type":"logTransfer","params":[]},{"type":"xstate.raise","event":{"type":"finish","params":[]},"params":[]}]}],"exit":[],"entry":[],"initial":"","states":[]},{"name":"moved","on":[{"name":"finish","target":"#ERC20.unlocked.active","guard":"","actions":[]}],"exit":[],"entry":[],"initial":"","states":[]}]}]},{"name":"locked","on":[{"name":"unlock","target":"unlocked","guard":"isAdmin","actions":[]}],"exit":[],"entry":[],"initial":"","states":[]}]}`

	data := []byte(fmt.Sprintf(`{"create":%s}`, config))
	res := appA.ExecuteContract(owner, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Contains(string(res.Data), `{"id":0}`)

	data = []byte(`{"getCurrentState":{"id":0}}`)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("active", string(qres))

	data = []byte(fmt.Sprintf(`{"run":{"id":0,"event":{"type":"mint","params":[{"key":"to","value": "%s"},{"key":"amount","value":"100"}]}}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes()))))
	appA.ExecuteContract(owner, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(fmt.Sprintf(`{"getContextValue":{"id":0,"key":"balance_0_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("100", string(qres))

	data = []byte(fmt.Sprintf(`{"run":{"id":0,"event":{"type":"transfer","params":[{"key":"to","value": "%s"},{"key":"amount","value":"10"}]}}}`, hex.EncodeToString(types.PaddLeftTo32(sender.Address.Bytes()))))
	appA.ExecuteContract(owner, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(fmt.Sprintf(`{"getContextValue":{"id":0,"key":"balance_0_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(sender.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("10", string(qres))

	data = []byte(fmt.Sprintf(`{"getContextValue":{"id":0,"key":"balance_0_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("90", string(qres))

	data = []byte(`{"getCurrentState":{"id":0}}`)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("#ERC20.unlocked.active", string(qres))

	// Approve tokens
	data = []byte(fmt.Sprintf(`{"run":{"id":0,"event":{"type":"approve","params":[{"key":"spender","value": "%s"},{"key":"amount","value":"10"}]}}}`, hex.EncodeToString(types.PaddLeftTo32(sender2.Address.Bytes()))))
	appA.ExecuteContract(owner, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(fmt.Sprintf(`{"getContextValue":{"id":0,"key":"allowance_0_%s_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes())), hex.EncodeToString(types.PaddLeftTo32(sender2.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("10", string(qres))

	// TransferFrom tokens from owner to sender
	data = []byte(fmt.Sprintf(`{"run":{"id":0,"event":{"type":"transferFrom","params":[{"key":"from","value": "%s"},{"key":"to","value": "%s"},{"key":"amount","value":"10"}]}}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes())), hex.EncodeToString(types.PaddLeftTo32(sender.Address.Bytes()))))
	appA.ExecuteContract(sender2, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// check admin balance
	data = []byte(fmt.Sprintf(`{"getContextValue":{"id":0,"key":"balance_0_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("80", string(qres))

	// check sender balance
	data = []byte(fmt.Sprintf(`{"getContextValue":{"id":0,"key":"balance_0_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(sender.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("20", string(qres))
}

func (suite *KeeperTestSuite) TestWasmxStateMachineTimer() {
	wasmbin := wasmxStateMachine
	owner := suite.GetRandomAccount()
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(100_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), owner.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	fmt.Println("---TestWasmxStateMachineTimer---")
	codeId := appA.StoreCode(owner, wasmbin, nil)
	fmt.Println("---TestWasmxStateMachineTimer-InstantiateCode--")
	contractAddress := appA.InstantiateCode(owner, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "stateMachine", nil)

	config := `{"context":[{"key":"data","value":"hello"},{"key":"address","value":"0.0.0.0:8091"}],"id":"AB-Req-Res-timer","initial":"uninitialized","states":[{"name":"uninitialized","after":[],"on":[{"name":"initialize","target":"active","guard":"","actions":[]}],"exit":[],"entry":[],"initial":"","states":[]},{"name":"active","after":[],"on":[{"name":"receiveRequest","target":"IReceivedTheRequest","guard":"","actions":[]},{"name":"send","target":"sender","guard":"","actions":[]}],"exit":[],"entry":[],"initial":"","states":[]},{"name":"IReceivedTheRequest","after":[{"name":"1000","target":"#AB-Req-Res-timer.active","guard":"","actions":[]}],"on":[],"exit":[],"entry":[],"initial":"","states":[]},{"name":"sender","after":[{"name":"5000","target":"#AB-Req-Res-timer.sending request","guard":"","actions":[{"type":"xstate.raise","event":{"type":"sendRequest","params":[{"key":"data","value":"data"},{"key":"address","value":"address"}]},"params":[]}]}],"on":[],"exit":[],"entry":[],"initial":"","states":[]},{"name":"sending request","after":[],"on":[{"name":"sendRequest","target":"sender","guard":"","actions":[{"type":"sendRequest","params":[]}]}],"exit":[],"entry":[],"initial":"","states":[]}]}`

	data := []byte(fmt.Sprintf(`{"create":%s}`, config))
	res := appA.ExecuteContract(owner, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Contains(string(res.Data), `{"id":0}`)

	data = []byte(`{"getCurrentState":{"id":0}}`)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("active", string(qres))

	fmt.Println("---TestWasmxStateMachineTimer-run--")

	data = []byte(`{"run":{"id":0,"event":{"type":"send","params":[]}}}`)
	appA.ExecuteContract(owner, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// Wait enough time to ensure all goroutines have time to run
	time.Sleep(20 * time.Second)
	fmt.Println("Main function finished")
}
