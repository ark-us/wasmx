package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	imap "github.com/emersion/go-imap/v2"

	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	vmsmtp "github.com/loredanacirstea/wasmx-vmsmtp"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	tinygo "github.com/loredanacirstea/mythos-tests/testdata/tinygo"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
)

type BuildAndSendMailRequest struct {
	Id      string   `json:"id"`
	From    string   `json:"from"`
	To      []string `json:"to"`
	Cc      []string `json:"cc"`
	Bcc     []string `json:"bcc"`
	Subject string   `json:"subject"`
	Body    []byte   `json:"body"`
}

type CalldataTestSmptTinygo struct {
	CalldataTestSmpt
	BuildAndSend *BuildAndSendMailRequest `json:"BuildAndSend,omitempty"`
}

func (suite *KeeperTestSuite) TestEmailTinygoImap() {
	wasmbin := tinygo.ImapTestWrapSdk
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "EmailTestWrapSdk", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "someemailrole", contractAddress, sender)

	msg := &Calldata{
		ConnectWithPassword: &vmimap.ImapConnectionSimpleRequest{
			Id:            "conn1",
			ImapServerUrl: "mail.mail.provable.dev:993",
			Username:      suite.emailUsername,
			Password:      suite.emailPassword,
		}}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resc := &vmimap.ImapConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resc)
	suite.Require().NoError(err)
	suite.Require().Equal("", resc.Error)

	msg = &Calldata{
		ListMailboxes: &vmimap.ListMailboxesRequest{
			Id: "conn1",
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qrespm := &vmimap.ListMailboxesResponse{}
	err = json.Unmarshal(qres, qrespm)
	suite.Require().NoError(err)
	suite.Require().Equal(qrespm.Error, "")
	suite.Require().Greater(len(qrespm.Mailboxes), 1)
	suite.Require().Contains(qrespm.Mailboxes, "INBOX")

	msg = &Calldata{
		Fetch: &vmimap.ImapFetchRequest{
			Id:          "conn1",
			Folder:      "INBOX",
			SeqSet:      imap.SeqSetNum(1),
			UidSet:      make(imap.UIDSet, 0),
			FetchFilter: nil,
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp := &vmimap.ImapFetchResponse{}
	err = json.Unmarshal(qres, qresp)
	suite.Require().NoError(err)
	suite.Require().Equal(qresp.Error, "")
	suite.Require().Equal(1, len(qresp.Data))

	msg = &Calldata{
		CreateFolder: &vmimap.ImapCreateFolderRequest{
			Id:   "conn1",
			Path: "INBOX2/mysubfolder",
		}}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rescf := &vmimap.ImapCreateFolderResponse{}
	err = appA.DecodeExecuteResponse(res, rescf)
	suite.Require().NoError(err)
	if rescf.Error != "" {
		fmt.Println(rescf.Error)
	}
	// suite.Require().Equal("", rescf.Error)

	msg = &Calldata{
		Close: &vmimap.ImapCloseRequest{
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

func (suite *KeeperTestSuite) TestEmailTinyGoSmtp() {
	wasmbin := tinygo.SmtpTestWrapSdk
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "smtptest", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "someemailrole", contractAddress, sender)

	msg := &CalldataTestSmpt{}
	if suite.isOAuth2 {
		msg.ConnectOAuth2 = &vmsmtp.SmtpConnectionOauth2Request{
			Id:                    "conn1",
			SmtpServerUrlSTARTTLS: "smtp.gmail.com:587",
			Username:              suite.emailUsername,
			AccessToken:           suite.emailPassword,
		}
	} else {
		msg.ConnectWithPassword = &vmsmtp.SmtpConnectionSimpleRequest{
			Id:                    "conn1",
			SmtpServerUrlSTARTTLS: "mail.mail.provable.dev:587",
			Username:              suite.emailUsername,
			Password:              suite.emailPassword,
		}
	}

	data, err := json.Marshal(msg)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resc := &vmsmtp.SmtpConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resc)
	suite.Require().NoError(err)
	suite.Require().Equal("", resc.Error)

	msg2 := &CalldataTestSmptTinygo{
		BuildAndSend: &BuildAndSendMailRequest{
			Id:      "conn1",
			From:    suite.emailUsername,
			To:      []string{"seth.one.info@gmail.com"},
			Body:    []byte("Some content"),
			Subject: "hello test smtp email tinygo",
		}}
	data, err = json.Marshal(msg2)
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
