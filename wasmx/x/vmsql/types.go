package vmsql

import (
	"database/sql"
	"fmt"
	"sync"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
)

const (
	// ModuleName defines the module name
	ModuleName = "vmsql"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

const HOST_WASMX_ENV_SQL_VER1 = "wasmx_sql_1"

const HOST_WASMX_ENV_SQL_EXPORT = "wasmx_sql_"

const HOST_WASMX_ENV_SQL = "sql"

type ContextKey string

const SqlContextKey ContextKey = "sql-context"

type Context struct {
	*vmtypes.Context
}

type SqlOpenConnection struct {
	Connection      string
	Db              *sql.DB
	OpenSavepointTx *sql.Tx
	SavePointMap    map[string]bool
	Closed          chan struct{}
}

func (conn *SqlOpenConnection) hasSavePoint(savepoint string) bool {
	sv, ok := conn.SavePointMap[savepoint]
	return ok && sv
}

type SqlContext struct {
	mtx           sync.Mutex
	DbConnections map[string]*SqlOpenConnection
}

func (p *SqlContext) GetConnection(id string) (*SqlOpenConnection, bool) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	db, found := p.DbConnections[id]
	return db, found
}

func (p *SqlContext) SetConnection(id string, connection string, db *sql.DB, closed chan struct{}) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	_, found := p.DbConnections[id]
	if found {
		return fmt.Errorf("cannot overwrite sql connection: %s", id)
	}
	p.DbConnections[id] = &SqlOpenConnection{Db: db, Connection: connection, Closed: closed, SavePointMap: make(map[string]bool, 0)}
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

type SqlExecuteRequest struct {
	Id     string `json:"id"`
	Query  string `json:"query"`
	Params []byte `json:"params"`
}

type SqlQueryParam struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type SqlQueryParams struct {
	Params []SqlQueryParam `json:"params"`
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
	Params []byte `json:"params"`
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
