package vmsql

import (
	"database/sql"
	"fmt"
	"sync"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

const HOST_WASMX_ENV_SQL_VER1 = "wasmx_sql_1"

const HOST_WASMX_ENV_SQL_EXPORT = "wasmx_sql_"

const HOST_WASMX_ENV_SQL = "sql"

type ContextKey string

const SqlContextKey ContextKey = "sql-context"

type Context struct {
	*vmtypes.Context
}

type SqlContext struct {
	mtx           sync.Mutex
	DbConnections map[string]*sql.DB
}

func (p *SqlContext) GetConnection(id string) (*sql.DB, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	db, found := p.DbConnections[id]
	return db, found
}

func (p *SqlContext) SetConnection(id string, db *sql.DB) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.DbConnections[id]
	if found {
		return fmt.Errorf("cannot overwrite sql connection: %s", id)
	}
	p.DbConnections[id] = db
	return nil
}

func (p *SqlContext) DeleteConnection(id string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	delete(p.DbConnections, id)
}

type SqlConnectionRequest struct {
	Driver     string `json:"driver"`
	Connection string `json:"connection"`
	Id         string `json:"id"`
}

type SqlConnectionResponse struct {
	Error string `json:"error"`
}

type SqlPingRequest struct {
	Id string `json:"id"`
}

type SqlPingResponse struct {
	Error string `json:"error"`
}

type SqlExecuteRequest struct {
	Id    string `json:"id"`
	Query string `json:"query"`
}

type SqlExecuteResponse struct {
	Error             string `json:"error"`
	LastInsertId      int64  `json:"last_insert_id"`
	LastInsertIdError string `json:"last_insert_id_error"`
	RowsAffected      int64  `json:"rows_affected"`
	RowsAffectedError string `json:"rows_affected_error"`
}

type SqlQueryRequest struct {
	Id    string `json:"id"`
	Query string `json:"query"`
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
