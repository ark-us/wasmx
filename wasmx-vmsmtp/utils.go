package vmsmtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"
	gosmtp "github.com/emersion/go-smtp"

	"golang.org/x/oauth2"
)

func connectSmtpClient(
	ctx context.Context,
	serverUrl string,
	startTls bool,
	networkType string,
	auth *ConnectionAuth,
	tlsConfig *TlsConfig,
) (sclient *gosmtp.Client, err error) {
	cfg, err := getTlsConfig(tlsConfig)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{}
	var conn net.Conn
	if cfg == nil {
		conn, err = dialer.DialContext(ctx, networkType, serverUrl)
	} else {
		tlsDialer := tls.Dialer{
			NetDialer: dialer,
			Config:    cfg,
		}
		conn, err = tlsDialer.Dial(networkType, serverUrl)
	}
	if err != nil {
		return nil, err
	}

	if startTls {
		sclient, err = gosmtp.NewClientStartTLS(conn, cfg)
	} else {
		sclient = gosmtp.NewClient(conn)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SMTP: %v", err)
	}
	if auth != nil {
		authClient := getAuthClient(*auth)
		if err = sclient.Auth(authClient); err != nil {
			return nil, fmt.Errorf("authentication failed: %v", err)
		}
	}
	return sclient, nil
}

func getTlsConfig(cfg *TlsConfig) (*tls.Config, error) {
	if cfg == nil {
		return nil, nil
	}
	config := &tls.Config{InsecureSkipVerify: false}
	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading TLS cert: %v", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}
	if cfg.ServerName != "" {
		config.ServerName = cfg.ServerName
	}
	return config, nil
}

func getAuthClient(auth ConnectionAuth) sasl.Client {
	switch auth.AuthType {
	case "password":
		return sasl.NewPlainClient(auth.Identity, auth.Username, auth.Password)
	case "oauth2":
		return &OAuth2Authenticator{username: auth.Username, accessToken: auth.Password}
	}
	return nil
}

// BuildRawEmail builds a full MIME email from an Email struct.
func BuildRawEmail(e Email) (string, error) {
	var buf bytes.Buffer

	// Build the top-level headers
	hdr := mail.Header{}

	// From
	addrs := make([]string, len(e.Envelope.From))
	for i, addr := range e.Envelope.From {
		if addr.Name != "" {
			addrs[i] = fmt.Sprintf("%s <%s@%s>", addr.Name, addr.Mailbox, addr.Host)
		} else {
			addrs[i] = fmt.Sprintf("%s@%s", addr.Mailbox, addr.Host)
		}
	}
	hdr.Set("From", strings.Join(addrs, ", "))

	// To
	toAddrs := make([]string, len(e.Envelope.To))
	for i, addr := range e.Envelope.To {
		if addr.Name != "" {
			toAddrs[i] = fmt.Sprintf("%s <%s@%s>", addr.Name, addr.Mailbox, addr.Host)
		} else {
			toAddrs[i] = fmt.Sprintf("%s@%s", addr.Mailbox, addr.Host)
		}
	}
	hdr.Set("To", strings.Join(toAddrs, ", "))

	// Subject
	hdr.Set("Subject", e.Envelope.Subject)

	// Date
	hdr.Set("Date", time.Now().UTC().Format(time.RFC1123Z))

	// Copy any extra header fields
	for k, vals := range e.Header {
		// Skip keys we've already set above
		lc := strings.ToLower(k)
		if lc == "from" || lc == "to" || lc == "subject" || lc == "date" || lc == "message-id" {
			continue
		}
		for _, v := range vals {
			hdr.Add(k, v)
		}
	}

	// Create the mail writer
	mw, err := mail.CreateWriter(&buf, hdr)
	if err != nil {
		return "", fmt.Errorf("mail.CreateWriter: %v", err)
	}

	// Write body and attachments
	// Always write something (even if empty) so the SMTP server sees a body part.
	header := mail.InlineHeader{}
	if len(e.Attachments) == 0 {
		header.SetContentType("text/plain", map[string]string{"charset": "UTF-8"})
	}
	bodyWriter, err := mw.CreateSingleInline(header)
	if err != nil {
		return "", fmt.Errorf("failed to create body writer: %v", err)
	}
	if _, err := bodyWriter.Write([]byte(e.Body)); err != nil {
		return "", fmt.Errorf("failed to write body: %v", err)
	}
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

func refreshToken(goCtx context.Context, refreshToken string, oauthConfig *oauth2.Config) string {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	newToken, err := oauthConfig.TokenSource(goCtx, token).Token()
	if err != nil {
		log.Fatalf("Failed to refresh token: %v", err)
	}

	return newToken.AccessToken
}
