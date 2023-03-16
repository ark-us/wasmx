package keeper_test

import (
	_ "embed"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmxtypes "wasmx/x/wasmx/types"
	"wasmx/x/websrv/types"
)

var (
	//go:embed testdata/classic/webserver.wasm
	webserverwasm []byte

	//go:embed testdata/classic/mythos.wasm
	mythoswasm []byte
)

func (suite *KeeperTestSuite) TestSimpleWebServer() {
	wasmbin := mythoswasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, wasmxtypes.WasmxExecutionMessage{Data: []byte{}}, "contract with interpreter", nil)

	res := appA.DeliverTx(sender, &types.MsgRegisterRoute{
		Sender:          sender.Address.String(),
		Path:            "/",
		ContractAddress: contractAddress.String(),
	})
	s.Require().True(res.IsOK(), res.GetLog())
	s.Commit()

	req := types.HttpRequest{Header: []types.HeaderItem{{HeaderType: types.HeaderOption_PathInfo, Value: "/"}}}
	response, err := appA.App.WebsrvKeeper.HandleContractRoute(req)
	s.Require().NoError(err)
	// s.Require().Equal("text/html", contentType)
	s.Require().Equal("Hello from contract. Path: /", string(response.Content))

	res = appA.DeliverTx(sender, &types.MsgRegisterRoute{
		Sender:          sender.Address.String(),
		Path:            "/arg1/arg2",
		ContractAddress: contractAddress.String(),
	})
	s.Require().True(res.IsOK(), res.GetLog())
	s.Commit()

	req = types.HttpRequest{Header: []types.HeaderItem{{HeaderType: types.HeaderOption_PathInfo, Value: "/arg1/arg2"}}}
	response, err = appA.App.WebsrvKeeper.HandleContractRoute(req)
	s.Require().NoError(err)
	// s.Require().Equal("text/html", contentType)
	s.Require().Equal("Hello from contract. Path: /arg1/arg2", string(response.Content))

	req = types.HttpRequest{Header: []types.HeaderItem{{HeaderType: types.HeaderOption_PathInfo, Value: "/arg1/arg2/arg3"}}}
	response, err = appA.App.WebsrvKeeper.HandleContractRoute(req)
	s.Require().NoError(err)
	// s.Require().Equal("text/html", contentType)
	s.Require().Equal("Hello from contract. Path: /arg1/arg2/arg3", string(response.Content))
}
