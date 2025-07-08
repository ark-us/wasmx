package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"

	"github.com/loredanacirstea/mailverif/dkim"
	"github.com/loredanacirstea/mailverif/dns"
	"github.com/loredanacirstea/mailverif/utils"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
	vmsmtp "github.com/loredanacirstea/wasmx-env-smtp"
)

// TODO remove me
// BuildRawEmail builds a full MIME email from an Email struct.
func BuildRawEmail(e vmimap.Email) (string, error) {
	var buf bytes.Buffer

	// Build the top-level headers
	hdr := mail.Header{}
	// Copy any extra header fields
	for _, h := range e.Headers {
		hdr.Add(h.Key, h.Value)
	}

	// Create the mail writer
	mw, err := mail.CreateWriter(&buf, hdr)
	if err != nil {
		return "", fmt.Errorf("mail.CreateWriter: %v", err)
	}

	fmt.Println("==========body???==========")
	fmt.Println(e.Body)
	fmt.Println("===============")
	fmt.Println(hdr.Get(vmimap.HEADER_CONTENT_TYPE))
	fmt.Println("===============")

	// Write body and attachments
	// Always write something (even if empty) so the SMTP server sees a body part.
	header := mail.InlineHeader{}
	if len(e.Attachments) == 0 && !hdr.Has(vmimap.HEADER_CONTENT_TYPE) {
		header.SetContentType("text/plain", map[string]string{"charset": "UTF-8"})
	}
	bodyWriter, err := mw.CreateSingleInline(header)
	if err != nil {
		return "", fmt.Errorf("failed to create body writer: %v", err)
	}
	// if _, err := bodyWriter.Write([]byte(e.Body)); err != nil {
	// 	return "", fmt.Errorf("failed to write body: %v", err)
	// }
	if err := bodyWriter.Close(); err != nil {
		return "", fmt.Errorf("failed to close body writer: %v", err)
	}

	// Add each attachment
	for _, att := range e.Attachments {
		attachHeader := mail.AttachmentHeader{}
		attachHeader.SetFilename(att.Filename)
		attachHeader.Set("Content-Type", att.ContentType)
		aw, err := mw.CreateAttachment(attachHeader)
		if err != nil {
			return "", fmt.Errorf("create attachment part: %v", err)
		}
		if _, err := aw.Write(att.Data); err != nil {
			return "", fmt.Errorf("write attachment %q: %v", att.Filename, err)
		}
		if err := aw.Close(); err != nil {
			return "", fmt.Errorf("close attachment %q: %v", att.Filename, err)
		}
	}

	if err := mw.Close(); err != nil {
		return "", fmt.Errorf("close mail writer: %v", err)
	}

	return buf.String(), nil
}

func SerializeEnvelope(envelope *vmimap.Envelope, hdr mail.Header) mail.Header {
	if len(envelope.From) > 0 {
		hdr.Set(vmimap.HEADER_FROM, vmimap.SerializeAddresses(envelope.From))
	}
	if len(envelope.To) > 0 {
		hdr.Set(vmimap.HEADER_TO, vmimap.SerializeAddresses(envelope.To))
	}
	if len(envelope.Subject) > 0 {
		hdr.Set(vmimap.HEADER_SUBJECT, envelope.Subject)
	}
	if len(envelope.Bcc) > 0 {
		hdr.Set(vmimap.HEADER_BCC, vmimap.SerializeAddresses(envelope.Bcc))
	}
	if len(envelope.Cc) > 0 {
		hdr.Set(vmimap.HEADER_CC, vmimap.SerializeAddresses(envelope.Cc))
	}
	if len(envelope.ReplyTo) > 0 {
		hdr.Set(vmimap.HEADER_REPLY_TO, vmimap.SerializeAddresses(envelope.ReplyTo))
	}
	if len(envelope.MessageID) > 0 {
		hdr.Set(vmimap.HEADER_MESSAGE_ID, vmimap.SerializeMessageId(envelope.MessageID))
	}
	if len(envelope.InReplyTo) > 0 {
		hdr.Set(vmimap.HEADER_IN_REPLY_TO, vmimap.SerializeMessageIds(envelope.InReplyTo))
	}
	hdr.Set(vmimap.HEADER_DATE, time.Now().UTC().Format(time.RFC1123Z))
	return hdr
}

