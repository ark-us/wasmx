package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	imap "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-msgauth/dkim"

	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	vmsmtp "github.com/loredanacirstea/wasmx-vmsmtp"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	tinygo "github.com/loredanacirstea/mythos-tests/testdata/tinygo"
	testdata "github.com/loredanacirstea/mythos-tests/vmemail/testdata"
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

type VerifyDKIMTestRequest struct {
	VerifyDKIM *VerifyDKIMTestData `json:"VerifyDKIM,omitempty"`
}

type VerifyDKIMTestData struct {
	EmailRaw string `json:"email_raw"`
}

type DKIMVerification struct {
	Domain     string   `json:"domain"`
	Selector   string   `json:"selector"`
	Valid      bool     `json:"valid"`
	Error      string   `json:"error,omitempty"`
	Algorithm  string   `json:"algorithm"`
	HeaderKeys []string `json:"header_keys"`
}

type VerifyDKIMResponse struct {
	Error         string             `json:"error"`
	Verifications []DKIMVerification `json:"verifications"`
	IsValid       bool               `json:"is_valid"`
}

func (suite *KeeperTestSuite) TestEmailTinyGoDKIM() {
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

	// Define email input for DKIM verification
	emailRaw := testdata.Email1

	// Prepare the VerifyDKIM request
	msg := &VerifyDKIMTestRequest{
		VerifyDKIM: &VerifyDKIMTestData{
			EmailRaw: emailRaw,
		},
	}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)

	// first verify email
	err = verifyEmail(emailRaw)
	suite.Require().NoError(err)

	// Execute the VerifyDKIM message
	res := appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 80000000, nil)
	resp := &VerifyDKIMResponse{}
	err = appA.DecodeExecuteResponse(res, resp)
	suite.Require().NoError(err)
	suite.Require().Equal(resp.Error, "")
	suite.Require().Greater(len(resp.Verifications), 0)
	fmt.Println("--verifications--", resp.Verifications)

	// Check that at least one verification is valid
	hasValidSignature := false
	for _, v := range resp.Verifications {
		if v.Valid && v.Error == "" {
			hasValidSignature = true
		}
	}
	suite.Require().True(hasValidSignature, "At least one DKIM signature should be valid")
	suite.Require().True(resp.IsValid, "Overall DKIM verification should be valid")
}

func verifyEmail(emailText string) error {
	// Convert the email string to an io.Reader
	reader := strings.NewReader(emailText)

	// Verify DKIM signatures
	verifications, err := dkim.Verify(reader)
	if err != nil {
		return err
	}
	verificationsbz, err := json.Marshal(verifications)
	fmt.Println("verifications", string(verificationsbz))

	// Process verification results
	for _, v := range verifications {
		if v.Err == nil {
			fmt.Printf("DKIM signature verified successfully for domain: %s\n", v.Domain)
		} else {
			fmt.Printf("DKIM verification failed for domain: %s: %v\n", v.Domain, v.Err)
			return v.Err
		}
	}
	return nil
}
