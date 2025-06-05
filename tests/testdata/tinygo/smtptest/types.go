package main

import (
	vmsmtp "github.com/loredanacirstea/wasmx-env-smtp"
)

type Calldata struct {
	ConnectWithPassword *vmsmtp.SmtpConnectionSimpleRequest `json:"ConnectWithPassword,omitempty"`
	ConnectOAuth2       *vmsmtp.SmtpConnectionOauth2Request `json:"ConnectOAuth2,omitempty"`
	Close               *vmsmtp.SmtpCloseRequest            `json:"Close,omitempty"`
	Quit                *vmsmtp.SmtpQuitRequest             `json:"Quit,omitempty"`
	Extension           *vmsmtp.SmtpExtensionRequest        `json:"Extension,omitempty"`
	Noop                *vmsmtp.SmtpNoopRequest             `json:"Noop,omitempty"`
	Verify              *vmsmtp.SmtpVerifyRequest           `json:"Verify,omitempty"`
	SupportsAuth        *vmsmtp.SmtpSupportsAuthRequest     `json:"SupportsAuth,omitempty"`
	MaxMessageSize      *vmsmtp.SmtpMaxMessageSizeRequest   `json:"MaxMessageSize,omitempty"`
	SendMail            *vmsmtp.SmtpSendMailRequest         `json:"SendMail,omitempty"`
	BuildAndSend        *BuildAndSendMailRequest            `json:"BuildAndSend,omitempty"`
}

type BuildAndSendMailRequest struct {
	Id      string   `json:"id"`
	From    string   `json:"from"`
	To      []string `json:"to"`
	Cc      []string `json:"cc"`
	Bcc     []string `json:"bcc"`
	Subject string   `json:"subject"`
	Body    []byte   `json:"body"`
}
