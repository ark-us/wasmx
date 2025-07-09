package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"

	"github.com/loredanacirstea/mailverif/arc"
	"github.com/loredanacirstea/mailverif/dkim"
	"github.com/loredanacirstea/mailverif/dns"
	"github.com/loredanacirstea/mailverif/utils"

	"github.com/loredanacirstea/wasmx-env"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
	vmsmtp "github.com/loredanacirstea/wasmx-env-smtp"
)

const ServerDomain = "provable.dev"
const MailServerDomain = "dmail.provable.dev"
const PortSmtpDirectAddr = "25"
const PortSmtpClientStartTls = "587"
const PortSmtpClientTls = "465"
const PortImapClientTls = "993"
const PortImapClient = "143"
const DefaultNetworkType = "tcp4"

func StartServer(req *StartServerRequest) {
	err := StoreDkimKey(req.Options)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}

	// SMTP 25
	req.Smtp.Addr = fmt.Sprintf("%s:%s", req.Smtp.Domain, PortSmtpDirectAddr)
	resp := vmsmtp.ServerStart(&vmsmtp.ServerStartRequest{
		ConnectionId: PortSmtpDirectAddr,
		ServerConfig: req.Smtp,
	})
	if resp.Error != "" {
		wasmx.Revert([]byte(resp.Error))
	}

	// SMTP 587
	req.Smtp.Addr = fmt.Sprintf("%s:%s", req.Smtp.Domain, PortSmtpClientStartTls)
	resp = vmsmtp.ServerStart(&vmsmtp.ServerStartRequest{
		ConnectionId: PortSmtpClientStartTls,
		ServerConfig: req.Smtp,
	})
	if resp.Error != "" {
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpDirectAddr})
		wasmx.Revert([]byte(resp.Error))
	}

	// IMAP 993
	req.Imap.Addr = PortImapClientTls // fmt.Sprintf("%s:%s", req.Imap.Domain, PortImapClientTls)
	resp2 := vmimap.ServerStart(&vmimap.ServerStartRequest{
		ConnectionId: PortImapClientTls,
		ServerConfig: req.Imap,
	})
	if resp2.Error != "" {
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpDirectAddr})
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpClientStartTls})
		wasmx.Revert([]byte(resp2.Error))
	}
}

