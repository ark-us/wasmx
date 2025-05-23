package keeper_test

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	regi "github.com/loredanacirstea/mythos-tests/utils/httpserver_registry"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	vmhttpserver "github.com/loredanacirstea/wasmx/x/vmhttpserver"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"
)

func (suite *KeeperTestSuite) TestEmailOauth2() {
	if !suite.runOAuth2 {
		suite.T().Skipf("Skipping listen test: TestEmailOauth2")
	}
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, err := utils.DeployDType(suite, appA, sender)
	suite.Require().NoError(err)

	wasmbin := precompiles.GetPrecompileByLabel(appA.AccBech32Codec(), types.EMAIL_v001)
	codeId := appA.StoreCode(sender, wasmbin, nil)

	wasmbinR := precompiles.GetPrecompileByLabel(appA.AccBech32Codec(), types.HTTPSERVER_REGISTRY_v001)
	codeIdR := appA.StoreCode(sender, wasmbinR, nil)
	contractAddressR := appA.InstantiateCode(sender, codeIdR, types.WasmxExecutionMessage{Data: []byte{}}, "httpserver_registry", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "httpserver_registry", contractAddressR, sender)

	msginit := &MsgInitializeRequest{
		Providers: []Provider{
			{
				Name:                  "provable",
				Domain:                "mail.provable.dev",
				ImapServerUrl:         "mail.mail.provable.dev:993",
				SmtpServerUrlStarttls: "mail.mail.provable.dev:587",
				SmtpServerUrlTls:      "mail.mail.provable.dev:465",
			},
		},
		Endpoints: []Endpoint{
			{
				Name:          "google",
				AuthURL:       "https://accounts.google.com/o/oauth2/auth",
				TokenURL:      "https://oauth2.googleapis.com/token",
				DeviceAuthURL: "https://oauth2.googleapis.com/device/code",
				AuthStyle:     0,
				UserInfoUrl:   "https://www.googleapis.com/oauth2/v3/userinfo",
			},
		},
		OAuth2Configs: []OAuth2ConfigToWrite{
			{
				ClientID:     suite.CLIENT_ID_WEB,
				ClientSecret: suite.CLIENT_SECRET_WEB,
				RedirectURL:  "",
				Scopes:       []string{"email", "profile", "https://mail.google.com/"},
				Provider:     suite.provider,
			},
		},
	}
	data, err := json.Marshal(msginit)
	suite.Require().NoError(err)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: data}, "emailtest", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "emailprover", contractAddress, sender)

	msg := &regi.CalldataTestHttpRegistry{
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
	suite.ParseQueryResponse(qres, qresp)
	suite.Require().Equal("", qresp.Error)

	suite.T().Log("Running websrv... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")

	httpmsg := &regi.CalldataTestHttpRegistry{
		Close: &vmhttpserver.CloseRequest{},
	}
	data, err = json.Marshal(httpmsg)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddressR, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp = &vmhttpserver.StartWebServerResponse{}
	suite.ParseQueryResponse(qres, qresp)
	suite.Require().Equal("", qresp.Error)
}
