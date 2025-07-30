package vmimap

import (
	imap "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
)

type ReentryCalldataServer struct {
	Login   *LoginRequest   `json:"Login"`
	Logout  *LogoutRequest  `json:"Logout"`
	Create  *CreateRequest  `json:"Create"`
	Delete  *DeleteRequest  `json:"Delete"`
	Rename  *RenameRequest  `json:"Rename"`
	Select  *SelectRequest  `json:"Select"`
	List    *ListRequest    `json:"List"`
	Status  *StatusRequest  `json:"Status"`
	Append  *AppendRequest  `json:"Append"`
	Expunge *ExpungeRequest `json:"Expunge"`
	Search  *SearchRequest  `json:"Search"`
	Fetch   *FetchRequest   `json:"Fetch"`
	Store   *StoreRequest   `json:"Store"`
	Copy    *CopyRequest    `json:"Copy"`
	// Subscribe *SubscribeRequest `json:"Subscribe"`
	// Unsubscribe *UnsubscribeRequest `json:"Unsubscribe"`
	// Poll *PollRequest `json:"Poll"`
	// Idle *IdleRequest `json:"Idle"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LogoutRequest struct {
	Username string `json:"username"`
}

type CreateRequest struct {
	Username string              `json:"username"`
	Mailbox  string              `json:"mailbox"`
	Options  *imap.CreateOptions `json:"options"`
}

type DeleteRequest struct {
	Username string `json:"username"`
	Mailbox  string `json:"mailbox"`
}

type RenameRequest struct {
	Username string `json:"username"`
	Mailbox  string `json:"mailbox"`
	NewName  string `json:"new_name"`
}

type ListRequest struct {
	Username string            `json:"username"`
	Patterns []string          `json:"patterns"`
	Options  *imap.ListOptions `json:"options"`
}

type StatusRequest struct {
	Username string              `json:"username"`
	Mailbox  string              `json:"mailbox"`
	Options  *imap.StatusOptions `json:"options"`
}

type AppendRequest struct {
	Username string              `json:"username"`
	Mailbox  string              `json:"mailbox"`
	EmailRaw []byte              `json:"email_raw"`
	Options  *imap.AppendOptions `json:"options"`
}

type ExpungeRequest struct {
	Username string       `json:"username"`
	Uids     *imap.UIDSet `json:"uids"`
}

type SearchRequest struct {
	Username string               `json:"username"`
	Mailbox  string               `json:"mailbox"`
	Kind     imapserver.NumKind   `json:"kind"`
	Criteria *imap.SearchCriteria `json:"criteria"`
	Options  *imap.SearchOptions  `json:"options"`
}

type FetchRequest struct {
	Username string             `json:"username"`
	Mailbox  string             `json:"mailbox"`
	UidSet   imap.UIDSet        `json:"uid_set"`
	SeqSet   imap.SeqSet        `json:"seq_set"`
	Options  *imap.FetchOptions `json:"options"`
}

type StoreRequest struct {
	Username string             `json:"username"`
	Mailbox  string             `json:"mailbox"`
	UidSet   imap.UIDSet        `json:"uid_set"`
	SeqSet   imap.SeqSet        `json:"seq_set"`
	Flags    *imap.StoreFlags   `json:"flags"`
	Options  *imap.StoreOptions `json:"options"`
}

type CopyRequest struct {
	Username string      `json:"username"`
	Mailbox  string      `json:"mailbox"`
	UidSet   imap.UIDSet `json:"uid_set"`
	SeqSet   imap.SeqSet `json:"seq_set"`
	Dest     string      `json:"dest"`
}

type SelectRequest struct {
	Username string `json:"username"`
	Mailbox  string `json:"mailbox"`
}
