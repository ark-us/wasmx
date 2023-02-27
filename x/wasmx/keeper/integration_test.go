package keeper_test

import (
	_ "embed"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmeth "wasmx/x/wasmx/keeper/ewasm"
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
	contractAddress := appA.InstantiateCode(sender, codeId, wasmeth.WasmEthMessage{Readonly: false, Data: []byte{}})

	res := appA.ExecuteContract(sender, contractAddress, wasmeth.WasmEthMessage{Readonly: false, Data: []byte{11, 22, 33, 44, 55}}, nil)
	suite.Require().Equal("0000000000000000000000000000000000000000000000000000000000000005", hex.EncodeToString(res.Data))
}
