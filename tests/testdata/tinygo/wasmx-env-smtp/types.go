package smtp

import "time"

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
	Envelope    *Envelope           `json:"envelope"` // Header fields (From, To, Subject, etc.)
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
