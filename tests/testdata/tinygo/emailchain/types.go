package main

import (
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
)

type Calldata struct {
	ConnectWithPassword *vmimap.ImapConnectionSimpleRequest `json:"ConnectWithPassword,omitempty"`
	ConnectOAuth2       *vmimap.ImapConnectionOauth2Request `json:"ConnectOAuth2,omitempty"`
	Close               *vmimap.ImapCloseRequest            `json:"Close,omitempty"`
	Listen              *vmimap.ImapListenRequest           `json:"Listen,omitempty"`
	Count               *vmimap.ImapCountRequest            `json:"Count,omitempty"`
	UIDSearch           *vmimap.ImapUIDSearchRequest        `json:"UIDSearch,omitempty"`
	Fetch               *vmimap.ImapFetchRequest            `json:"Fetch,omitempty"`
	ListMailboxes       *vmimap.ListMailboxesRequest        `json:"ListMailboxes,omitempty"`
	CreateFolder        *vmimap.ImapCreateFolderRequest     `json:"CreateFolder,omitempty"`
}
