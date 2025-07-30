package smtp

import (
	"time"

	vmimap "github.com/loredanacirstea/wasmx-env-imap"
)

type TlsConfig struct {
	ServerName  string `json:"server_name"`
	TLSCertFile string `json:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file"`
}

type ConnectionAuth struct {
	AuthType string `json:"auth_type"` // "password", "oauth2"
	Username string `json:"username"`
	Password string `json:"password"`
	Identity string `json:"identity"`
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

type Address struct {
	Name    string
	Mailbox string
	Host    string
}

type Envelope struct {
	Date      time.Time
	Subject   string
	From      []Address
	Sender    []Address
	ReplyTo   []Address
	To        []Address
	Cc        []Address
	Bcc       []Address
	InReplyTo []string
	MessageID string
}

type Email struct {
	Envelope    *Envelope       `json:"envelope"` // Header fields (From, To, Subject, etc.)
	Headers     []vmimap.Header `json:"headers"`  // Parsed headers (future use)
	Body        string          `json:"body"`     // Body content (if separated)
	Attachments []Attachment    `json:"attachments"`
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

type ServerConfig struct {
	// The type of network, "tcp" or "unix".
	Network string `json:"network"`
	// TCP or Unix address to listen on.
	Addr string `json:"address"` // ":25"

	StartTLS   bool       `json:"start_tls"`
	EnableAuth bool       `json:"enable_auth"`
	TlsConfig  *TlsConfig `json:"tls_config"`

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
	EnableDELIVERBY bool `json:"enable_deliver_by"`

	// The minimum time (seconds precision) a client may specify in BY=.
	// Only used if DELIVERBY is enabled.
	MinimumDeliverByTime time.Duration `json:"minimum_deliver_by_time"`
}
