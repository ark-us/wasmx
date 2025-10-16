package keeper_test

import (
	_ "embed"
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

	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

func (suite *KeeperTestSuite) TestEmailKayrosBot() {
	if !suite.runEmailServer {
		suite.T().Skipf("Skipping email server test: TestEmailKayrosBot")
	}
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	contractAddress := appA.BytesToAccAddressPrefixed(types.AccAddressFromHex(types.ADDR_EMAIL_HANDLER))

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
	fmt.Println(string(data))
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	fmt.Println("started email server")

	// create test account1
	msg = &EmailChainCalldata{
		CreateAccount: &CreateAccountRequest{
			Username: "bot@dmail.provable.dev",
			Password: "123456",
		},
	}
	data, err = json.Marshal(msg)
	fmt.Println(string(data))
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)

	suite.T().Log("Running websrv... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")
}
