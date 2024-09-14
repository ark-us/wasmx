package keeper_test

import (
	_ "embed"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	testdata "mythos/v1/x/wasmx/keeper/testdata/taylor"
	"mythos/v1/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestInterpreterTaySimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.SimpleStorageTay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{}}`)}, "SimpleContractTay", nil)

	key := []byte("hello")
	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), contractAddress, key)
	s.Require().Equal([]byte("sammy"), value)

	data = []byte(`{"get":{"key":"hello"}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("sammy"), resp)
}

func (suite *KeeperTestSuite) TestInterpreterTayERC20() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	deps := []string{types.INTERPRETER_TAY}
	codeId := appA.StoreCode(sender, []byte(testdata.ERC20Tay), deps)

	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte(`{"instantiate":{"name":"Token","symbol":"TKN","decimals":8}}`)}, "ERC20Tay", nil)

	data := []byte(`{"getName":{}}`)
	resp := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("Token"), resp)

	data = []byte(`{"getSymbol":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal([]byte("TKN"), resp)

	data = []byte(`{"getDecimals":{}}`)
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(8, resp)

	data = []byte(`{"mint":{}}`)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	data = []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, senderPrefixed.String()))
	resp = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	s.Require().Equal(1000, resp)
}