func IncomingEmail(req *IncomingEmailRequest) {
	err := ConnectSql(ConnectionId)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	for _, to := range req.To {
		err = StoreEmail(MailServerDomain, req.From[0], to, req.EmailRaw, ConnectionId)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func SendEmail(req *BuildAndSendMailRequest) []string {
	errs := []string{}
	// from, err := mail.ParseAddress(req.From)
	// if err != nil {
	// 	wasmx.Revert([]byte(err.Error()))
	// }
	from := vmimap.AddressFromString(req.From, "")
	headers, err := SerializeEnvelope2(&vmimap.Envelope{
		Subject: req.Subject,
		From:    []vmimap.Address{from},
		To:      vmimap.AddressesFromString(req.To),
		Cc:      vmimap.AddressesFromString(req.Cc),
		Bcc:     vmimap.AddressesFromString(req.Bcc),
	}, mail.Header{})
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	emailstr, err := BuildRawEmail2(vmimap.Email{
		Headers: headers,
		Body: vmimap.EmailBody{
			Parts: []vmimap.BodyPart{{ContentType: "text/plain", Body: req.Body}},
		},
	})
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	fmt.Println("======SendEmail--")
	fmt.Println(emailstr)
	fmt.Println("====== END SendEmail--")

	opts := LoadDkimKey()
	if opts == nil {
		wasmx.Revert([]byte("no dkim keys"))
	}

	for _, to := range req.To {
		prepped, err := prepareEmailSend(*opts, emailstr, req.From)
		if err != nil {
			wasmx.Revert([]byte(err.Error()))
		}
		err = sendEmailInternal(req.From, to, prepped, MailServerDomain, DefaultNetworkType)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	fmt.Println("---sending err--", errs)
	return errs
}

func VerifyDKIM(req *VerifyDKIMRequest) VerifyDKIMResponse {
	resp := VerifyDKIMResponse{Error: ""}
	dnsResolver := NewDNSResolver()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	now := func() time.Time {
		return req.Timestamp
	}

	results, err := dkim.Verify2(logger, dnsResolver, false, dkim.DefaultPolicy, strings.NewReader(req.EmailRaw), false, true, now, req.PublicKey)
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
	identif := utils.Localpart(req.Options.Identifier)
	domain := dns.Domain{ASCII: req.Options.Domain}
	key := ToPrivateKey(req.Options.PrivateKeyType, req.Options.PrivateKey)
	sel := dkim.Selector{
		Hash:          "sha256",
		PrivateKey:    key,
		Headers:       req.Options.HeaderKeys,
		Domain:        dns.Domain{ASCII: req.Options.Selector},
		HeaderRelaxed: req.Options.HeaderRelaxed,
		BodyRelaxed:   req.Options.BodyRelaxed,
	}
	selectors := []dkim.Selector{sel}
	header, err := dkim.Sign2(logger, identif, domain, selectors, false, r, now)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.Header = utils.SerializeHeaders(header)

	return resp
}

func SignARC(req *SignARCRequest) SignARCResponse {
	resp := SignARCResponse{Error: ""}

	r := strings.NewReader(req.EmailRaw)
	now := func() time.Time {
		return req.Timestamp
	}
	resolver := NewDNSResolver()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	domain := dns.Domain{ASCII: req.Options.Domain}
	key := ToPrivateKey(req.Options.PrivateKeyType, req.Options.PrivateKey)
	sel := dkim.Selector{
		Hash:          "sha256",
		PrivateKey:    key,
		Headers:       req.Options.HeaderKeys,
		Domain:        dns.Domain{ASCII: req.Options.Selector},
		HeaderRelaxed: req.Options.HeaderRelaxed,
		BodyRelaxed:   req.Options.BodyRelaxed,
	}
	selectors := []dkim.Selector{sel}
	headers, err := arc.Sign(logger, resolver, domain, selectors, false, r, req.MailFrom, req.IP, req.MailServerDomain, false, false, now, nil)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.Header = utils.SerializeHeaders(headers)
	return resp
}

func VerifyARC(req *VerifyDKIMRequest) VerifyARCResponse {
	fmt.Println("--VerifyARC--")
	resp := VerifyARCResponse{Error: ""}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	msgr := strings.NewReader(req.EmailRaw)
	resolver := NewDNSResolver()
	now := func() time.Time {
		return req.Timestamp
	}
	res, err := arc.Verify(logger, resolver, false, msgr, false, true, now, req.PublicKey)
	fmt.Println("--VerifyArc err, res--", err, res)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.Response = res
	return resp
}

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

	// email = BuildForwardHeaders(email, req.From, req.To, req.Cc, req.Bcc, req.Options, req.Timestamp)
	emailstr, err := BuildRawEmail2(email)
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

// func BuildForwardHeaders(email vmimap.Email, from vmimap.Address, to []vmimap.Address, cc []vmimap.Address, bcc []vmimap.Address, opts SignOptions, timestamp time.Time) vmimap.Email {
// 	options := opts.toLib()

// 	// same body, same bh
// 	// unchanged headers:
// 	// MIME-Version
// 	// Content-Type
// 	// Content-Transfer-Encoding
// 	// existing DKIM-Signature
// 	// existing ARC headers

// 	// changed headers
// 	updatedHeaders := []string{
// 		vmimap.HEADER_FROM,
// 		vmimap.HEADER_TO,
// 		vmimap.HEADER_CC,
// 		vmimap.HEADER_BCC,
// 		vmimap.HEADER_DATE,
// 		vmimap.HEADER_MESSAGE_ID,
// 		vmimap.HEADER_SUBJECT,
// 		vmimap.HEADER_REFERENCES,
// 		vmimap.HEADER_IN_REPLY_TO,
// 	}
// 	// addedHeaders := updatedHeaders
// 	messageId := ""
// 	dkimCtxParams := make(map[string]string, 0)
// 	headers := make([]vmimap.Header, 0)

// 	for _, h := range email.Headers {
// 		switch strings.ToLower(h.Key) {
// 		case vmimap.HEADER_LOW_MESSAGE_ID:
// 			messageId = h.Value
// 			dkimCtxParams[vmimap.HEADER_MESSAGE_ID] = h.Value
// 		}
// 	}

// 	// we replace these headers
// 	for _, h := range email.Headers {
// 		switch strings.ToLower(h.Key) {
// 		case vmimap.HEADER_LOW_SUBJECT:
// 			h.Value = "Re: " + h.Value // TODO add from.toAddress()
// 			h.Raw = []byte{}
// 		case vmimap.HEADER_LOW_IN_REPLY_TO:
// 			h.Value = messageId
// 			h.Raw = []byte{}
// 		case vmimap.HEADER_LOW_REFERENCES:
// 			h.Value = messageId
// 			h.Raw = []byte{}
// 		case vmimap.HEADER_LOW_FROM:
// 			h.Value = vmimap.SerializeAddresses([]vmimap.Address{from})
// 			h.Raw = []byte{}
// 		case vmimap.HEADER_LOW_TO:
// 			h.Value = vmimap.SerializeAddresses(to)
// 			h.Raw = []byte{}
// 		case vmimap.HEADER_LOW_CC:
// 			h.Value = vmimap.SerializeAddresses(cc)
// 			h.Raw = []byte{}
// 		case vmimap.HEADER_LOW_BCC:
// 			h.Value = vmimap.SerializeAddresses(bcc)
// 			h.Raw = []byte{}
// 		case vmimap.HEADER_LOW_DATE:
// 			h.Value = timestamp.UTC().Format(time.RFC1123Z)
// 			h.Raw = []byte{}
// 		case vmimap.HEADER_LOW_MESSAGE_ID:
// 			continue
// 		case vmimap.HEADER_LOW_MIME_VERSION:
// 			h.Key = vmimap.HEADER_MIME_VERSION
// 		}

// 		if slices.Contains(updatedHeaders, h.Key) {
// 			if _, ok := dkimCtxParams[h.Key]; !ok {
// 				dkimCtxParams[h.Key] = h.Value
// 			}
// 		}
// 		// if slices.Contains(addedHeaders, h.Key) {
// 		// 	headers = append(headers, h)
// 		// }
// 		headers = append(headers, h)

// 		fmt.Println("--BuildForwardHeaders h.Value--", h.Value)
// 	}
// 	// replace headers
// 	email.Headers = headers

// 	dnsRegistryParams := make(map[string]string, 0)
// 	dnsRegistryParams["chain.id"] = wasmx.GetChainId()
// 	dnsRegistryFormatted := dkimS.FormatHeaderParams(HEADER_PROVABLE_DNS_REGISTRY, dnsRegistryParams)
// 	fmt.Println("--dnsRegistryFormatted--", dnsRegistryFormatted)
// 	dnsRegistryFormatted = strings.TrimRight(dnsRegistryFormatted, crlf) + crlf
// 	fmt.Println("--dnsRegistryFormatted--", dnsRegistryFormatted)

// 	emailRegistryParams := make(map[string]string, 0)
// 	emailRegistryParams["chain.id"] = wasmx.GetChainId()
// 	emailRegistryFormatted := dkimS.FormatHeaderParams(HEADER_PROVABLE_EMAIL_REGISTRY, emailRegistryParams)
// 	fmt.Println("--emailRegistryFormatted--", emailRegistryFormatted)
// 	emailRegistryFormatted = strings.TrimRight(emailRegistryFormatted, crlf) + crlf
// 	fmt.Println("--emailRegistryFormatted--", emailRegistryFormatted)

// 	dkimCtxFormatted := dkimS.FormatHeaderParams(HEADER_PROVABLE_FORWARD_ORIGIN_DKIM_CONTEXT, dkimCtxParams)
// 	fmt.Println("--dkimCtxFormatted--", dkimCtxFormatted)
// 	dkimCtxFormatted = strings.TrimRight(dkimCtxFormatted, crlf) + crlf
// 	fmt.Println("--dkimCtxFormatted--", dkimCtxFormatted)

// 	forwardSigParams := make(map[string]string, 0)
// 	forwardSigParams["v"] = "1"
// 	forwardSigParams["i"] = "1" // TODO fixme
// 	forwardSigParams["f"] = from.ToAddress()
// 	// forwardSigParams["a"] = "1" // NewSignerForward fills it in
// 	forwardSigParams["x"] = "1" // TODO DKIM-Signature hash ?
// 	forwardSigParams["h"] = strings.Join(updatedHeaders, ":")
// 	forwardSigParams["b"] = ""

// 	signer, err := dkimS.NewSignerForward(HEADER_PROVABLE_FORWARD_CHAIN_SIGNATURE, options, []byte(email.Raw), forwardSigParams)
// 	if err != nil {
// 		wasmx.Revert([]byte(err.Error()))
// 	}

// 	fmt.Println("--signature--", signer.Signature())
// 	fmt.Println("--signature b--", signer.B())
// 	forwardSigParams["b"] = signer.B()

// 	forwardSigFormatted := dkimS.FormatHeaderParams(HEADER_PROVABLE_FORWARD_CHAIN_SIGNATURE, forwardSigParams)
// 	fmt.Println("--forwardSigFormatted--", forwardSigFormatted)

// 	email.Headers.AppendTop(vmimap.Header{Key: HEADER_PROVABLE_DNS_REGISTRY, Raw: []byte(dnsRegistryFormatted)})
// 	email.Headers.AppendTop(vmimap.Header{Key: HEADER_PROVABLE_EMAIL_REGISTRY, Raw: []byte(emailRegistryFormatted)})
// 	email.Headers.AppendTop(vmimap.Header{Key: HEADER_PROVABLE_FORWARD_ORIGIN_DKIM_CONTEXT, Raw: []byte(dkimCtxFormatted)})
// 	email.Headers.AppendTop(vmimap.Header{Key: HEADER_PROVABLE_FORWARD_CHAIN_SIGNATURE, Raw: []byte(forwardSigFormatted)})

// 	return email
// }
