package main

import (
	"crypto"
	"crypto/x509"
	"time"

	"encoding/json"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ed25519"

	vmimap "github.com/loredanacirstea/wasmx-env-imap"

	dkimS "github.com/emersion/go-msgauth/dkim"
	dkimMox "github.com/loredanacirstea/mailverif/dkim"
	dkim "github.com/redsift/dkim"
)

type Calldata struct {
	ConnectWithPassword *ConnectionSimpleRequest        `json:"ConnectWithPassword,omitempty"`
	ConnectOAuth2       *ConnectionOauth2Request        `json:"ConnectOAuth2,omitempty"`
	Close               *CloseRequest                   `json:"Close,omitempty"`
	SendEmail           *vmimap.ImapCreateFolderRequest `json:"SendEmail,omitempty"`
	BuildAndSend        *BuildAndSendMailRequest        `json:"BuildAndSend,omitempty"`
	VerifyDKIM          *VerifyDKIMRequest              `json:"VerifyDKIM,omitempty"`
	VerifyARC           *VerifyDKIMRequest              `json:"VerifyARC,omitempty"`
	SignDKIM            *SignDKIMRequest                `json:"SignDKIM,omitempty"`
	SignARC             *SignARCRequest                 `json:"SignARC,omitempty"`
	ForwardEmail        *ForwardEmailRequest            `json:"ForwardEmail,omitempty"`
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

type BuildAndSendMailRequest struct {
	Id      string   `json:"id"`
	From    string   `json:"from"`
	To      []string `json:"to"`
	Cc      []string `json:"cc"`
	Bcc     []string `json:"bcc"`
	Subject string   `json:"subject"`
	Body    []byte   `json:"body"`
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

func (v SignOptions) toLib() *dkimS.SignOptions {
	var signer crypto.Signer
	var err error
	if v.PrivateKeyType == "rsa" {
		// we expect privatekey in PEM format
		block, _ := pem.Decode(v.PrivateKey)
		var err error
		signer, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			panic(err)
		}
	} else {
		signer, err = loadPrivateKey(v.PrivateKey)
		if err != nil {
			panic(err)
		}
	}

	return &dkimS.SignOptions{
		Domain:                 v.Domain,
		Selector:               v.Selector,
		Identifier:             v.Identifier,
		Signer:                 signer,
		Hash:                   crypto.Hash(v.Hash),
		HeaderCanonicalization: v.HeaderCanonicalization,
		BodyCanonicalization:   v.BodyCanonicalization,
		HeaderKeys:             v.HeaderKeys,
		Expiration:             v.Expiration,
		QueryMethods:           v.QueryMethods,
	}
}

type SignDKIMRequest struct {
	EmailRaw  string      `json:"email_raw"`
	Options   SignOptions `json:"options"`
	Timestamp time.Time   `json:"timestamp"`
}

type SignARCRequest struct {
	MailFrom  string      `json:"mailfrom"`
	IP        string      `json:"ip"`
	EmailRaw  string      `json:"email_raw"`
	Options   SignOptions `json:"options"`
	Timestamp time.Time   `json:"timestamp"`
}

type SignDKIMResponse struct {
	Error  string `json:"error"`
	Header string `json:"header"`
}

type SignARCResponse struct {
	Error       string `json:"error"`
	SignedEmail string `json:"signed_email"`
}

type VerifyDKIMRequest struct {
	EmailRaw  string          `json:"email_raw"`
	PublicKey *dkimMox.Record `json:"public_key,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

type VerifyDKIMResponse struct {
	Error    string           `json:"error"`
	Response []dkimMox.Result `json:"response"`
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

type ForwardEmailRequest struct {
	ConnectionId string           `json:"connection_id"`
	Folder       string           `json:"folder"`
	Uid          uint32           `json:"uid"`
	MessageId    string           `json:"message_id"`
	From         vmimap.Address   `json:"from"`
	To           []vmimap.Address `json:"to"`
	Cc           []vmimap.Address `json:"cc"`
	Bcc          []vmimap.Address `json:"bcc"`
	Options      SignOptions      `json:"options"`
	Timestamp    time.Time        `json:"timestamp"`
	SendEmail    bool             `json:"send_email"`
}

type ForwardEmailResponse struct {
	Error    string `json:"error"`
	EmailRaw string `json:"email_raw"`
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

type Result struct {
	Order     int                     `json:"order"`
	Code      string                  `json:"code"`
	Error     *dkim.VerificationError `json:"error,omitempty"`
	Signature *Signature              `json:"signature,omitempty"`
	Key       *dkim.PublicKey         `json:"key,omitempty"`
	Timestamp time.Time               `json:"timestamp"`
}

func (v Result) String() string {
	vbz, err := json.Marshal(&v)
	if err == nil {
		return string(vbz)
	}
	errstr := ""
	signature := ""
	key := ""
	if v.Error != nil {
		if v.Error.Err != nil {
			errstr = v.Error.Err.Error()
		} else {
			errstr = v.Error.Explanation
		}
	}
	return fmt.Sprintf(`order:%d code:%s error:%s signature:%s key:%s timestamp:%s`, v.Order, v.Code, errstr, signature, key, v.Timestamp.UTC().String())
}

func (v *Signature) FromLib(val *dkim.Signature) {
	fmt.Println("--Signature.FromLib--")
	fmt.Println("--Signature.FromLib val.AlgorithmID--", val.AlgorithmID)
	hashbz, err := val.AlgorithmID.MarshalText()
	fmt.Println("--Signature.FromLib2--")
	fmt.Println("--Signature.FromLib3--", err, string(hashbz))
	if err != nil {
		panic("cannot marshal AlgorithmID: " + err.Error())
	}
	v.Header = val.Header
	v.Raw = val.Raw
	v.AlgorithmID = string(hashbz)
	v.Hash = val.Hash
	v.BodyHash = val.BodyHash
	v.RelaxedHeader = val.RelaxedHeader
	v.RelaxedBody = val.RelaxedBody
	v.SignerDomain = val.SignerDomain
	v.Headers = val.Headers
	v.UserIdentifier = val.UserIdentifier
	v.ArcInstance = val.ArcInstance
	v.Length = val.Length
	v.Selector = val.Selector
	v.Timestamp = val.Timestamp
	v.Expiration = val.Expiration
	v.CopiedHeaders = val.CopiedHeaders
	v.ArcCV = val.ArcCV.String()
	v.Dmarc = val.Dmarc.String()
	v.Dkim = val.Dkim.String()
}

func (v *Result) FromLib(val *dkim.Result) {
	fmt.Println("--Result.FromLib--")
	v.Order = val.Order
	v.Code = val.Code.String()
	v.Error = val.Error
	fmt.Println("--Result.FromLib val.Signature--", val.Signature)
	if val.Signature != nil {
		v.Signature = &Signature{}
		v.Signature.FromLib(val.Signature)
		fmt.Println("--Result.FromLib v.Signature--", v.Signature)
	}
	v.Key = val.Key
	v.Timestamp = val.Timestamp
}

func (v *Result) FromLibArcResult(val *dkim.ArcResult) {
	fmt.Println("--Result.FromLibArcResult--")
	v.Order = val.Order
	v.Code = val.Code.String()
	v.Error = val.Error
	fmt.Println("--Result.FromLibArcResult val.Signature--", val.Signature)
	if val.Signature != nil {
		fmt.Println("--val.Signature not nil--")
		v.Signature = &Signature{}
		v.Signature.FromLib(val.Signature)
		fmt.Println("--Result.FromLibArcResult v.Signature--", v.Signature)
	}
	fmt.Println("--Result.FromLibArcResult post val.Signature--")
	fmt.Println("--Result.FromLibArcResult val.Key--", val.Key)
	if val.Key != nil {
		v.Key = val.Key
	}
	fmt.Println("--Result.FromLibArcResult val.Timestamp--", val.Timestamp)
	v.Timestamp = val.Timestamp
	fmt.Println("--Result.FromLibArcResult post val.Timestamp--")
}

func ResultArrFromLib(vv []*dkim.Result) []Result {
	r := make([]Result, 0)
	for _, v := range vv {
		if v != nil {
			val := &Result{}
			val.FromLib(v)
			r = append(r, *val)
		}
	}
	return r
}

func (v *ArcResult) FromLib(val *dkim.ArcResult) {
	fmt.Println("--ArcResult.FromLib0--", val)
	r := &Result{}
	r.FromLibArcResult(val)
	fmt.Println("--ArcResult.FromLib Result--")
	fmt.Println("--ArcResult.FromLib Result2--", r)
	v.Result = *r
	fmt.Println("--ArcResult.FromLib v.Result--", v.Result)
	chain := make([]ArcSetResult, len(val.Chain))
	fmt.Println("--ArcResult.FromLib chain--", chain)
	for i, res := range val.Chain {
		fmt.Println("--ArcResult.FromLib i,res--", i, res)
		chain[i] = ArcSetResult{
			Instance: res.Instance,
			Spf:      res.Spf.String(),
			Dkim:     res.Dkim.String(),
			Dmarc:    res.Dmarc.String(),
			AMSValid: res.AMSValid,
			ASValid:  res.ASValid,
			CV:       res.CV.String(),
		}
	}
	fmt.Println("--ArcResult.FromLib chain--", chain)
	v.Chain = chain
}

func loadPrivateKey(keyBytes []byte) (ed25519.PrivateKey, error) {
	if len(keyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid key length: got %d, want %d", len(keyBytes), ed25519.PrivateKeySize)
	}
	return ed25519.PrivateKey(keyBytes), nil
}

// func hashStringToLib(hashstr string) crypto.Hash {
// 	switch hashstr {
// 	case "MD4":
// 		return crypto.MD4
// 	case "MD5":
// 		return crypto.MD5
// 	case "SHA-1":
// 		return crypto.SHA1
// 	case "SHA-224":
// 		return crypto.SHA224
// 	case "SHA-256":
// 		return crypto.SHA256
// 	case "SHA-384":
// 		return crypto.SHA384
// 	case "SHA-512":
// 		return crypto.SHA512
// 	case "MD5+SHA1":
// 		return crypto.MD5SHA1
// 	case "RIPEMD-160":
// 		return crypto.RIPEMD160
// 	case "SHA3-224":
// 		return crypto.SHA3_224
// 	case "SHA3-256":
// 		return crypto.SHA3_256
// 	case "SHA3-384":
// 		return crypto.SHA3_384
// 	case "SHA3-512":
// 		return crypto.SHA3_512
// 	case "SHA-512/224":
// 		return crypto.SHA512_224
// 	case "SHA-512/256":
// 		return crypto.SHA512_256
// 	case "BLAKE2s-256":
// 		return crypto.BLAKE2s_256
// 	case "BLAKE2b-256":
// 		return crypto.BLAKE2b_256
// 	case "BLAKE2b-384":
// 		return crypto.BLAKE2b_384
// 	case "BLAKE2b-512":
// 		return crypto.BLAKE2b_512
// 	default:
// 		return 100
// 	}
// }

const (
	HEADER_PROVABLE_DNS_REGISTRY                = "Provable-DNS-Registry"
	HEADER_PROVABLE_EMAIL_REGISTRY              = "Provable-Email-Registry"
	HEADER_PROVABLE_FORWARD_ORIGIN_DKIM_CONTEXT = "Provable-Forward-Origin-DKIM-Context"
	HEADER_PROVABLE_FORWARD_CHAIN_SIGNATURE     = "Provable-Forward-Chain-Signature"
)
