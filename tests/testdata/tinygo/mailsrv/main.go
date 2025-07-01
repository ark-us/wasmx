// main.go
package main

import (
	"crypto/tls"
	"io"
	"log"
	"time"

	smtp "github.com/emersion/go-smtp"
)

// --- SMTP backend ------------------------------------------------------------

type backend struct{}

//	func (*backend) Login(_ *smtp.ConnectionState, _, _ string) (smtp.Session, error) {
//		return &session{}, nil // we don’t require AUTH to receive
//	}
//
//	func (*backend) AnonymousLogin(_ *smtp.ConnectionState) (smtp.Session, error) {
//		return &session{}, nil
//	}
func (*backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &session{}, nil
}

type session struct{}

func (*session) Mail(from string, _ *smtp.MailOptions) error { return nil }
func (*session) Rcpt(to string, _ *smtp.RcptOptions) error   { return nil }
func (*session) Data(r io.Reader) error {
	msg, _ := io.ReadAll(r)
	log.Printf("=== New message ===\n%s\n", msg) // write to journald/syslog
	return nil
}
func (*session) Reset()        {}
func (*session) Logout() error { return nil }

// --- Main --------------------------------------------------------------------

func main() {
	s := smtp.NewServer(&backend{})
	s.Addr = ":25"
	s.Domain = "provable.dev"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 10 << 20 // 10 MB
	s.MaxRecipients = 100
	s.AllowInsecureAuth = false // we’re receive-only; no AUTH offered
	// s.AuthDisabled = true

	// TLS (STARTTLS is advertised automatically)
	cert, err := tls.LoadX509KeyPair(
		"/etc/letsencrypt/live/dmail.provable.dev/fullchain.pem",
		"/etc/letsencrypt/live/dmail.provable.dev/privkey.pem")
	if err != nil {
		log.Fatal(err)
	}
	s.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}

	log.Printf("SMTP server listening on %s", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
