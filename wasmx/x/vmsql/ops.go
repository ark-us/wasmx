package vmsql

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// TODO this API is only for priviledged contracts
func Connect(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req SqlConnectionRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSqlContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SqlConnectionResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)

	conn, found := vctx.GetConnection(connId)
	if found {
		if conn.Connection == req.Connection {
			// we test the connection with a ping
			// if not successful, we try one more time to open it, below
			err := conn.Db.Ping()
			if err == nil {
				return prepareResponse(rnh, response)
			}
		} else {
			response.Error = "connection id already in use"
			return prepareResponse(rnh, response)
		}
	}
	// TODO req.Connection - should we restrict this path and make it relative to our DataDirectory? or introduce a list of allowed directories that WASMX can modify and make sure the path is within these directories.
	db, err := sql.Open(req.Driver, req.Connection)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	closedChannel := make(chan struct{})

	ctx.GoRoutineGroup.Go(func() error {
		select {
		case <-ctx.GoContextParent.Done():
			ctx.Ctx.Logger().Info(fmt.Sprintf("parent context was closed, closing database connection: %s", connId))
			err := db.Close()
			if err != nil {
				ctx.Ctx.Logger().Error(fmt.Sprintf(`database close error for connection id "%s": %v`, connId, err))
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

	err = vctx.SetConnection(connId, req.Connection, db, closedChannel)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	if req.Driver == "sqlite3" {
		db.Exec("PRAGMA journal_mode = WAL")
		db.Exec("PRAGMA foreign_keys = ON")
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
	var req SqlCloseRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSqlContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SqlCloseResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	db, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "sql connection not found"
		return prepareResponse(rnh, response)
	}
	err = db.Db.Close()
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	close(db.Closed) // signal closing the database
	return prepareResponse(rnh, response)
}

func Ping(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req SqlPingRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSqlContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SqlPingResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	db, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "sql connection not found"
		return prepareResponse(rnh, response)
	}
	err = db.Db.PingContext(ctx.Ctx)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

// to have flexibility, we need to allow contracts to create the full sql query
// but this has security issues that should be addressed on the contract side
// and provide JSON-encoded arguments that the host can use to construct the query
// Embedding values directly in a SQL query is dangerous and should be avoided
// unless you're doing it safely and are sure the data is not user-controlled.
// Always prefer parameter binding (?) to avoid SQL injection.
func Execute(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req SqlExecuteRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSqlContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SqlExecuteResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	db, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "sql connection not found"
		return prepareResponse(rnh, response)
	}

	if db.OpenSavepointTx == nil {
		err := beginDbTx(db, ctx)
		if err != nil {
			return nil, err
		}
	}

	reqparams, err := parseRequestParams(req.Params)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	qparams := parseSqlQueryParams(reqparams)
	res, err := db.OpenSavepointTx.ExecContext(ctx.Ctx, req.Query, qparams...)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	prepareExecutionResponse(res, response)

	return prepareResponse(rnh, response)
}

func prepareExecutionResponse(res sql.Result, response *SqlExecuteResponse) {
	id, err := res.LastInsertId()
	if err != nil {
		response.LastInsertIdError = err.Error()
	}
	response.LastInsertId = id
	rows, err := res.RowsAffected()
	if err != nil {
		response.RowsAffectedError = err.Error()
	}
	response.RowsAffected = rows
}

func BatchAtomic(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req SqlExecuteBatchRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSqlContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SqlExecuteBatchResponse{Error: "", Responses: make([]SqlExecuteResponse, 0)}
	connId := buildConnectionId(req.Id, ctx)
	db, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "sql connection not found"
		return prepareResponse(rnh, response)
	}

	if db.OpenSavepointTx == nil {
		err := beginDbTx(db, ctx)
		if err != nil {
			return nil, err
		}
	}

	// to make this atomic, we add a savepoint
	savepoint := "batchatomic"
	_, err = db.OpenSavepointTx.Exec(fmt.Sprintf("SAVEPOINT %s", savepoint))
	if err != nil {
		ctx.Logger(ctx.Ctx).Error(fmt.Sprintf("cannot add savepoint: %s, %s", savepoint, err.Error()))
		return nil, err
	}

	var txerr error
	for _, cmd := range req.Commands {
		reqparams, err := parseRequestParams(cmd.Params)
		if err != nil {
			txerr = err
			break
		}

		qparams := parseSqlQueryParams(reqparams)
		res, err := db.OpenSavepointTx.ExecContext(ctx.Ctx, cmd.Query, qparams...)
		if err != nil {
			txerr = err
			break
		}
		execResp := SqlExecuteResponse{}
		prepareExecutionResponse(res, &execResp)
		response.Responses = append(response.Responses, execResp)
	}
	if txerr != nil {
		_, err = db.OpenSavepointTx.Exec(fmt.Sprintf("ROLLBACK TO %s", savepoint))
		if err != nil {
			ctx.Logger(ctx.Ctx).Error(fmt.Sprintf("cannot rollback to savepoint: %s, %s", savepoint, err.Error()))
			return nil, err
		}
		response.Error = txerr.Error()
		return prepareResponse(rnh, response)
	}

	_, err = db.OpenSavepointTx.Exec(fmt.Sprintf("RELEASE %s", savepoint))
	if err != nil {
		ctx.Logger(ctx.Ctx).Error(fmt.Sprintf("cannot release savepoint: %s, %s", savepoint, err.Error()))
		return nil, err
	}
	return prepareResponse(rnh, response)
}

func Query(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req SqlQueryRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetSqlContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &SqlQueryResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	db, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "sql connection not found"
		return prepareResponse(rnh, response)
	}

	if db.OpenSavepointTx == nil {
		err := beginDbTx(db, ctx)
		if err != nil {
			return nil, err
		}
	}

	reqparams, err := parseRequestParams(req.Params)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	qparams := parseSqlQueryParams(reqparams)
	rows, err := db.OpenSavepointTx.QueryContext(ctx.Ctx, req.Query, qparams...)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	defer rows.Close()

	if rows == nil {
		response.Data = []byte(`[]`)
		return prepareResponse(rnh, response)
	}

	resp, err := RowsToJSON(rows)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	response.Data = resp
	return prepareResponse(rnh, response)
}

