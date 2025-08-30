package keeper_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	vmsmtp "github.com/loredanacirstea/wasmx-vmsmtp"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	imap "github.com/emersion/go-imap/v2"

	tinygo "github.com/loredanacirstea/mythos-tests/testdata/tinygo"
	"github.com/loredanacirstea/mythos-tests/vmemail/testdata"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

func (suite *KeeperTestSuite) TestIncomingEmail() {
	defer os.Remove("emailchain.db")
	defer os.Remove("emailchain.db-shm")
	defer os.Remove("emailchain.db-wal")
	wasmbin := tinygo.EmailChain
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	// Store the emailchain contract and instantiate it
	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "emailchain", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "emailprover", contractAddress, sender)

	raw, err := base64.StdEncoding.DecodeString(testdata.IncomingEmailRaw)
	suite.Require().NoError(err)
	msg := &EmailChainCalldata{
		IncomingEmail: &vmsmtp.Session{
			From:     []string{"test@mail.provable.dev"},
			To:       []string{"test@dmail.provable.dev", "test2@dmail.provable.dev"},
			EmailRaw: raw,
		},
	}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)

	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
}

func (suite *KeeperTestSuite) TestEmailSmtpServer() {
	if !suite.runEmailServer {
		suite.T().Skipf("Skipping email server test: TestEmailSmtpServer")
	}
	wasmbin := tinygo.EmailChain
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	// Store the emailchain contract and instantiate it
	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "emailchain", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "emailprover", contractAddress, sender)

	// tlsPath := "~/dev/letsencrypt/"
	tlsPath := "/etc/letsencrypt/live/"

	// Prepare the VerifyDKIM request
	msg := &EmailChainCalldata{
		StartServer: &StartServerRequest{
			SignOptions: SignOptions{
				Domain:         "dmail.provable.dev",
				Selector:       "2025a",
				PrivateKeyType: "rsa",
				PrivateKey:     []byte(testPrivateKeyPEM),
				Identifier:     "",
			},
			Smtp: vmsmtp.ServerConfig{
				Network: "tcp4",
				Domain:  "dmail.provable.dev",
				TlsConfig: &vmsmtp.TlsConfig{
					TLSCertFile: tlsPath + "dmail.provable.dev/fullchain.pem",
					TLSKeyFile:  tlsPath + "dmail.provable.dev/privkey.pem",
					ServerName:  "dmail.provable.dev",
				},
			},
			Imap: vmimap.ServerConfig{
				TlsConfig: &vmimap.TlsConfig{
					TLSCertFile: tlsPath + "dmail.provable.dev/fullchain.pem",
					TLSKeyFile:  tlsPath + "dmail.provable.dev/privkey.pem",
					ServerName:  "dmail.provable.dev",
				},
				Network: "tcp4",
			},
		},
	}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	fmt.Println("started email server")

	var res *abci.ExecTxResult

	// create test account1
	msg = &EmailChainCalldata{
		CreateAccount: &CreateAccountRequest{
			Username: "test@dmail.provable.dev",
			Password: "123456",
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)

	// create test account2
	msg = &EmailChainCalldata{
		CreateAccount: &CreateAccountRequest{
			Username: "test2@dmail.provable.dev",
			Password: "123456",
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)

	// create test account3
	msg = &EmailChainCalldata{
		CreateAccount: &CreateAccountRequest{
			Username: "test3@dmail.provable.dev",
			Password: "123456",
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)

	// send email from account1 to account2
	msg = &EmailChainCalldata{
		BuildAndSend: &BuildAndSendMailRequest{
			From:    "test@dmail.provable.dev",
			To:      []string{"test2@dmail.provable.dev"},
			Subject: "what a day!",
			Body:    []byte(`Just demoed a provable forwarding protocol today!`),
			Date:    time.Now(),
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)

	// wait for email to be received
	time.Sleep(time.Second * 3)

	// forward email from account2 to account3
	msg = &EmailChainCalldata{
		ForwardEmail: &ForwardEmailRequest{
			From:              AddressFromString("test2@dmail.provable.dev", "Test2 Test2"),
			To:                []imap.Address{AddressFromString("test3@dmail.provable.dev", "Test3 Test3")},
			Folder:            "INBOX",
			Uid:               1,
			Timestamp:         time.Now(),
			AdditionalSubject: "whoop!",
			SendEmail:         true,
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	res = appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 958152876, nil)
	resc := &ForwardEmailResponse{}
	err = appA.DecodeExecuteResponse(res, resc)
	suite.Require().NoError(err)
	suite.Require().Equal("", resc.Error)

	// wait for email to be received
	time.Sleep(time.Second * 3)

	// forward email from account3 to gmail
	msg = &EmailChainCalldata{
		ForwardEmail: &ForwardEmailRequest{
			From: AddressFromString("test3@dmail.provable.dev", "Test3 Test3"),
			// To:                []imap.Address{AddressFromString("test@dmail.provable.dev", "Test Test")},
			To:                []imap.Address{AddressFromString("seth.one.info@gmail.com", "Seth Info")},
			Folder:            "INBOX",
			Uid:               1,
			Timestamp:         time.Now(),
			AdditionalSubject: "not bad!",
			SendEmail:         true,
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	res = appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 958152876, nil)
	resc = &ForwardEmailResponse{}
	err = appA.DecodeExecuteResponse(res, resc)
	suite.Require().NoError(err)
	suite.Require().Equal("", resc.Error)

	// wait for email to be received
	time.Sleep(time.Second * 3)

	// forward fake email from account3 to account1
	msg = &EmailChainCalldata{
		ForwardEmail: &ForwardEmailRequest{
			From:              AddressFromString("test3@dmail.provable.dev", "Test3 Test3"),
			To:                []imap.Address{AddressFromString("test@dmail.provable.dev", "Test Test")},
			Folder:            "INBOX",
			Uid:               1,
			Timestamp:         time.Now(),
			AdditionalSubject: "fake forward",
			SendEmail:         false,
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	res = appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 958152876, nil)
	resc = &ForwardEmailResponse{}
	err = appA.DecodeExecuteResponse(res, resc)
	suite.Require().NoError(err)
	suite.Require().Equal("", resc.Error)

	emailraw := strings.Replace(resc.EmailRaw, "provable forwarding protocol", "provable forwarding protocol [modified]", 1)

	// send email from account1 to account2
	msg = &EmailChainCalldata{
		SendEmail: &SendMailRequest{
			From:     AddressFromString("test3@dmail.provable.dev", "Test3 Test3"),
			To:       []imap.Address{AddressFromString("test@dmail.provable.dev", "Test Test")},
			EmailRaw: []byte(emailraw),
			Date:     time.Now(),
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)

	// // Prepare the VerifyDKIM request
	// msg = &EmailChainCalldata{
	// 	BuildAndSend: &BuildAndSendMailRequest{
	// 		From: "test@dmail.provable.dev",
	// 		// To:      []string{"seth.one.info@gmail.com"},
	// 		To:      []string{"test@mail.provable.dev"},
	// 		Subject: "this is a subject",
	// 		Body:    []byte(`hei hei hei`),
	// 		Date:    time.Now(),
	// 	},
	// }
	// data, err = json.Marshal(msg)
	// suite.Require().NoError(err)
	// res = appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	// fmt.Println("--send email--", string(res.Data))

	suite.T().Log("Running websrv... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")
}
