package vmsmtp

import (
	"context"
	"fmt"
	"sync"

	"github.com/emersion/go-imap/v2"
	gosmtp "github.com/emersion/go-smtp"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

const (
	// ModuleName defines the module name
	ModuleName = "vmsmtp"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const HOST_WASMX_ENV_SMTP_i32_VER1 = "wasmx_smtp_i32_1"
const HOST_WASMX_ENV_SMTP_i64_VER1 = "wasmx_smtp_i64_1"

const HOST_WASMX_ENV_SMTP_EXPORT = "wasmx_smtp_"

const HOST_WASMX_ENV_SMTP = "smtp"

type ContextKey string

const SmtpContextKey ContextKey = "smtp-context"

type Context struct {
	*vmtypes.Context
}

type SmtpOpenConnection struct {
	mtx                   sync.Mutex
	GoContextParent       context.Context
	Username              string
	SmtpServerUrlSTARTTLS string `json:"smtp_server_url_starttls"`
	SmtpServerUrlTLS      string `json:"smtp_server_url_tls"`
	Client                *gosmtp.Client
	Closed                chan struct{}
	GetClient             func() (*gosmtp.Client, error)
}

type SmtpContext struct {
	mtx           sync.Mutex
	DbConnections map[string]*SmtpOpenConnection
}

func (p *SmtpContext) GetConnection(id string) (*SmtpOpenConnection, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	db, found := p.DbConnections[id]
	return db, found
}

func (p *SmtpContext) SetConnection(id string, conn *SmtpOpenConnection) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.DbConnections[id]
	if found {
		return fmt.Errorf("cannot overwrite SMTP connection: %s", id)
	}
	p.DbConnections[id] = conn
	return nil
}

func (p *SmtpContext) DeleteConnection(id string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	delete(p.DbConnections, id)
}

type SmtpConnectionSimpleRequest struct {
	Id                    string `json:"id"`
	SmtpServerUrlSTARTTLS string `json:"smtp_server_url_starttls"`
	SmtpServerUrlTLS      string `json:"smtp_server_url_tls"`
	Username              string `json:"username"`
	Password              string `json:"password"`
}

type SmtpConnectionOauth2Request struct {
	Id                    string `json:"id"`
	SmtpServerUrlSTARTTLS string `json:"smtp_server_url_starttls"`
	SmtpServerUrlTLS      string `json:"smtp_server_url_tls"`
	Username              string `json:"username"`
	AccessToken           string `json:"access_token"`
}

type SmtpConnectionResponse struct {
	Error string `json:"error"`
}

type SmtpCloseRequest struct {
	Id string `json:"id"`
}

type SmtpCloseResponse struct {
	Error string `json:"error"`
}

type SmtpQuitRequest struct {
	Id string `json:"id"`
}

type SmtpQuitResponse struct {
	Error string `json:"error"`
}

type SmtpSendMailRequest struct {
	Id    string   `json:"id"`
	From  string   `json:"from"`
	To    []string `json:"to"`
	Email []byte   `json:"email"`
}

type SmtpSendMailResponse struct {
	Error string `json:"error"`
}

type SmtpVerifyRequest struct {
	Id      string `json:"id"`
	Address string `json:"address"`
}

type SmtpVerifyResponse struct {
	Error string `json:"error"`
}

type SmtpNoopRequest struct {
	Id string `json:"id"`
}

type SmtpNoopResponse struct {
	Error string `json:"error"`
}

type SmtpExtensionRequest struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type SmtpExtensionResponse struct {
	Error  string `json:"error"`
	Found  bool   `json:"found"`
	Params string `json:"params"`
}

type SmtpMaxMessageSizeRequest struct {
	Id string `json:"id"`
}

type SmtpMaxMessageSizeResponse struct {
	Error string `json:"error"`
	Size  int64  `json:"size"`
	Ok    bool   `json:"ok"`
}

type SmtpSupportsAuthRequest struct {
	Id        string `json:"id"`
	Mechanism string `json:"mechanism"`
}

type SmtpSupportsAuthResponse struct {
	Error string `json:"error"`
	Found bool   `json:"found"`
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type Email struct {
	Envelope    *imap.Envelope      `json:"envelope"` // Header fields (From, To, Subject, etc.)
	Header      map[string][]string `json:"header"`   // Parsed headers (future use)
	Body        string              `json:"body"`     // Body content (if separated)
	Attachments []Attachment        `json:"attachments"`
}

type SmtpBuildMailRequest struct {
	Email Email `json:"email"`
}

type SmtpBuildMailResponse struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}
