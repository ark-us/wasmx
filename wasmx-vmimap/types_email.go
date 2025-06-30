package vmimap

import (
	"time"

	imap "github.com/emersion/go-imap/v2"
)

type UserInfo struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	Sub           string `json:"sub"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}

type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Raw   []byte `json:"raw"`
}

type BodyPart struct {
	ContentType string `json:"content_type"`
	Body        []byte `json:"body"`
}

type EmailBody struct {
	Boundary string     `json:"boundary,omitempty"`
	Parts    []BodyPart `json:"parts"`
}

type Email struct {
	UID          imap.UID       `json:"uid"`          // Unique identifier
	Flags        []imap.Flag    `json:"flags"`        // Flags like \Seen, \Answered
	InternalDate time.Time      `json:"internalDate"` // Date received by server
	RFC822Size   int64          `json:"rfc822Size"`   // Size in bytes
	Envelope     *imap.Envelope `json:"envelope"`     // Header fields (From, To, Subject, etc.)
	// topmost headers are last
	Headers     []Header     `json:"headers"`
	Body        EmailBody    `json:"body"`
	Attachments []Attachment `json:"attachments"`
	Raw         string       `json:"raw"` // Entire email as a string
	Bh          string       `json:"bh"`  // extracted body hash

	// BodyStructure *imap.BodyStructure `json:"bodyStructure"` // MIME structure
	// BodySection       []imap.FetchBodySectionBuffer
	// BinarySection     []imap.FetchBinarySectionBuffer
	// BinarySectionSize []imap.FetchItemDataBinarySectionSize
	// ModSeq            uint64 // requires CONDSTORE
}

type EmailPartial struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}
