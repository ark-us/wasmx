package imap

import (
	"strings"
	"time"
)

// all email headers
const (
	HEADER_FROM        = "From"
	HEADER_TO          = "To"
	HEADER_CC          = "Cc"
	HEADER_BCC         = "Bcc"
	HEADER_REPLY_TO    = "Reply-To"
	HEADER_SENDER      = "Sender"
	HEADER_RETURN_PATH = "Return-Path"

	HEADER_DATE        = "Date"
	HEADER_MESSAGE_ID  = "Message-ID"
	HEADER_IN_REPLY_TO = "In-Reply-To"
	HEADER_REFERENCES  = "References"

	HEADER_SUBJECT                   = "Subject"
	HEADER_CONTENT_TYPE              = "Content-Type"
	HEADER_CONTENT_TRANSFER_ENCODING = "Content-Transfer-Encoding"
	HEADER_MIME_VERSION              = "MIME-Version"

	HEADER_DKIM_SIGNATURE             = "DKIM-Signature"
	HEADER_ARC_SEAL                   = "ARC-Seal"
	HEADER_ARC_MESSAGE_SIGNATURE      = "ARC-Message-Signature"
	HEADER_ARC_AUTHENTICATION_RESULTS = "ARC-Authentication-Results"
	HEADER_AUTHENTICATION_RESULTS     = "Authentication-Results"
	HEADER_RECEIVED                   = "Received"
	HEADER_RECEIVED_SPF               = "Received-SPF"
	HEADER_DELIVERED_TO               = "Delivered-To"

	// additional
	HEADER_USER_AGENT              = "User-Agent"
	HEADER_X_MAILER                = "X-Mailer"
	HEADER_X_ORIGINATING_IP        = "X-Originating-IP"
	HEADER_X_GOOGLE_DKIM_SIGNATURE = "X-Google-DKIM-Signature"
	HEADER_X_GM_MESSAGE_STATE      = "X-Gm-Message-State"
	HEADER_X_FORWARDED_FOR         = "X-Forwarded-For"
	HEADER_X_FORWARDED_TO          = "X-Forwarded-To"
	HEADER_X_FORWARDED_ENCRYPTED   = "X-Forwarded-Encrypted"
)

var (
	HEADER_LOW_FROM        = strings.ToLower(HEADER_FROM)
	HEADER_LOW_TO          = strings.ToLower(HEADER_TO)
	HEADER_LOW_CC          = strings.ToLower(HEADER_CC)
	HEADER_LOW_BCC         = strings.ToLower(HEADER_BCC)
	HEADER_LOW_REPLY_TO    = strings.ToLower(HEADER_REPLY_TO)
	HEADER_LOW_SENDER      = strings.ToLower(HEADER_SENDER)
	HEADER_LOW_RETURN_PATH = strings.ToLower(HEADER_RETURN_PATH)

	HEADER_LOW_DATE        = strings.ToLower(HEADER_DATE)
	HEADER_LOW_MESSAGE_ID  = strings.ToLower(HEADER_MESSAGE_ID)
	HEADER_LOW_IN_REPLY_TO = strings.ToLower(HEADER_IN_REPLY_TO)
	HEADER_LOW_REFERENCES  = strings.ToLower(HEADER_REFERENCES)

	HEADER_LOW_SUBJECT                   = strings.ToLower(HEADER_SUBJECT)
	HEADER_LOW_CONTENT_TYPE              = strings.ToLower(HEADER_CONTENT_TYPE)
	HEADER_LOW_CONTENT_TRANSFER_ENCODING = strings.ToLower(HEADER_CONTENT_TRANSFER_ENCODING)
	HEADER_LOW_MIME_VERSION              = strings.ToLower(HEADER_MIME_VERSION)

	HEADER_LOW_DKIM_SIGNATURE             = strings.ToLower(HEADER_DKIM_SIGNATURE)
	HEADER_LOW_ARC_SEAL                   = strings.ToLower(HEADER_ARC_SEAL)
	HEADER_LOW_ARC_MESSAGE_SIGNATURE      = strings.ToLower(HEADER_ARC_MESSAGE_SIGNATURE)
	HEADER_LOW_ARC_AUTHENTICATION_RESULTS = strings.ToLower(HEADER_ARC_AUTHENTICATION_RESULTS)
	HEADER_LOW_AUTHENTICATION_RESULTS     = strings.ToLower(HEADER_AUTHENTICATION_RESULTS)
	HEADER_LOW_RECEIVED                   = strings.ToLower(HEADER_RECEIVED)
	HEADER_LOW_RECEIVED_SPF               = strings.ToLower(HEADER_RECEIVED_SPF)
	HEADER_LOW_DELIVERED_TO               = strings.ToLower(HEADER_DELIVERED_TO)

	// additional
	HEADER_LOW_USER_AGENT              = strings.ToLower(HEADER_USER_AGENT)
	HEADER_LOW_X_MAILER                = strings.ToLower(HEADER_X_MAILER)
	HEADER_LOW_X_ORIGINATING_IP        = strings.ToLower(HEADER_X_ORIGINATING_IP)
	HEADER_LOW_X_GOOGLE_DKIM_SIGNATURE = strings.ToLower(HEADER_X_GOOGLE_DKIM_SIGNATURE)
	HEADER_LOW_X_GM_MESSAGE_STATE      = strings.ToLower(HEADER_X_GM_MESSAGE_STATE)
	HEADER_LOW_X_FORWARDED_FOR         = strings.ToLower(HEADER_X_FORWARDED_FOR)
	HEADER_LOW_X_FORWARDED_TO          = strings.ToLower(HEADER_X_FORWARDED_TO)
	HEADER_LOW_X_FORWARDED_ENCRYPTED   = strings.ToLower(HEADER_X_FORWARDED_ENCRYPTED)
)

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Raw   []byte `json:"raw"`
}

