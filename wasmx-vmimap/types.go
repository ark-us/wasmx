package vmimap

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	imap "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"

	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

const (
	// ModuleName defines the module name
	ModuleName = "vmimap"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const HOST_WASMX_ENV_IMAP_i32_VER1 = "wasmx_imap_i32_1"
const HOST_WASMX_ENV_IMAP_i64_VER1 = "wasmx_imap_i64_1"

const HOST_WASMX_ENV_IMAP_EXPORT = "wasmx_imap_"

const HOST_WASMX_ENV_IMAP = "imap"

const ENTRY_POINT_IMAP = "imap_update"

type ContextKey string

const ImapContextKey ContextKey = "imap-context"

type Context struct {
	*vmtypes.Context
}

// IMAPListener manages an IMAP connection for real-time updates
type IMAPListener struct {
	Folder string
	Client *imapclient.Client // blocking listener IMAP client
}

type ImapOpenConnection struct {
	mtx             sync.Mutex
	GoContextParent context.Context
	Username        string
	ImapServerUrl   string             `json:"imap_server_url"`
	Client          *imapclient.Client // no mailbox/folder selected
	Closed          chan struct{}
	listeners       map[string]*IMAPListener // mailbox/folder => listener
	GetClient       func(opts *imapclient.Options) (*imapclient.Client, error)
}

func (p *ImapOpenConnection) GetListener(folder string) (*IMAPListener, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	db, found := p.listeners[folder]
	return db, found
}

func (p *ImapOpenConnection) SetListener(folder string, listener *IMAPListener) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.listeners[folder]
	if found {
		return fmt.Errorf("cannot overwrite IMAP listener: %s", folder)
	}
	p.listeners[folder] = listener
	return nil
}

func (p *ImapOpenConnection) DeleteListener(folder string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	delete(p.listeners, folder)
}

type ImapContext struct {
	mtx           sync.Mutex
	DbConnections map[string]*ImapOpenConnection
}

func (p *ImapContext) GetConnection(id string) (*ImapOpenConnection, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	db, found := p.DbConnections[id]
	return db, found
}

func (p *ImapContext) SetConnection(id string, conn *ImapOpenConnection) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.DbConnections[id]
	if found {
		return fmt.Errorf("cannot overwrite IMAP connection: %s", id)
	}
	p.DbConnections[id] = conn
	return nil
}

func (p *ImapContext) DeleteConnection(id string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	delete(p.DbConnections, id)
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
	Limit   uint32               `json:"limit"`
	Start   uint32               `json:"start"`
	Search  *imap.SearchCriteria `json:"search"`
	From    string               `json:"from"`
	To      string               `json:"to"`
	Subject string               `json:"subject"`
	Content string               `json:"content"`
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
	Error string      `json:"error"`
	UIDs  imap.UIDSet `json:"uids"`
	Count int64       `json:"count"`
}

type ImapFetchRequest struct {
	Id          string                     `json:"id"`
	Folder      string                     `json:"folder"`
	SeqSet      imap.SeqSet                `json:"seq_set"`
	UidSet      imap.UIDSet                `json:"uid_set"`
	FetchFilter *FetchFilter               `json:"fetch_filter"`
	Options     *imap.FetchOptions         `json:"options"`
	BodySection *imap.FetchItemBodySection `json:"bodySection"`
	Reverse     bool                       `json:"reverse"`
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
	Id      string              `json:"id"`
	Path    string              `json:"path"`
	Options *imap.CreateOptions `json:"options"`
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

func (c *Context) HandleIncomingEmail(owner string, folder string, uid uint32, seq uint32) {
	// TODO remove in production
	c.Ctx.Logger().Debug("email received", "owner", owner, "folder", folder, "uid", uid, "seq", seq)

	msg := &ReentryCalldata{
		IncomingEmail: &MsgIncomingEmail{
			Owner:  owner,
			Folder: folder,
			UID:    uid,
			SeqNum: seq,
		}}
	c.handleReentry(msg)
}

func (c *Context) handleExpunge(owner string, folder string, seq uint32) {
	// TODO remove in production
	c.Ctx.Logger().Debug("email expunge", "owner", owner, "folder", folder, "seq", seq)

	msg := &ReentryCalldata{
		Expunge: &MsgExpunge{
			Owner:  owner,
			Folder: folder,
			SeqNum: seq,
		}}
	c.handleReentry(msg)
}

func (c *Context) handleMetadata(owner string, folder string, entries []string) {
	// TODO remove in production
	c.Ctx.Logger().Debug("email metadata", "owner", owner, "folder", folder)

	msg := &ReentryCalldata{
		Metadata: &MsgMetadata{
			Owner:   owner,
			Folder:  folder,
			Entries: entries,
		}}
	c.handleReentry(msg)
}

func (c *Context) handleReentry(msg *ReentryCalldata) {
	msgbz, err := json.Marshal(msg)
	if err != nil {
		c.Ctx.Logger().Error("cannot marshal Expunge", "error", err.Error())
		return
	}

	contractAddress := c.Env.Contract.Address

	msgtosend := &networktypes.MsgReentryWithGoRoutine{
		Sender:     contractAddress.String(),
		Contract:   contractAddress.String(),
		EntryPoint: ENTRY_POINT_IMAP,
		Msg:        msgbz,
	}
	_, _, err = c.Context.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		c.Ctx.Logger().Error(err.Error())
	}
}
