package imap

import "time"

type Email struct {
	UID          UID                 `json:"uid"`          // Unique identifier
	Flags        []Flag              `json:"flags"`        // Flags like \Seen, \Answered
	InternalDate time.Time           `json:"internalDate"` // Date received by server
	RFC822Size   int64               `json:"rfc822Size"`   // Size in bytes
	Envelope     *Envelope           `json:"envelope"`     // Header fields (From, To, Subject, etc.)
	Header       map[string][]string `json:"header"`       // Parsed headers (future use)
	Body         string              `json:"body"`         // Body content (if separated)
	Attachments  []Attachment        `json:"attachments"`
	Raw          string              `json:"raw"` // Entire email as a string
	Bh           string              `json:"bh"`  // extracted body hash

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
