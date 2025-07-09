package vmimap

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
)

type TlsConfig struct {
	ServerName  string `json:"server_name"`
	TLSCertFile string `json:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file"`
}

type ServerConfig struct {
	TlsConfig *TlsConfig `json:"tls_config"`
	Addr      string     `json:"address"`
	// The type of network, "tcp", "tcp4", or "unix".
	Network string `json:"network"`
}

type Session struct {
	ctx *Context
}

func (s *Session) Close() error {
	fmt.Println("-Session.Close--")
	return nil
}

// Not authenticated state
func (s *Session) Login(username, password string) error {
	fmt.Println("-Session.Login--")
	return nil
}

// Authenticated state
func (s *Session) Select(mailbox string, options *imap.SelectOptions) (*imap.SelectData, error) {
	fmt.Println("-Session.Select--")
	return nil, nil
}
func (s *Session) Create(mailbox string, options *imap.CreateOptions) error {
	fmt.Println("-Session.Create--")
	return nil
}
func (s *Session) Delete(mailbox string) error {
	fmt.Println("-Session.Delete--", mailbox)
	return nil
}
func (s *Session) Rename(mailbox, newName string) error {
	fmt.Println("-Session.Rename--", mailbox, newName)
	return nil
}
func (s *Session) Subscribe(mailbox string) error {
	fmt.Println("-Session.Subscribe--", mailbox)
	return nil
}
func (s *Session) Unsubscribe(mailbox string) error {
	fmt.Println("-Session.Unsubscribe--")
	return nil
}
func (s *Session) List(w *imapserver.ListWriter, ref string, patterns []string, options *imap.ListOptions) error {
	fmt.Println("-Session.List--")
	return nil
}
func (s *Session) Status(mailbox string, options *imap.StatusOptions) (*imap.StatusData, error) {
	fmt.Println("-Session.Status--", mailbox)
	return nil, nil
}
func (s *Session) Append(mailbox string, r imap.LiteralReader, options *imap.AppendOptions) (*imap.AppendData, error) {
	fmt.Println("-Session.Append--")
	return nil, nil
}
func (s *Session) Poll(w *imapserver.UpdateWriter, allowExpunge bool) error {
	fmt.Println("-Session.Poll--")
	return nil
}
func (s *Session) Idle(w *imapserver.UpdateWriter, stop <-chan struct{}) error {
	fmt.Println("-Session.Idle--")
	return nil
}

// Selected state
func (s *Session) Unselect() error {
	fmt.Println("-Session.Unselect--")
	return nil
}
func (s *Session) Expunge(w *imapserver.ExpungeWriter, uids *imap.UIDSet) error {
	fmt.Println("-Session.Expunge--")
	return nil
}
func (s *Session) Search(kind imapserver.NumKind, criteria *imap.SearchCriteria, options *imap.SearchOptions) (*imap.SearchData, error) {
	fmt.Println("-Session.Search--")
	return nil, nil
}
func (s *Session) Fetch(w *imapserver.FetchWriter, numSet imap.NumSet, options *imap.FetchOptions) error {
	fmt.Println("-Session.Fetch--")
	return nil
}
func (s *Session) Store(w *imapserver.FetchWriter, numSet imap.NumSet, flags *imap.StoreFlags, options *imap.StoreOptions) error {
	fmt.Println("-Session.Store--")
	return nil
}
func (s *Session) Copy(numSet imap.NumSet, dest string) (*imap.CopyData, error) {
	fmt.Println("-Session.Copy--")
	return nil, nil
}

type backend struct {
	ctx *Context
}

func (b *backend) NewSession(conn *imapserver.Conn) (imapserver.Session, *imapserver.GreetingData, error) {
	fmt.Println("--backend.NewSession--", conn.NetConn().LocalAddr(), conn.NetConn().RemoteAddr())
	return &Session{ctx: b.ctx}, &imapserver.GreetingData{PreAuth: true}, nil
}

func NewServer(cfg ServerConfig, ctx *Context) (*imapserver.Server, error) {
	b := &backend{ctx: ctx}
	opts := &imapserver.Options{
		NewSession: b.NewSession,
		Caps:       nil, // IMAP4rev1 only – fine for PoC
		// Logger: ctx.Ctx.Logger(),
		InsecureAuth: false,
	}

	tlsCfg, err := getTlsConfig(cfg.TlsConfig)
	if err != nil {
		return nil, err
	}
	if tlsCfg != nil {
		opts.TLSConfig = tlsCfg
	}

	s := imapserver.New(opts)
	startfn := func() error {
		return ListenAndServe(s, cfg.Addr, cfg.Network, tlsCfg)
	}
	log.Printf("IMAP listening on %s", cfg.Addr)

	srvDone := make(chan struct{}, 1)
	ctx.GoRoutineGroup.Go(func() error {
		err := startGoRoutine(ctx, s, startfn, cfg, srvDone)
		if err != nil {
			ctx.Ctx.Logger().Error(err.Error())
		}
		return err
	})

	if err := s.ListenAndServe(cfg.Addr); err != nil {
		return nil, err
	}
	return s, nil
}

func startGoRoutine(
	ctx *Context,
	s *imapserver.Server,
	startfn func() error,
	cfg ServerConfig,
	srvDone chan struct{},
) error {
	errCh := make(chan error, 1)
	go func() {
		if err := startfn(); err != nil {
			if err == net.ErrClosed {
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
		ctx.Ctx.Logger().Info("stopping imap server...", "addr", cfg.Addr, "network", cfg.Network)
		err := s.Close()
		if err != nil {
			ctx.Ctx.Logger().Error("stopping imap server error: ", err.Error())
		}
		close(errCh)
		return nil
	case err := <-errCh:
		ctx.Ctx.Logger().Error("failed to boot imap server", "error", err.Error())
		return err
	}
}

func ListenAndServe(s *imapserver.Server, addr string, network string, tlsCfg *tls.Config) error {
	if addr == "" {
		addr = ":143"
		if tlsCfg != nil {
			addr = ":993"
		}
	}
	var ln net.Listener
	var err error
	if tlsCfg == nil {
		ln, err = net.Listen(network, addr)
	} else {
		ln, err = tls.Listen(network, addr, tlsCfg)
	}
	if err != nil {
		return err
	}
	return s.Serve(ln)
}
