package keeper_test

// import (
// 	_ "embed"
// 	"encoding/hex"
// 	"wasmx/x/wasmx/types"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// )

// var (

// 	//go:embed testdata/contract.wasm
// 	contractcw []byte
// )

// func (suite *KeeperTestSuite) TestEwasmContract() {
// 	wasmbin := contractcw
// 	sender := suite.GetRandomAccount()
// 	initBalance := sdk.NewInt(1000_000_000)

// 	appA := s.GetAppContext(s.chainA)
// 	appA.faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.denom, initBalance))
// 	suite.Commit()

// 	codeId := appA.StoreCode(sender, wasmbin)
// 	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "contract with interpreter")

// 	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte{11, 22, 33, 44, 55}}, nil, nil)
// 	suite.Require().Equal("124d0a272f7761736d782e7761736d782e4d736745786563757465436f6e7472616374526573706f6e736512220a200000000000000000000000000000000000000000000000000000000000000005", hex.EncodeToString(res.Data))
// }
