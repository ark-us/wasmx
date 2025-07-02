package main

import (
	"fmt"

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
    last_uid INTEGER NOT NULL,
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
    UNIQUE (folder, uid),
	UNIQUE (folder, seq_num)
);`

// -- Indexes for fast search and UID operations
const DefIndexes1 = `CREATE UNIQUE INDEX idx_emails_owner_folder_uid ON emails(owner, folder, uid);`
const DefIndexes2 = `CREATE INDEX idx_emails_owner_folder_seq_num ON emails(owner, folder, seq_num);`
const DefIndexes3 = `CREATE UNIQUE INDEX idx_emails_message_id ON emails(message_id);`
const DefIndexes4 = `CREATE INDEX idx_emails_subject ON emails(owner,subject);`
const DefIndexes5 = `CREATE INDEX idx_emails_internal_date ON emails(internal_date);`

const ExecUpdateUid = `UPDATE folder_state
SET last_uid = last_uid + 1
WHERE owner = ? AND folder = ?`
const ExecGetUid = `SELECT last_uid FROM folder_state
WHERE owner = ? AND folder = ?`
const ExecGetSeq = `SELECT COALESCE(MAX(seq_num), 0) + 1 FROM emails
WHERE owner = ? AND folder = ?`
const ExecInsertEmail = `INSERT INTO emails (
	owner, folder, uid, seq_num, message_id, subject, internal_date,
	flags, raw_email, size, headers, body
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

func InitializeTables(connId string) {
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

func StoreEmail(req *IncomingEmail, connId string) error {
	resp := sql.Execute(&sql.SqlExecuteRequest{
		Id:    connId,
		Query: ExecUpdateUid,
	})
	if resp.Error != "" {
		return fmt.Errorf(resp.Error)
	}
	qresp := sql.Query(&sql.SqlQueryRequest{
		Id:    connId,
		Query: ExecGetUid,
	})
	fmt.Println("--uid---", string(qresp.Data))
	// uids := []int64
	// json.Unmarshal(qresp.Data, )
	// uid := 1

	qresp = sql.Query(&sql.SqlQueryRequest{
		Id:    connId,
		Query: ExecGetSeq,
	})
	fmt.Println("--seq---", string(qresp.Data))

	resp = sql.Execute(&sql.SqlExecuteRequest{
		Id:     connId,
		Query:  ExecInsertEmail,
		Params: sql.Params{},
	})

	// sql.BatchAtomic(&sql.SqlExecuteBatchRequest{
	// 	Id: connId,
	// 	Commands: []sql.SqlExecuteCommand{
	// 		{
	// 			Query: ``
	// 		},
	// 	},
	// })
	return nil
}
