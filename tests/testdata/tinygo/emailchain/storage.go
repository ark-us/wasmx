package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	// "github.com/emersion/go-message/mail"

	"github.com/loredanacirstea/mailverif/dkim"
	"github.com/loredanacirstea/mailverif/utils"
	wasmx "github.com/loredanacirstea/wasmx-env"
	imap "github.com/loredanacirstea/wasmx-env-imap"
	sql "github.com/loredanacirstea/wasmx-env-sql"
)

// -- Table: owners (users/accounts)
const DefTableOwners = `CREATE TABLE owners (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    address TEXT UNIQUE NOT NULL
);`

const DefTableFolders = `CREATE TABLE folder_state (
	uid_validity INTEGER PRIMARY KEY AUTOINCREMENT,
    owner TEXT NOT NULL,
    folder TEXT NOT NULL,
    last_uid INTEGER NOT NULL DEFAULT 1,
    UNIQUE (owner, folder)
);`

// -- Table: emails (each email is one row, per folder/owner)
const DefTableEmails = `CREATE TABLE emails (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner TEXT NOT NULL,
    folder TEXT NOT NULL,
    uid INTEGER NOT NULL,
    seq_num INTEGER NOT NULL,
    message_id TEXT NOT NULL,
    subject TEXT NOT NULL DEFAULT '',
    internal_date INTEGER NOT NULL,
    flags TEXT NOT NULL DEFAULT '',
    raw_email BLOB NOT NULL,
    size INTEGER NOT NULL,
    headers TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
	bh VARCHAR NOT NULL,
	envelope TEXT,
    UNIQUE (owner, folder, uid),
	UNIQUE (owner, folder, seq_num)
);`

// -- Indexes for fast search and UID operations
const DefIndexes1 = `CREATE UNIQUE INDEX idx_emails_owner_folder_uid ON emails(owner, folder, uid);`
const DefIndexes2 = `CREATE INDEX idx_emails_owner_folder_seq_num ON emails(owner, folder, seq_num);`
const DefIndexes3 = `CREATE UNIQUE INDEX idx_emails_message_id ON emails(message_id);`
const DefIndexes4 = `CREATE INDEX idx_emails_subject ON emails(owner,subject);`
const DefIndexes5 = `CREATE INDEX idx_emails_internal_date ON emails(internal_date);`

