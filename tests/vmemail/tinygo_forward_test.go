package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/emersion/go-imap/v2"

	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	tinygo "github.com/loredanacirstea/mythos-tests/testdata/tinygo"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

func (suite *KeeperTestSuite) TestEmailTinyGoForwardCustom() {
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

	msg := &EmailChainCalldata{
		ConnectWithPassword: &ConnectionSimpleRequest{
			Id:                    "conn1",
			ImapServerUrl:         "mail.mail.provable.dev:993",
			SmtpServerUrlSTARTTLS: "mail.mail.provable.dev:587",
			Username:              suite.emailUsername,
			Password:              suite.emailPassword,
		},
	}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resc := &vmimap.ImapConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resc)
	suite.Require().NoError(err)
	suite.Require().Equal("", resc.Error)

	// publicKey := &dkim.PublicKey{
	// 	Version:    "DKIM1",
	// 	KeyType:    "rsa",
	// 	Algorithms: []string{"rsa-sha256"},
	// 	Revoked:    false,
	// 	Testing:    false,
	// 	Strict:     false,
	// 	Services:   []string{"email"},
	// 	Key:        &testPrivateKey.PublicKey,
	// 	Data:       testPrivateKey.PublicKey.N.Bytes(),
	// }

	fromParts := strings.Split(suite.emailUsername, "@")
	fromAddr := imap.Address{Mailbox: fromParts[0], Host: fromParts[1]}

	msg = &EmailChainCalldata{
		ForwardEmail: &ForwardEmailRequest{
			ConnectionId: "conn1",
			Folder:       "INBOX",
			From:         fromAddr,
			To:           []imap.Address{{Mailbox: "seth.one.info", Host: "gmail.com"}},
			MessageId:    "CADMWPsW+XVDARBqGcdPLcv_K68WMMcnDE2UueSKDcwsd-EpADQ@mail.gmail.com",
			Options: SignOptions{
				// PrivateKeyType: "ed25519",
				// PrivateKey:     testEd25519PrivateKey,
				PrivateKeyType: "rsa",
				PrivateKey:     []byte(testPrivateKeyPEM),
			},
			Timestamp: time.Unix(424242, 0),
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)

	res = appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 380000000, nil)
	fmt.Println("--ForwardEmail--", string(res.Data))
	resp3 := &ForwardEmailResponse{}
	err = appA.DecodeExecuteResponse(res, resp3)
	suite.Require().NoError(err)
	fmt.Println("=============ForwardEmail EmailRaw")
	fmt.Println(resp3.EmailRaw)
	fmt.Println("=============END ForwardEmail EmailRaw")
	suite.Require().Equal(resp3.Error, "")
	// suite.Require().Equal(signedMailString, resp3.EmailRaw)

	// resDKIM, _, err := verifyEmail(resp3.SignedEmail, publicKey)
	// suite.Require().NoError(err)
	// suite.Require().Equal(1, len(resDKIM))
	// fmt.Println("--resDKIM.Error--", resDKIM[0].Error)
	// fmt.Println("--resDKIM--", *resDKIM[0])
	// suite.Require().Nil(resDKIM[0].Error)
	// suite.Require().Equal(dkim.Pass, resDKIM[0].Code)

	// // sign ARC
	// msg = &EmailChainCalldata{
	// 	SignARC: &SignARCRequest{
	// 		EmailRaw: resp3.SignedEmail,
	// 		Options: SignOptions{
	// 			Domain:   "example.org",
	// 			Selector: "brisbane",
	// 			// PrivateKeyType: "ed25519",
	// 			// PrivateKey:     testEd25519PrivateKey,
	// 			PrivateKeyType: "rsa",
	// 			PrivateKey:     []byte(testPrivateKeyPEM),
	// 		},
	// 		Timestamp: time.Unix(424242, 0),
	// 	},
	// }
	// data, err = json.Marshal(msg)
	// suite.Require().NoError(err)

	// res = appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 380000000, nil)
	// fmt.Println("--SignARCResponse--", string(res.Data))
	// resp4 := &SignARCResponse{}
	// err = appA.DecodeExecuteResponse(res, resp4)
	// suite.Require().NoError(err)
	// suite.Require().Equal(resp4.Error, "")
	// fmt.Println("===============SignARCResponse SignedEmail--")
	// fmt.Println(resp4.SignedEmail)
	// fmt.Println("===============END SignARCResponse SignedEmail--")
	// // suite.Require().Greater(len(resp4.SignedEmail), 128)

	// // verify signed email
	// resDKIM, resARC, err := verifyEmail(resp4.SignedEmail, publicKey)
	// suite.Require().NoError(err)
	// suite.Require().Equal(1, len(resDKIM))
	// fmt.Println("--resDKIM.Error--", resDKIM[0].Error)
	// fmt.Println("--resDKIM--", *resDKIM[0])
	// suite.Require().Nil(resDKIM[0].Error)
	// suite.Require().Equal(dkim.Pass, resDKIM[0].Code)
	// suite.Require().NoError(resARC.Error)
	// suite.Require().Equal(dkim.Pass, resARC.Code)
	// suite.Require().Equal(1, len(resARC.Chain))
	// suite.Require().True(resARC.Chain[0].AMSValid)
	// suite.Require().True(resARC.Chain[0].ASValid)
	// suite.Require().Equal("pass", resARC.Chain[0].CV)
	// suite.Require().Equal("pass", resARC.Chain[0].Dkim)
	// suite.Require().Equal("pass", resARC.Chain[0].Dmarc)
	// suite.Require().Equal("pass", resARC.Chain[0].Spf)
	// suite.Require().Equal(1, resARC.Chain[0].Instance)

	msg = &EmailChainCalldata{
		Close: &CloseRequest{
			Id: "conn1",
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rescl := &vmimap.ImapCloseResponse{}
	err = appA.DecodeExecuteResponse(res, rescl)
	suite.Require().NoError(err)
	suite.Require().Equal("", rescl.Error)
}
