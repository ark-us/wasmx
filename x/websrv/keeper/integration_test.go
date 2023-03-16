package keeper_test

import (
	_ "embed"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmeth "wasmx/x/wasmx/ewasm"
	wasmxtypes "wasmx/x/wasmx/types"
	"wasmx/x/websrv/types"
)

var (
	//go:embed testdata/classic/webserver.wasm
	webserverwasm []byte

	//go:embed testdata/classic/webserver_test.wasm
	testserverwasm []byte
)

func (suite *KeeperTestSuite) TestSimpleWebServer() {
	wasmbin := testserverwasm
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

	req := types.HttpRequest{Header: []types.HeaderItem{{HeaderType: types.Path_Info, Value: "/"}}}
	response, err := appA.App.WebsrvKeeper.HandleContractRoute(req)
	s.Require().NoError(err)
	s.Require().Equal(types.Content_Type, response.Header[0].HeaderType)
	s.Require().Equal("text/html", response.Header[0].Value)
	s.Require().Equal("Hello from contract. Path: /", string(response.Content))

	res = appA.DeliverTx(sender, &types.MsgRegisterRoute{
		Sender:          sender.Address.String(),
		Path:            "/arg1/arg2",
		ContractAddress: contractAddress.String(),
	})
	s.Require().True(res.IsOK(), res.GetLog())
	s.Commit()

	req = types.HttpRequest{Header: []types.HeaderItem{{HeaderType: types.Path_Info, Value: "/arg1/arg2"}}}
	response, err = appA.App.WebsrvKeeper.HandleContractRoute(req)
	s.Require().NoError(err)
	s.Require().Equal(types.Content_Type, response.Header[0].HeaderType)
	s.Require().Equal("text/html", response.Header[0].Value)
	s.Require().Equal("Hello from contract. Path: /arg1/arg2", string(response.Content))

	req = types.HttpRequest{Header: []types.HeaderItem{{HeaderType: types.Path_Info, Value: "/arg1/arg2/arg3"}}}
	response, err = appA.App.WebsrvKeeper.HandleContractRoute(req)
	s.Require().NoError(err)
	s.Require().Equal(types.Content_Type, response.Header[0].HeaderType)
	s.Require().Equal("text/html", response.Header[0].Value)
	s.Require().Equal("Hello from contract. Path: /arg1/arg2/arg3", string(response.Content))
}

func (suite *KeeperTestSuite) TestWebServer() {
	wasmbin := testserverwasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(1000_000_000)

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, wasmxtypes.WasmxExecutionMessage{Data: []byte{}}, "contract with interpreter", nil)
	contractAddressHex := strings.ToLower(wasmeth.EvmAddressFromAcc(contractAddress).Hex())

	wasmbinRoot := webserverwasm
	codeIdRoot := appA.StoreCode(sender, wasmbinRoot)
	contractAddressRoot := appA.InstantiateCode(sender, codeIdRoot, wasmxtypes.WasmxExecutionMessage{Data: []byte{}}, "contract with interpreter", nil)
	deps := []string{contractAddressHex}

	res := appA.DeliverTx(sender, &types.MsgRegisterRoute{
		Sender:          sender.Address.String(),
		Path:            "/",
		ContractAddress: contractAddressRoot.String(),
	})
	s.Require().True(res.IsOK(), res.GetLog())
	s.Commit()

	resp, err := appA.App.WebsrvKeeper.ContractByRoute(appA.Context(), &types.QueryContractByRouteRequest{Path: "/"})
	s.Require().NoError(err)
	s.Require().Equal(contractAddressRoot.String(), resp.ContractAddress)

	handler := appA.App.WebsrvKeeper.GetMostSpecificRouteToContract(appA.Context(), "/")
	s.Require().Equal(contractAddressRoot.String(), handler.String())
	handler = appA.App.WebsrvKeeper.GetMostSpecificRouteToContract(appA.Context(), "/testserver")
	s.Require().Equal(contractAddressRoot.String(), handler.String())

	// setPage /testserver
	appA.ExecuteContract(sender, contractAddressRoot, wasmxtypes.WasmxExecutionMessage{Data: appA.Hex2bz("baafbf770000000000000000000000000000000000000000000000000000000000000040000000000000000000000000" + contractAddressHex[2:] + "000000000000000000000000000000000000000000000000000000000000000b2f74657374736572766572000000000000000000000000000000000000000000")}, nil, deps)

	// query pages /testserver
	qres := appA.EwasmQuery(sender, contractAddressRoot, wasmxtypes.WasmxExecutionMessage{Data: appA.Hex2bz("918a4fd40000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000b2f74657374736572766572000000000000000000000000000000000000000000")}, nil, nil)
	s.Require().Equal("000000000000000000000000"+contractAddressHex[2:], qres)

	req := types.HttpRequest{Header: []types.HeaderItem{{HeaderType: types.Path_Info, Value: "/testserver"}}}
	response, err := appA.App.WebsrvKeeper.HandleContractRoute(req)
	s.Require().NoError(err)
	s.Require().Equal(types.Content_Type, response.Header[0].HeaderType)
	s.Require().Equal("text/html", response.Header[0].Value)
	s.Require().Equal("Hello from contract. Path: /testserver", string(response.Content))
}
