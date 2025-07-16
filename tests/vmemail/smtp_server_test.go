package keeper_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	vmsmtp "github.com/loredanacirstea/wasmx-vmsmtp"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"

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
	suite.Commit()

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
	wasmbin := tinygo.EmailChain
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	// Store the emailchain contract and instantiate it
	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "emailchain", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "emailprover", contractAddress, sender)

	tlsPath := "/etc/letsencrypt/live/"

	// Prepare the VerifyDKIM request
	msg := &EmailChainCalldata{
		StartServer: &StartServerRequest{
			SignOptions: SignOptions{
				Domain:         "provable.dev",
				Selector:       "dmail",
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
	res := appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)

	// // Prepare the VerifyDKIM request
	// msg = &EmailChainCalldata{
	// 	SendEmail: &BuildAndSendMailRequest{
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
