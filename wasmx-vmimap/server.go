package vmimap

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
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
	Network  string `json:"network"`
	StartTLS bool   `json:"start_tls"`
}

type Session struct {
	ctx           *Context
	username      string
	mailbox       string
	selectOptions *imap.SelectOptions
}

func (s *Session) Close() error {

	msg := &ReentryCalldataServer{
		Logout: &LogoutRequest{Username: s.username},
	}
	_, err := s.ctx.HandleServerReentry(msg)
	return err
}

// Not authenticated state
func (s *Session) Login(username, password string) error {
	s.username = username
	msg := &ReentryCalldataServer{
		Login: &LoginRequest{Username: s.username, Password: password},
	}
	_, err := s.ctx.HandleServerReentry(msg)
	return err
}

// Authenticated state
func (s *Session) Select(mailbox string, options *imap.SelectOptions) (*imap.SelectData, error) {
	s.mailbox = mailbox
	s.selectOptions = options
	msg := &ReentryCalldataServer{
		Select: &SelectRequest{Username: s.username, Mailbox: mailbox},
	}
	res, err := s.ctx.HandleServerReentry(msg)
	if err != nil {
		return nil, err
	}
	resp := &imap.SelectData{}
	err = json.Unmarshal(res, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (s *Session) Create(mailbox string, options *imap.CreateOptions) error {
	msg := &ReentryCalldataServer{
		Create: &CreateRequest{Username: s.username, Mailbox: mailbox, Options: options},
	}
	_, err := s.ctx.HandleServerReentry(msg)
	return err
}
func (s *Session) Delete(mailbox string) error {
	msg := &ReentryCalldataServer{
		Delete: &DeleteRequest{Username: s.username, Mailbox: mailbox},
	}
	_, err := s.ctx.HandleServerReentry(msg)
	return err
}
func (s *Session) Rename(mailbox, newName string) error {
	msg := &ReentryCalldataServer{
		Rename: &RenameRequest{Username: s.username, Mailbox: mailbox},
	}
	_, err := s.ctx.HandleServerReentry(msg)
	return err
}
func (s *Session) Subscribe(mailbox string) error {
	fmt.Println("-imap.Session.Subscribe NOT IMPLEMENTED--", mailbox)
	return nil
}
func (s *Session) Unsubscribe(mailbox string) error {
	fmt.Println("-imap.Session.Unsubscribe NOT IMPLEMENTED--", mailbox)
	return nil
}

// Lists visible mailboxes
func (s *Session) List(w *imapserver.ListWriter, ref string, patterns []string, options *imap.ListOptions) error {
	msg := &ReentryCalldataServer{
		List: &ListRequest{
			Username: s.username,
			Patterns: patterns,
			Options:  options,
		},
	}
	res, err := s.ctx.HandleServerReentry(msg)
	if err != nil {
		return err
	}
	resp := []imap.ListData{}
	err = json.Unmarshal(res, &resp)
	if err != nil {
		return err
	}
	for _, l := range resp {
		err = w.WriteList(&l)
		if err != nil {
			fmt.Println("error WriteList", err)
		}
	}
	return nil
}
func (s *Session) Status(mailbox string, options *imap.StatusOptions) (*imap.StatusData, error) {
	msg := &ReentryCalldataServer{
		Status: &StatusRequest{Username: s.username, Mailbox: mailbox, Options: options},
	}
	res, err := s.ctx.HandleServerReentry(msg)
	if err != nil {
		return nil, err
	}
	resp := &imap.StatusData{}
	err = json.Unmarshal(res, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Adds a new message to a mailbox (e.g., saving a sent message). r is the raw MIME.
func (s *Session) Append(mailbox string, r imap.LiteralReader, options *imap.AppendOptions) (*imap.AppendData, error) {
	emailRaw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	msg := &ReentryCalldataServer{
		Append: &AppendRequest{
			Username: s.username,
			Mailbox:  mailbox,
			Options:  options,
			EmailRaw: emailRaw,
		},
	}
	res, err := s.ctx.HandleServerReentry(msg)
	if err != nil {
		return nil, err
	}
	resp := &imap.AppendData{}
	err = json.Unmarshal(res, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Used for push-like update checks (like for IDLE). If you support updates or notifications, hook in here.
func (s *Session) Poll(w *imapserver.UpdateWriter, allowExpunge bool) error {
	fmt.Println("-imap.Session.Poll NOT IMPLEMENTED--", allowExpunge)
	return nil
}

// Long-lived command where Thunderbird waits for updates. You must write mailbox changes using w.Write(...) if anything changes while idle.
func (s *Session) Idle(w *imapserver.UpdateWriter, stop <-chan struct{}) error {
	fmt.Println("-imap.Session.Idle NOT IMPLEMENTED--")
	return nil
}

// Deselect current mailbox; ends "selected" state.
func (s *Session) Unselect() error {
	s.mailbox = ""
	s.selectOptions = nil
	return nil
}

// Permanently deletes messages marked as \Deleted. Often called automatically after Store sets \Deleted.
func (s *Session) Expunge(w *imapserver.ExpungeWriter, uids *imap.UIDSet) error {
	msg := &ReentryCalldataServer{
		Expunge: &ExpungeRequest{Username: s.username, Uids: uids},
	}
	res, err := s.ctx.HandleServerReentry(msg)
	if err != nil {
		return err
	}
	// Expecting back a list of expunged messages with their sequence numbers
	var resp struct {
		ExpungedSeqNums []uint32 `json:"expunged_seq_nums"`
	}
	if err := json.Unmarshal(res, &resp); err != nil {
		return err
	}

	// Notify client of each expunged message
	for _, seqNum := range resp.ExpungedSeqNums {
		err = w.WriteExpunge(seqNum)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *Session) Search(kind imapserver.NumKind, criteria *imap.SearchCriteria, options *imap.SearchOptions) (*imap.SearchData, error) {
	msg := &ReentryCalldataServer{
		Search: &SearchRequest{
			Username: s.username,
			Mailbox:  s.mailbox,
			Kind:     kind,
			Criteria: criteria,
			Options:  options,
		},
	}
	res, err := s.ctx.HandleServerReentry(msg)
	if err != nil {
		return nil, err
	}
	resp := &imap.SearchData{}
	err = json.Unmarshal(res, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (s *Session) Fetch(w *imapserver.FetchWriter, numSet imap.NumSet, options *imap.FetchOptions) error {
	bz, _ := json.Marshal(options)
	options.Envelope = true
	options.InternalDate = true
	options.BodySection = []*imap.FetchItemBodySection{
		&imap.FetchItemBodySection{},
	}

	uidSet, seqSet := fromNumSet(numSet)
	msg := &ReentryCalldataServer{
		Fetch: &FetchRequest{
			Username: s.username,
			Mailbox:  s.mailbox,
			UidSet:   uidSet,
			SeqSet:   seqSet,
			Options:  options,
		},
	}
	res, err := s.ctx.HandleServerReentry(msg)
	if err != nil {
		return err
	}
	var messages []struct {
		SeqNum       uint32        `json:"seq_num"`
		UID          uint32        `json:"uid"`
		Flags        []imap.Flag   `json:"flags"`
		InternalDate string        `json:"internal_date"`
		RFC822Size   int64         `json:"rfc822size"`
		Body         string        `json:"body"` // assume it's a raw string or base64
		Envelope     imap.Envelope `json:"envelope"`
		RawEmail     []byte        `json:"raw_email"`
	}
	if err := json.Unmarshal(res, &messages); err != nil {
		return err
	}

	// Write each fetched message
	for _, m := range messages {
		msg := w.CreateMessage(m.SeqNum)
		if options.UID {
			msg.WriteUID(imap.UID(m.UID))
		}
		if options.Flags {
			msg.WriteFlags(m.Flags)
		}
		if options.InternalDate {
			t, err := time.Parse(time.RFC1123Z, m.InternalDate)
			if err != nil {
				fmt.Println("vmimap.Fetch.InternalDate", err)
			}
			msg.WriteInternalDate(t)
		}
		if options.RFC822Size {
			msg.WriteRFC822Size(m.RFC822Size)
		}
		if options.Envelope {
			msg.WriteEnvelope(&m.Envelope)
		}
		// Write requested body sections
		for _, section := range options.BodySection {
			if len(section.Part) == 0 && len(section.HeaderFields) == 0 && len(section.HeaderFieldsNot) == 0 {
				writer := msg.WriteBodySection(section, int64(len(m.RawEmail)))
				_, err := writer.Write(m.RawEmail)
				writer.Close()
				if err != nil {
					return err
				}
				continue
			}

			writer := msg.WriteBodySection(section, int64(len(m.Body)))
			if _, err := io.Copy(writer, strings.NewReader(m.Body)); err != nil {
				writer.Close()
				return err
			}
			err = writer.Close()
			if err != nil {
				return err
			}
		}
		err = msg.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// Sets or unsets flags (e.g., \Seen, \Deleted) on messages.
func (s *Session) Store(w *imapserver.FetchWriter, numSet imap.NumSet, flags *imap.StoreFlags, options *imap.StoreOptions) error {
	uidSet, seqSet := fromNumSet(numSet)
	msg := &ReentryCalldataServer{
		Store: &StoreRequest{
			Username: s.username,
			Mailbox:  s.mailbox,
			UidSet:   uidSet,
			SeqSet:   seqSet,
			Flags:    flags,
			Options:  options,
		},
	}
	res, err := s.ctx.HandleServerReentry(msg)
	if err != nil {
		return err
	}

	// Expect updated flags for each message
	var updates []struct {
		SeqNum uint32      `json:"seq_num"`
		UID    uint32      `json:"uid"`
		Flags  []imap.Flag `json:"flags"`
	}
	if err := json.Unmarshal(res, &updates); err != nil {
		return fmt.Errorf("imap.Store: failed to decode response: %w", err)
	}

	// Report changes to client
	for _, u := range updates {
		msg := w.CreateMessage(u.SeqNum)
		msg.WriteUID(imap.UID(u.UID))
		msg.WriteFlags(u.Flags)
	}

	return nil
}

// Copies messages to another mailbox. Used for archiving or moving to Sent.
func (s *Session) Copy(numSet imap.NumSet, dest string) (*imap.CopyData, error) {
	uidSet, seqSet := fromNumSet(numSet)
	msg := &ReentryCalldataServer{
		Copy: &CopyRequest{
			Username: s.username,
			Mailbox:  s.mailbox,
			UidSet:   uidSet,
			SeqSet:   seqSet,
			Dest:     dest,
		},
	}
	res, err := s.ctx.HandleServerReentry(msg)
	if err != nil {
		return nil, err
	}

	// Optional: if backend returns mapping of source to new UIDs
	var data imap.CopyData
	if err := json.Unmarshal(res, &data); err != nil {
		return nil, fmt.Errorf("imap.Copy: failed to decode response: %w", err)
	}

	return &data, nil
}

type backend struct {
	ctx *Context
}

func (b *backend) NewSession(conn *imapserver.Conn) (imapserver.Session, *imapserver.GreetingData, error) {
	// PreAuth is true if session already exists and user is authenticated
	data := &imapserver.GreetingData{PreAuth: false}
	return &Session{ctx: b.ctx}, data, nil
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
		return ListenAndServe(s, cfg.Addr, cfg.Network, cfg.StartTLS, tlsCfg)
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
				ctx.Ctx.Logger().Info("closing imap server", "message", err.Error())
				close(srvDone)
				return
			}
			ctx.Ctx.Logger().Error("failed to serve imap server", "error", err.Error())
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

func ListenAndServe(s *imapserver.Server, addr string, network string, startTls bool, tlsCfg *tls.Config) error {
	if addr == "" {
		addr = ":143" // starttls
		if tlsCfg != nil && !startTls {
			addr = ":993" // tls
		}
	}
	var ln net.Listener
	var err error
	if tlsCfg == nil || startTls {
		fmt.Println("imap.ListenAndServe net", addr)
		ln, err = net.Listen(network, addr)
	} else {
		fmt.Println("imap.ListenAndServe tls", addr)
		ln, err = tls.Listen(network, addr, tlsCfg)
	}
	if err != nil {
		return err
	}
	return s.Serve(ln)
}

func (ctx *Context) HandleServerReentry(msg *ReentryCalldataServer) ([]byte, error) {
	msgbz, err := json.Marshal(msg)
	if err != nil {
		ctx.Ctx.Logger().Error("cannot marshal Expunge", "error", err.Error())
		return nil, err
	}

	contractAddress := ctx.Env.Contract.Address

	msgtosend := &networktypes.MsgReentry{
		Sender:     contractAddress.String(),
		Contract:   contractAddress.String(),
		EntryPoint: ENTRY_POINT_IMAP_SERVER,
		Msg:        msgbz,
	}
	_, resp, err := ctx.Context.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		ctx.Ctx.Logger().Error(err.Error())
		return nil, err
	}

	var rres networktypes.MsgReentryResponse
	err = rres.Unmarshal(resp)
	if err != nil {
		return nil, err
	}

	return rres.Data, nil
}

func fromNumSet(numSet imap.NumSet) (imap.UIDSet, imap.SeqSet) {
	uidSet, ok := numSet.(imap.UIDSet)
	if !ok {
		uidSet = imap.UIDSet{}
	}
	seqSet, ok := numSet.(imap.SeqSet)
	if !ok {
		seqSet = imap.SeqSet{}
	}
	return uidSet, seqSet
}
