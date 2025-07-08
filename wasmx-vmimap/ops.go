package vmimap

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	imap "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

func Connect(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ImapConnectionRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetImapContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &ImapConnectionResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)

	conn, found := vctx.GetConnection(connId)
	if found {
		if conn.Info.ImapServerUrl == req.ImapServerUrl {
			err := conn.Client.Noop().Wait()
			if err == nil {
				return prepareResponse(rnh, response)
			} else {
				_ = closeConnection(vctx, conn, connId)
			}
		} else {
			response.Error = "connection id already in use"
			return prepareResponse(rnh, response)
		}
	}

	getClient := func(opts *imapclient.Options) (*imapclient.Client, error) {
		c, err := connectToIMAP(req.ImapServerUrl, req.Auth, opts)
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	return connectCommon(ctx, rnh, vctx, getClient, response, connId, req)
}

func connectCommon(
	ctx *Context,
	rnh memc.RuntimeHandler,
	vctx *ImapContext,
	getClient func(opts *imapclient.Options) (*imapclient.Client, error),
	response *ImapConnectionResponse,
	connId string,
	info ImapConnectionRequest,
) ([]interface{}, error) {
	client, err := getClient(nil)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	closedChannel := make(chan struct{})

	// TODO this should be done in 1 cleanup hook per vm extension
	ctx.GoRoutineGroup.Go(func() error {
		select {
		case <-ctx.GoContextParent.Done():
			ctx.Ctx.Logger().Info(fmt.Sprintf("parent context was closed, closing database connection: %s", connId))
			err := client.Close()
			if err != nil {
				ctx.Ctx.Logger().Error(fmt.Sprintf(`imap close error for connection id "%s": %v`, connId, err))
			}
			close(closedChannel)
			return nil
		case <-closedChannel:
			// when close signal is received from Close() API
			// database is already closed, so we exit this goroutine
			ctx.Ctx.Logger().Info(fmt.Sprintf("database connection closed: %s", connId))
			return nil
		}
	})

	conn := &ImapOpenConnection{
		GoContextParent: ctx.GoContextParent,
		Info:            info,
		Client:          client,
		Closed:          closedChannel,
		GetClient:       getClient,
		listeners:       make(map[string]*IMAPListener, 0),
	}

	err = vctx.SetConnection(connId, conn)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func Close(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ImapCloseRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetImapContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &ImapCloseResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "IMAP connection not found"
		return prepareResponse(rnh, response)
	}
	err = closeConnection(vctx, conn, connId)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func closeConnection(vctx *ImapContext, conn *ImapOpenConnection, connId string) error {
	err := conn.Client.Close()
	close(conn.Closed) // signal closing the database
	vctx.DeleteConnection(connId)
	// TODO listeners?
	return err
}

func Listen(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ImapListenRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetImapContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &ImapListenResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)

	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "IMAP connection not found"
		return prepareResponse(rnh, response)
	}

	dataHandler := callbackMailboxChange(ctx, req.Folder, connId, conn.Info.Auth.Username)
	err = conn.StartListener(ctx.GoContextParent, req.Folder, dataHandler)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func Count(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ImapCountRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetImapContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &ImapCountResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)

	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "IMAP connection not found"
		return prepareResponse(rnh, response)
	}

	folder, err := conn.Client.Select(req.Folder, nil).Wait()
	if err != nil {
		response.Error = "failed to select email folder"
		return prepareResponse(rnh, response)
	}
	response.Count = int64(folder.NumMessages)
	return prepareResponse(rnh, response)
}

func UIDSearch(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ImapUIDSearchRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetImapContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &ImapUIDSearchResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)

	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "IMAP connection not found"
		return prepareResponse(rnh, response)
	}

	folder, err := conn.Client.Select(req.Folder, nil).Wait()
	if err != nil {
		response.Error = "failed to select email folder"
		return prepareResponse(rnh, response)
	}
	if req.FetchFilter == nil {
		response.UIDs = make(imap.UIDSet, 0)
		response.Count = 0
		return prepareResponse(rnh, response)
	}

	numset, count, err := fetchEmailIds(conn.Client, folder, conn.Info.Auth.Username, *req.FetchFilter)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	response.UIDs = numset.(imap.UIDSet)
	response.Count = int64(count)
	return prepareResponse(rnh, response)
}

