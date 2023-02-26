package keeper_test

import (
	_ "embed"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (

	//go:embed testdata/contract.wasm
	contractcw []byte
)

func (suite *KeeperTestSuite) TestEwasmContract() {
	wasmbin := contractcw
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, `{"readonly":false,"data":"0x"}`)

	res := appA.ExecuteContract(sender, contractAddress, `{"readonly":false,"data":"0x1122334455"}`, nil)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", hex.EncodeToString(res.Data))
}