func BuildRawEmail2(e vmimap.Email, writeCrlfHeaders bool) (string, error) {
	headers := e.Headers
	bodyParts := e.Body.Parts
	boundary := e.Body.Boundary
	attachments := e.Attachments
	var b strings.Builder

	// crlf := "\r\n"
	// if !writeCrlfHeaders {
	// 	crlf = ""
	// }

	// Write headers
	for _, h := range headers {
		// fmt.Println("--BuildRawEmail2 h.Value--", h.Raw)
		// fmt.Fprintf(&b, "%s: %s%s", h.Key, h.Value, crlf)
		if len(h.Raw) > 0 {
			fmt.Fprintf(&b, string(h.Raw))
		} else {
			fmt.Fprintf(&b, "%s: %s\r\n", h.Key, h.Value)
		}
	}
	b.WriteString("\r\n") // end of headers

	// Determine boundary from headers
	// var boundary string
	// for _, h := range headers {
	// 	if strings.ToLower(h.Key) == "content-type" {
	// 		_, params, _ := mime.ParseMediaType(h.Value)
	// 		boundary = params["boundary"]
	// 		break
	// 	}
	// }
	// if boundary == "" {
	// 	return "", fmt.Errorf("no c found in Content-Type header")
	// }

	// Write body parts
	for _, bp := range bodyParts {
		fmt.Fprintf(&b, "--%s\r\n", boundary)
		fmt.Fprintf(&b, "Content-Type: %s\r\n\r\n", bp.ContentType)
		b.Write(bp.Body)
		// if !strings.HasSuffix(string(bp.Body), "\r\n") {
		// 	b.WriteString("\r\n")
		// }
		b.WriteString("\r\n")
	}

	for _, att := range attachments {
		fmt.Fprintf(&b, "--%s\r\n", boundary)
		fmt.Fprintf(&b, "Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename)
		fmt.Fprintf(&b, "Content-Type: %s\r\n", att.ContentType)
		fmt.Fprintf(&b, "Content-Transfer-Encoding: base64\r\n\r\n")
		encoded := base64.StdEncoding.EncodeToString(att.Data)
		b.WriteString(encoded + "\r\n")
	}

	if boundary != "" {
		fmt.Fprintf(&b, "--%s--\r\n", boundary)
	}

	return b.String(), nil
}

func prepareEmailSend(
	emailstr string,
	dkimSelector string,
	from string,
) (string, error) {
	parts := strings.Split(from, "@")
	fromUsername := parts[0]
	fromDomain := parts[1]
	date := time.Now().UTC().Format(time.RFC1123Z)
	emailstr = fmt.Sprintf("Date: %s\r\n", date) + emailstr
	emailstr, err := signDkim(emailstr, fromUsername, fromDomain, dkimSelector)
	if err != nil {
		return "", fmt.Errorf("signDkim: %s", err.Error())
	}
	return emailstr, nil
}

func signDkim(emailstr string, username string, domainstr string, selectorstr string) (string, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	r := strings.NewReader(emailstr)
	identif := utils.Localpart(username)
	domain := dns.Domain{ASCII: domainstr}
	key := ToPrivateKey("rsa", []byte(testPrivateKeyPEM))
	sel := dkim.Selector{
		Hash:          "sha256",
		PrivateKey:    key,
		Headers:       strings.Split("From,To,Cc,Bcc,Reply-To,References,In-Reply-To,Subject,Date,Message-ID,Content-Type", ","),
		Domain:        dns.Domain{ASCII: selectorstr},
		HeaderRelaxed: true,
		BodyRelaxed:   true,
	}
	selectors := []dkim.Selector{sel}
	header, err := dkim.Sign2(logger, identif, domain, selectors, false, r, time.Now)
	if err != nil {
		return "", fmt.Errorf("dkim.Sign2: %s", err.Error())
	}
	headerstr := utils.SerializeHeaders(header)
	return headerstr + emailstr, nil
}

func sendEmailInternal(
	from string, to string,
	emailstr string,
	mailServerDomain string,
	networkType string, // tcp, tcp4
) error {
	var err error
	at := strings.LastIndex(to, "@")
	if at == -1 {
		return fmt.Errorf("invalid recipient address")
	}
	toDomain := to[at+1:]

	// Step 2: lookup MX records
	mxRecords, err := net.LookupMX(toDomain)
	if err != nil || len(mxRecords) == 0 {
		return fmt.Errorf("no MX records found for domain %s", toDomain)
	}

	// Try each MX record in order
	for _, mx := range mxRecords {
		mxHost := strings.TrimSuffix(mx.Host, ".")
		addr := fmt.Sprintf("%s:25", mxHost)

		log.Printf("Trying to send to %s...", addr)

		tlsConfig := &vmsmtp.TlsConfig{
			ServerName: mxHost,
		}
		connResp := vmsmtp.ClientConnect(&vmsmtp.SmtpConnectionRequest{
			Id:          req.ConnectionId,
			ServerUrl:   addr,
			StartTLS:    true,
			NetworkType: networkType,
			TlsConfig:   tlsConfig,
		})
		if connResp.Error != "" {
			log.Printf("Failed to connect to %s: %v", addr, err)
			continue
		}
		hresp := vmsmtp.Hello(&vmsmtp.SmtpHelloRequest{Id: req.ConnectionId, LocalName: mailServerDomain})
		if hresp.Error != "" {
			log.Printf("EHLO/HELO failed: %v", err)
			vmsmtp.Quit(&vmsmtp.SmtpQuitRequest{Id: req.ConnectionId})
			continue
		}

		sendresp := vmsmtp.SendMail(&vmsmtp.SmtpSendMailRequest{
			Id:    req.ConnectionId,
			From:  from,
			To:    []string{to},
			Email: []byte(emailstr),
		})
		if sendresp.Error != "" {
			log.Println("Email send failed: ", err)
			vmsmtp.Quit(&vmsmtp.SmtpQuitRequest{Id: req.ConnectionId})
			return err
		}
		vmsmtp.Quit(&vmsmtp.SmtpQuitRequest{Id: req.ConnectionId})
		log.Println("Email sent successfully to", to)
		return nil
	}
	return fmt.Errorf("could not deliver email to any MX server for %s", toDomain)
}
