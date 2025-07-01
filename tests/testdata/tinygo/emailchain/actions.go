package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	dkimS "github.com/emersion/go-msgauth/dkim"
	dkimMox "github.com/loredanacirstea/mailverif/dkim"
	dnsMox "github.com/loredanacirstea/mailverif/dns"
	utilsMox "github.com/loredanacirstea/mailverif/utils"

	"github.com/loredanacirstea/wasmx-env"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
	vmsmtp "github.com/loredanacirstea/wasmx-env-smtp"
)

const DKIM_HEADER = "DKIM-Signature"

func StartServer() {
	resp := vmsmtp.ServerStart(&vmsmtp.ServerStartRequest{})
	if resp.Error != "" {
		wasmx.Revert([]byte(resp.Error))
	}
}

func SaveEmail(req *IncomingEmail) {
	// sql.
}

func VerifyDKIM(req *VerifyDKIMRequest) VerifyDKIMResponse {
	resp := VerifyDKIMResponse{Error: ""}
	dnsResolver := NewDNSResolver()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	now := func() time.Time {
		return req.Timestamp
	}

	results, err := dkimMox.Verify2(logger, dnsResolver, false, dkimMox.DefaultPolicy, strings.NewReader(req.EmailRaw), false, true, now, req.PublicKey)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.Response = results
	return resp
}

