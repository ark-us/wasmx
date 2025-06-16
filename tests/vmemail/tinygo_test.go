package keeper_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	imap "github.com/emersion/go-imap/v2"
	"github.com/stretchr/testify/require"

	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	vmsmtp "github.com/loredanacirstea/wasmx-vmsmtp"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	tinygo "github.com/loredanacirstea/mythos-tests/testdata/tinygo"
	testdata "github.com/loredanacirstea/mythos-tests/vmemail/testdata"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"

	dkimS "github.com/emersion/go-msgauth/dkim"
	dkim "github.com/redsift/dkim"
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

type EmailChainCalldata struct {
	ConnectWithPassword *ConnectionSimpleRequest `json:"ConnectWithPassword,omitempty"`
	ConnectOAuth2       *ConnectionOauth2Request `json:"ConnectOAuth2,omitempty"`
	Close               *CloseRequest            `json:"Close,omitempty"`
	VerifyDKIM          *VerifyDKIMTestData      `json:"VerifyDKIM,omitempty"`
	VerifyARC           *VerifyARCTestData       `json:"VerifyARC,omitempty"`
	SignDKIM            *SignDKIMRequest         `json:"SignDKIM,omitempty"`
	SignARC             *SignARCRequest          `json:"SignARC,omitempty"`
	ForwardEmail        *ForwardEmailRequest     `json:"ForwardEmail,omitempty"`
}

type ConnectionSimpleRequest struct {
	Id                    string `json:"id"`
	ImapServerUrl         string `json:"imap_server_url"`
	SmtpServerUrlSTARTTLS string `json:"smtp_server_url_starttls"`
	SmtpServerUrlTLS      string `json:"smtp_server_url_tls"`
	Username              string `json:"username"`
	Password              string `json:"password"`
}

type ConnectionOauth2Request struct {
	Id                    string `json:"id"`
	ImapServerUrl         string `json:"imap_server_url"`
	SmtpServerUrlSTARTTLS string `json:"smtp_server_url_starttls"`
	SmtpServerUrlTLS      string `json:"smtp_server_url_tls"`
	Username              string `json:"username"`
	AccessToken           string `json:"access_token"`
}

type CloseRequest struct {
	Id string `json:"id"`
}

type VerifyDKIMTestData struct {
	EmailRaw  string          `json:"email_raw"`
	PublicKey *dkim.PublicKey `json:"public_key,omitempty"`
}

type DKIMVerification struct {
	Domain     string   `json:"domain"`
	Selector   string   `json:"selector"`
	Valid      bool     `json:"valid"`
	Error      string   `json:"error,omitempty"`
	Algorithm  string   `json:"algorithm"`
	HeaderKeys []string `json:"header_keys"`
}

type VerifyARCTestData struct {
	EmailRaw  string          `json:"email_raw"`
	PublicKey *dkim.PublicKey `json:"public_key,omitempty"`
}

type VerifyDKIMResponse struct {
	Error    string   `json:"error"`
	Response []Result `json:"response"`
}

type VerifyARCResponse struct {
	Error    string     `json:"error"`
	Response *ArcResult `json:"response"`
}

type ArcResult struct {
	Result Result `json:"result"`
	// Result data at each part of the chain until failure
	Chain []ArcSetResult `json:"chain"`
}

type ArcSetResult struct {
	Instance int    `json:"instance"`
	Spf      string `json:"spf"`
	Dkim     string `json:"dkim"`
	Dmarc    string `json:"dmarc"`
	AMSValid bool   `json:"ams-vaild"`
	ASValid  bool   `json:"as-valid"`
	CV       string `json:"cv"`
}

type Result struct { // size=64 (0x40)
	Order     int                     `json:"order"`
	Code      string                  `json:"code"`
	Error     *dkim.VerificationError `json:"error,omitempty"`
	Signature *Signature              `json:"signature,omitempty"`
	Key       *dkim.PublicKey         `json:"key,omitempty"`
	Timestamp time.Time               `json:"timestamp"`
}

