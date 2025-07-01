package vmsmtp

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	smtp "github.com/emersion/go-smtp"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
)

type ServerConfig struct {
	// The type of network, "tcp" or "unix".
	Network string `json:"network"`
	// TCP or Unix address to listen on.
	Addr        string `json:"address"` // ":25"
	TLSCertFile string `json:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file"`

	// Enable LMTP mode, as defined in RFC 2033.
	LMTP bool `json:"lmtp"`

	Domain            string        `json:"domain"`
	MaxRecipients     int           `json:"max_recipients"`
	MaxMessageBytes   int64         `json:"max_message_bytes"`
	MaxLineLength     int           `json:"max_line_length"`
	AllowInsecureAuth bool          `json:"allow_insecure_auth"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`

	// Advertise SMTPUTF8 (RFC 6531) capability.
	EnableSMTPUTF8 bool `json:"enable_smtp_utf8"`
	// Advertise REQUIRETLS (RFC 8689) capability.
	EnableREQUIRETLS bool `json:"enable_require_tls"`
	// Advertise BINARYMIME (RFC 3030) capability.
	EnableBINARYMIME bool `json:"enable_binary_mime"`
	// Advertise DSN (RFC 3461) capability.
	EnableDSN bool `json:"enable_dsn"`
	// Advertise RRVS (RFC 7293) capability.
	EnableRRVS bool `json:"enable_rrvs"`
	// Advertise DELIVERBY (RFC 2852) capability.
	// EnableDELIVERBY bool `json:"enable_deliver_by"`

	// The minimum time (seconds precision) a client may specify in BY=.
	// Only used if DELIVERBY is enabled.
	// MinimumDeliverByTime time.Duration `json:"minimum_deliver_by_time"`
}

type backend struct {
	ctx *Context
}

//	func (*backend) Login(_ *smtp.ConnectionState, _, _ string) (smtp.Session, error) {
//		return &Session{}, nil // we don’t require AUTH to receive
//	}
//
//	func (*backend) AnonymousLogin(_ *smtp.ConnectionState) (smtp.Session, error) {
//		return &Session{}, nil
//	}
func (b *backend) NewSession(conn *smtp.Conn) (smtp.Session, error) {
	fmt.Println("--backend.NewSession--", conn.Hostname(), conn.Server().Addr, conn.Server().Network)
	return &Session{ctx: b.ctx}, nil
}

type Session struct {
	ctx      *Context
	From     []string `json:"from"`
	To       []string `json:"to"`
	EmailRaw []byte   `json:"email_raw"`
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	fmt.Println("--Session.Mail--", from, opts)
	if opts != nil {
		fmt.Println("--Session.Mail opts--", opts.EnvelopeID, opts.Auth)
	}
	s.From = append(s.From, from)
	return nil
}
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	fmt.Println("--Session.Rcpt--", to, opts)
	if opts != nil {
		fmt.Println("--Session.Rcpt opts--", opts.OriginalRecipient, opts.OriginalRecipientType, opts.Notify)
	}
	s.To = append(s.To, to)
	return nil
}
func (s *Session) Data(r io.Reader) error {
	msg, _ := io.ReadAll(r)
	log.Printf("=== New message ===\n%s\n", msg) // write to journald/syslog
	s.EmailRaw = msg
	return nil
}
func (*Session) Reset() {
	fmt.Println("--Session.Reset--")
}
func (s *Session) Logout() error {
	fmt.Println("--Session.Logout--")
	s.ctx.HandleIncomingEmail(*s)
	return nil
}

func NewServer(cfg ServerConfig, ctx *Context) (*smtp.Server, error) {
	s := smtp.NewServer(&backend{ctx: ctx})
	s.Network = "tcp"
	s.Addr = ":25"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 10 << 20 // 10 MiB
	s.MaxRecipients = 100

	if cfg.Network != "" {
		s.Network = cfg.Network
	}
	if cfg.Addr != "" {
		s.Addr = cfg.Addr
	}
	if cfg.Domain != "" {
		s.Domain = cfg.Domain
	}
	if cfg.ReadTimeout != 0 {
		s.ReadTimeout = cfg.ReadTimeout
	}
	if cfg.WriteTimeout != 0 {
		s.WriteTimeout = cfg.WriteTimeout
	}
	if cfg.MaxMessageBytes != 0 {
		s.MaxMessageBytes = int64(cfg.MaxMessageBytes)
	}
	if cfg.MaxRecipients != 0 {
		s.MaxRecipients = cfg.MaxRecipients
	}
	if cfg.MaxLineLength != 0 {
		s.MaxLineLength = cfg.MaxLineLength
	}

	s.AllowInsecureAuth = cfg.AllowInsecureAuth // plain AUTH over cleartext?
	s.LMTP = cfg.LMTP

	s.EnableSMTPUTF8 = cfg.EnableSMTPUTF8
	s.EnableREQUIRETLS = cfg.EnableREQUIRETLS
	s.EnableBINARYMIME = cfg.EnableBINARYMIME
	s.EnableDSN = cfg.EnableDSN
	s.EnableRRVS = cfg.EnableRRVS
	// s.EnableDELIVERBY = cfg.EnableDELIVERBY
	// s.MinimumDeliverByTime = cfg.MinimumDeliverByTime

	startfn := s.ListenAndServe

	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			log.Fatalf("loading TLS cert: %v", err)
		}
		s.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
		startfn = s.ListenAndServeTLS
	}

	log.Printf("SMTP server listening on %s (%s)", s.Addr, s.Network)

	srvDone := make(chan struct{}, 1)
	ctx.GoRoutineGroup.Go(func() error {
		err := startGoRoutine(ctx, s, startfn, srvDone)
		if err != nil {
			ctx.Ctx.Logger().Error(err.Error())
		}
		return err
	})
	return s, nil
}

func startGoRoutine(
	ctx *Context,
	s *smtp.Server,
	startfn func() error,
	srvDone chan struct{},
) error {
	errCh := make(chan error, 1)
	go func() {
		if err := startfn(); err != nil {
			if err == smtp.ErrServerClosed {
				ctx.Ctx.Logger().Info("closing email server", "message", err.Error())
				close(srvDone)
				return
			}
			ctx.Ctx.Logger().Error("failed to serve email server", "error", err.Error())
			errCh <- err
		}
	}()

	select {
	case <-ctx.GoContextParent.Done():
		// The calling process canceled or closed the provided context, so we must
		// gracefully stop the websrv server.
		ctx.Ctx.Logger().Info("stopping email server...", "addr", s.Addr, "network", s.Network)
		err := s.Shutdown(ctx.GoContextParent)
		if err != nil {
			ctx.Ctx.Logger().Error("stopping email server error: ", err.Error())
		}
		close(errCh)
		return nil
	case err := <-errCh:
		ctx.Ctx.Logger().Error("failed to boot email server", "error", err.Error())
		return err
	}
}

type ReentryCalldata struct {
	IncomingEmail *Session `json:"IncomingEmail"`
}

func (ctx *Context) HandleIncomingEmail(s Session) {
	msg := &ReentryCalldata{
		IncomingEmail: &s}

	msgbz, err := json.Marshal(msg)
	if err != nil {
		ctx.Ctx.Logger().Error("cannot marshal Expunge", "error", err.Error())
		return
	}

	contractAddress := ctx.Env.Contract.Address

	msgtosend := &networktypes.MsgReentryWithGoRoutine{
		Sender:     contractAddress.String(),
		Contract:   contractAddress.String(),
		EntryPoint: ENTRY_POINT_SMTP,
		Msg:        msgbz,
	}
	_, _, err = ctx.Context.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error())
	}
}
