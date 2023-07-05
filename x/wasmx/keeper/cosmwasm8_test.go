package keeper_test

import (
	_ "embed"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

var (
	//go:embed testdata/cw8/simple_contract.wasm
	cwSimpleContract []byte
)

func (suite *KeeperTestSuite) TestWasmxSimpleContract() {
	wasmbin := cwSimpleContract
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	expectedDeps := []string{types.CW_ENV_8}

	codeId := appA.StoreCode(sender, wasmbin, nil)
	codeInfo := appA.App.WasmxKeeper.GetCodeInfo(appA.Context(), codeId)
	s.Require().ElementsMatch(expectedDeps, codeInfo.Deps, "wrong deps")

	value := 2
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{}`)}, "cwSimpleContract", nil)

	data := []byte(`{"increase":{}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	value += 1

	keybz := []byte("counter")
	queryres := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, keybz)
	suite.Require().Equal(fmt.Sprintf("%d", value), string(queryres))

	data = []byte(`{"value":{}}`)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(string(qres), value)
}
