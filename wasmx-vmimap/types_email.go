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
	Filename    string
	ContentType string
	Data        []byte
}

type Email struct {
	UID          imap.UID            `json:"uid"`          // Unique identifier
	Flags        []imap.Flag         `json:"flags"`        // Flags like \Seen, \Answered
	InternalDate time.Time           `json:"internalDate"` // Date received by server
	RFC822Size   int64               `json:"rfc822Size"`   // Size in bytes
	Envelope     *imap.Envelope      `json:"envelope"`     // Header fields (From, To, Subject, etc.)
	Header       map[string][]string `json:"header"`       // Parsed headers (future use)
	Body         string              `json:"body"`         // Body content (if separated)
	Attachments  []Attachment        `json:"attachments"`
	Raw          string              `json:"raw"` // Entire email as a string
	Bh           string              `json:"bh"`  // extracted body hash

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
