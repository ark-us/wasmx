package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"

	"github.com/loredanacirstea/mailverif/arc"
	"github.com/loredanacirstea/mailverif/dkim"
	"github.com/loredanacirstea/mailverif/dns"
	"github.com/loredanacirstea/mailverif/forward"
	"github.com/loredanacirstea/mailverif/utils"
	sql "github.com/loredanacirstea/wasmx-env-sql"

	"github.com/loredanacirstea/emailchain/imap"

	"github.com/loredanacirstea/wasmx-env"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
	vmsmtp "github.com/loredanacirstea/wasmx-env-smtp"
)

// const ServerDomain = "provable.dev"
const MailServerDomain = "dmail.provable.dev"
const PortSmtpDirectAddr = "25"
const PortSmtpClientStartTls = "587"
const PortSmtpClientTls = "465"
const PortImapClientTls = "993"
const PortImapClientStartTls = "143"
const DefaultNetworkType = "tcp4"
const FolderInbox = "INBOX"
const FolderSent = "SENT"
const FolderDraft = "DRAFTS"
const FolderTrash = "TRASH"
const FolderJunk = "JUNK"
const FolderSpam = "SPAM"
const FolderArchive = "ARCHIVE"

func StartServer(req *StartServerRequest) {
	err := StoreDkimKey(req.SignOptions)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	smtpCfg := requiredSmtpDefaults(req.Smtp)

	// SMTP 25
	// smtpCfg.Addr = fmt.Sprintf("%s:%s", smtpCfg.Domain, PortSmtpDirectAddr)
	smtpCfg.Addr = fmt.Sprintf("%s:%s", "", PortSmtpDirectAddr)
	smtpCfg.StartTLS = true
	smtpCfg.EnableAuth = false
	resp := vmsmtp.ServerStart(&vmsmtp.ServerStartRequest{
		ConnectionId: PortSmtpDirectAddr,
		ServerConfig: requiredSmtpDefaults(smtpCfg),
	})
	if resp.Error != "" {
		wasmx.Revert([]byte(resp.Error))
	}

	// SMTP
	// smtpCfg.Addr = fmt.Sprintf("%s:%s", smtpCfg.Domain, PortSmtpClientTls)
	smtpCfg.Addr = fmt.Sprintf("%s:%s", "", PortSmtpClientTls)
	smtpCfg.StartTLS = false
	smtpCfg.EnableAuth = true
	resp = vmsmtp.ServerStart(&vmsmtp.ServerStartRequest{
		ConnectionId: PortSmtpClientTls,
		ServerConfig: smtpCfg,
	})
	if resp.Error != "" {
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpDirectAddr})
		wasmx.Revert([]byte(resp.Error))
	}

	// SMTP
	smtpCfg.Addr = fmt.Sprintf("%s:%s", "", PortSmtpClientStartTls)
	smtpCfg.StartTLS = true
	smtpCfg.EnableAuth = true
	resp = vmsmtp.ServerStart(&vmsmtp.ServerStartRequest{
		ConnectionId: PortSmtpClientStartTls,
		ServerConfig: smtpCfg,
	})
	if resp.Error != "" {
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpDirectAddr})
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpClientTls})
		wasmx.Revert([]byte(resp.Error))
	}

	// IMAP
	// req.Imap.Addr = fmt.Sprintf("%s:%s", req.Imap.Domain, PortImapClientTls)
	req.Imap.Addr = fmt.Sprintf("%s:%s", "", PortImapClientTls)
	req.Imap.StartTLS = false
	resp2 := vmimap.ServerStart(&vmimap.ServerStartRequest{
		ConnectionId: PortImapClientTls,
		ServerConfig: req.Imap,
	})
	if resp2.Error != "" {
		fmt.Println("---IMAP--", resp2.Error)
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpDirectAddr})
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpClientTls})
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpClientStartTls})
		wasmx.Revert([]byte(resp2.Error))
	}

	req.Imap.Addr = fmt.Sprintf("%s:%s", "", PortImapClientStartTls)
	req.Imap.StartTLS = true
	resp2 = vmimap.ServerStart(&vmimap.ServerStartRequest{
		ConnectionId: PortImapClientStartTls,
		ServerConfig: req.Imap,
	})
	if resp2.Error != "" {
		fmt.Println("---IMAP--", resp2.Error)
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpDirectAddr})
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpClientTls})
		vmsmtp.ServerClose(&vmsmtp.ServerCloseRequest{ConnectionId: PortSmtpClientStartTls})
		vmimap.ServerClose(&vmimap.ServerCloseRequest{ConnectionId: PortImapClientTls})
		wasmx.Revert([]byte(resp2.Error))
	}
}

