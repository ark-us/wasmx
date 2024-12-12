package keeper_test

import (
	_ "embed"
	"encoding/json"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	testutil "github.com/loredanacirstea/wasmx/v1/testutil/wasmx"
	wasmxtypes "github.com/loredanacirstea/wasmx/v1/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/v1/x/websrv/types"
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
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	// websrv := websrvserver.NewWebsrvServer(nil, appA.App.Logger(), appA.ClientCtx, config.DefaultWebsrvConfigConfig())

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, wasmxtypes.WasmxExecutionMessage{Data: []byte{}}, "contract with interpreter", nil)

	// Register route proposal
	title := "Register /"
	description := "because"
	registerRouteProposal := &types.MsgRegisterRoute{Title: title, Description: description, Path: "/", ContractAddress: contractAddress.String()}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{registerRouteProposal}, "", title, description, false)

	req := types.HttpRequest{Header: []types.HeaderItem{{HeaderType: types.Path_Info, Value: "/"}}}
	response, err := HandleContractRoute(appA, req)
	s.Require().NoError(err)
	s.Require().Equal(types.Content_Type, response.Header[0].HeaderType)
	s.Require().Equal("text/html", response.Header[0].Value)
	s.Require().Equal("Hello from contract. Path: /", string(response.Content))

	// Register route proposal
	title = "Register /arg1/arg2"
	description = "because"
	registerRouteProposal = &types.MsgRegisterRoute{Title: title, Description: description, Path: "/arg1/arg2", ContractAddress: contractAddress.String()}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{registerRouteProposal}, "", title, description, false)

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
	initBalance := sdkmath.NewInt(2_000_000_000_000)
	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, wasmxtypes.WasmxExecutionMessage{Data: []byte{}}, "contract with interpreter", nil)
	contractAddressHex := strings.ToLower(wasmxtypes.EvmAddressFromAcc(contractAddress.Bytes()).Hex())

	wasmbinRoot := webserverwasm
	codeIdRoot := appA.StoreCode(sender, wasmbinRoot, nil)
	contractAddressRoot := appA.InstantiateCode(sender, codeIdRoot, wasmxtypes.WasmxExecutionMessage{Data: []byte{}}, "contract with interpreter", nil)
	deps := []string{contractAddressHex}

	// Register route proposal
	title := "Register /"
	description := "because"
	registerRouteProposal := &types.MsgRegisterRoute{Title: title, Description: description, Path: "/", ContractAddress: contractAddressRoot.String()}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{registerRouteProposal}, "", title, description, false)

	resp, err := appA.App.WebsrvKeeper.ContractByRoute(appA.Context(), &types.QueryContractByRouteRequest{Path: "/"})
	s.Require().NoError(err)
	s.Require().Equal(contractAddressRoot.String(), resp.ContractAddress)

	handler := appA.App.WebsrvKeeper.GetMostSpecificRouteToContract(appA.Context(), "/")
	s.Require().Equal(contractAddressRoot.String(), appA.MustAccAddressToString(handler))
	handler = appA.App.WebsrvKeeper.GetMostSpecificRouteToContract(appA.Context(), "/testserver")
	s.Require().Equal(contractAddressRoot.String(), appA.MustAccAddressToString(handler))

	// setPage /testserver
	appA.ExecuteContract(sender, contractAddressRoot, wasmxtypes.WasmxExecutionMessage{Data: appA.Hex2bz("baafbf770000000000000000000000000000000000000000000000000000000000000040000000000000000000000000" + contractAddressHex[2:] + "000000000000000000000000000000000000000000000000000000000000000b2f74657374736572766572000000000000000000000000000000000000000000")}, nil, deps)

	// query pages /testserver
	qres := appA.WasmxQuery(sender, contractAddressRoot, wasmxtypes.WasmxExecutionMessage{Data: appA.Hex2bz("918a4fd40000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000b2f74657374736572766572000000000000000000000000000000000000000000")}, nil, nil)
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
