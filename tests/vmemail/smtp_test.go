package keeper_test

import (
	_ "embed"
	"encoding/json"

	"github.com/emersion/go-imap/v2"
	_ "github.com/mattn/go-sqlite3"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/mythos-tests/vmemail/testdata"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	vmsmtp "github.com/loredanacirstea/wasmx-vmsmtp"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

type CalldataTestSmpt struct {
	ConnectWithPassword *vmsmtp.SmtpConnectionSimpleRequest `json:"ConnectWithPassword"`
	ConnectOAuth2       *vmsmtp.SmtpConnectionOauth2Request `json:"ConnectOAuth2"`
	Close               *vmsmtp.SmtpCloseRequest            `json:"Close"`
	Quit                *vmsmtp.SmtpQuitRequest             `json:"Quit"`
	Extension           *vmsmtp.SmtpExtensionRequest        `json:"Extension"`
	Noop                *vmsmtp.SmtpNoopRequest             `json:"Noop"`
	SendMail            *vmsmtp.SmtpSendMailRequest         `json:"SendMail"`
	Verify              *vmsmtp.SmtpVerifyRequest           `json:"Verify"`
	SupportsAuth        *vmsmtp.SmtpSupportsAuthRequest     `json:"SupportsAuth"`
	MaxMessageSize      *vmsmtp.SmtpMaxMessageSizeRequest   `json:"MaxMessageSize"`
}

func (suite *KeeperTestSuite) TestSmtp() {
	wasmbin := testdata.WasmxTestSmtp
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "smtptest", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "someemailrole", contractAddress, sender)

	msg := &CalldataTestSmpt{
		ConnectWithPassword: &vmsmtp.SmtpConnectionSimpleRequest{
			Id:                    "conn1",
			SmtpServerUrlSTARTTLS: "mail.mail.provable.dev:587",
			Username:              suite.emailUsername,
			Password:              suite.emailPassword,
		}}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resc := &vmsmtp.SmtpConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resc)
	suite.Require().NoError(err)
	suite.Require().Equal("", resc.Error)

	msg = &CalldataTestSmpt{
		Verify: &vmsmtp.SmtpVerifyRequest{
			Id:      "conn1",
			Address: "test",
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp := &vmsmtp.SmtpVerifyResponse{}
	err = json.Unmarshal(qres, qresp)
	suite.Require().NoError(err)
	suite.Require().Contains(qresp.Error, "SMTP error 252")

	email := vmsmtp.Email{
		Envelope: &imap.Envelope{
			Subject: "Hello",
			From: []imap.Address{
				{Host: "mail.provable.dev", Mailbox: "test", Name: "Test"},
			},
			To: []imap.Address{
				{Host: "gmail.com", Mailbox: "seth.one.info", Name: "Seth One"},
			},
		},
		Header:      make(map[string][]string, 0),
		Body:        "Some content",
		Attachments: []vmsmtp.Attachment{},
	}

	emailraw, err := vmsmtp.BuildRawEmail(email)
	suite.Require().NoError(err)

	msg = &CalldataTestSmpt{
		SendMail: &vmsmtp.SmtpSendMailRequest{
			Id:    "conn1",
			From:  "test@mail.provable.dev",
			To:    []string{"seth.one.info@gmail.com"},
			Email: []byte(emailraw),
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	ress := &vmsmtp.SmtpSendMailResponse{}
	err = appA.DecodeExecuteResponse(res, ress)
	suite.Require().NoError(err)
	suite.Require().Equal("", ress.Error)

	msg = &CalldataTestSmpt{
		Close: &vmsmtp.SmtpCloseRequest{
			Id: "conn1",
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rescl := &vmsmtp.SmtpCloseResponse{}
	err = appA.DecodeExecuteResponse(res, rescl)
	suite.Require().NoError(err)
	suite.Require().Equal("", rescl.Error)
}
