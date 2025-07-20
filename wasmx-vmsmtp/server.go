package vmsmtp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/emersion/go-sasl"

	smtp "github.com/emersion/go-smtp"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
)

type backend struct {
	ctx          *Context
	auth         bool
	connectionId string
}

func (b *backend) NewSession(conn *smtp.Conn) (smtp.Session, error) {
	// TODO implement reject policy
	// conn.Reject()
	fmt.Println("--smtpbackend.NewSession--", conn.Hostname(), conn.Server().Addr, conn.Server().Network)
	sess := Session{ctx: b.ctx, ConnectionId: b.connectionId}
	if b.auth {
		return &AuthSession{Session: sess}, nil
	}
	return &sess, nil
}

type Session struct {
	ConnectionId string
	ctx          *Context
	From         []string `json:"from"`
	To           []string `json:"to"`
	EmailRaw     []byte   `json:"email_raw"`
}

type AuthSession struct {
	Session
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	fmt.Println("--smtp.Session.Mail--", from, opts)
	if opts != nil {
		fmt.Println("--Session.Mail opts--", opts.EnvelopeID, opts.Auth)
	}
	s.From = append(s.From, from)
	return nil
}
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	fmt.Println("--smtp.Session.Rcpt--", to, opts)
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
	fmt.Println("--smtp.Session.Reset--")
}
func (s *Session) Logout() error {
	fmt.Println("--smtp.Session.Logout--")
	s.ctx.HandleIncomingEmail(*s)
	return nil
}

// support AuthSession
// Implement AuthMechanisms
func (s *AuthSession) AuthMechanisms() []string {
	fmt.Println("--smtp.AuthMechanisms--")
	return []string{"PLAIN"}
}

// Implement Auth
func (s *AuthSession) Auth(mech string) (sasl.Server, error) {
	fmt.Println("--smtp.Auth--", mech)
	switch mech {
	case sasl.Plain:
		return sasl.NewPlainServer(func(identity, username, password string) error {
			msg := &ReentryCalldata{
				Login: &LoginRequest{Username: username, Password: password},
			}
			_, err := s.ctx.HandleServerReentry(msg)
			if err != nil {
				return fmt.Errorf("invalid credentials: %w", err)
			}
			return nil
		}), nil
	case sasl.OAuthBearer:
		// TODO implement me
		return nil, fmt.Errorf("unsupported auth mechanism")
	default:
		return nil, fmt.Errorf("unsupported auth mechanism")
	}
}

func NewServer(cfg ServerConfig, ctx *Context, connectionId string) (*smtp.Server, error) {
	s := smtp.NewServer(&backend{ctx: ctx, auth: cfg.EnableAuth, connectionId: connectionId})
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

	fmt.Println("--NewServer AllowInsecureAuth--", cfg.AllowInsecureAuth)

	startfn := s.ListenAndServe

	tlsCfg, err := getTlsConfig(cfg.TlsConfig)
	if err != nil {
		return nil, err
	}
	if tlsCfg != nil {
		s.TLSConfig = tlsCfg
		fmt.Println("---Smtp.StartTLS--", cfg.StartTLS)
		if !cfg.StartTLS {
			startfn = s.ListenAndServeTLS
		}
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

func (ctx *Context) HandleIncomingEmail(s Session) {
	fmt.Println("--smtp.HandleIncomingEmail--")
	if s.From == nil || s.To == nil || s.EmailRaw == nil {
		return
	}
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

func (ctx *Context) HandleServerReentry(msg *ReentryCalldata) ([]byte, error) {
	fmt.Println("--smtp.HandleServerReentry--")
	msgbz, err := json.Marshal(msg)
	if err != nil {
		ctx.Ctx.Logger().Error("cannot marshal reentry request", "error", err.Error())
		return nil, err
	}

	contractAddress := ctx.Env.Contract.Address

	fmt.Println("--smtp.HandleServerReentry msgbz---", string(msgbz))

	msgtosend := &networktypes.MsgReentry{
		Sender:     contractAddress.String(),
		Contract:   contractAddress.String(),
		EntryPoint: ENTRY_POINT_SMTP_SERVER,
		Msg:        msgbz,
	}
	_, resp, err := ctx.Context.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	fmt.Println("--smtp.HandleServerReentry resp---", err, string(resp))
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error())
		return nil, err
	}

	var rres networktypes.MsgReentryResponse
	err = rres.Unmarshal(resp)
	if err != nil {
		return nil, err
	}
	fmt.Println("--smtp.HandleServerReentry resp2---", string(rres.Data))

	return rres.Data, nil
}
