package keeper_test

import (
	_ "embed"
	"encoding/json"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	wasmeth "wasmx/v1/x/wasmx/ewasm"
	"wasmx/v1/x/wasmx/keeper/testutil"
	wasmxtypes "wasmx/v1/x/wasmx/types"
	"wasmx/v1/x/websrv/types"
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
	initBalance := sdk.NewInt(1_000_000_000_000_000_000)
	valAccount := simulation.Account{
		PrivKey: s.chainA.SenderPrivKey,
		PubKey:  s.chainA.SenderPrivKey.PubKey(),
		Address: s.chainA.SenderAccount.GetAddress(),
	}

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), valAccount.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	// websrv := websrvserver.NewWebsrvServer(nil, appA.App.Logger(), appA.ClientCtx, config.DefaultWebsrvConfigConfig())

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, wasmxtypes.WasmxExecutionMessage{Data: []byte{}}, "contract with interpreter", nil)

	// Register route proposal
	registerRouteProposal := types.NewRegisterRouteProposal("Register /", "because", "/", contractAddress.String())
	appA.PassGovProposal(valAccount, sender, registerRouteProposal)

	req := types.HttpRequest{Header: []types.HeaderItem{{HeaderType: types.Path_Info, Value: "/"}}}
	response, err := HandleContractRoute(appA, req)
	s.Require().NoError(err)
	s.Require().Equal(types.Content_Type, response.Header[0].HeaderType)
	s.Require().Equal("text/html", response.Header[0].Value)
	s.Require().Equal("Hello from contract. Path: /", string(response.Content))

	// Register route proposal
	registerRouteProposal = types.NewRegisterRouteProposal("Register /arg1/arg2", "because", "/arg1/arg2", contractAddress.String())
	appA.PassGovProposal(valAccount, sender, registerRouteProposal)

	req = types.HttpRequest{Header: []types.HeaderItem{{HeaderType: types.Path_Info, Value: "/arg1/arg2"}}}
	response, err = HandleContractRoute(appA, req)
	s.Require().NoError(err)
	s.Require().Equal(types.Content_Type, response.Header[0].HeaderType)
	s.Require().Equal("text/html", response.Header[0].Value)
	s.Require().Equal("Hello from contract. Path: /arg1/arg2", string(response.Content))

	req = types.HttpRequest{Header: []types.HeaderItem{{HeaderType: types.Path_Info, Value: "/arg1/arg2/arg3"}}}
	response, err = HandleContractRoute(appA, req)
	s.Require().NoError(err)
	s.Require().Equal(types.Content_Type, response.Header[0].HeaderType)
	s.Require().Equal("text/html", response.Header[0].Value)
	s.Require().Equal("Hello from contract. Path: /arg1/arg2/arg3", string(response.Content))
}

func (suite *KeeperTestSuite) TestWebServer() {
	wasmbin := testserverwasm
	sender := suite.GetRandomAccount()
	initBalance := sdk.NewInt(2_000_000_000_000)
	valAccount := simulation.Account{
		PrivKey: s.chainA.SenderPrivKey,
		PubKey:  s.chainA.SenderPrivKey.PubKey(),
		Address: s.chainA.SenderAccount.GetAddress(),
	}

	appA := s.GetAppContext(s.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), valAccount.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin)
	contractAddress := appA.InstantiateCode(sender, codeId, wasmxtypes.WasmxExecutionMessage{Data: []byte{}}, "contract with interpreter", nil)
	contractAddressHex := strings.ToLower(wasmeth.EvmAddressFromAcc(contractAddress).Hex())

	wasmbinRoot := webserverwasm
	codeIdRoot := appA.StoreCode(sender, wasmbinRoot)
	contractAddressRoot := appA.InstantiateCode(sender, codeIdRoot, wasmxtypes.WasmxExecutionMessage{Data: []byte{}}, "contract with interpreter", nil)
	deps := []string{contractAddressHex}

	// Register route proposal
	registerRouteProposal := types.NewRegisterRouteProposal("Register /", "because", "/", contractAddressRoot.String())
	appA.PassGovProposal(valAccount, sender, registerRouteProposal)

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
	response, err := HandleContractRoute(appA, req)
	s.Require().NoError(err)
	s.Require().Equal(types.Content_Type, response.Header[0].HeaderType)
	s.Require().Equal("text/html", response.Header[0].Value)
	s.Require().Equal("Hello from contract. Path: /testserver", string(response.Content))
}

func HandleContractRoute(app testutil.AppContext, httpReq types.HttpRequest) (*types.HttpResponse, error) {
	httpReqBz, err := json.Marshal(httpReq)
	if err != nil {
		return nil, err
	}

	req := &types.QueryHttpRequestGet{HttpRequest: httpReqBz}
	reqResp, err := app.App.WebsrvKeeper.HttpGet(app.Context(), req)
	if err != nil {
		return nil, err
	}

	var response types.HttpResponse
	err = json.Unmarshal(reqResp.Data, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
