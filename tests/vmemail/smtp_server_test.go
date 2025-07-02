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

	vmsmtp "github.com/loredanacirstea/wasmx-vmsmtp"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	tinygo "github.com/loredanacirstea/mythos-tests/testdata/tinygo"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

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

	// Prepare the VerifyDKIM request
	msg := &EmailChainCalldata{
		StartServer: &vmsmtp.ServerConfig{
			Addr:        ":25",
			Domain:      "dmail.provable.dev",
			TLSCertFile: "/etc/letsencrypt/live/dmail.provable.dev/fullchain.pem",
			TLSKeyFile:  "/etc/letsencrypt/live/dmail.provable.dev/privkey.pem",
		},
	}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)

	// Execute the VerifyDKIM message
	res := appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	resp := &vmsmtp.ServerStartResponse{}
	fmt.Println("--DKIM result--", string(res.Data))
	err = appA.DecodeExecuteResponse(res, resp)
	suite.Require().NoError(err)
	// suite.Require().Equal(resp.Error, "")
	// suite.Require().Greater(len(resp.Response), 0)
	// suite.Require().NoError(resp.Response[0].Err)
	// suite.Require().Equal("pass", string(resp.Response[0].Status), "DKIM result not pass")

	suite.T().Log("Running websrv... Press Ctrl+C to exit")

	// Create a channel to listen for interrupt/terminate signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	suite.T().Log("Received exit signal. Test ending.")
}