type Signature struct {
	Header         string            `json:"header"`                  // Header of the signature
	Raw            string            `json:"raw"`                     // Raw value of the signature
	AlgorithmID    string            `json:"algorithmId"`             // 3 (SHA1) or 5 (SHA256)
	Hash           []byte            `json:"hash"`                    // 'h' tag value
	BodyHash       []byte            `json:"bodyHash"`                // 'bh' tag value
	RelaxedHeader  bool              `json:"relaxedHeader"`           // header canonicalization algorithm
	RelaxedBody    bool              `json:"relaxedBody"`             // body canonicalization algorithm
	SignerDomain   string            `json:"signerDomain"`            // 'd' tag value
	Headers        []string          `json:"headers"`                 // parsed 'h' tag value
	UserIdentifier string            `json:"userId"`                  // 'i' tag value
	ArcInstance    int               `json:"arcInstance"`             // 'i' tag value (only in arc headers)
	Length         int64             `json:"length"`                  // 'l' tag value
	Selector       string            `json:"selector"`                // 's' tag value
	Timestamp      time.Time         `json:"ts"`                      // 't' tag value as time.Time
	Expiration     time.Time         `json:"exp"`                     // 'x' tag value as time.Time
	CopiedHeaders  map[string]string `json:"copiedHeaders,omitempty"` // parsed 'z' tag value

	// Arc related fields
	ArcCV string `json:"arcCv"` // 'cv' tag, chain validation value for arc seal
	Spf   string `json:"spf"`   // spf value for ARC-Authentication-Results
	Dmarc string `json:"dmarc"` // dmarc value for ARC-Authentication-Results
	Dkim  string `json:"dkim"`  // dkim value for ARC-Authentication-Results
}

type SignOptions struct {
	// The SDID claiming responsibility for an introduction of a message into the
	// mail stream. Hence, the SDID value is used to form the query for the public
	// key. The SDID MUST correspond to a valid DNS name under which the DKIM key
	// record is published.
	//
	// This can't be empty.
	Domain string `json:"domain"`
	// The selector subdividing the namespace for the domain.
	//
	// This can't be empty.
	Selector string `json:"selector"`
	// The Agent or User Identifier (AUID) on behalf of which the SDID is taking
	// responsibility.
	//
	// This is optional.
	Identifier string `json:"identifier"`

	// The key used to sign the message.
	//
	// Supported Signer.Public() values are *rsa.PublicKey and
	// ed25519.PublicKey.
	PrivateKeyType string `json:"private_key_type"` // rsa or ed25519
	PrivateKey     []byte `json:"private_key"`

	// The hash algorithm used to sign the message. If zero, a default hash will
	// be chosen.
	//
	// The only supported hash algorithm is crypto.SHA256.
	Hash uint `json:"hash"`

	// Header and body canonicalization algorithms.
	//
	// If empty, CanonicalizationSimple is used.
	HeaderCanonicalization dkimS.Canonicalization `json:"header_canonicalization"`
	BodyCanonicalization   dkimS.Canonicalization `json:"body_canonicalization"`

	// A list of header fields to include in the signature. If nil, all headers
	// will be included. If not nil, "From" MUST be in the list.
	//
	// See RFC 6376 section 5.4.1 for recommended header fields.
	HeaderKeys []string `json:"header_keys"`

	// The expiration time. A zero value means no expiration.
	Expiration time.Time `json:"expiration"`

	// A list of query methods used to retrieve the public key.
	//
	// If nil, it is implicitly defined as QueryMethodDNSTXT.
	QueryMethods []dkimS.QueryMethod `json:"query_methods"`
}

type SignDKIMRequest struct {
	EmailRaw  string      `json:"email_raw"`
	Options   SignOptions `json:"options"`
	Timestamp time.Time   `json:"timestamp"`
}

type SignARCRequest struct {
	EmailRaw  string      `json:"email_raw"`
	Options   SignOptions `json:"options"`
	Timestamp time.Time   `json:"timestamp"`
}

type SignDKIMResponse struct {
	Error       string `json:"error"`
	SignedEmail string `json:"signed_email"`
}

type SignARCResponse struct {
	Error       string `json:"error"`
	SignedEmail string `json:"signed_email"`
}