func prepareResponse(rnh memc.RuntimeHandler, response interface{}) ([]interface{}, error) {
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	// fmt.Println("--prepareResponse--", string(responsebz))
	return rnh.AllocateWriteMem(responsebz)
}

func beginDbTx(db *SqlOpenConnection, ctx *Context) error {
	tx, err := db.Db.BeginTx(ctx.Ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot begin atomic db transaction: %v", err)
	}
	db.OpenSavepointTx = tx
	db.SavePointMap["sp0"] = true
	_, err = tx.Exec("SAVEPOINT sp0")
	if err != nil {
		return fmt.Errorf("cannot add savepoint sp0: %v", err)
	}
	return nil
}

func parseRequestParams(params [][]byte) ([]SqlQueryParam, error) {
	reqparams := []SqlQueryParam{}
	for _, param := range params {
		reqp := SqlQueryParam{}
		err := json.Unmarshal(param, &reqp)
		if err != nil {
			return nil, fmt.Errorf("invalid query params: %s", err.Error())
		}
		reqparams = append(reqparams, reqp)
	}
	return reqparams, nil
}

func parseSqlQueryParams(params []SqlQueryParam) []interface{} {
	qparams := []interface{}{}
	for _, param := range params {
		typeLower := strings.ToLower(param.Type)
		switch typeLower {
		case "":
			qparams = append(qparams, param.Value)
		case "blob":
			// expect base64-encoded string
			// otherwise, just insert the value as is
			v, ok := param.Value.(string)
			if !ok {
				qparams = append(qparams, param.Value)
				break
			}
			value, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				qparams = append(qparams, param.Value)
				break
			}
			qparams = append(qparams, value)
		default:
			qparams = append(qparams, param.Value)
		}
	}
	return qparams
}

func buildConnectionId(id string, ctx *Context) string {
	return fmt.Sprintf("%s_%s", ctx.Env.Contract.Address.String(), id)
}
