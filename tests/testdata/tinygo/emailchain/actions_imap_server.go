package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	imap "github.com/loredanacirstea/emailchain/imap"
	vmimap "github.com/loredanacirstea/wasmx-env-imap"
	sql "github.com/loredanacirstea/wasmx-env-sql"
)

func HandleLogin(req *LoginRequest) ([]byte, error) {
	// err := ConnectSql(ConnectionId)
	// if err != nil {
	// 	return nil, err
	// }
	// params, err := paramsMarshal([]sql.SqlQueryParam{
	// 	{Type: "text", Value: req.Username},
	// })
	// resp := sql.Execute(&sql.SqlExecuteRequest{
	// 	Id:     ConnectionId,
	// 	Query:  `INSERT OR IGNORE INTO owners (address) VALUES (?);`,
	// 	Params: params,
	// })
	// if resp.Error != "" {
	// 	return nil, fmt.Errorf(resp.Error)
	// }
	return nil, nil
}

func HandleLogout(req *LogoutRequest) ([]byte, error) {
	return nil, nil
}

func HandleCreate(req *CreateRequest) ([]byte, *vmimap.Error, error) {
	fmt.Println("--HandleCreate--", req.Username, req.Mailbox)

	exists, err := checkFolderExists(ConnectionId, req.Username, req.Mailbox)
	if err != nil {
		return nil, nil, err
	}

	if exists {
		// return nil, err
		return nil, &vmimap.Error{
			Type: vmimap.StatusResponseTypeNo,
			Code: vmimap.ResponseCodeAlreadyExists,
			Text: vmimap.ErrMailboxAlreadyExists,
		}, nil
	}
	// Insert initial folder state for the user if not exists
	query := `
		INSERT INTO folder_state (owner, folder, last_uid)
		VALUES (?, ?, 0)
		ON CONFLICT(owner, folder) DO NOTHING;
	`

	params, err := paramsMarshal([]sql.SqlQueryParam{
		{Type: "text", Value: req.Username},
		{Type: "text", Value: req.Mailbox},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create: param marshal error: %s", err.Error())
	}

	res := sql.Execute(&sql.SqlExecuteRequest{
		Id:     ConnectionId,
		Query:  query,
		Params: params,
	})

	if res.Error != "" {
		return nil, nil, fmt.Errorf("create: execute error: %s", res.Error)
	}

	return []byte(`{}`), nil, nil
}

func HandleDelete(req *DeleteRequest) ([]byte, error) {
	// TODO: requires deleting mailbox folder state + all messages in it
	return nil, nil
}

func HandleRename(req *RenameRequest) ([]byte, error) {
	// TODO: requires updating folder field in emails and folder_state
	return nil, nil
}

func HandleSelect(req *SelectRequest) ([]byte, error) {
	fmt.Println("--HandleSelect--")
	val, err := GetNumMessages(req.Username, req.Mailbox)
	if err != nil {
		return nil, err
	}
	sel := &imap.SelectData{
		NumMessages: uint32(val),
	}
	bz, err := json.Marshal(sel)
	if err != nil {
		return nil, fmt.Errorf("select: result marshal error: %s", err.Error())
	}
	fmt.Println("--SelectData resp--", string(bz))
	return bz, nil
}

func HandleList(req *ListRequest) ([]byte, error) {
	params, err := paramsMarshal([]sql.SqlQueryParam{
		{Type: "text", Value: req.Username},
	})
	if err != nil {
		return nil, err
	}
	res := sql.Query(&sql.SqlQueryRequest{
		Id:     ConnectionId,
		Query:  ExecGetFolders,
		Params: params,
	})
	folders := []FolderState{}
	err = json.Unmarshal(res.Data, &folders)
	if err != nil {
		return nil, err
	}
	var out []imap.ListData
	for _, f := range folders {
		attrs := GetAttrs(f.Folder)
		status, err := GetFolderStatus(req.Username, f.Folder, &f)
		if err != nil {
			return nil, err
		}
		out = append(out, imap.ListData{
			Attrs:   attrs,
			Mailbox: f.Folder,
			Status:  status,
		})
	}
	bz, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func HandleStatus(req *StatusRequest) ([]byte, error) {
	status, err := GetFolderStatus(req.Username, req.Mailbox, nil)
	if err != nil {
		return nil, err
	}
	bz, err := json.Marshal(status)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func HandleAppend(req *AppendRequest) ([]byte, error) {
	err := StoreEmail(req.Username, []string{}, req.EmailRaw, ConnectionId, req.Mailbox)
	if err != nil {
		return nil, err
	}
	q := `SELECT uid FROM emails WHERE owner = ? AND folder = ? ORDER BY uid DESC LIMIT 1`
	params, err := paramsMarshal([]sql.SqlQueryParam{
		{Type: "text", Value: req.Username},
		{Type: "text", Value: req.Mailbox},
	})
	if err != nil {
		return nil, err
	}
	qres := sql.Query(&sql.SqlQueryRequest{
		Id:     ConnectionId,
		Query:  q,
		Params: params,
	})
	var result []struct {
		UID imap.UID `json:"uid"`
	}
	err = json.Unmarshal(qres.Data, &result)
	if err != nil {
		return nil, err
	}
	res, err := json.Marshal(&imap.AppendData{UID: result[0].UID})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func HandleExpunge(req *ExpungeRequest) ([]byte, error) {
	var expunged []uint32
	for _, r := range *req.Uids {
		for i := r.Start; i <= r.Stop; i++ {
			params, err := paramsMarshal([]sql.SqlQueryParam{
				{Type: "text", Value: req.Username},
				{Type: "integer", Value: i},
			})
			if err != nil {
				return nil, err
			}
			resp := sql.Execute(&sql.SqlExecuteRequest{
				Id:     ConnectionId,
				Query:  `DELETE FROM emails WHERE owner = ? AND uid = ?`,
				Params: params,
			})
			if resp.Error == "" {
				expunged = append(expunged, uint32(i))
			}
		}
	}
	bz, err := json.Marshal(expunged)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func HandleSearch(req *SearchRequest) ([]byte, error) {
	if req.Criteria == nil {
		return nil, fmt.Errorf("missing search criteria")
	}

	where := []string{"owner = ?"}
	params := []sql.SqlQueryParam{
		{Type: "text", Value: req.Username},
	}

	// Match known header fields (e.g. From, To, Subject)
	for _, h := range req.Criteria.Header {
		lkey := strings.ToLower(h.Key)
		switch lkey {
		case "from", "to", "cc", "bcc":
			where = append(where, "headers LIKE ?")
			params = append(params, sql.SqlQueryParam{
				Type:  "text",
				Value: "%" + h.Key + ": " + h.Value + "%",
			})
		case "subject":
			where = append(where, "subject LIKE ?")
			params = append(params, sql.SqlQueryParam{
				Type:  "text",
				Value: "%" + h.Value + "%",
			})
		default:
			// fallback: generic match in headers
			where = append(where, "headers LIKE ?")
			params = append(params, sql.SqlQueryParam{
				Type:  "text",
				Value: "%" + h.Key + ": " + h.Value + "%",
			})
		}
	}

	// Date filters
	if !req.Criteria.Since.IsZero() {
		where = append(where, "internal_date >= ?")
		params = append(params, sql.SqlQueryParam{
			Type:  "integer",
			Value: req.Criteria.Since.Unix(),
		})
	}
	if !req.Criteria.Before.IsZero() {
		where = append(where, "internal_date < ?")
		params = append(params, sql.SqlQueryParam{
			Type:  "integer",
			Value: req.Criteria.Before.Unix(),
		})
	}

	// Size filters
	if req.Criteria.Smaller > 0 {
		where = append(where, "size < ?")
		params = append(params, sql.SqlQueryParam{
			Type:  "integer",
			Value: req.Criteria.Smaller,
		})
	}
	if req.Criteria.Larger > 0 {
		where = append(where, "size > ?")
		params = append(params, sql.SqlQueryParam{
			Type:  "integer",
			Value: req.Criteria.Larger,
		})
	}

	query := fmt.Sprintf(`
		SELECT uid FROM emails
		WHERE %s
		ORDER BY uid ASC
	`, strings.Join(where, " AND "))

	paramBytes, err := paramsMarshal(params)
	if err != nil {
		return nil, err
	}

	resp := sql.Query(&sql.SqlQueryRequest{
		Id:     ConnectionId,
		Query:  query,
		Params: paramBytes,
	})
	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}

	var rows []struct {
		SeqNum uint32 `json:"seq_num"`
		Uid    uint32 `json:"uid"`
	}
	if err := json.Unmarshal(resp.Data, &rows); err != nil {
		return nil, err
	}

	// Convert to imap.SeqSet (your NumSet interface)
	var ranges imap.UIDSet
	for _, row := range rows {
		ranges = append(ranges, imap.UIDRange{
			Start: imap.UID(row.Uid),
			Stop:  imap.UID(row.Uid),
		})
	}

	result := &imap.SearchData{
		All: ranges,
	}

	return json.Marshal(result)
}

type Range struct {
	Start, Stop int
}

func HandleFetch(req *FetchRequest) ([]byte, error) {
	if len(req.UidSet) == 0 && len(req.SeqSet) == 0 {
		return json.Marshal([]interface{}{}) // empty result
	}
	field := "uid"
	ranges := make([]Range, 0)
	if len(req.UidSet) > 0 {
		for _, v := range req.UidSet {
			ranges = append(ranges, Range{Start: int(v.Start), Stop: int(v.Stop)})
		}
	} else {
		field = "seq_num"
		for _, v := range req.SeqSet {
			ranges = append(ranges, Range{Start: int(v.Start), Stop: int(v.Stop)})
		}
	}
	rangeClause, rangeParams, err := buildRangeClause(req.Username, req.Mailbox, field, ranges)
	if err != nil {
		return nil, err
	}

	params := []sql.SqlQueryParam{
		{Type: "text", Value: req.Username},
	}
	params = append(params, rangeParams...)

	if len(rangeClause) > 0 {
		rangeClause = fmt.Sprintf(`AND (%s)`, rangeClause)
	}

	query := fmt.Sprintf(`
		SELECT seq_num, uid, flags, internal_date, size, body, subject, envelope, raw_email
		FROM emails
		WHERE owner = ? %s
		ORDER BY %s ASC;
	`, rangeClause, field)

	pbz, err := paramsMarshal(params)
	if err != nil {
		return nil, err
	}

	resp := sql.Query(&sql.SqlQueryRequest{
		Id:     ConnectionId,
		Query:  query,
		Params: pbz,
	})
	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}

	var emails []struct {
		SeqNum       int    `json:"seq_num"`
		UID          int    `json:"uid"`
		Flags        string `json:"flags"`
		InternalDate int64  `json:"internal_date"`
		Size         int    `json:"size"`
		Body         string `json:"body"`
		Subject      string `json:"subject"`
		Envelope     string `json:"envelope"`
		RawEmail     []byte `json:"raw_email"`
	}
	// var emails []EmailWrite
	fmt.Println("--tinygo.HandleFetch--", string(resp.Data))
	if err := json.Unmarshal(resp.Data, &emails); err != nil {
		fmt.Println("--tinygo.HandleFetch unmarshal--", err)
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(emails))
	for _, e := range emails {
		fmt.Println("--tinygo.HandleFetch email--", e)
		flags := []string{}
		if e.Flags != "" {
			flags = strings.Split(e.Flags, " ")
		}
		t := time.Unix(0, e.InternalDate*int64(time.Millisecond)).UTC()
		envelope := vmimap.Envelope{}
		fmt.Println("--tinygo.HandleFetch envelope--", e.Envelope)
		err = json.Unmarshal([]byte(e.Envelope), &envelope)
		fmt.Println("--tinygo.HandleFetch envelope unmarshal err--", err)
		if err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"seq_num":       e.SeqNum,
			"uid":           e.UID,
			"flags":         flags,
			"internal_date": t.Format(time.RFC1123Z),
			"rfc822size":    e.Size,
			"body":          e.Body,
			"envelope":      envelope,
			"raw_email":     e.RawEmail,
		})
		fmt.Println("--tinygo.HandleFetch results--", results)
	}
	fmt.Println("--tinygo.HandleFetch results--", len(results))
	return json.Marshal(results)
}

func HandleStore(req *StoreRequest) ([]byte, error) {
	// TODO: update flags in emails table based on req.NumSet and req.Flags
	return nil, nil
}

func HandleCopy(req *CopyRequest) ([]byte, error) {
	if len(req.UidSet) == 0 && len(req.SeqSet) == 0 {
		return json.Marshal([]interface{}{}) // empty result
	}
	field := "uid"
	ranges := make([]Range, 0)
	if len(req.UidSet) > 0 {
		for _, v := range req.UidSet {
			ranges = append(ranges, Range{Start: int(v.Start), Stop: int(v.Stop)})
		}
	} else {
		field = "seq_num"
		for _, v := range req.SeqSet {
			ranges = append(ranges, Range{Start: int(v.Start), Stop: int(v.Stop)})
		}
	}
	rangeClause, rangeParams, err := buildRangeClause(req.Username, req.Mailbox, field, ranges)
	if err != nil {
		return nil, err
	}

	srcParams := []sql.SqlQueryParam{
		{Type: "text", Value: req.Username},
	}
	srcParams = append(srcParams, rangeParams...)

	query := fmt.Sprintf(`
		SELECT uid, seq_num, message_id, subject, internal_date, bh, body, raw_email, size
		FROM emails
		WHERE owner = ? AND (%s)
	`, rangeClause)

	pbz, err := paramsMarshal(srcParams)
	if err != nil {
		return nil, err
	}

	resp := sql.Query(&sql.SqlQueryRequest{
		Id:     ConnectionId,
		Query:  query,
		Params: pbz,
	})
	if resp.Error != "" {
		return nil, fmt.Errorf(resp.Error)
	}

	var srcEmails []EmailWrite
	if err := json.Unmarshal(resp.Data, &srcEmails); err != nil {
		return nil, err
	}

	// var destUIDs []imap.UID
	destUIDs := imap.UIDRange{Start: 0, Stop: 0}

	for _, email := range srcEmails {
		// Update UID
		paramsOwnerFolder, _ := paramsMarshal([]sql.SqlQueryParam{
			{Type: "text", Value: req.Username},
			{Type: "text", Value: req.Dest},
		})
		resp := sql.Execute(&sql.SqlExecuteRequest{
			Id:     ConnectionId,
			Query:  ExecUpdateUid,
			Params: paramsOwnerFolder,
		})
		if resp.Error != "" {
			return nil, fmt.Errorf(resp.Error)
		}

		qresp := sql.Query(&sql.SqlQueryRequest{
			Id:     ConnectionId,
			Query:  ExecGetUid,
			Params: paramsOwnerFolder,
		})
		var uidRes []UidResponse
		_ = json.Unmarshal(qresp.Data, &uidRes)
		uid := uidRes[0].LastUid
		// destUIDs = append(destUIDs, imap.UID(uid))
		if destUIDs.Start == 0 {
			destUIDs.Start = imap.UID(uid)
		} else {
			destUIDs.Stop = imap.UID(uid)
		}

		qresp = sql.Query(&sql.SqlQueryRequest{
			Id:     ConnectionId,
			Query:  ExecGetSeq,
			Params: paramsOwnerFolder,
		})
		var seqRes []SeqNumResponse
		_ = json.Unmarshal(qresp.Data, &seqRes)
		seq := seqRes[0].NextSeqNum

		// Insert copied email
		paramsInsert, err := paramsMarshal([]sql.SqlQueryParam{
			{Type: "text", Value: req.Username},
			{Type: "text", Value: req.Dest},
			{Type: "integer", Value: uid},
			{Type: "integer", Value: seq},
			{Type: "text", Value: email.MessageID},
			{Type: "text", Value: email.Subject},
			{Type: "integer", Value: email.InternalDate.Unix()},
			{Type: "text", Value: email.Bh},
			{Type: "text", Value: email.Body},
			{Type: "blob", Value: email.RawEmail},
			{Type: "text", Value: email.Size},
		})
		if err != nil {
			return nil, err
		}
		resp = sql.Execute(&sql.SqlExecuteRequest{
			Id:     ConnectionId,
			Query:  ExecInsertEmail,
			Params: paramsInsert,
		})
		if resp.Error != "" {
			return nil, fmt.Errorf(resp.Error)
		}
	}

	return json.Marshal(&imap.CopyData{
		UIDValidity: 1,
		SourceUIDs:  req.UidSet,
		DestUIDs:    imap.UIDSet{destUIDs},
	})
}

func buildRangeClause(username, mailbox, column string, ranges []Range) (string, []sql.SqlQueryParam, error) {
	fmt.Println("--buildRangeClause--", username, mailbox, column, ranges)
	var clauses []string
	var params []sql.SqlQueryParam
	var err error

	ranges, err = resolveUIDRanges(username, mailbox, ranges)
	if err != nil {
		return "", nil, err
	}
	fmt.Println("--buildRangeClause.ranges--", ranges)
	if len(ranges) == 0 {
		return "", nil, nil
	}

	for _, r := range ranges {
		clauses = append(clauses, fmt.Sprintf("(%s BETWEEN ? AND ?)", column))
		params = append(params,
			sql.SqlQueryParam{Type: "integer", Value: r.Start},
			sql.SqlQueryParam{Type: "integer", Value: r.Stop},
		)
	}
	return strings.Join(clauses, " OR "), params, nil
}

func resolveUIDRanges(owner string, folder string, ranges []Range) ([]Range, error) {
	var out []Range
	maxUID := uint32(0)

	for _, r := range ranges {
		start := r.Start
		stop := r.Stop

		if stop == 0 {
			if maxUID == 0 {
				// only fetch max UID once
				params, _ := paramsMarshal([]sql.SqlQueryParam{
					{Type: "text", Value: owner},
					{Type: "text", Value: folder},
				})
				qres := sql.Query(&sql.SqlQueryRequest{
					Id:     ConnectionId,
					Query:  `SELECT MAX(uid) as max_uid FROM emails WHERE owner = ? AND folder = ?`,
					Params: params,
				})
				if qres.Error != "" {
					return nil, fmt.Errorf(qres.Error)
				}
				var res []struct {
					MaxUID uint32 `json:"max_uid"`
				}
				if err := json.Unmarshal(qres.Data, &res); err != nil {
					return nil, err
				}
				maxUID = res[0].MaxUID
			}
			stop = int(maxUID)
		}

		if start > stop {
			continue // invalid range, skip
		}

		out = append(out, Range{Start: start, Stop: stop})
	}

	return out, nil
}

func GetNumMessages(username string, mailbox string) (int, error) {
	q := `SELECT COUNT(*) as messages FROM emails WHERE owner = ? AND folder = ?`
	params, err := paramsMarshal([]sql.SqlQueryParam{
		{Type: "text", Value: username},
		{Type: "text", Value: mailbox},
	})
	if err != nil {
		return 0, err
	}
	res := sql.Query(&sql.SqlQueryRequest{
		Id:     ConnectionId,
		Query:  q,
		Params: params,
	})
	var result []struct {
		Messages int `json:"messages"`
	}
	err = json.Unmarshal(res.Data, &result)
	if err != nil {
		return 0, err
	}
	return result[0].Messages, nil
}

func GetFolderStatus(username string, mailbox string, folderState *FolderState) (*imap.StatusData, error) {
	val, err := GetNumMessages(username, mailbox)
	if err != nil {
		return nil, err
	}
	v := uint32(val)

	if folderState == nil {
		params, err := paramsMarshal([]sql.SqlQueryParam{
			{Type: "text", Value: username},
			{Type: "text", Value: mailbox},
		})
		if err != nil {
			return nil, err
		}
		res := sql.Query(&sql.SqlQueryRequest{
			Id:     ConnectionId,
			Query:  ExecGetFolder,
			Params: params,
		})
		fres := []FolderState{}
		err = json.Unmarshal(res.Data, &fres)
		if err != nil {
			return nil, err
		}
		if len(fres) == 0 {
			return nil, fmt.Errorf("folder not found")
		}
		folderState = &fres[0]
	}
	lastUid := uint32(folderState.LastUid)
	return &imap.StatusData{
		Mailbox:        mailbox,
		NumMessages:    &v,
		UIDNext:        imap.UID(lastUid + 1),
		UIDValidity:    folderState.UidValidity,
		NumUnseen:      PtrUint32(uint32(0)),
		NumDeleted:     PtrUint32(uint32(0)),
		Size:           PtrInt64(int64(0)),
		AppendLimit:    PtrUint32(uint32(0)),
		DeletedStorage: PtrInt64(int64(0)),
		HighestModSeq:  uint64(0),
	}, nil
}
