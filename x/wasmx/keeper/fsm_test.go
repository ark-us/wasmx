package keeper_test

import (
	_ "embed"
	"encoding/hex"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	testdata "mythos/v1/x/wasmx/keeper/testdata/fsm"
	"mythos/v1/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestFSM_Semaphore() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	deps := []string{types.INTERPRETER_FSM}
	codeId := appA.StoreCode(sender, []byte(testdata.Semaphore), deps)
	var data []byte

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{"initialState":"uninitialized","context":[]}}`)}, "simpleStorageJsInterpret", nil)

	data = []byte(`{"getCurrentState":{}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("#Semaphore.red"), resp)

	data = []byte(`{"run":{"event":{"type":"next","params":[]}}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"getCurrentState":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("#Semaphore.blue"), resp)
}

func (suite *KeeperTestSuite) TestFSM_ERC20() {
	owner := suite.GetRandomAccount()
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(100_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), owner.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	deps := []string{types.INTERPRETER_FSM}
	codeId := appA.StoreCode(sender, []byte(testdata.ERC20), deps)
	var data []byte
	var qres []byte

	contractAddress := appA.InstantiateCode(owner, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{"initialState":"uninitialized","context":[{"key":"admin","value":"010101010101"},{"key":"supply","value":"0"},{"key":"tokenName","value":"Token"},{"key":"tokenSymbol","value":"TKN"}]}}`)}, "simpleStorageJsInterpret", nil)

	data = []byte(`{"getCurrentState":{}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("#ERC20.unlocked.active"), resp)

	data = []byte(fmt.Sprintf(`{"run":{"event":{"type":"mint","params":[{"key":"to","value": "%s"},{"key":"amount","value":"100"}]}}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes()))))
	appA.ExecuteContract(owner, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(fmt.Sprintf(`{"getContextValue":{,"key":"balance_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("100", string(qres))

	data = []byte(fmt.Sprintf(`{"run":{,"event":{"type":"transfer","params":[{"key":"to","value": "%s"},{"key":"amount","value":"10"}]}}}`, hex.EncodeToString(types.PaddLeftTo32(sender.Address.Bytes()))))
	appA.ExecuteContract(owner, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(fmt.Sprintf(`{"getContextValue":{,"key":"balance_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(sender.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("10", string(qres))

	data = []byte(fmt.Sprintf(`{"getContextValue":{,"key":"balance_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("90", string(qres))

	data = []byte(`{"getCurrentState":{}}`)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("#ERC20.unlocked.active", string(qres))

	// Approve tokens
	data = []byte(fmt.Sprintf(`{"run":{,"event":{"type":"approve","params":[{"key":"spender","value": "%s"},{"key":"amount","value":"10"}]}}}`, hex.EncodeToString(types.PaddLeftTo32(sender2.Address.Bytes()))))
	appA.ExecuteContract(owner, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(fmt.Sprintf(`{"getContextValue":{,"key":"allowance_%s_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes())), hex.EncodeToString(types.PaddLeftTo32(sender2.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("10", string(qres))

	// TransferFrom tokens from owner to sender
	data = []byte(fmt.Sprintf(`{"run":{,"event":{"type":"transferFrom","params":[{"key":"from","value": "%s"},{"key":"to","value": "%s"},{"key":"amount","value":"10"}]}}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes())), hex.EncodeToString(types.PaddLeftTo32(sender.Address.Bytes()))))
	appA.ExecuteContract(sender2, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// check admin balance
	data = []byte(fmt.Sprintf(`{"getContextValue":{,"key":"balance_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(owner.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("80", string(qres))

	// check sender balance
	data = []byte(fmt.Sprintf(`{"getContextValue":{,"key":"balance_%s"}}`, hex.EncodeToString(types.PaddLeftTo32(sender.Address.Bytes()))))
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("20", string(qres))
}

func (suite *KeeperTestSuite) TestFSM_Timer() {
	owner := suite.GetRandomAccount()
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(100_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), owner.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	var data []byte
	var qres []byte
	deps := []string{types.INTERPRETER_FSM}
	codeId := appA.StoreCode(owner, []byte(testdata.TimedGrpc), deps)
	contractAddress := appA.InstantiateCode(owner, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"data","value":"aGVsbG8="},{"key":"address","value":"0.0.0.0:8091"}],"initialState":"uninitialized"}}`)}, "stateMachine", nil)

	data = []byte(`{"getCurrentState":{}}`)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal("#AB-Req-Res-timer.active", string(qres))

	data = []byte(`{"run":{"event":{"type":"send","params":[]}}}`)
	appA.ExecuteContract(owner, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// Wait enough time to ensure all goroutines have time to run
	time.Sleep(10 * time.Second)
	fmt.Println("Main function finished")
}
