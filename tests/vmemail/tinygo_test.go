package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
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

	"github.com/loredanacirstea/mailverif/arc"
	"github.com/loredanacirstea/mailverif/dkim"
	"github.com/loredanacirstea/mailverif/dns"
	dkimUtils "github.com/loredanacirstea/mailverif/utils"
)

type SendMailRequest struct {
	From     imap.Address   `json:"from"`
	To       []imap.Address `json:"to"`
	EmailRaw []byte         `json:"email_raw"`
	Date     time.Time      `json:"date"`
}

type BuildAndSendMailRequest struct {
	Id      string    `json:"id"`
	From    string    `json:"from"`
	To      []string  `json:"to"`
	Cc      []string  `json:"cc"`
	Bcc     []string  `json:"bcc"`
	Subject string    `json:"subject"`
	Body    []byte    `json:"body"`
	Date    time.Time `json:"date"`
}

type CalldataTestSmptTinygo struct {
	CalldataTestSmpt
	BuildAndSend *BuildAndSendMailRequest `json:"BuildAndSend,omitempty"`
}

func (suite *KeeperTestSuite) TestEmailTinygoImap() {
	SkipNoPasswordTests(suite.T(), "TestEmailTinygoImap")
	wasmbin := tinygo.ImapTestWrapSdk
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "EmailTestWrapSdk", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "someemailrole", contractAddress, sender)

	msg := &Calldata{
		Connect: &vmimap.ImapConnectionRequest{
			Id:            "conn1",
			ImapServerUrl: "mail.mail.provable.dev:993",
			Auth: vmimap.ConnectionAuth{
				AuthType: vmimap.ConnectionAuthTypePassword,
				Username: suite.emailUsername,
				Password: suite.emailPassword,
			},
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
	SkipNoPasswordTests(suite.T(), "TestEmailTinyGoSmtp")
	wasmbin := tinygo.SmtpTestWrapSdk
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "smtptest", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "someemailrole", contractAddress, sender)

	msg := &CalldataTestSmpt{}
	if suite.isOAuth2 {
		msg.Connect = &vmsmtp.SmtpConnectionRequest{
			Id:        "conn1",
			ServerUrl: "smtp.gmail.com:587",
			StartTLS:  true,
			Auth: &vmsmtp.ConnectionAuth{
				AuthType: vmsmtp.ConnectionAuthTypeOAuth2,
				Username: suite.emailUsername,
				Password: suite.emailPassword,
			},
		}
	} else {
		msg.Connect = &vmsmtp.SmtpConnectionRequest{
			Id:        "conn1",
			ServerUrl: "mail.mail.provable.dev:587",
			StartTLS:  true,
			Auth: &vmsmtp.ConnectionAuth{
				AuthType: vmsmtp.ConnectionAuthTypePassword,
				Username: suite.emailUsername,
				Password: suite.emailPassword,
			},
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
	CreateAccount       *CreateAccountRequest    `json:"CreateAccount,omitempty"`
	ConnectWithPassword *ConnectionSimpleRequest `json:"ConnectWithPassword,omitempty"`
	ConnectOAuth2       *ConnectionOauth2Request `json:"ConnectOAuth2,omitempty"`
	Close               *CloseRequest            `json:"Close,omitempty"`
	VerifyDKIM          *VerifyDKIMTestData      `json:"VerifyDKIM,omitempty"`
	VerifyARC           *VerifyDKIMRequest       `json:"VerifyARC,omitempty"`
	SignDKIM            *SignDKIMRequest         `json:"SignDKIM,omitempty"`
	SignARC             *SignARCRequest          `json:"SignARC,omitempty"`
	ForwardEmail        *ForwardEmailRequest     `json:"ForwardEmail,omitempty"`
	StartServer         *StartServerRequest      `json:"StartServer,omitempty"`
	IncomingEmail       *vmsmtp.Session          `json:"IncomingEmail,omitempty"`
	BuildAndSend        *BuildAndSendMailRequest `json:"BuildAndSend,omitempty"`
	SendEmail           *SendMailRequest         `json:"SendEmail,omitempty"`
}

type CreateAccountRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type StartServerRequest struct {
	SignOptions SignOptions         `json:"options"`
	Smtp        vmsmtp.ServerConfig `json:"smtp"`
	Imap        vmimap.ServerConfig `json:"imap"`
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
	EmailRaw string `json:"email_raw"`
	// PublicKey *dkim.PublicKey `json:"public_key,omitempty"`
	Pubkey    []byte // Public key, as base64 in record
	PublicKey any    `json:"-"` // Parsed form of public key, an *rsa.PublicKey or ed25519.PublicKey.
	Timestamp time.Time
}

type VerifyDKIMRequest struct {
	EmailRaw  string       `json:"email_raw"`
	PublicKey *dkim.Record `json:"public_key,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

type VerifyDKIMResponse struct {
	Error    string        `json:"error"`
	Response []dkim.Result `json:"response"`
}

type VerifyARCResponse struct {
	Error    string         `json:"error"`
	Response *arc.ArcResult `json:"response"`
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
	PrivateKeyType string `json:"private_key_type"` // "rsa" or "ed25519"
	PrivateKey     []byte `json:"private_key"`

	// The hash algorithm used to sign the message. If zero, a default hash will
	// be chosen.
	//
	// The only supported hash algorithm is crypto.SHA256.
	Hash uint `json:"hash"`

	// Header and body canonicalization algorithms.
	//
	// If empty, CanonicalizationSimple is used.
	HeaderRelaxed bool `json:"header_relaxed"`
	BodyRelaxed   bool `json:"body_relaxed"`

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
	// QueryMethods []dkimS.QueryMethod `json:"query_methods"`
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
	Error  string `json:"error"`
	Header string `json:"header"`
}

type SignARCResponse struct {
	Error  string `json:"error"`
	Header string `json:"header"`
}

type ForwardEmailRequest struct {
	ConnectionId      string         `json:"connection_id"`
	Folder            string         `json:"folder"`
	Uid               uint32         `json:"uid"`
	MessageId         string         `json:"message_id"`
	AdditionalSubject string         `json:"additional_subject"`
	From              imap.Address   `json:"from"`
	To                []imap.Address `json:"to"`
	Options           SignOptions    `json:"options"`
	Timestamp         time.Time      `json:"timestamp"`
	SendEmail         bool           `json:"send_email"`
}

type ForwardEmailResponse struct {
	Error    string `json:"error"`
	EmailRaw string `json:"email_raw"`
}

var now = func() time.Time {
	return time.Unix(424242, 0)
}

func AddressFromString(account string, name string) imap.Address {
	parts := strings.Split(account, "@")
	return imap.Address{Name: name, Mailbox: parts[0], Host: parts[1]}
}

func AddressesFromString(accounts []string) []imap.Address {
	addrs := []imap.Address{}
	for _, v := range accounts {
		addrs = append(addrs, AddressFromString(v, ""))
	}
	return addrs
}

func TestSignARCSync1(t *testing.T) {
	options := &SignOptions{
		Domain:         "example.org",
		Selector:       "brisbane",
		PrivateKeyType: "rsa",
		PrivateKey:     []byte(testPrivateKeyPEM),
		Identifier:     "joe",
	}

	pubk := &dkim.Record{
		Version:   "DKIM1",
		Key:       "rsa",
		Hashes:    []string{"sha256"},
		Services:  []string{"email"},
		PublicKey: &testPrivateKey.PublicKey,
		Pubkey:    testPrivateKey.PublicKey.N.Bytes(),
	}

	_, resARC := ARCSignAndVerify(t, options, signedMailString, "joe@football.example.com", "85.215.130.119", "example.org", pubk)
	require.NoError(t, resARC.Result.Err)
	require.Equal(t, dkim.StatusPass, resARC.Result.Status)
	require.Equal(t, 1, len(resARC.Chain))

	require.True(t, resARC.Chain[0].AMSValid)
	require.True(t, resARC.Chain[0].ASValid)
	require.Equal(t, 1, resARC.Chain[0].Instance)
	require.Equal(t, dkim.StatusNone, resARC.Chain[0].CV)
	require.Equal(t, dkim.StatusPass, resARC.Chain[0].Dkim)
	require.Equal(t, dkim.StatusPass, resARC.Chain[0].Dmarc)
	require.Equal(t, dkim.StatusPass, resARC.Chain[0].Spf)
}

func TestSignARCSync(t *testing.T) {
	options := &SignOptions{
		Domain:         "example.org",
		Selector:       "brisbane",
		PrivateKeyType: "rsa",
		PrivateKey:     []byte(testPrivateKeyPEM),
		Identifier:     "joe",
	}

	pubk := &dkim.Record{
		Version:   "DKIM1",
		Key:       "rsa",
		Hashes:    []string{"sha256"},
		Services:  []string{"email"},
		PublicKey: &testPrivateKey.PublicKey,
		Pubkey:    testPrivateKey.PublicKey.N.Bytes(),
	}

	emailStr := testdata.EmailARC3
	// emailStr := signedMailString
	newemail, resARC := ARCSignAndVerify(t, options, emailStr, "seth.one.info@gmail.com", "209.85.214.177", "example.org", pubk)

	require.NoError(t, resARC.Result.Err)
	require.Equal(t, dkim.StatusPass, resARC.Result.Status)
	require.Equal(t, 1, len(resARC.Chain))

	require.True(t, resARC.Chain[0].AMSValid)
	require.True(t, resARC.Chain[0].ASValid)
	require.Equal(t, 1, resARC.Chain[0].Instance)
	require.Equal(t, dkim.StatusNone, resARC.Chain[0].CV)
	require.Equal(t, dkim.StatusFail, resARC.Chain[0].Dkim)
	require.Equal(t, dkim.StatusFail, resARC.Chain[0].Dmarc)
	require.Equal(t, dkim.StatusFail, resARC.Chain[0].Spf)

	// sign another time
	_, resARC = ARCSignAndVerify(t, options, newemail, "test@provable.dev", "85.215.130.119", "provable.dev", pubk)
	require.Error(t, resARC.Result.Err)
	require.Contains(t, resARC.Result.Err.Error(), "ARC-Seal reported failure, the chain is terminated")
	require.Equal(t, dkim.StatusFail, resARC.Result.Status)
	require.Equal(t, 2, len(resARC.Chain))

	require.True(t, resARC.Chain[0].AMSValid)
	require.False(t, resARC.Chain[0].ASValid)
	require.Equal(t, 2, resARC.Chain[0].Instance)
	require.Equal(t, dkim.StatusFail, resARC.Chain[0].CV)
	require.Equal(t, dkim.StatusFail, resARC.Chain[0].Dkim)
	require.Equal(t, dkim.StatusFail, resARC.Chain[0].Dmarc)
	require.Equal(t, dkim.StatusFail, resARC.Chain[0].Spf)

	require.True(t, resARC.Chain[1].AMSValid)
	require.True(t, resARC.Chain[1].ASValid)
	require.Equal(t, 1, resARC.Chain[1].Instance)
	require.Equal(t, dkim.StatusNone, resARC.Chain[1].CV)
	require.Equal(t, dkim.StatusFail, resARC.Chain[1].Dkim)
	require.Equal(t, dkim.StatusFail, resARC.Chain[1].Dmarc)
	require.Equal(t, dkim.StatusFail, resARC.Chain[1].Spf)
}

func ARCSignAndVerify(t *testing.T, options *SignOptions, emailStr string, mailfrom string, ip string, mailServerDomain string, pubk *dkim.Record) (string, *arc.ArcResult) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	domain := dns.Domain{ASCII: options.Domain}
	key := ToPrivateKey(options.PrivateKeyType, options.PrivateKey)
	sel := dkim.Selector{
		Hash:          "sha256",
		PrivateKey:    key,
		Headers:       options.HeaderKeys,
		Domain:        dns.Domain{ASCII: options.Selector},
		HeaderRelaxed: options.HeaderRelaxed,
		BodyRelaxed:   options.BodyRelaxed,
	}
	selectors := []dkim.Selector{sel}
	headers, err := arc.Sign(logger, &DNSResolver{}, domain, selectors, false, []byte(emailStr), mailfrom, ip, false, false, now, pubk)
	require.NoError(t, err, "Expected no error while signing mail")
	slices.Reverse(headers)
	signedEmail := dkimUtils.SerializeHeaders(headers) + emailStr

	fmt.Println("=======ARCSigned")
	fmt.Println(signedEmail)
	fmt.Println("=======END ARCSigned")

	resARC, err := arc.Verify(logger, &DNSResolver{}, false, []byte(signedEmail), false, true, now, pubk)
	require.NoError(t, err)
	return signedEmail, resARC
}

func (suite *KeeperTestSuite) TestEmailTinyGoDKIM() {
	SkipFixmeTests(suite.T(), "TestEmailTinyGoDKIM")
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

	// DKIM with simple canon
	emailRaw := testdata.EmailDkim2

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
	res := appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	resp := &VerifyDKIMResponse{}
	fmt.Println("--DKIM result--", string(res.Data))
	err = appA.DecodeExecuteResponse(res, resp)
	suite.Require().NoError(err)
	suite.Require().Equal(resp.Error, "")
	suite.Require().Greater(len(resp.Response), 0)
	suite.Require().NoError(resp.Response[0].Err)
	suite.Require().Equal(dkim.StatusPass, string(resp.Response[0].Status), "DKIM result not pass")

	// Define email input for DKIM verification
	emailRaw = testdata.EmailARC3

	// Prepare the VerifyDKIM request
	msg = &EmailChainCalldata{
		VerifyDKIM: &VerifyDKIMTestData{
			EmailRaw: emailRaw,
		},
	}
	data, err = json.Marshal(msg)
	suite.Require().NoError(err)

	// first verify email
	// _, _, err = verifyEmail(emailRaw, nil)
	// suite.Require().NoError(err)

	// Execute the VerifyDKIM message
	res = appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 280000000, nil)
	resp = &VerifyDKIMResponse{}
	fmt.Println("--DKIM result--", string(res.Data))
	err = appA.DecodeExecuteResponse(res, resp)
	suite.Require().NoError(err)
	suite.Require().Equal(resp.Error, "")
	suite.Require().Greater(len(resp.Response), 0)
	suite.Require().NoError(resp.Response[0].Err)
	suite.Require().Equal(dkim.StatusPass, string(resp.Response[0].Status), "DKIM result not pass")

	publicKey := &dkim.Record{
		Version:   "DKIM1",
		Key:       "rsa",
		Hashes:    []string{"sha256"},
		Services:  []string{"email"},
		PublicKey: &testPrivateKey.PublicKey,
		Pubkey:    testPrivateKey.PublicKey.N.Bytes(),
	}

	// sign DKIM
	msg = &EmailChainCalldata{
		SignDKIM: &SignDKIMRequest{
			EmailRaw: mailString,
			Options: SignOptions{
				Domain:   "example.org",
				Selector: "football",
				// PrivateKeyType: "ed25519",
				// PrivateKey:     testEd25519PrivateKey,
				PrivateKeyType: "rsa",
				PrivateKey:     []byte(testPrivateKeyPEM),
				Identifier:     "joe",
				HeaderKeys:     strings.Split("From,To,Cc,Bcc,Reply-To,References,In-Reply-To,Subject,Date,Message-ID,Content-Type", ","),
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
	fmt.Println(resp3.Header)
	fmt.Println("=============END SignDKIMResponse SignedEmail")
	// suite.Require().Equal(mailStringDkim, resp3.Header)

	newemailstr := resp3.Header + mailString
	resDKIM, _, err := verifyEmail(newemailstr, publicKey)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(resDKIM))
	fmt.Println("--resDKIM.Error--", resDKIM[0].Err)
	fmt.Println("--resDKIM--", resDKIM[0])
	suite.Require().Nil(resDKIM[0].Err)
	suite.Require().Equal(dkim.StatusPass, string(resDKIM[0].Status))

	// ARC
	msg = &EmailChainCalldata{
		VerifyARC: &VerifyDKIMRequest{
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
	suite.Require().Equal(dkim.StatusPass, resp2.Response.Result.Status, "ARC result not pass")
	suite.Require().Greater(len(resp2.Response.Chain), 0)
	suite.Require().True(resp2.Response.Chain[0].AMSValid)
	suite.Require().True(resp2.Response.Chain[0].ASValid)
	suite.Require().Equal(dkim.StatusPass, resp2.Response.Chain[0].CV)
	suite.Require().Equal(dkim.StatusPass, resp2.Response.Chain[0].Dkim)
	suite.Require().Equal(dkim.StatusPass, resp2.Response.Chain[0].Dmarc)
	suite.Require().Equal(dkim.StatusPass, resp2.Response.Chain[0].Spf)

	// sign ARC
	msg = &EmailChainCalldata{
		SignARC: &SignARCRequest{
			EmailRaw: signedMailString,
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
	fmt.Println(resp4.Header)
	fmt.Println("===============END SignARCResponse SignedEmail--")
	// suite.Require().Greater(len(resp4.SignedEmail), 128)
	newemail := resp4.Header + signedMailString

	// verify signed email
	resDKIM, resARC, err := verifyEmail(newemail, publicKey)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(resDKIM))
	fmt.Println("--resDKIM.Error--", resDKIM[0].Err)
	fmt.Println("--resDKIM--", resDKIM[0])
	suite.Require().Nil(resDKIM[0].Err)
	suite.Require().Equal(dkim.StatusPass, resDKIM[0].Status)
	suite.Require().NoError(resARC.Result.Err)
	suite.Require().Equal(dkim.StatusPass, resARC.Result.Status)
	suite.Require().Equal(1, len(resARC.Chain))
	suite.Require().True(resARC.Chain[0].AMSValid)
	suite.Require().True(resARC.Chain[0].ASValid)
	suite.Require().Equal(dkim.StatusPass, resARC.Chain[0].CV)
	suite.Require().Equal(dkim.StatusPass, resARC.Chain[0].Dkim)
	suite.Require().Equal(dkim.StatusPass, resARC.Chain[0].Dmarc)
	suite.Require().Equal(dkim.StatusPass, resARC.Chain[0].Spf)
	suite.Require().Equal(1, resARC.Chain[0].Instance)

	// verify signed DKIM
	msg = &EmailChainCalldata{
		VerifyDKIM: &VerifyDKIMTestData{
			EmailRaw:  newemail,
			PublicKey: publicKey,
			Timestamp: time.Now(),
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
	suite.Require().NoError(resp.Response[0].Err)
	suite.Require().Equal(dkim.StatusPass, string(resp.Response[0].Status), "DKIM result2 not pass")

	// verify signed ARC
	msg = &EmailChainCalldata{
		VerifyARC: &VerifyDKIMRequest{
			EmailRaw:  newemail,
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
	suite.Require().Equal(dkim.StatusPass, resp2.Response.Result.Status, "ARC result not pass")
	suite.Require().Greater(len(resp2.Response.Chain), 0)
	suite.Require().True(resp2.Response.Chain[0].AMSValid)
	suite.Require().True(resp2.Response.Chain[0].ASValid)
	suite.Require().Equal(dkim.StatusPass, resp2.Response.Chain[0].CV)
	suite.Require().Equal(dkim.StatusPass, resp2.Response.Chain[0].Dkim)
	suite.Require().Equal(dkim.StatusPass, resp2.Response.Chain[0].Dmarc)
	suite.Require().Equal(dkim.StatusPass, resp2.Response.Chain[0].Spf)
}

var timeNow = func() time.Time {
	return time.Now()
}

func verifyEmail(emailText string, pubk *dkim.Record) ([]dkim.Result, *arc.ArcResult, error) {
	fmt.Println("--DKIM verify--")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	results, err := dkim.Verify2(logger, &DNSResolver{}, false, dkim.DefaultPolicy, []byte(emailText), false, false, true, timeNow, pubk)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("--DKIM verify END--")

	fmt.Println("--ARC verify--")
	resARC, err := arc.Verify(logger, &DNSResolver{}, false, []byte(emailText), false, true, now, pubk)
	if err != nil {
		return nil, nil, err
	}
	return results, resARC, nil
}
