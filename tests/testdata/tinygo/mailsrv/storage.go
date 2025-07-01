package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	sql "github.com/loredanacirstea/wasmx-env-sql"
)

type EmailRecord struct {
	Raw       string `json:"raw"`
	Owner     string `json:"owner"`
	MessageID string `json:"message_id"`
	UID       int64  `json:"uid"`
	Folder    string `json:"folder"`
}

// StoreEmail inserts a new email into the database.
func StoreEmail(email EmailRecord) error {
	query := `INSERT INTO emails (raw, owner, message_id, uid, folder) VALUES (?, ?, ?, ?, ?)`
	params := sql.Params{
		[]byte(email.Raw),
		[]byte(email.Owner),
		[]byte(email.MessageID),
		[]byte(strconv.FormatInt(email.UID, 10)),
		[]byte(email.Folder),
	}
	req := &sql.SqlExecuteRequest{
		Id:     "main",
		Query:  query,
		Params: params,
	}
	resp := sql.Execute(req)
	if resp.Error != "" {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

// GetEmail retrieves a single email by message_id.
func GetEmail(owner string, messageID string) (EmailRecord, error) {
	query := `SELECT raw, owner, message_id, uid, folder FROM emails WHERE owner = ? AND message_id = ? LIMIT 1`
	params := sql.Params{
		[]byte(owner),
		[]byte(messageID),
	}
	req := &sql.SqlQueryRequest{
		Id:     "main",
		Query:  query,
		Params: params,
	}
	resp := sql.Query(req)
	if resp.Error != "" {
		return EmailRecord{}, fmt.Errorf(resp.Error)
	}

	var emails []EmailRecord
	err := json.Unmarshal(resp.Data, &emails)
	if err != nil {
		return EmailRecord{}, err
	}
	if len(emails) == 0 {
		return EmailRecord{}, fmt.Errorf("not found")
	}
	return emails[0], nil
}

// ListEmailsByFolder lists emails by folder with pagination.
func ListEmailsByFolder(owner string, folder string, limit int, offset int) ([]EmailRecord, error) {
	query := `SELECT raw, owner, message_id, uid, folder FROM emails WHERE owner = ? AND folder = ? ORDER BY uid DESC LIMIT ? OFFSET ?`
	params := sql.Params{
		[]byte(owner),
		[]byte(folder),
		[]byte(strconv.Itoa(limit)),
		[]byte(strconv.Itoa(offset)),
	}
	req := &sql.SqlQueryRequest{
		Id:     "main",
		Query:  query,
		Params: params,
	}
	resp := sql.Query(req)
	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}
	var emails []EmailRecord
	err := json.Unmarshal(resp.Data, &emails)
	if err != nil {
		return nil, err
	}
	return emails, nil
}
