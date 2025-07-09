package vmsmtp

import (
	"context"
	"fmt"
	"sync"

	"github.com/emersion/go-imap/v2"
	gosmtp "github.com/emersion/go-smtp"

	mcodec "github.com/loredanacirstea/wasmx/codec"
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

const ENTRY_POINT_SMTP = "smtp_update"

type ContextKey string

const SmtpContextKey ContextKey = "smtp-context"

type Context struct {
	*vmtypes.Context
}

type SmtpOpenConnection struct {
	mtx             sync.Mutex
	GoContextParent context.Context
	Info            SmtpConnectionRequest
	Client          *gosmtp.Client
	Closed          chan struct{}
	GetClient       func() (*gosmtp.Client, error)
}

type SmtpServerConnection struct {
	GoContextParent context.Context
	Server          *gosmtp.Server
	ContractAddress mcodec.AccAddressPrefixed
}

type SmtpContext struct {
	mtx               sync.Mutex
	DbConnections     map[string]*SmtpOpenConnection
	ServerConnections map[string]*SmtpServerConnection
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

func (p *SmtpContext) GetServerConnection(id string) (*SmtpServerConnection, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	db, found := p.ServerConnections[id]
	return db, found
}

func (p *SmtpContext) SetServerConnection(id string, conn *SmtpServerConnection) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.ServerConnections[id]
	if found {
		return fmt.Errorf("cannot overwrite SMTP connection: %s", id)
	}
	p.ServerConnections[id] = conn
	return nil
}

func (p *SmtpContext) DeleteServerConnection(id string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	delete(p.ServerConnections, id)
}

type ConnectionAuthType string

const (
	ConnectionAuthTypePassword ConnectionAuthType = "password"
	ConnectionAuthTypeOAuth2   ConnectionAuthType = "oauth2"
)

type ConnectionAuth struct {
	AuthType ConnectionAuthType `json:"auth_type"` // "password", "oauth2"
	Username string             `json:"username"`
	Password string             `json:"password"`
	Identity string             `json:"identity"`
}

type SmtpConnectionRequest struct {
	Id          string          `json:"id"`
	ServerUrl   string          `json:"server_url"`
	StartTLS    bool            `json:"start_tls"`
	NetworkType string          `json:"network_type"` // "tcp", "tcp4", "udp"
	Auth        *ConnectionAuth `json:"auth"`
	TlsConfig   *TlsConfig      `json:"tls_config"`
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

type SmtpHelloRequest struct {
	Id        string `json:"id"`
	LocalName string `json:"local_name"`
}

type SmtpHelloResponse struct {
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

type ServerStartRequest struct {
	ConnectionId string       `json:"connection_id"`
	ServerConfig ServerConfig `json:"server_config"`
}

type ServerStartResponse struct {
	Error string `json:"error"`
}

type ServerCloseRequest struct {
	ConnectionId string `json:"connection_id"`
}

type ServerCloseResponse struct {
	Error string `json:"error"`
}

type ServerShutdownRequest struct {
	ConnectionId string `json:"connection_id"`
}

type ServerShutdownResponse struct {
	Error string `json:"error"`
}
