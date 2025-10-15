package sql

import (
	"encoding/base64"
	"encoding/json"
)

type SqlConnectionRequest struct {
	Driver     string `json:"driver"`
	Connection string `json:"connection"`
	Id         string `json:"id"`
}

type SqlConnectionResponse struct {
	Error string `json:"error"`
}

type SqlCloseRequest struct {
	Id string `json:"id"`
}

type SqlCloseResponse struct {
	Error string `json:"error"`
}

type SqlPingRequest struct {
	Id string `json:"id"`
}

type SqlPingResponse struct {
	Error string `json:"error"`
}

type Params [][]byte

// MarshalJSON - marshal Params as array of base64 strings
func (p Params) MarshalJSON() ([]byte, error) {
	strs := make([]string, len(p))
	for i, b := range p {
		strs[i] = base64.StdEncoding.EncodeToString(b)
	}
	return json.Marshal(strs)
}

// UnmarshalJSON - unmarshal Params from array of base64 strings
func (p *Params) UnmarshalJSON(data []byte) error {
	var strs []string
	if err := json.Unmarshal(data, &strs); err != nil {
		return err
	}
	decoded := make([][]byte, len(strs))
	for i, s := range strs {
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return err
		}
		decoded[i] = b
	}
	*p = decoded
	return nil
}

type SqlExecuteRequest struct {
	Id     string `json:"id"`
	Query  string `json:"query"`
	Params Params `json:"params"`
}

type SqlExecuteCommand struct {
	Query  string `json:"query"`
	Params Params `json:"params"`
}

type SqlExecuteBatchRequest struct {
	Id       string              `json:"id"`
	Commands []SqlExecuteCommand `json:"commands"`
}

type SqlExecuteBatchResponse struct {
	Error     string               `json:"error"`
	Responses []SqlExecuteResponse `json:"responses"`
}

type SqlQueryParam struct {
	Type string `json:"type"`
	// TODO consider using []byte instead of interface{} -> encode-decode for each supported type
	Value interface{} `json:"value"`
}

type SqlExecuteResponse struct {
	Error             string `json:"error"`
	LastInsertId      int64  `json:"last_insert_id"`
	LastInsertIdError string `json:"last_insert_id_error"`
	RowsAffected      int64  `json:"rows_affected"`
	RowsAffectedError string `json:"rows_affected_error"`
}

type SqlQueryRequest struct {
	Id     string `json:"id"`
	Query  string `json:"query"`
	Params Params `json:"params"`
}

type SqlQueryResponse struct {
	Error string `json:"error"`
	Data  []byte `json:"data"`
}

type SqlQueryRowRequest struct {
	Id string `json:"id"`
}

type SqlQueryRowResponse struct {
	Error string `json:"error"`
}
