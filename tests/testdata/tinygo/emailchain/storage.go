package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	// "github.com/emersion/go-message/mail"

	"github.com/loredanacirstea/mailverif/dkim"
	"github.com/loredanacirstea/mailverif/utils"
	wasmx "github.com/loredanacirstea/wasmx-env"
	sql "github.com/loredanacirstea/wasmx-env-sql"
)

// -- Table: owners (users/accounts)
const DefTableOwners = `CREATE TABLE owners (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    address TEXT UNIQUE NOT NULL
);`

const DefTableFolders = `CREATE TABLE folder_state (
    owner TEXT NOT NULL,
    folder TEXT NOT NULL,
    last_uid INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (owner, folder)
);`

// -- Table: emails (each email is one row, per folder/owner)
const DefTableEmails = `CREATE TABLE emails (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner TEXT NOT NULL,
    folder TEXT NOT NULL,
    uid INTEGER NOT NULL,
    seq_num INTEGER NOT NULL,
    message_id TEXT,
    subject TEXT,
    internal_date INTEGER NOT NULL,
    flags TEXT,
    raw_email BLOB NOT NULL,
    size INTEGER NOT NULL,
    headers TEXT,
    body TEXT,
	bh VARCHAR,
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
const ExecGetSeq = `SELECT COALESCE(MAX(seq_num), 0) + 1 AS next_seq_num
FROM emails
WHERE owner = ? AND folder = ?;`
const ExecInsertEmail = `INSERT INTO emails (
	owner, folder, uid, seq_num, message_id, subject, internal_date, bh, body, raw_email, size
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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

func StoreEmail(domain string, from string, to string, emailRaw []byte, connId string) error {
	folder := "INBOX"
	paramsbz, err := paramsMarshal([]sql.SqlQueryParam{
		{Type: "text", Value: to},     // owner
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

	email, err := extractEmail(emailRaw)
	if err != nil {
		return err
	}
	mId, err := GenerateMessageID(domain)
	if err != nil {
		return err
	}
	email.Folder = folder
	email.UID = int64(uid)
	email.SeqNum = int64(seq)
	email.Owner = to
	email.MessageID = mId

	paramsbz2, err := paramsMarshal([]sql.SqlQueryParam{
		{Type: "text", Value: email.Owner},
		{Type: "text", Value: email.Folder},
		{Type: "integer", Value: email.UID},
		{Type: "integer", Value: email.SeqNum},
		{Type: "text", Value: email.MessageID},
		{Type: "text", Value: email.Subject},
		{Type: "text", Value: email.InternalDate},
		{Type: "text", Value: email.Bh},
		{Type: "text", Value: email.Body},
		{Type: "text", Value: string(email.RawEmail)},
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

// GenerateMessageID generates a unique RFC 5322-compliant Message-ID.
// Example: <e4cfd38a7bce4fda9a2a4cc21f24a3b2@yourdomain.com>
func GenerateMessageID(domain string) (string, error) {
	// 16 bytes of randomness
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	timestamp := time.Now().UnixNano()
	localPart := fmt.Sprintf("%x.%d", buf, timestamp)

	return fmt.Sprintf("<%s@%s>", localPart, domain), nil
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
	timestamp := time.Now()
	for _, h := range hdrs {
		switch h.LKey {
		case "subject":
			subject = h.GetValueTrimmed()
		case "date":
			t, err := time.Parse(time.RFC3339, h.GetValueTrimmed())
			if err != nil {
				timestamp = t
			}
		case "dkim-signature":
			parts := strings.Split(string(h.GetValueTrimmed()), "; ")
			for _, p := range parts {
				if p[0:3] == "bh=" {
					bh = p[3:]
				}
			}
		default:
			continue
		}
	}
	return &EmailWrite{
		Subject:      subject,
		InternalDate: timestamp.Unix(),
		RawEmail:     raw,
		Body:         string(rawBody),
		Bh:           bh,
	}, nil
}
