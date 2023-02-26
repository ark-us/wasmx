package keeper_test

import (
	_ "embed"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (

	//go:embed testdata/contract.wasm
	contractcw []byte
)

func (suite *KeeperTestSuite) TestEwasmContractInstantiation() {
	wasmbin := contractcw
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, `{"readonly":false,"data":"0x"}`)

	fmt.Println("contractAddressStr", contractAddress.String())
	suite.Require().True(false)
}
