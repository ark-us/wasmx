package vmsmtp

import (
	"crypto/tls"
	"io"
	"log"
	"time"

	smtp "github.com/emersion/go-smtp"
)

type Server struct {
	// The type of network, "tcp" or "unix".
	Network string `json:"network"`
	// TCP or Unix address to listen on.
	Addr        string `json:"address"`
	TLSCertFile string `json:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file"`
	// Enable LMTP mode, as defined in RFC 2033.
	LMTP bool

	Domain            string
	MaxRecipients     int
	MaxMessageBytes   int64
	MaxLineLength     int
	AllowInsecureAuth bool
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration

	// Advertise SMTPUTF8 (RFC 6531) capability.
	// Should be used only if backend supports it.
	EnableSMTPUTF8 bool

	// Advertise REQUIRETLS (RFC 8689) capability.
	// Should be used only if backend supports it.
	EnableREQUIRETLS bool

	// Advertise BINARYMIME (RFC 3030) capability.
	// Should be used only if backend supports it.
	EnableBINARYMIME bool

	// Advertise DSN (RFC 3461) capability.
	// Should be used only if backend supports it.
	EnableDSN bool

	// Advertise RRVS (RFC 7293) capability.
	// Should be used only if backend supports it.
	EnableRRVS bool

	// Advertise DELIVERBY (RFC 2852) capability.
	// Should be used only if backend supports it.
	EnableDELIVERBY bool
	// The minimum time, with seconds precision, that a client
	// may specify in the BY argument with return mode.
	// A zero value indicates no set minimum.
	// Only use if DELIVERBY is enabled.
	MinimumDeliverByTime time.Duration
}

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

func NewServer() {
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
