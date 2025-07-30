package vmsmtp

import "time"

type TlsConfig struct {
	ServerName  string `json:"server_name"`
	TLSCertFile string `json:"tls_cert_file"`
	TLSKeyFile  string `json:"tls_key_file"`
}

type ServerConfig struct {
	// The type of network, "tcp", "tcp4", or "unix".
	Network string `json:"network"`
	// TCP or Unix address to listen on.
	Addr       string     `json:"address"` // ":25"
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
	// EnableDELIVERBY bool `json:"enable_deliver_by"`

	// The minimum time (seconds precision) a client may specify in BY=.
	// Only used if DELIVERBY is enabled.
	// MinimumDeliverByTime time.Duration `json:"minimum_deliver_by_time"`
}

type ReentryCalldata struct {
	Login         *LoginRequest  `json:"Login"`
	Logout        *LogoutRequest `json:"Logout"`
	IncomingEmail *Session       `json:"IncomingEmail"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LogoutRequest struct {
	Username string `json:"username"`
}