func Fetch(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ImapFetchRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetImapContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &ImapFetchResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)

	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "IMAP connection not found"
		return prepareResponse(rnh, response)
	}

	// bodySection := &imap.FetchItemBodySection{Specifier: imap.PartSpecifierHeader}
	bodySection := req.BodySection
	if bodySection == nil {
		bodySection = &imap.FetchItemBodySection{}
	}

	options := req.Options
	if options == nil {
		options = &imap.FetchOptions{
			Flags:         true,
			Envelope:      true,
			InternalDate:  true,
			RFC822Size:    true,
			UID:           true,
			BodyStructure: &imap.FetchItemBodyStructure{Extended: true},
			BodySection:   []*imap.FetchItemBodySection{bodySection},
		}
	}

	folder, err := conn.Client.Select(req.Folder, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to select email folder: %v", err.Error())
	}
	response.Count = int64(folder.NumMessages)
	emails := []Email{}
	if len(req.SeqSet) > 0 {
		e, err := imapFetch(conn.Client, ctx.Ctx.Logger(), req.SeqSet, options, bodySection)
		if err != nil {
			response.Error = err.Error()
			return prepareResponse(rnh, response)
		}
		emails = e
	}
	if len(req.UidSet) > 0 {
		e, err := imapFetch(conn.Client, ctx.Ctx.Logger(), req.UidSet, options, bodySection)
		if err != nil {
			response.Error = err.Error()
			return prepareResponse(rnh, response)
		}
		emails = append(emails, e...)
	}
	if req.FetchFilter != nil {
		numset, count, err := fetchEmailIds(conn.Client, folder, conn.Info.Auth.Username, *req.FetchFilter)
		if err != nil {
			response.Error = err.Error()
			return prepareResponse(rnh, response)
		}
		if count > 0 {
			e, err := imapFetch(conn.Client, ctx.Ctx.Logger(), numset, options, bodySection)
			if err != nil {
				response.Error = err.Error()
				return prepareResponse(rnh, response)
			}
			emails = append(emails, e...)
		}
	}
	if req.Reverse {
		slices.Reverse(emails)
	}
	response.Data = emails
	return prepareResponse(rnh, response)
}

func ListMailboxes(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ListMailboxesRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetImapContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &ListMailboxesResponse{Error: "", Mailboxes: []string{}}
	connId := buildConnectionId(req.Id, ctx)

	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "IMAP connection not found"
		return prepareResponse(rnh, response)
	}

	// List all mailboxes (use the empty string for the reference and "*" for the mailbox pattern)
	cmd := conn.Client.List("", "*", nil)
	defer cmd.Close()
	for {
		d := cmd.Next()
		if d == nil {
			break
		}
		response.Mailboxes = append(response.Mailboxes, d.Mailbox)
	}
	return prepareResponse(rnh, response)
}

func CreateFolder(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req ImapCreateFolderRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetImapContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &ImapCreateFolderResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)

	conn, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "IMAP connection not found"
		return prepareResponse(rnh, response)
	}

	err = conn.Client.Create(req.Path, req.Options).Wait()
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

// TODO call contract
func callbackMailboxChange(ctx *Context, folder string, sessionId string, sessionUsername string) *imapclient.UnilateralDataHandler {
	return &imapclient.UnilateralDataHandler{
		Expunge: func(seqNum uint32) {
			ctx.Ctx.Logger().Info("message seqnum %v has been expunged\n", seqNum)
			ctx.handleExpunge(sessionUsername, folder, seqNum)
		},
		Mailbox: func(data *imapclient.UnilateralDataMailbox) {
			if data.NumMessages != nil {
				seqNum := *data.NumMessages
				ctx.Ctx.Logger().Info("a new message has been received seqnum: %d \n", seqNum)
			}
		},
		Fetch: func(msg *imapclient.FetchMessageData) {
			if msg == nil {
				return
			}
			ctx.Ctx.Logger().Info("new message seqnum: %d \n", msg.SeqNum)

			// TODO use Next() for big attachments and maybe restrict size of data forwarded to the contract if needed (TBD)
			data, err := msg.Collect()
			if err != nil {
				ctx.Ctx.Logger().Info("new message collect error: seqnum %d: %v\n", msg.SeqNum, err)
				return
			}
			if data == nil {
				return
			}
			ctx.Ctx.Logger().Info("new message seqnum: %d, uid: %d \n", msg.SeqNum, data.UID)
			ctx.HandleIncomingEmail(sessionUsername, folder, uint32(data.UID), msg.SeqNum)
		},

		// requires ENABLE METADATA or ENABLE SERVER-METADATA
		Metadata: func(mailbox string, entries []string) {
			ctx.Ctx.Logger().Info("new message metadata for mailbox: %s, entries: %s \n", mailbox, strings.Join(entries, ","))
			ctx.handleMetadata(sessionUsername, mailbox, entries)
		},
	}
}

func prepareResponse(rnh memc.RuntimeHandler, response interface{}) ([]interface{}, error) {
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

// per session
func buildConnectionId(id string, ctx *Context) string {
	return fmt.Sprintf("%s_%s", ctx.Env.Contract.Address.String(), id)
}
