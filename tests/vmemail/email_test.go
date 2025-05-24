package keeper_test

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	imap "github.com/emersion/go-imap/v2"

	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"
)

type Provider struct {
	Id                    string `json:"id"`
	Name                  string `json:"name"`
	Domain                string `json:"domain"`
	ImapServerUrl         string `json:"imap_server_url"`
	SmtpServerUrlStarttls string `json:"smtp_server_url_starttls"`
	SmtpServerUrlTls      string `json:"smtp_server_url_tls"`
}

type Endpoint struct {
	Name          string           `json:"name"`
	AuthURL       string           `json:"auth_url"`
	DeviceAuthURL string           `json:"device_auth_url"`
	TokenURL      string           `json:"token_url"`
	AuthStyle     oauth2.AuthStyle `json:"auth_style"`
	UserInfoUrl   string           `json:"user_info_url"`
}

type MsgResponseDefault struct {
	Error string `json:"error"`
}

type OAuth2ConfigToWrite struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
	Provider     string   `json:"provider"`
}

type Config struct {
	SessionExpirationMs int64  `json:"session_expiration_ms"`
	JWTSecret           []byte `json:"jwt_secret"`
}

type MsgInitializeRequest struct {
	Config        Config                `json:"config"`
	Providers     []Provider            `json:"providers"`
	Endpoints     []Endpoint            `json:"endpoints"`
	OAuth2Configs []OAuth2ConfigToWrite `json:"outh2_configs"`
}

type MsgRegisterProviderRequest struct {
	Providers []Provider `json:"providers"`
	Endpoints []Endpoint `json:"endpoints"`
}

type MsgConnectUserRequest struct {
	Username   string `json:"username"`
	Secret     string `json:"secret"`
	SecretType string `json:"secret_type"`
}

type MsgCacheEmailRequest struct {
	UserId      int64       `json:"user_id"`
	Username    string      `json:"username"`
	EmailFolder string      `json:"email_folder"`
	SeqRange    imap.SeqSet `json:"seq_range"`
	UidRange    imap.UIDSet `json:"uid_range"`
}

type MsgListenEmailRequest struct {
	UserId      int64  `json:"user_id"`
	Username    string `json:"username"`
	EmailFolder string `json:"email_folder"`
}

type MsgSendEmailRequest struct {
	UserId   int64    `json:"user_id"`
	Username string   `json:"username"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
	To       []string `json:"to"`
	Cc       []string `json:"cc"`
	Bcc      []string `json:"bcc"`
}

type CalldataEmailProver struct {
	Initialize        *MsgInitializeRequest       `json:"Initialize"`
	RegisterProviders *MsgRegisterProviderRequest `json:"RegisterProviders"`
	ConnectUser       *MsgConnectUserRequest      `json:"ConnectUser"`
	CacheEmail        *MsgCacheEmailRequest       `json:"CacheEmail"`
	ListenEmail       *MsgListenEmailRequest      `json:"ListenEmail"`
	SendEmail         *MsgSendEmailRequest        `json:"SendEmail"`
}

func (suite *KeeperTestSuite) TestEmail() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	_, err := utils.DeployDType(suite, appA, sender)
	suite.Require().NoError(err)

	wasmbin := precompiles.GetPrecompileByLabel(appA.AccBech32Codec(), types.EMAIL_v001)
	codeId := appA.StoreCode(sender, wasmbin, nil)

	msginit := &MsgInitializeRequest{
		Config: Config{
			SessionExpirationMs: 10000000000,
			JWTSecret:           []byte("jdjfhjfhdskjifjeijklkfjngjnfjksnlkldkadjffskfsd"),
		},
		Providers: []Provider{
			{
				Name:                  "provable",
				Domain:                "mail.provable.dev",
				ImapServerUrl:         "mail.mail.provable.dev:993",
				SmtpServerUrlStarttls: "mail.mail.provable.dev:587",
				SmtpServerUrlTls:      "mail.mail.provable.dev:465",
			},
		},
		Endpoints:     []Endpoint{},
		OAuth2Configs: []OAuth2ConfigToWrite{},
	}
	data, err := json.Marshal(msginit)
	suite.Require().NoError(err)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: data}, "emailtest", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "emailprover", contractAddress, sender)

	msg := &CalldataEmailProver{
		ConnectUser: &MsgConnectUserRequest{
			Username:   suite.emailUsername,
			Secret:     suite.emailPassword,
			SecretType: "password",
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	msg = &CalldataEmailProver{
		CacheEmail: &MsgCacheEmailRequest{
			Username:    suite.emailUsername,
			EmailFolder: "INBOX",
			UidRange:    imap.UIDSetNum(6, 7, 8),
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 1000000000, nil)
}

func (suite *KeeperTestSuite) TestEmailListen() {
	if !suite.runListen {
		suite.T().Skipf("Skipping listen test: TestEmailListen")
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
		Endpoints:     []Endpoint{},
		OAuth2Configs: []OAuth2ConfigToWrite{},
	}
	data, err := json.Marshal(msginit)
	suite.Require().NoError(err)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: data}, "emailtest", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "emailprover", contractAddress, sender)

	msg := &CalldataEmailProver{
		ConnectUser: &MsgConnectUserRequest{
			Username:   suite.emailUsername,
			Secret:     suite.emailPassword,
			SecretType: "password",
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	msg = &CalldataEmailProver{
		ListenEmail: &MsgListenEmailRequest{
			Username:    suite.emailUsername,
			EmailFolder: "INBOX",
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	ctx := vmimap.Context{
		Context: &vmtypes.Context{
			GoRoutineGroup:  appA.App.GetGoRoutineGroup(),
			GoContextParent: appA.App.GetGoContextParent(),
			Ctx:             appA.Context(),
			Logger:          appA.App.AccountKeeper.Logger,
			Env: &types.Env{
				Contract: types.EnvContractInfo{
					Address: contractAddress,
				},
			},
			CosmosHandler: appA.App.WasmxKeeper.NewCosmosHandler(appA.Context(), contractAddress),
		},
	}
	ctx.HandleIncomingEmail(suite.emailUsername, "INBOX", 3, 3)

	suite.T().Log("Running... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")
}