type ForwardEmailRequest struct {
	ConnectionId string         `json:"connection_id"`
	Folder       string         `json:"folder"`
	Uid          uint32         `json:"uid"`
	MessageId    string         `json:"message_id"`
	From         imap.Address   `json:"from"`
	To           []imap.Address `json:"to"`
	Options      SignOptions    `json:"options"`
	Timestamp    time.Time      `json:"timestamp"`
	SendEmail    bool           `json:"send_email"`
}

type ForwardEmailResponse struct {
	Error    string `json:"error"`
	EmailRaw string `json:"email_raw"`
}

var now = func() time.Time {
	return time.Unix(424242, 0)
}

func TestSignARCSync(t *testing.T) {
	options := &dkimS.SignOptions{
		Domain:    "example.org",
		Selector:  "brisbane",
		Signer:    testPrivateKey,
		LookupTXT: net.LookupTXT,
	}

	pubk := &dkim.PublicKey{
		Version:    "DKIM1",
		KeyType:    "rsa",
		Algorithms: []string{"rsa-sha256"},
		Revoked:    false,
		Testing:    false,
		Strict:     false,
		Services:   []string{"email"},
		Key:        &testPrivateKey.PublicKey,
		Data:       testPrivateKey.PublicKey.N.Bytes(),
	}

	emailStr := testdata.EmailARC3
	// emailStr := signedMailString
	newemail, resARC := ARCSignAndVerify(t, emailStr, options, "seth.one.info@gmail.com", "209.85.214.177", pubk)

	require.Nil(t, resARC.Error)
	require.Equal(t, dkim.Pass, resARC.Code)
	require.Equal(t, 1, len(resARC.Chain))

	require.True(t, resARC.Chain[0].AMSValid)
	require.True(t, resARC.Chain[0].ASValid)
	require.Equal(t, 1, resARC.Chain[0].Instance)
	require.Equal(t, "none", resARC.Chain[0].CV.String())
	require.Equal(t, "fail", resARC.Chain[0].Dkim.String())
	require.Equal(t, "fail", resARC.Chain[0].Dmarc.String())
	require.Equal(t, "fail", resARC.Chain[0].Spf.String())

	// sign another time
	_, resARC = ARCSignAndVerify(t, newemail, options, "test@provable.dev", "85.215.130.119", pubk)
	require.NotNil(t, resARC.Error)
	require.Contains(t, resARC.Error.Error(), "ARC-Seal reported failure, the chain is terminated")
	require.Equal(t, dkim.Fail, resARC.Code)
	require.Equal(t, 2, len(resARC.Chain))

	require.True(t, resARC.Chain[0].AMSValid)
	require.False(t, resARC.Chain[0].ASValid)
	require.Equal(t, 2, resARC.Chain[0].Instance)
	require.Equal(t, "fail", resARC.Chain[0].CV.String())
	require.Equal(t, "fail", resARC.Chain[0].Dkim.String())
	require.Equal(t, "fail", resARC.Chain[0].Dmarc.String())
	require.Equal(t, "fail", resARC.Chain[0].Spf.String())

	require.True(t, resARC.Chain[1].AMSValid)
	require.True(t, resARC.Chain[1].ASValid)
	require.Equal(t, 1, resARC.Chain[1].Instance)
	require.Equal(t, "none", resARC.Chain[1].CV.String())
	require.Equal(t, "fail", resARC.Chain[1].Dkim.String())
	require.Equal(t, "fail", resARC.Chain[1].Dmarc.String())
	require.Equal(t, "fail", resARC.Chain[1].Spf.String())
}

