package main

import (
	"github.com/emersion/go-msgauth/dkim"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"

	dkim2 "github.com/redsift/dkim"
)

type Calldata struct {
	ConnectWithPassword *vmimap.ImapConnectionSimpleRequest `json:"ConnectWithPassword,omitempty"`
	ConnectOAuth2       *vmimap.ImapConnectionOauth2Request `json:"ConnectOAuth2,omitempty"`
	Close               *vmimap.ImapCloseRequest            `json:"Close,omitempty"`
	SendEmail           *vmimap.ImapCreateFolderRequest     `json:"SendEmail,omitempty"`
	BuildAndSend        *BuildAndSendMailRequest            `json:"BuildAndSend,omitempty"`
	SignDKIM            *SignDKIMRequest                    `json:"SignDKIM,omitempty"`
	VerifyDKIM          *VerifyDKIMRequest                  `json:"VerifyDKIM,omitempty"`
	VerifyARC           *VerifyDKIMRequest                  `json:"VerifyARC,omitempty"`
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

type SignDKIMRequest struct {
}

type SignDKIMResponse struct {
	Error     string `json:"error"`
	Signature []byte `json:"signature"`
}

type VerifyDKIMRequest struct {
	EmailRaw string `json:"email_raw"`
}

type VerifyDKIMResponse struct {
	Error string `json:"error"`
	// Verifications []DKIMVerification `json:"verifications"`
	Verifications []*dkim.Verification `json:"verifications"`
	IsValid       bool                 `json:"is_valid"`
}

type VerifyARCResponse struct {
	Error    string           `json:"error"`
	Response *dkim2.ArcResult `json:"response"`
}