type Headers []Header

// Has checks if a header with the given key exists (case-insensitive)
func (h Headers) Has(key string) bool {
	key = strings.ToLower(key)
	for _, header := range h {
		if strings.ToLower(header.Key) == key {
			return true
		}
	}
	return false
}

// Get header with the given key exists (case-insensitive)
func (h Headers) Get(key string) *Header {
	key = strings.ToLower(key)
	for _, header := range h {
		if strings.ToLower(header.Key) == key {
			return &header
		}
	}
	return nil
}

// Set sets the value for a header key, replacing any existing entry with the same key (case-insensitive)
func (h *Headers) Set(header Header) {
	keyLower := strings.ToLower(header.Key)
	for i, header := range *h {
		if strings.ToLower(header.Key) == keyLower {
			(*h)[i] = header
			return
		}
	}
	*h = append(*h, header)
}

// Append adds a new header entry, even if one with the same key exists
func (h *Headers) Append(header Header) {
	*h = append(*h, header)
}

func (h *Headers) AppendTop(header Header) {
	*h = append([]Header{header}, *h...)
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
	UID          UID          `json:"uid"`          // Unique identifier
	Flags        []Flag       `json:"flags"`        // Flags like \Seen, \Answered
	InternalDate time.Time    `json:"internalDate"` // Date received by server
	RFC822Size   int64        `json:"rfc822Size"`   // Size in bytes
	Envelope     *Envelope    `json:"envelope"`     // Header fields (From, To, Subject, etc.)
	Headers      Headers      `json:"headers"`      // Parsed headers (future use)
	Body         EmailBody    `json:"body"`         // Body content (if separated)
	Attachments  []Attachment `json:"attachments"`
	Raw          string       `json:"raw"` // Entire email as a string
	Bh           string       `json:"bh"`  // extracted body hash

}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type ImapConnectionSimpleRequest struct {
	Id            string `json:"id"`
	ImapServerUrl string `json:"imap_server_url"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

type ImapConnectionOauth2Request struct {
	Id            string `json:"id"`
	ImapServerUrl string `json:"imap_server_url"`
	Username      string `json:"username"`
	AccessToken   string `json:"access_token"`
}

type ImapConnectionResponse struct {
	Error string `json:"error"`
}

type ImapCloseRequest struct {
	Id string `json:"id"`
}

type ImapCloseResponse struct {
	Error string `json:"error"`
}

type ImapListenRequest struct {
	Id     string `json:"id"`
	Folder string `json:"folder"`
}

type ImapListenResponse struct {
	Error string `json:"error"`
}

type FetchFilter struct {
	Limit   uint32          `json:"limit"`
	Start   uint32          `json:"start"`
	Search  *SearchCriteria `json:"search"`
	From    string          `json:"from"`
	To      string          `json:"to"`
	Subject string          `json:"subject"`
	Content string          `json:"content"`
}

type ImapCountRequest struct {
	Id     string `json:"id"`
	Folder string `json:"folder"`
}

type ImapCountResponse struct {
	Error string `json:"error"`
	Count int64  `json:"count"`
}

type ImapUIDSearchRequest struct {
	Id          string       `json:"id"`
	Folder      string       `json:"folder"`
	FetchFilter *FetchFilter `json:"fetch_filter"`
}

type ImapUIDSearchResponse struct {
	Error string `json:"error"`
	UIDs  UIDSet `json:"uids"`
	Count int64  `json:"count"`
}

type ImapFetchRequest struct {
	Id          string                `json:"id"`
	Folder      string                `json:"folder"`
	SeqSet      SeqSet                `json:"seq_set"`
	UidSet      UIDSet                `json:"uid_set"`
	FetchFilter *FetchFilter          `json:"fetch_filter"`
	Options     *FetchOptions         `json:"options"`
	BodySection *FetchItemBodySection `json:"bodySection"`
	Reverse     bool                  `json:"reverse"`
}

type ImapFetchResponse struct {
	Error string  `json:"error"`
	Data  []Email `json:"data"`
	Count int64   `json:"count"`
}

type ListMailboxesRequest struct {
	Id string `json:"id"`
}

type ListMailboxesResponse struct {
	Error     string   `json:"error"`
	Mailboxes []string `json:"mailboxes"`
}

type ImapCreateFolderRequest struct {
	Id      string         `json:"id"`
	Path    string         `json:"path"`
	Options *CreateOptions `json:"options"`
}

type ImapCreateFolderResponse struct {
	Error string `json:"error"`
}

type MsgIncomingEmail struct {
	Folder string `json:"folder"`
	UID    uint32 `json:"uid"`
	SeqNum uint32 `json:"seq_num"`
	Owner  string `json:"owner"`
}

type MsgExpunge struct {
	Owner  string `json:"owner"`
	Folder string `json:"folder"`
	SeqNum uint32 `json:"seq_num"`
}

type MsgMetadata struct {
	Owner   string   `json:"owner"`
	Folder  string   `json:"folder"`
	Entries []string `json:"entries"`
}

type ReentryCalldata struct {
	IncomingEmail *MsgIncomingEmail `json:"IncomingEmail"`
	Expunge       *MsgExpunge       `json:"Expunge"`
	Metadata      *MsgMetadata      `json:"IncominMetadatagEmail"`
}