func ARCSignAndVerify(t *testing.T, emailStr string, options *dkimS.SignOptions, mailfrom string, ip string, pubk *dkim.PublicKey) (string, *dkim.ArcResult) {
	r := strings.NewReader(emailStr)
	var b bytes.Buffer
	err := dkimS.SignARCSync(&b, r, options, now, mailfrom, ip)
	require.NoError(t, err, "Expected no error while signing mail")

	signedEmail := b.String()

	fmt.Println("=======ARCSigned")
	fmt.Println(signedEmail)
	fmt.Println("=======END ARCSigned")

	msg, err := dkim.ParseMessage(signedEmail)
	require.NoError(t, err)

	resARC, err := dkim.VerifyArc(net.LookupTXT, pubk, msg)
	require.NoError(t, err)
	return signedEmail, resARC
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
	emailRaw := testdata.EmailARC3

	// Prepare the VerifyDKIM request
	msg := &EmailChainCalldata{
		VerifyDKIM: &VerifyDKIMTestData{
			EmailRaw: emailRaw,
		},
	}
	data, err := json.Marshal(msg)
	suite.Require().NoError(err)

	// first verify email
	_, _, err = verifyEmail(emailRaw, nil)
	suite.Require().NoError(err)

	// Execute the VerifyDKIM message
	res := appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 80000000, nil)
	resp := &VerifyDKIMResponse{}
	fmt.Println("--DKIM result--", string(res.Data))
	err = appA.DecodeExecuteResponse(res, resp)
	suite.Require().NoError(err)
	suite.Require().Equal(resp.Error, "")
	suite.Require().Greater(len(resp.Response), 0)
	suite.Require().Equal("pass", resp.Response[0].Code, "DKIM result not pass")

	// ARC
	msg = &EmailChainCalldata{
		VerifyARC: &VerifyARCTestData{
			EmailRaw: emailRaw,
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)

	// first verify email
	_, _, err = verifyEmail(emailRaw, nil)
	suite.Require().NoError(err)

	res = appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 380000000, nil)
	fmt.Println("--ARC result--", string(res.Data))
	resp2 := &VerifyARCResponse{}
	err = appA.DecodeExecuteResponse(res, resp2)
	suite.Require().NoError(err)
	suite.Require().Equal(resp2.Error, "")
	suite.Require().Equal("pass", resp2.Response.Result.Code, "ARC result not pass")
	suite.Require().Greater(len(resp2.Response.Chain), 0)
	suite.Require().True(resp2.Response.Chain[0].AMSValid)
	suite.Require().True(resp2.Response.Chain[0].ASValid)
	suite.Require().Equal("pass", resp2.Response.Chain[0].CV)
	suite.Require().Equal("pass", resp2.Response.Chain[0].Dkim)
	suite.Require().Equal("pass", resp2.Response.Chain[0].Dmarc)
	suite.Require().Equal("pass", resp2.Response.Chain[0].Spf)

	publicKey := &dkim.PublicKey{
		Version:    "DKIM1",
		KeyType:    "rsa",
		Algorithms: []string{"rsa-sha256"},
		Revoked:    false,
		Testing:    false,
		Strict:     false,
		Services:   []string{"email"},
		Key:        &testPrivateKey.PublicKey,
		Data:       testPrivateKey.PublicKey.N.Bytes(),
	}

	// sign DKIM
	msg = &EmailChainCalldata{
		SignDKIM: &SignDKIMRequest{
			EmailRaw: mailString,
			Options: SignOptions{
				Domain:   "example.org",
				Selector: "brisbane",
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
	fmt.Println("--SignDKIMResponse--", string(res.Data))
	resp3 := &SignDKIMResponse{}
	err = appA.DecodeExecuteResponse(res, resp3)
	suite.Require().NoError(err)
	suite.Require().Equal(resp3.Error, "")
	fmt.Println("=============SignDKIMResponse SignedEmail")
	fmt.Println(resp3.SignedEmail)
	fmt.Println("=============END SignDKIMResponse SignedEmail")
	suite.Require().Equal(signedMailString, resp3.SignedEmail)

	resDKIM, _, err := verifyEmail(resp3.SignedEmail, publicKey)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(resDKIM))
	fmt.Println("--resDKIM.Error--", resDKIM[0].Error)
	fmt.Println("--resDKIM--", *resDKIM[0])
	suite.Require().Nil(resDKIM[0].Error)
	suite.Require().Equal(dkim.Pass, resDKIM[0].Code)

	// sign ARC
	msg = &EmailChainCalldata{
		SignARC: &SignARCRequest{
			EmailRaw: resp3.SignedEmail,
			Options: SignOptions{
				Domain:   "example.org",
				Selector: "brisbane",
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
	fmt.Println("--SignARCResponse--", string(res.Data))
	resp4 := &SignARCResponse{}
	err = appA.DecodeExecuteResponse(res, resp4)
	suite.Require().NoError(err)
	suite.Require().Equal(resp4.Error, "")
	fmt.Println("===============SignARCResponse SignedEmail--")
	fmt.Println(resp4.SignedEmail)
	fmt.Println("===============END SignARCResponse SignedEmail--")
	// suite.Require().Greater(len(resp4.SignedEmail), 128)

	// verify signed email
	resDKIM, resARC, err := verifyEmail(resp4.SignedEmail, publicKey)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(resDKIM))
	fmt.Println("--resDKIM.Error--", resDKIM[0].Error)
	fmt.Println("--resDKIM--", *resDKIM[0])
	suite.Require().Nil(resDKIM[0].Error)
	suite.Require().Equal(dkim.Pass, resDKIM[0].Code)
	suite.Require().NoError(resARC.Error)
	suite.Require().Equal(dkim.Pass, resARC.Code)
	suite.Require().Equal(1, len(resARC.Chain))
	suite.Require().True(resARC.Chain[0].AMSValid)
	suite.Require().True(resARC.Chain[0].ASValid)
	suite.Require().Equal("pass", resARC.Chain[0].CV)
	suite.Require().Equal("pass", resARC.Chain[0].Dkim)
	suite.Require().Equal("pass", resARC.Chain[0].Dmarc)
	suite.Require().Equal("pass", resARC.Chain[0].Spf)
	suite.Require().Equal(1, resARC.Chain[0].Instance)

	// verify signed DKIM
	msg = &EmailChainCalldata{
		VerifyDKIM: &VerifyDKIMTestData{
			EmailRaw:  resp4.SignedEmail,
			PublicKey: publicKey,
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)

	// Execute the VerifyDKIM message in the contract
	res = appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 80000000, nil)
	resp = &VerifyDKIMResponse{}
	fmt.Println("--DKIM result2--", string(res.Data))
	err = appA.DecodeExecuteResponse(res, resp)
	suite.Require().NoError(err)
	suite.Require().Equal(resp.Error, "")
	suite.Require().Greater(len(resp.Response), 0)
	suite.Require().Equal("pass", resp.Response[0].Code, "DKIM result2 not pass")

	// verify signed ARC
	msg = &EmailChainCalldata{
		VerifyARC: &VerifyARCTestData{
			EmailRaw:  resp4.SignedEmail,
			PublicKey: publicKey,
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)

	res = appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 380000000, nil)
	fmt.Println("--ARC result--", string(res.Data))
	resp2 = &VerifyARCResponse{}
	err = appA.DecodeExecuteResponse(res, resp2)
	suite.Require().NoError(err)
	suite.Require().Equal(resp2.Error, "")
	suite.Require().Equal("pass", resp2.Response.Result.Code, "ARC result not pass")
	suite.Require().Greater(len(resp2.Response.Chain), 0)
	suite.Require().True(resp2.Response.Chain[0].AMSValid)
	suite.Require().True(resp2.Response.Chain[0].ASValid)
	suite.Require().Equal("pass", resp2.Response.Chain[0].CV)
	suite.Require().Equal("pass", resp2.Response.Chain[0].Dkim)
	suite.Require().Equal("pass", resp2.Response.Chain[0].Dmarc)
	suite.Require().Equal("pass", resp2.Response.Chain[0].Spf)
}

func verifyEmail(emailText string, pubk *dkim.PublicKey) ([]*dkim.Result, *dkim.ArcResult, error) {
	msg, err := dkim.ParseMessage(emailText)
	if err != nil {
		return nil, nil, err
	}

	resDKIM, err := dkim.Verify("DKIM-Signature", msg, net.LookupTXT, pubk)
	if err != nil {
		return nil, nil, err
	}

	resARC, err := dkim.VerifyArc(net.LookupTXT, pubk, msg)
	if err != nil {
		return nil, nil, err
	}
	return resDKIM, resARC, nil
}
