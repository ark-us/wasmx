package keeper_test

import (
	_ "embed"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	testdata "mythos/v1/x/wasmx/keeper/testdata/fsm"
	"mythos/v1/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestFSMCounter() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	deps := []string{types.INTERPRETER_FSM}
	codeId := appA.StoreCode(sender, []byte(testdata.Semaphore), deps)
	var data []byte

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{"initialState":"uninitialized","params":[]}}`)}, "simpleStorageJsInterpret", nil)

	data = []byte(`{"getCurrentState":{}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("red"), resp)

	data = []byte(`{"run":{"event":{"type":"next","params":[]}}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(`{"getCurrentState":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("blue"), resp)
}