func CreateAccount(req *CreateAccountRequest) {
	fmt.Println("--CreateAccount--", req.Username)

	err := ConnectSql(ConnectionId)
	if err != nil {
		wasmx.Revert([]byte("CreateAccount: DB connection failed: " + err.Error()))
	}

	// Insert into owners table
	params, err := paramsMarshal([]sql.SqlQueryParam{
		{Type: "text", Value: req.Username},
	})
	if err != nil {
		wasmx.Revert([]byte("CreateAccount: marshal error: " + err.Error()))
	}
	res := sql.Execute(&sql.SqlExecuteRequest{
		Id:     ConnectionId,
		Query:  `INSERT OR IGNORE INTO owners (address) VALUES (?)`,
		Params: params,
	})
	if res.Error != "" {
		wasmx.Revert([]byte("CreateAccount: insert owner failed: " + res.Error))
	}

	// TODO: Store req.Password securely (e.g., add password column to owners)

	// Create default folders: INBOX, SENT
	defaultFolders := []string{FolderInbox, FolderSent, FolderArchive, FolderDraft, FolderJunk, FolderSpam, FolderTrash}
	for i, folder := range defaultFolders {
		uidv := uint32(100 + i)
		params, err := paramsMarshal([]sql.SqlQueryParam{
			{Type: "text", Value: req.Username},
			{Type: "text", Value: folder},
			{Type: "integer", Value: uidv},
		})
		if err != nil {
			wasmx.Revert([]byte("CreateAccount: param marshal failed: " + err.Error()))
		}
		res := sql.Execute(&sql.SqlExecuteRequest{
			Id: ConnectionId,
			Query: `INSERT INTO folder_state (owner, folder, last_uid)
			         VALUES (?, ?, 0)
			         ON CONFLICT(owner, folder) DO NOTHING`,
			Params: params,
		})
		if res.Error != "" {
			wasmx.Revert([]byte("CreateAccount: create folder failed: " + res.Error))
		}
	}
}

func IncomingEmail(req *IncomingEmailRequest) {
	if len(req.From) == 0 {
		wasmx.Revert([]byte("incoming email: empty from"))
	}
	err := ConnectSql(ConnectionId)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}

	if req.ConnectionId == PortSmtpDirectAddr {
		// reject emails without messageId/date
		// reject emails with unregistered sender
		for _, to := range req.To {
			owners, err := GetAccount(to)
			if err != nil {
				wasmx.Revert([]byte(err.Error()))
				return
			}
			if len(owners) == 0 {
				wasmx.Revert([]byte("invalid account"))
				return
			}
		}
		email, err := extractEmail(req.EmailRaw)
		if err != nil {
			wasmx.Revert([]byte(err.Error()))
			return
		}
		// TODO validate
		if email.MessageID == "" {
			wasmx.Revert([]byte("invalid MessageID"))
			return
		}
		if email.InternalDate.UTC().Unix() == 0 {
			wasmx.Revert([]byte("empty date"))
			return
		}
		// if this a forwarded email, we do a check and add a header with the result
		timestamp := time.Unix(req.Timestamp, 0).UTC()
		emailraw := ApplyForwardCheck(req.EmailRaw, timestamp)
		// TODO get ipfrom from Received headers added by MTAs (Gmail, etc.)
		for _, to := range req.To {
			err = StoreEmail(to, req.From, emailraw, req.IpFrom, ConnectionId, FolderInbox)
			if err != nil {
				fmt.Println(err)
			}
		}
	} else {
		folder := FolderSent
		opts := LoadDkimKey()
		if opts == nil {
			fmt.Println("no dkim keys")
			folder = FolderDraft
		} else {
			// this is an email sent by an email client to our smtp server
			errs := SendRawEmail(req.From[0], req.To, req.EmailRaw, *opts)
			if err != nil {
				fmt.Println(errs)
				folder = FolderDraft
			}
		}
		err = StoreEmail(req.From[0], []string{}, req.EmailRaw, "", ConnectionId, folder)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func SendRawEmail(from string, tos []string, emailRaw []byte, opts SignOptions) []error {
	errs := []error{}
	for _, to := range tos {
		vals, err := extractHeaders(emailRaw, []string{vmimap.HEADER_MESSAGE_ID})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		generateMessageId := len(vals) == 0
		prepped, err := prepareEmailSend(opts, string(emailRaw), from, time.Now(), generateMessageId)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		err = sendEmailInternal(from, to, prepped, MailServerDomain, DefaultNetworkType)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = StoreEmail(from, []string{}, []byte(prepped), "", ConnectionId, FolderSent)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func SendEmail(req *SendMailRequest) []string {
	errs := []string{}
	err := ConnectSql(ConnectionId)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}

	opts := LoadDkimKey()
	if opts == nil {
		wasmx.Revert([]byte("no dkim keys"))
	}
	for _, to := range req.To {
		err = sendEmailInternal(req.From.ToAddress(), to.ToAddress(), string(req.EmailRaw), MailServerDomain, DefaultNetworkType)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		err = StoreEmail(req.From.ToString(), []string{}, req.EmailRaw, "", ConnectionId, FolderSent)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	return errs
}

func BuildAndSend(req *BuildAndSendMailRequest) []string {
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
		Date:    req.Date,
	}, mail.Header{})
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	emailstr, err := BuildRawEmail2(vmimap.Email{
		Headers: headers,
		Body: vmimap.EmailBody{
			ContentType: "text/plain",
			Parts:       []vmimap.BodyPart{{ContentType: "text/plain", Body: req.Body}},
		},
	})
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}

	err = ConnectSql(ConnectionId)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}

	opts := LoadDkimKey()
	if opts == nil {
		wasmx.Revert([]byte("no dkim keys"))
	}
	for _, to := range req.To {
		prepped, err := prepareEmailSend(*opts, emailstr, req.From, req.Date, true)
		if err != nil {
			errs = append(errs, err.Error())
			fmt.Println("---prepareEmailSend err-----", err)
			continue
		}
		err = sendEmailInternal(req.From, to, prepped, MailServerDomain, DefaultNetworkType)
		if err != nil {
			errs = append(errs, err.Error())
			fmt.Println("---sendEmailInternal err-----", err)
			continue
		}

		err = StoreEmail(req.From, []string{}, []byte(prepped), "", ConnectionId, FolderSent)
		if err != nil {
			fmt.Println("--StoreEmail--", err)
			errs = append(errs, err.Error())
		}
	}
	return errs
}

