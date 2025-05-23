package keeper_test

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/mythos-tests/vmhttp/testdata"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/vmhttpserver"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	regi "github.com/loredanacirstea/mythos-tests/utils/httpserver_registry"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"
)

type CalldataTestHttpServer struct {
	StartWebServer     *vmhttpserver.StartWebServerRequest     `json:"StartWebServer"`
	SetRouteHandler    *vmhttpserver.SetRouteHandlerRequest    `json:"SetRouteHandler"`
	RemoveRouteHandler *vmhttpserver.RemoveRouteHandlerRequest `json:"RemoveRouteHandler"`
	Close              *vmhttpserver.CloseRequest              `json:"Close"`
}

func (suite *KeeperTestSuite) TestHttpServer() {
	wasmbin := testdata.WasmxTestHttp
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "httpserver", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "httpserver", contractAddress, sender)

	wasmbinR := precompiles.GetPrecompileByLabel(appA.AccBech32Codec(), types.HTTPSERVER_REGISTRY_v001)
	codeIdR := appA.StoreCode(sender, wasmbinR, nil)
	contractAddressR := appA.InstantiateCode(sender, codeIdR, types.WasmxExecutionMessage{Data: []byte{}}, "httpserver_registry", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "httpserver_registry", contractAddress, sender)

	routemap := map[string]string{}
	routemap["/route1"] = contractAddressR.String()

	msg := &CalldataTestHttpServer{
		StartWebServer: &vmhttpserver.StartWebServerRequest{
			Config: vmhttpserver.WebsrvConfig{
				EnableOAuth:            true,
				Address:                "0.0.0.0:9999",
				CORSAllowedOrigins:     []string{"*"},
				CORSAllowedMethods:     []string{},
				CORSAllowedHeaders:     []string{},
				MaxOpenConnections:     1000,
				RouteToContractAddress: routemap,
				RequestBodyMaxSize:     1000000000,
			},
		}}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp := &vmhttpserver.StartWebServerResponse{}
	suite.parseQueryResponse(qres, qresp)
	suite.Require().Equal("", qresp.Error)

	msg = &CalldataTestHttpServer{
		SetRouteHandler: &vmhttpserver.SetRouteHandlerRequest{
			Route:           "/hello1",
			ContractAddress: contractAddressR.String(),
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp2 := &vmhttpserver.SetRouteHandlerResponse{}
	suite.parseQueryResponse(qres, qresp2)
	suite.Require().Equal("", qresp2.Error)

	msg = &CalldataTestHttpServer{
		SetRouteHandler: &vmhttpserver.SetRouteHandlerRequest{
			Route:           "/hello2",
			ContractAddress: contractAddressR.String(),
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp2 = &vmhttpserver.SetRouteHandlerResponse{}
	suite.parseQueryResponse(qres, qresp2)
	suite.Require().Equal("", qresp2.Error)

	msg = &CalldataTestHttpServer{
		RemoveRouteHandler: &vmhttpserver.RemoveRouteHandlerRequest{
			Route: "/hello2",
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp3 := &vmhttpserver.RemoveRouteHandlerResponse{}
	suite.parseQueryResponse(qres, qresp3)
	suite.Require().Equal("", qresp3.Error)

	suite.T().Log("Running websrv... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")

	msg = &CalldataTestHttpServer{
		Close: &vmhttpserver.CloseRequest{},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp = &vmhttpserver.StartWebServerResponse{}
	suite.parseQueryResponse(qres, qresp)
	suite.Require().Equal("", qresp.Error)
}

func (suite *KeeperTestSuite) TestHttpServerRegistry() {
	wasmbin := testdata.WasmxTestHttp
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, err := utils.DeployDType(suite, appA, sender)
	suite.Require().NoError(err)

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "httpserver", nil)

	// set a role to have access to protected APIs
	// utils.RegisterRole(suite, appA, "httpserver", contractAddress, sender)

	wasmbinR := precompiles.GetPrecompileByLabel(appA.AccBech32Codec(), types.HTTPSERVER_REGISTRY_v001)
	codeIdR := appA.StoreCode(sender, wasmbinR, nil)
	contractAddressR := appA.InstantiateCode(sender, codeIdR, types.WasmxExecutionMessage{Data: []byte{}}, "httpserver_registry", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "httpserver_registry", contractAddressR, sender)

	msg := &regi.CalldataTestHttpRegistry{
		SetRoute: &vmhttpserver.SetRouteHandlerRequest{
			Route:           "/hello1",
			ContractAddress: contractAddress.String(),
		},
	}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddressR, types.WasmxExecutionMessage{Data: data}, nil, nil)

	msg = &regi.CalldataTestHttpRegistry{
		SetRoute: &vmhttpserver.SetRouteHandlerRequest{
			Route:           "/hello2",
			ContractAddress: contractAddressR.String(),
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddressR, types.WasmxExecutionMessage{Data: data}, nil, nil)

	msg = &regi.CalldataTestHttpRegistry{
		SetRoute: &vmhttpserver.SetRouteHandlerRequest{
			Route:           "/hello3",
			ContractAddress: contractAddress.String(),
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddressR, types.WasmxExecutionMessage{Data: data}, nil, nil)

	msg = &regi.CalldataTestHttpRegistry{
		RemoveRoute: &vmhttpserver.RemoveRouteHandlerRequest{
			Route: "/hello3",
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddressR, types.WasmxExecutionMessage{Data: data}, nil, nil)

	msg = &regi.CalldataTestHttpRegistry{
		StartWebServer: &vmhttpserver.StartWebServerRequest{
			Config: vmhttpserver.WebsrvConfig{
				EnableOAuth:        true,
				Address:            "0.0.0.0:9999",
				CORSAllowedOrigins: []string{"*"},
				CORSAllowedMethods: []string{},
				CORSAllowedHeaders: []string{},
				MaxOpenConnections: 1000,
				RequestBodyMaxSize: 1000000000,
			},
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddressR, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp := &vmhttpserver.StartWebServerResponse{}
	suite.parseQueryResponse(qres, qresp)
	suite.Require().Equal("", qresp.Error)

	suite.T().Log("Running websrv... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")

	msg = &regi.CalldataTestHttpRegistry{
		Close: &vmhttpserver.CloseRequest{},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddressR, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp = &vmhttpserver.StartWebServerResponse{}
	suite.parseQueryResponse(qres, qresp)
	suite.Require().Equal("", qresp.Error)
}
