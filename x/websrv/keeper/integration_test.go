package keeper_test

import (
	_ "embed"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmxtypes "wasmx/x/wasmx/types"
	"wasmx/x/websrv/types"
)

var (

	//go:embed testdata/classic/webserver.wasm
	webserver []byte
)

func (suite *KeeperTestSuite) TestWebServerGet() {
	wasmbin := webserver
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

	req := types.HttpRequestGet{Url: &types.RequestUrl{Path: "/"}}
	resp, err := appA.App.WebsrvKeeper.HandleContractRoute(req)
	s.Require().NoError(err)
	s.Require().Equal("Hello from contract. Path: /", resp)

	res = appA.DeliverTx(sender, &types.MsgRegisterRoute{
		Sender:          sender.Address.String(),
		Path:            "/arg1/arg2",
		ContractAddress: contractAddress.String(),
	})
	s.Require().True(res.IsOK(), res.GetLog())
	s.Commit()

	req = types.HttpRequestGet{Url: &types.RequestUrl{Path: "/arg1/arg2"}}
	resp, err = appA.App.WebsrvKeeper.HandleContractRoute(req)
	s.Require().NoError(err)
	s.Require().Equal("Hello from contract. Path: /arg1/arg2", resp)

	req = types.HttpRequestGet{Url: &types.RequestUrl{Path: "/arg1/arg2/arg3"}}
	resp, err = appA.App.WebsrvKeeper.HandleContractRoute(req)
	s.Require().NoError(err)
	s.Require().Equal("Hello from contract. Path: /arg1/arg2/arg3", resp)
}