func VerifyDKIM(req *VerifyDKIMRequest) VerifyDKIMResponse {
	resp := VerifyDKIMResponse{Error: ""}
	dnsResolver := NewDNSResolver()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	now := func() time.Time {
		return req.Timestamp
	}

	results, err := dkim.Verify2(logger, dnsResolver, false, dkim.DefaultPolicy, []byte(req.EmailRaw), false, false, true, now, req.PublicKey)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.Response = results
	return resp
}

func SignDKIM(req *SignDKIMRequest) SignDKIMResponse {
	resp := SignDKIMResponse{Error: ""}
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
	header, err := dkim.Sign2(logger, identif, domain, selectors, false, []byte(req.EmailRaw), now)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.Header = utils.SerializeHeaders(header)

	return resp
}

func SignARC(req *SignARCRequest) SignARCResponse {
	resp := SignARCResponse{Error: ""}
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
	headers, err := arc.Sign(logger, resolver, domain, selectors, false, []byte(req.EmailRaw), req.MailFrom, req.IP, false, false, now, nil)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	resp.Header = utils.SerializeHeaders(headers)
	return resp
}

func VerifyARC(req *VerifyDKIMRequest) VerifyARCResponse {
	resp := VerifyARCResponse{Error: ""}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	resolver := NewDNSResolver()
	now := func() time.Time {
		return req.Timestamp
	}
	res, err := arc.Verify(logger, resolver, false, []byte(req.EmailRaw), false, true, now, req.PublicKey)
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
	getreq := &FetchRequest{
		Username: req.From.ToAddress(),
		Mailbox:  folder,
	}
	if req.Uid > 0 {
		getreq.UidSet = imap.UIDSet{imap.UIDRange{Start: imap.UID(req.Uid), Stop: imap.UID(req.Uid)}}
	}
	// TODO
	// if req.MessageId != "" {
	// 	criteria := &vmimap.SearchCriteria{}
	// 	criteria.Header = []vmimap.SearchCriteriaHeaderField{{
	// 		Key:   "Message-ID",
	// 		Value: fmt.Sprintf("<%s>", req.MessageId),
	// 	}}
	// 	getreq.FetchFilter = &imap.FetchFilter{
	// 		Search: criteria,
	// 	}
	// }

	emails, err := GetEmails(getreq)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	if len(emails) == 0 {
		resp.Error = "email not found"
		return resp
	}
	email, err := emails[0].ToEmail()
	if err != nil {
		resp.Error = err.Error()
		return resp
	}


	opts := LoadDkimKey()
	if opts == nil {
		resp.Error = "no dkim keys"
		return resp
	}
	resolver := NewDNSResolver()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	domain := dns.Domain{ASCII: opts.Domain}
	timeNow := func() time.Time {
		return req.Timestamp
	}
	key := ToPrivateKey(opts.PrivateKeyType, opts.PrivateKey)
	sel := dkim.Selector{
		Hash:          "sha256",
		PrivateKey:    key,
		Headers:       opts.HeaderKeys,
		Domain:        dns.Domain{ASCII: opts.Selector},
		HeaderRelaxed: opts.HeaderRelaxed,
		BodyRelaxed:   opts.BodyRelaxed,
	}
	selectors := []dkim.Selector{sel}
	dkimSel := dkim.Selector{
		Hash:          "sha256",
		PrivateKey:    key,
		Domain:        dns.Domain{ASCII: opts.Selector},
		HeaderRelaxed: opts.HeaderRelaxed,
		BodyRelaxed:   opts.BodyRelaxed,
		Headers:       strings.Split("From,To,Cc,Bcc,Reply-To,Subject,Date", ","),
		SealHeaders:   true,
	}
	mailfrom := email.Envelope.From[0].ToAddress()
	timestamp := req.Timestamp
	ipfrom := email.IpFrom
	subjectAddl := req.AdditionalSubject
	from := &mail.Address{Name: req.From.Name, Address: req.From.ToAddress()}
	to := make([]*mail.Address, len(req.To))
	for i, v := range req.To {
		to[i] = &mail.Address{Name: v.Name, Address: v.ToAddress()}
	}
	messageId, err := BuildMessageID(*opts, timestamp)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	header, br, err := forward.Forward(
		logger, resolver, domain, selectors,
		false, []byte(email.RawEmail), mailfrom, ipfrom,
		from, to, nil, nil,
		subjectAddl, timestamp,
		messageId, false, true, timeNow, nil,
	)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	bodyBytes, err := io.ReadAll(br)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	// also add dkim signature for this instance
	dkimHeaders, err := dkim.Sign(logger, utils.Localpart(req.From.Mailbox), domain, []dkim.Selector{dkimSel}, false, header, bufio.NewReader(bytes.NewReader(bodyBytes)), timeNow)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	header = append(dkimHeaders, header...)

	// compute new email
	emailstr := utils.SerializeHeaders(header) + "\r\n" + string(bodyBytes)

	resp.EmailRaw = emailstr
	if req.SendEmail {
		folder := FolderSent
		errs := []string{}
		// TODO ideally first save in DRAFTS, and after sending, store in SENT
		for _, to := range req.To {
			err = sendEmailInternal(req.From.ToAddress(), to.ToAddress(), emailstr, MailServerDomain, DefaultNetworkType)
			if err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(errs) > 0 {
			folder = FolderDraft
		}
		err = StoreEmail(req.From.ToAddress(), []string{}, []byte(emailstr), "", ConnectionId, folder)
		if err != nil {
			errs = append(errs, err.Error())
		}
		resp.Error = strings.Join(errs, "; ")
		return resp
	}
	return resp
}

func requiredSmtpDefaults(cfg vmsmtp.ServerConfig) vmsmtp.ServerConfig {
	cfg.AllowInsecureAuth = false
	// cfg.MaxLineLength
	// cfg.ReadTimeout = 10 * time.Second
	// cfg.WriteTimeout = 10 * time.Second
	// cfg.MaxMessageBytes = 10 << 20 // 10 MiB
	cfg.MaxRecipients = 100
	cfg.EnableSMTPUTF8 = false
	// TODO eventually this should be true
	cfg.EnableREQUIRETLS = false
	return cfg
}

func ApplyForwardCheck(emailraw []byte, timestamp time.Time) []byte {
	resolver := NewDNSResolver()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	timeNow := func() time.Time {
		return timestamp
	}

	res, err := forward.Verify(logger, resolver, false, emailraw, false, false, timeNow, nil)
	if err != nil {
		fmt.Println("forward verify: ", err)
		return emailraw
	}
	if res.Result.Err == forward.ErrMsgNotSigned {
		return emailraw
	}

	status := string(res.Result.Status)
	if res.Result.Err != nil {
		status += fmt.Sprintf(" (%s)", res.Result.Err.Error())
	}
	// add header
	HEADER_FORWARD_CHECK := "Provable-Forward-Check"
	header := utils.Header{
		Key:   HEADER_FORWARD_CHECK,
		Value: []byte(fmt.Sprintf(` %s%s`, status, utils.CRLF)),
	}
	header.RebuildRaw()
	return append(header.Raw, emailraw...)
}
