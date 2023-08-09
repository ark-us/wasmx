package keeper_test

import (
	_ "embed"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestInterpreterPythonSimpleStorage() {
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	pyInterpreterAddress := types.AccAddressFromHex("0x0000000000000000000000000000000000000026")

	data := []byte(`print("hello")`)
	appA.ExecuteContract(sender, pyInterpreterAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// TODO not pyInterpreterAddress
	key := []byte("pystore")
	value := appA.App.WasmxKeeper.QueryRaw(appA.Context(), pyInterpreterAddress, key)
	fmt.Println("value", value)
	s.Require().Equal([]byte("222"), value)
}
