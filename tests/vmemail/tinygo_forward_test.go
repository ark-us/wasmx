package keeper_test

import (
	"crypto"
	"crypto/ed25519"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/emersion/go-imap/v2"
	"github.com/stretchr/testify/require"

	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"

	tinygo "github.com/loredanacirstea/mythos-tests/testdata/tinygo"
	testdata "github.com/loredanacirstea/mythos-tests/vmemail/testdata"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"

	dkimMox "github.com/loredanacirstea/mailverif/dkim"
	dnsMox "github.com/loredanacirstea/mailverif/dns"
	utilsMox "github.com/loredanacirstea/mailverif/utils"
)

type DNSResolver struct{}

func (r *DNSResolver) LookupTXT(name string) ([]string, dnsMox.Result, error) {
	res, err := net.LookupTXT(name)
	return res, dnsMox.Result{Authentic: true}, err
}

func TestEmailTinyGoVerifyDKIM(t *testing.T) {
	dkimres, arcres, err := verifyEmail(testdata.EmailDkim1, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(dkimres))
	require.Nil(t, dkimres[0].Err)
	require.Equal(t, "pass", string(dkimres[0].Status))
	require.Nil(t, dkimres[1].Err)
	require.Equal(t, "pass", string(dkimres[1].Status))
	require.NoError(t, arcres.Result.Err)
	require.Equal(t, "pass", arcres.Result.Status)

	dkimres, arcres, err = verifyEmail(testdata.EmailDkim2, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(dkimres))
	require.Nil(t, dkimres[0].Err)
	require.Equal(t, "pass", string(dkimres[0].Status))
	require.Nil(t, dkimres[1].Err)
	require.Equal(t, "pass", string(dkimres[1].Status))
	require.NoError(t, arcres.Result.Err)
	require.Equal(t, "pass", arcres.Result.Status)

	publicKey := &dkimMox.Record{
		Version:   "DKIM1",
		Key:       "rsa",
		Hashes:    []string{"sha256"},
		Services:  []string{"email"},
		PublicKey: &testPrivateKey.PublicKey,
		Pubkey:    testPrivateKey.PublicKey.N.Bytes(),
	}

	now := func() time.Time {
		return time.Unix(424242, 0)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	identif := utilsMox.Localpart("joe")
	domain := dnsMox.Domain{ASCII: "example.org"}
	key := ToPrivateKey("rsa", []byte(testPrivateKeyPEM))
	sel := dkimMox.Selector{
		Hash:       "sha256",
		PrivateKey: key,
		Headers:    strings.Split("From,To,Cc,Bcc,Reply-To,References,In-Reply-To,Subject,Date,Message-ID,Content-Type", ","),
		Domain:     dnsMox.Domain{ASCII: "football"},
	}
	selectors := []dkimMox.Selector{sel}
	header, err := dkimMox.Sign2(logger, identif, domain, selectors, false, []byte(mailString), now)
	require.NoError(t, err)

	newemailstr := utilsMox.SerializeHeaders(header) + mailString
	dkimres, arcres, err = verifyEmail(newemailstr, publicKey)
	require.NoError(t, err)
	require.Equal(t, 1, len(dkimres))
	require.Nil(t, dkimres[0].Err)
	require.Equal(t, "pass", string(dkimres[0].Status))
	require.NoError(t, arcres.Result.Err)
	require.Equal(t, "pass", arcres.Result.Status)
}

// func TestEmailTinyGoForwardSignature(t *testing.T) {
// 	// options := &dkimS.SignOptions{
// 	// 	Domain:    "example.org",
// 	// 	Selector:  "brisbane",
// 	// 	Signer:    testPrivateKey,
// 	// 	LookupTXT: net.LookupTXT,
// 	// }

// 	// pubk := &dkim.PublicKey{
// 	// 	Version:    "DKIM1",
// 	// 	KeyType:    "rsa",
// 	// 	Algorithms: []string{"rsa-sha256"},
// 	// 	Revoked:    false,
// 	// 	Testing:    false,
// 	// 	Strict:     false,
// 	// 	Services:   []string{"email"},
// 	// 	Key:        &testPrivateKey.PublicKey,
// 	// 	Data:       testPrivateKey.PublicKey.N.Bytes(),
// 	// }
// 	// fromParts := strings.Split("test@mail.provable.dev", "@")
// 	// fromAddr := imap.Address{Mailbox: fromParts[0], Host: fromParts[1]}
// 	// toAddrs := []imap.Address{{Mailbox: "seth.one.info", Host: "gmail.com"}}
// 	// timestamp := time.Unix(424242, 0)

// 	// email := emailchain.BuildForwardHeaders(testdata.EmailForwarded0, fromAddr, toAddrs, []imap.Address{}, []imap.Address{}, options, timestamp)
// 	// emailstr, err := emailchain.BuildRawEmail2(email, true)
// 	// require.NoError(t, err)
// 	// fmt.Println("--BuildRawEmail2-", emailstr)

// 	dkimres0, arcres0, err := verifyEmail(testdata.EmailForwarded0, nil)
// 	require.NoError(t, err)
// 	fmt.Println("--dkimres0--", len(dkimres0), dkimres0)
// 	for _, v := range dkimres0 {
// 		fmt.Println("--dkimres0--", v.Err, v.Status, v)
// 	}
// 	fmt.Println("--arcres0--", arcres0.Error, arcres0.Code, arcres0)

// 	dkimres1, arcres1, err := verifyEmail(testdata.EmailDkim1, nil)
// 	require.NoError(t, err)
// 	fmt.Println("--dkimres1--", len(dkimres1), dkimres1)
// 	for _, v := range dkimres1 {
// 		fmt.Println("--dkimres1--", v.Err, v.Status, v)
// 	}
// 	fmt.Println("--arcres1--", arcres1.Error, arcres1.Code, arcres1)

// 	options := dkimS.VerifyOptions{LookupTXT: net.LookupTXT}
// 	verifications, err := dkimS.VerifyWithOptions(strings.NewReader(testdata.EmailDkim1), &options)
// 	fmt.Println("--QQ dkimres1 err--", len(verifications), err)
// 	for _, v := range verifications {
// 		fmt.Println("--QQ dkimres1--", v.Expired, v.Err, v.Domain, v)
// 	}

// 	dkimres2, arcres2, err := verifyEmail(testdata.EmailForwarded1, nil)
// 	require.NoError(t, err)

// 	fmt.Println("--dkimres2--", len(dkimres2), dkimres2)
// 	for _, v := range dkimres2 {
// 		fmt.Println("--dkimres2--", v.Err, v.Status, v)
// 	}
// 	fmt.Println("--arcres2--", arcres2.Error, arcres2.Code, arcres2)
// }

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
			// Timestamp: time.Unix(424242, 0),
			Timestamp: time.Now(),
			SendEmail: true,
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

	dkimres, arcres, err := verifyEmail(resp3.EmailRaw, nil) // publicKey
	suite.Require().NoError(err)

	fmt.Println("--dkimres--", len(dkimres), dkimres)
	for _, v := range dkimres {
		fmt.Println("--dkimres--", v.Err, v.Status, v)
	}
	fmt.Println("--arcres--", arcres.Result.Err, arcres.Result.Status, arcres)

	dkimres2, arcres2, err := verifyEmail(testdata.EmailForwarded1, nil)
	suite.Require().NoError(err)

	fmt.Println("--dkimres2--", len(dkimres2), dkimres2)
	for _, v := range dkimres2 {
		fmt.Println("--dkimres2--", v.Err, v.Status, v)
	}
	fmt.Println("--arcres2--", arcres2.Result.Err, arcres2.Result.Status, arcres2)

	// TODO test
	// forwarded email same bh as original email
	// extract/recover original email + headers & verify it - DKIM & optional ARC
	// verify forwarded signature

	// TODO
	// forward again, see instance number change

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

func ToPrivateKey(keyType string, pk []byte) crypto.Signer {
	var signer crypto.Signer
	var err error
	if keyType == "rsa" {
		// we expect privatekey in PEM format
		block, _ := pem.Decode(pk)
		var err error
		signer, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			panic(err)
		}
	} else {
		signer, err = loadPrivateKey(pk)
		if err != nil {
			panic(err)
		}
	}
	return signer
}

func loadPrivateKey(keyBytes []byte) (ed25519.PrivateKey, error) {
	if len(keyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid key length: got %d, want %d", len(keyBytes), ed25519.PrivateKeySize)
	}
	return ed25519.PrivateKey(keyBytes), nil
}