const ExecUpdateUid = `INSERT INTO folder_state (owner, folder, last_uid)
VALUES (?, ?, 1)
ON CONFLICT(owner, folder)
DO UPDATE SET last_uid = last_uid + 1;`
const ExecGetUid = `SELECT last_uid FROM folder_state
WHERE owner = ? AND folder = ?`
const ExecGetFolder = `SELECT * FROM folder_state
WHERE owner = ? AND folder = ?`
const ExecGetFolders = `SELECT * FROM folder_state
WHERE owner = ?`
const ExecGetSeq = `SELECT COALESCE(MAX(seq_num), 0) + 1 AS next_seq_num
FROM emails
WHERE owner = ? AND folder = ?;`
const ExecInsertEmail = `INSERT INTO emails (
	owner, folder, uid, seq_num, message_id, subject, internal_date, bh, body, envelope, raw_email, size
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

// flags, size, headers,

func ConnectSql(connId string) error {
	resp := sql.Connect(&sql.SqlConnectionRequest{Id: connId, Driver: "sqlite3", Connection: "emailchain.db"})
	if resp.Error != "" {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func InitializeTables(connId string) {
	err := ConnectSql(connId)
	if err != nil {
		wasmx.Revert([]byte(err.Error()))
	}
	createTable := &sql.SqlExecuteBatchRequest{
		Id: connId,
		Commands: []sql.SqlExecuteCommand{
			{
				Query:  DefTableOwners,
				Params: sql.Params{},
			},
			{
				Query:  DefTableEmails,
				Params: sql.Params{},
			},
			{
				Query:  DefTableFolders,
				Params: sql.Params{},
			},
			{
				Query:  DefIndexes1,
				Params: sql.Params{},
			},
			{
				Query:  DefIndexes2,
				Params: sql.Params{},
			},
			{
				Query:  DefIndexes3,
				Params: sql.Params{},
			},
			{
				Query:  DefIndexes4,
				Params: sql.Params{},
			},
			{
				Query:  DefIndexes5,
				Params: sql.Params{},
			},
		},
	}
	respexec := sql.BatchAtomic(createTable)
	if respexec.Error != "" {
		wasmx.Revert([]byte(`could not create table: ` + respexec.Error))
	}
}

func checkFolderExists(connId, username, folder string) (bool, error) {
	q := `SELECT 1 FROM folder_state WHERE owner = ? AND folder = ? LIMIT 1`
	paramsbz, err := paramsMarshal([]sql.SqlQueryParam{
		{Type: "text", Value: username},
		{Type: "text", Value: folder},
	})
	qresp := sql.Query(&sql.SqlQueryRequest{
		Id:     connId,
		Query:  q,
		Params: paramsbz,
	})
	if qresp.Error != "" {
		return false, err
	}
	// [{"1":1}]
	resp := []map[string]int64{}
	err = json.Unmarshal(qresp.Data, &resp)
	if err != nil {
		return false, err
	}
	return len(resp) > 0, nil
}

func StoreEmail(owner string, mailfrom []string, emailRaw []byte, connId string, folder string) error {
	email, err := extractEmail(emailRaw)
	if err != nil {
		return err
	}
	if len(mailfrom) > 0 {
		email.Envelope.Sender = imap.AddressesFromString(mailfrom)
	} else {
		email.Envelope.Sender = email.Envelope.From
	}
	paramsbz, err := paramsMarshal([]sql.SqlQueryParam{
		{Type: "text", Value: owner},  // owner
		{Type: "text", Value: folder}, // folder
	})
	if err != nil {
		return err
	}
	resp := sql.Execute(&sql.SqlExecuteRequest{
		Id:     connId,
		Query:  ExecUpdateUid,
		Params: paramsbz,
	})
	if resp.Error != "" {
		return fmt.Errorf(resp.Error)
	}
	qresp := sql.Query(&sql.SqlQueryRequest{
		Id:     connId,
		Query:  ExecGetUid,
		Params: paramsbz,
	})
	uidResp := []UidResponse{}
	err = json.Unmarshal(qresp.Data, &uidResp)
	if err != nil {
		return err
	}
	uid := uidResp[0].LastUid

	qresp = sql.Query(&sql.SqlQueryRequest{
		Id:     connId,
		Query:  ExecGetSeq,
		Params: paramsbz,
	})
	seqResp := []SeqNumResponse{}
	err = json.Unmarshal(qresp.Data, &seqResp)
	if err != nil {
		return err
	}
	seq := seqResp[0].NextSeqNum

	email.Folder = folder
	email.UID = int64(uid)
	email.SeqNum = int64(seq)
	email.Owner = owner
	envbz, err := json.Marshal(&email.Envelope)
	if err != nil {
		return err
	}

	paramsbz2, err := paramsMarshal([]sql.SqlQueryParam{
		{Type: "text", Value: email.Owner},
		{Type: "text", Value: email.Folder},
		{Type: "integer", Value: email.UID},
		{Type: "integer", Value: email.SeqNum},
		{Type: "text", Value: email.MessageID},
		{Type: "text", Value: email.Subject},
		{Type: "integer", Value: email.InternalDate.Unix()},
		{Type: "text", Value: email.Bh},
		{Type: "text", Value: email.Body},
		{Type: "text", Value: string(envbz)},
		{Type: "blob", Value: email.RawEmail},
		{Type: "text", Value: len(email.RawEmail)},
		// {Type: "text", Value: "INBOX"}, // size
		// {Type: "text", Value: "INBOX"}, // headers
		// {Type: "text", Value: email.Flags},

	})
	if err != nil {
		return err
	}
	resp = sql.Execute(&sql.SqlExecuteRequest{
		Id:     connId,
		Query:  ExecInsertEmail,
		Params: paramsbz2,
	})
	if resp.Error != "" {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

func paramsMarshal(params []sql.SqlQueryParam) ([][]byte, error) {
	res := sql.Params{}
	for _, param := range params {
		paramsbz, err := json.Marshal(&param)
		if err != nil {
			return nil, err
		}
		res = append(res, paramsbz)
	}
	return res, nil
}

func extractEmail(raw []byte) (*EmailWrite, error) {
	envelope := imap.Envelope{}
	msg := strings.NewReader(string(raw))
	hdrs, bodyOffset, err := utils.ParseHeaders(bufio.NewReader(&utils.AtReader{R: msg}))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", dkim.ErrHeaderMalformed, err)
	}
	rawBody, err := io.ReadAll(bufio.NewReader(&utils.AtReader{R: msg, Offset: int64(bodyOffset)}))
	if err != nil {
		return nil, err
	}
	subject := ""
	bh := ""
	messageId := ""
	timestamp := time.Now()
	for _, h := range hdrs {
		switch h.LKey {
		case imap.HEADER_LOW_SUBJECT:
			subject = h.GetValueTrimmed()
			envelope.Subject = subject
		case imap.HEADER_LOW_DATE:
			v := h.GetValueTrimmed()
			fmt.Println("--date---", v)
			t, err := ParseEmailDate(v)
			fmt.Println("--date2---", err, t)
			if err != nil {
				fmt.Println("tinygo.emailchain.extractEmail.Date", err)
			}
			timestamp = t
			envelope.Date = timestamp
			fmt.Println("--date2---", timestamp)
		case imap.HEADER_LOW_MESSAGE_ID:
			v := h.GetValueTrimmed()
			messageId = strings.Trim(v, "<>")
			envelope.MessageID = messageId
		case imap.HEADER_LOW_DKIM_SIGNATURE:
			parts := strings.Split(string(h.GetValueTrimmed()), "; ")
			for _, p := range parts {
				if p[0:3] == "bh=" {
					bh = p[3:]
				}
			}
		case imap.HEADER_LOW_FROM:
			valuestr := h.GetValueTrimmed()
			v, err := imap.ParseEmailAddresses(valuestr)
			if err != nil {
				return nil, err
			}
			envelope.From = v
		case imap.HEADER_LOW_TO:
			valuestr := h.GetValueTrimmed()
			v, err := imap.ParseEmailAddresses(valuestr)
			if err != nil {
				return nil, err
			}
			envelope.To = v
		case imap.HEADER_LOW_CC:
			valuestr := h.GetValueTrimmed()
			v, err := imap.ParseEmailAddresses(valuestr)
			if err != nil {
				return nil, err
			}
			envelope.Cc = v
		case imap.HEADER_LOW_BCC:
			valuestr := h.GetValueTrimmed()
			v, err := imap.ParseEmailAddresses(valuestr)
			if err != nil {
				return nil, err
			}
			envelope.Bcc = v
		case imap.HEADER_LOW_IN_REPLY_TO:
			v := h.GetValueTrimmed()
			fmt.Println("--HEADER_LOW_IN_REPLY_TO--", v)
			inReplyTo := strings.Trim(v, "<>")
			envelope.InReplyTo = []string{inReplyTo}
			fmt.Println("--envelope.InReplyTo--", envelope.InReplyTo)
		case imap.HEADER_LOW_REPLY_TO:
			valuestr := h.GetValueTrimmed()
			fmt.Println("--HEADER_LOW_REPLY_TO--", valuestr)
			v, err := imap.ParseEmailAddresses(valuestr)
			if err != nil {
				return nil, err
			}
			envelope.ReplyTo = v
			fmt.Println("--envelope.ReplyTo--", envelope.ReplyTo)
		default:
			continue
		}
	}
	return &EmailWrite{
		Subject:      subject,
		InternalDate: timestamp,
		RawEmail:     raw,
		Body:         string(rawBody),
		Bh:           bh,
		MessageID:    messageId,
		Envelope:     envelope,
	}, nil
}
