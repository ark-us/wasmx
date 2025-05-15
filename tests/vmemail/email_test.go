package keeper_test

import (
	_ "embed"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	imap "github.com/emersion/go-imap/v2"

	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"

	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
)

type Provider struct {
	Id                    string `json:"id"`
	Name                  string `json:"name"`
	Domain                string `json:"domain"`
	ImapServerUrl         string `json:"imap_server_url"`
	SmtpServerUrlStarttls string `json:"smtp_server_url_starttls"`
	SmtpServerUrlTls      string `json:"smtp_server_url_tls"`
}

type MsgResponseDefault struct {
	Error string `json:"error"`
}

type MsgInitializeRequest struct {
	Providers []Provider `json:"providers"`
}

type MsgRegisterProviderRequest struct {
	Provider Provider `json:"provider"`
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
	UserId      Provider `json:"user_id"`
	Username    string   `json:"username"`
	EmailFolder string   `json:"email_folder"`
}

type MsgSendEmailRequest struct {
	UserId   Provider `json:"user_id"`
	Username string   `json:"username"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
	To       []string `json:"to"`
	Cc       []string `json:"cc"`
	Bcc      []string `json:"bcc"`
}

type CalldataEmailProver struct {
	Initialize       *MsgInitializeRequest       `json:"Initialize"`
	RegisterProvider *MsgRegisterProviderRequest `json:"RegisterProvider"`
	ConnectUser      *MsgConnectUserRequest      `json:"ConnectUser"`
	CacheEmail       *MsgCacheEmailRequest       `json:"CacheEmail"`
	ListenEmail      *MsgListenEmailRequest      `json:"ListenEmail"`
	SendEmail        *MsgSendEmailRequest        `json:"SendEmail"`
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
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "emailtest", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "emailprover", contractAddress, sender)

	msg := &CalldataEmailProver{
		Initialize: &MsgInitializeRequest{
			Providers: []Provider{
				{
					Name:                  "provable",
					Domain:                "mail.provable.dev",
					ImapServerUrl:         "mail.mail.provable.dev:993",
					SmtpServerUrlStarttls: "mail.mail.provable.dev:587",
					SmtpServerUrlTls:      "mail.mail.provable.dev:465",
				},
			},
		},
	}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 100000000, nil)

	msg = &CalldataEmailProver{
		ConnectUser: &MsgConnectUserRequest{
			Username:   "test@mail.provable.dev",
			Secret:     "uwsawW3A6**yB^kp",
			SecretType: "password",
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	msg = &CalldataEmailProver{
		CacheEmail: &MsgCacheEmailRequest{
			Username:    "test@mail.provable.dev",
			EmailFolder: "INBOX",
			UidRange:    imap.UIDSetNum(6, 7, 8),
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 1000000000, nil)

	// qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: []byte(`{"name":{}}`)}, nil, nil)
	// var respName struct {
	// 	Name string `json:"name"`
	// }
	// err = json.Unmarshal(qres, &respName)
	// suite.Require().NoError(err, string(qres))
	// suite.Require().Equal("token", respName.Name)
}