func SignDKIM(req *SignDKIMRequest) SignDKIMResponse {
	fmt.Println("--SignDKIM--")
	resp := SignDKIMResponse{Error: ""}

	r := strings.NewReader(req.EmailRaw)
	now := func() time.Time {
		return req.Timestamp
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	identif := utilsMox.Localpart(req.Options.Identifier)
	domain := dnsMox.Domain{ASCII: req.Options.Domain}
	key := ToPrivateKey(req.Options.PrivateKeyType, req.Options.PrivateKey)
	sel := dkimMox.Selector{
		Hash:       "sha256",
		PrivateKey: key,
		// Headers:    strings.Split("From,To,Cc,Bcc,Reply-To,References,In-Reply-To,Subject,Date,Message-ID,Content-Type", ","),
		Headers: req.Options.HeaderKeys,
		Domain:  dnsMox.Domain{ASCII: req.Options.Selector},
	}
	selectors := []dkimMox.Selector{sel}
	header, err := dkimMox.Sign2(logger, identif, domain, selectors, false, r, now)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.Header = utilsMox.SerializeHeaders(header)

	return resp
}

// func SignDKIM(req *SignDKIMRequest) SignDKIMResponse {
// 	fmt.Println("--SignDKIM--")
// 	resp := SignDKIMResponse{Error: ""}

// 	r := strings.NewReader(req.EmailRaw)
// 	var b bytes.Buffer

// 	now := func() time.Time {
// 		return req.Timestamp
// 	}

// 	options := req.Options.toLib()
// 	fmt.Println("--SignSync options--")
// 	err := dkimS.SignSync(&b, r, options, now)
// 	fmt.Println("--SignSync--")
// 	fmt.Println("--SignSync err--", err)
// 	if err != nil {
// 		resp.Error = err.Error()
// 		return resp
// 	}
// 	resp.SignedEmail = b.String()
// 	return resp
// }

func SignARC(req *SignARCRequest) SignARCResponse {
	resp := SignARCResponse{Error: ""}

	r := strings.NewReader(req.EmailRaw)
	var b bytes.Buffer

	now := func() time.Time {
		return req.Timestamp
	}

	dnsResolver := NewDNSResolver()
	lookupTxt := func(name string) ([]string, error) {
		res, _, err := dnsResolver.LookupTXT(name)
		return res, err
	}

	options := req.Options.toLib()
	options.LookupTXT = lookupTxt
	fmt.Println("--SignSync options--")
	err := dkimS.SignARCSync(&b, r, options, now, req.MailFrom, req.IP)
	fmt.Println("--SignSync--")
	fmt.Println("--SignSync err--", err)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.SignedEmail = b.String()
	return resp
}

func VerifyARC(req *VerifyDKIMRequest) VerifyARCResponse {
	fmt.Println("--VerifyARC--")
	resp := VerifyARCResponse{Error: ""}
	// msg, err := dkim.ParseMessage(req.EmailRaw)
	// if err != nil {
	// 	resp.Error = err.Error()
	// 	return resp
	// }

	// dnsResolver := NewDNSResolver()
	// lookupTxt := func(name string) ([]string, error) {
	// 	res, _, err := dnsResolver.LookupTXT(name)
	// 	return res, err
	// }

	// res, err := dkim.VerifyArc(lookupTxt, req.PublicKey, msg)
	// fmt.Println("--VerifyArc err, res--", err, res)
	// if err != nil {
	// 	resp.Error = err.Error()
	// 	return resp
	// }

	// resbz, err := json.Marshal(res)
	// fmt.Println("--dkim.VerifyArc resbz--", err, string(resbz))

	// resp.Response = &ArcResult{}
	// fmt.Println("--VerifyArc FromLib pre--")
	// resp.Response.FromLib(res)
	// fmt.Println("--VerifyArc FromLib--")

	// resbz, err = json.Marshal(resp.Response)
	// fmt.Println("--dkim.VerifyArc resp.Response--", err, string(resbz))
	return resp
}

// func VerifyDKIM(req *VerifyDKIMRequest) VerifyDKIMResponse {
// 	fmt.Println("--VerifyDKIM--" + req.EmailRaw)
// 	resp := VerifyDKIMResponse{Error: ""}

// 	msg, err := dkim.ParseMessage(req.EmailRaw)
// 	if err != nil {
// 		resp.Error = err.Error()
// 		return resp
// 	}

// 	dnsResolver := NewDNSResolver()
// 	lookupTxt := func(name string) ([]string, error) {
// 		res, _, err := dnsResolver.LookupTXT(name)
// 		return res, err
// 	}

// 	// if we want to exclude domains
// 	// InvalidSigningEntityOption("com", "org", "net"),

// 	// if we want to fail if expiration date failed
// 	// SignatureTimingOption(5*time.Minute)
// 	res, err := dkim.Verify(DKIM_HEADER, msg, lookupTxt, req.PublicKey)
// 	fmt.Println("--dkim.Verify--", err, res)
// 	if err != nil {
// 		resp.Error = err.Error()
// 		return resp
// 	}

// 	resbz, err := json.Marshal(res)
// 	fmt.Println("--dkim.Verify resbz--", err, string(resbz))

// 	resp.Response = ResultArrFromLib(res)

// 	fmt.Println("--dkim.Verify post ResultArrFromLib--")

// 	resbz, err = json.Marshal(resp.Response)
// 	fmt.Println("--dkim.Verify resp.Response--", err, string(resbz))

// 	return resp
// }

func ForwardEmail(req *ForwardEmailRequest) ForwardEmailResponse {
	resp := ForwardEmailResponse{Error: ""}
	if len(req.To) == 0 {
		resp.Error = `missing To address`
		return resp
	}
	folder := req.Folder
	if folder == "" {
		folder = "INBOX"
	}
	getreq := &vmimap.ImapFetchRequest{
		Id:     req.ConnectionId,
		Folder: folder,
	}
	if req.Uid > 0 {
		getreq.UidSet = vmimap.UIDSet{vmimap.UIDRange{Start: vmimap.UID(req.Uid), Stop: vmimap.UID(req.Uid)}}
	}
	if req.MessageId != "" {
		criteria := &vmimap.SearchCriteria{}
		criteria.Header = []vmimap.SearchCriteriaHeaderField{{
			Key:   "Message-ID",
			Value: fmt.Sprintf("<%s>", req.MessageId),
		}}
		getreq.FetchFilter = &vmimap.FetchFilter{
			Search: criteria,
		}
	}

	emailresp := vmimap.Fetch(getreq)
	if emailresp.Error != "" {
		resp.Error = emailresp.Error
		return resp
	}
	if len(emailresp.Data) == 0 {
		resp.Error = "email not found"
		return resp
	}
	email := emailresp.Data[0]

	// fmt.Println("=================foundEmail")
	// fmt.Println(email.Raw)
	// fmt.Println("=====================")

	email = BuildForwardHeaders(email, req.From, req.To, req.Cc, req.Bcc, req.Options, req.Timestamp)
	emailstr, err := BuildRawEmail2(email, true)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	resp.EmailRaw = emailstr
	if req.SendEmail {
		sendresp := vmsmtp.SendMail(&vmsmtp.SmtpSendMailRequest{
			Id:    req.ConnectionId,
			From:  req.From.ToAddress(),
			To:    vmimap.ToAddresses(req.To),
			Email: []byte(emailstr),
		})
		fmt.Println("---forwarding err--", sendresp.Error)
		if sendresp.Error != "" {
			resp.Error = emailresp.Error
			return resp
		}
		fmt.Println("---forwarded--")
	}
	return resp
}

func BuildForwardHeaders(email vmimap.Email, from vmimap.Address, to []vmimap.Address, cc []vmimap.Address, bcc []vmimap.Address, opts SignOptions, timestamp time.Time) vmimap.Email {
	options := opts.toLib()

	// same body, same bh
	// unchanged headers:
	// MIME-Version
	// Content-Type
	// Content-Transfer-Encoding
	// existing DKIM-Signature
	// existing ARC headers

	// changed headers
	updatedHeaders := []string{
		vmimap.HEADER_FROM,
		vmimap.HEADER_TO,
		vmimap.HEADER_CC,
		vmimap.HEADER_BCC,
		vmimap.HEADER_DATE,
		vmimap.HEADER_MESSAGE_ID,
		vmimap.HEADER_SUBJECT,
		vmimap.HEADER_REFERENCES,
		vmimap.HEADER_IN_REPLY_TO,
	}
	// addedHeaders := updatedHeaders
	messageId := ""
	dkimCtxParams := make(map[string]string, 0)
	headers := make([]vmimap.Header, 0)

	for _, h := range email.Headers {
		switch strings.ToLower(h.Key) {
		case vmimap.HEADER_LOW_MESSAGE_ID:
			messageId = h.Value
			dkimCtxParams[vmimap.HEADER_MESSAGE_ID] = h.Value
		}
	}

	// we replace these headers
	for _, h := range email.Headers {
		switch strings.ToLower(h.Key) {
		case vmimap.HEADER_LOW_SUBJECT:
			h.Value = "Re: " + h.Value // TODO add from.toAddress()
			h.Raw = []byte{}
		case vmimap.HEADER_LOW_IN_REPLY_TO:
			h.Value = messageId
			h.Raw = []byte{}
		case vmimap.HEADER_LOW_REFERENCES:
			h.Value = messageId
			h.Raw = []byte{}
		case vmimap.HEADER_LOW_FROM:
			h.Value = vmimap.SerializeAddresses([]vmimap.Address{from})
			h.Raw = []byte{}
		case vmimap.HEADER_LOW_TO:
			h.Value = vmimap.SerializeAddresses(to)
			h.Raw = []byte{}
		case vmimap.HEADER_LOW_CC:
			h.Value = vmimap.SerializeAddresses(cc)
			h.Raw = []byte{}
		case vmimap.HEADER_LOW_BCC:
			h.Value = vmimap.SerializeAddresses(bcc)
			h.Raw = []byte{}
		case vmimap.HEADER_LOW_DATE:
			h.Value = timestamp.UTC().Format(time.RFC1123Z)
			h.Raw = []byte{}
		case vmimap.HEADER_LOW_MESSAGE_ID:
			continue
		case vmimap.HEADER_LOW_MIME_VERSION:
			h.Key = vmimap.HEADER_MIME_VERSION
		}

		if slices.Contains(updatedHeaders, h.Key) {
			if _, ok := dkimCtxParams[h.Key]; !ok {
				dkimCtxParams[h.Key] = h.Value
			}
		}
		// if slices.Contains(addedHeaders, h.Key) {
		// 	headers = append(headers, h)
		// }
		headers = append(headers, h)

		fmt.Println("--BuildForwardHeaders h.Value--", h.Value)
	}
	// replace headers
	email.Headers = headers

	dnsRegistryParams := make(map[string]string, 0)
	dnsRegistryParams["chain.id"] = wasmx.GetChainId()
	dnsRegistryFormatted := dkimS.FormatHeaderParams(HEADER_PROVABLE_DNS_REGISTRY, dnsRegistryParams)
	fmt.Println("--dnsRegistryFormatted--", dnsRegistryFormatted)
	dnsRegistryFormatted = strings.TrimRight(dnsRegistryFormatted, crlf) + crlf
	fmt.Println("--dnsRegistryFormatted--", dnsRegistryFormatted)

	emailRegistryParams := make(map[string]string, 0)
	emailRegistryParams["chain.id"] = wasmx.GetChainId()
	emailRegistryFormatted := dkimS.FormatHeaderParams(HEADER_PROVABLE_EMAIL_REGISTRY, emailRegistryParams)
	fmt.Println("--emailRegistryFormatted--", emailRegistryFormatted)
	emailRegistryFormatted = strings.TrimRight(emailRegistryFormatted, crlf) + crlf
	fmt.Println("--emailRegistryFormatted--", emailRegistryFormatted)

	dkimCtxFormatted := dkimS.FormatHeaderParams(HEADER_PROVABLE_FORWARD_ORIGIN_DKIM_CONTEXT, dkimCtxParams)
	fmt.Println("--dkimCtxFormatted--", dkimCtxFormatted)
	dkimCtxFormatted = strings.TrimRight(dkimCtxFormatted, crlf) + crlf
	fmt.Println("--dkimCtxFormatted--", dkimCtxFormatted)

	forwardSigParams := make(map[string]string, 0)
	forwardSigParams["v"] = "1"
	forwardSigParams["i"] = "1" // TODO fixme
	forwardSigParams["f"] = from.ToAddress()
	// forwardSigParams["a"] = "1" // NewSignerForward fills it in
	forwardSigParams["x"] = "1" // TODO DKIM-Signature hash ?
	forwardSigParams["h"] = strings.Join(updatedHeaders, ":")
	forwardSigParams["b"] = ""

	signer, err := dkimS.NewSignerForward(HEADER_PROVABLE_FORWARD_CHAIN_SIGNATURE, options, []byte(email.Raw), forwardSigParams)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}

	fmt.Println("--signature--", signer.Signature())
	fmt.Println("--signature b--", signer.B())
	forwardSigParams["b"] = signer.B()

	forwardSigFormatted := dkimS.FormatHeaderParams(HEADER_PROVABLE_FORWARD_CHAIN_SIGNATURE, forwardSigParams)
	fmt.Println("--forwardSigFormatted--", forwardSigFormatted)

	email.Headers.AppendTop(vmimap.Header{Key: HEADER_PROVABLE_DNS_REGISTRY, Raw: []byte(dnsRegistryFormatted)})
	email.Headers.AppendTop(vmimap.Header{Key: HEADER_PROVABLE_EMAIL_REGISTRY, Raw: []byte(emailRegistryFormatted)})
	email.Headers.AppendTop(vmimap.Header{Key: HEADER_PROVABLE_FORWARD_ORIGIN_DKIM_CONTEXT, Raw: []byte(dkimCtxFormatted)})
	email.Headers.AppendTop(vmimap.Header{Key: HEADER_PROVABLE_FORWARD_CHAIN_SIGNATURE, Raw: []byte(forwardSigFormatted)})

	return email
}
