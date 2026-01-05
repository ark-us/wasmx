package vmpostgresql

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	dbm "github.com/cosmos/cosmos-db"

	"github.com/jackc/pgx/v5/pgconn"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

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
	connId := buildConnectionId(ctx.Ctx.ChainID(), req.Id, ctx)

	conn, found := vctx.GetConnection(connId)
	if found {
		if conn.Connection == req.Connection {
			// we test the connection with a ping
			// if not successful, we try one more time to open it, below
			err := conn.Db.Pool().Ping(ctx.GoContextParent)
			if err == nil {
				return prepareResponse(rnh, response)
			}
		} else {
			response.Error = "connection id already in use"
			return prepareResponse(rnh, response)
		}
	}

	// TODO options
	db, err := dbm.NewPostgreSQLDbWithCtx(ctx.GoContextParent, req.DbName, req.Connection, nil)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	closedChannel := make(chan struct{})

	ctx.GoRoutineGroup.Go(func() error {
		select {
		case <-ctx.GoContextParent.Done():
			ctx.Ctx.Logger().Info(fmt.Sprintf("parent context was closed, closing database connection: %s", connId))
			db.Close()
			close(closedChannel)
			return nil
		case <-closedChannel:
			// when close signal is received from Close() API
			// database is already closed, so we exit this goroutine
			ctx.Ctx.Logger().Info(fmt.Sprintf("postgresql database connection closed: %s", connId))
			return nil
		}
	})

	err = vctx.SetConnection(connId, req.Connection, db, closedChannel)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	// PostgreSQL-specific settings can be configured via connection string
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
	connId := buildConnectionId(ctx.Ctx.ChainID(), req.Id, ctx)
	db, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "postgresql connection not found"
		return prepareResponse(rnh, response)
	}
	db.Db.Close()
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
	connId := buildConnectionId(ctx.Ctx.ChainID(), req.Id, ctx)
	db, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "postgresql connection not found"
		return prepareResponse(rnh, response)
	}
	err = db.Db.Pool().Ping(ctx.Ctx)
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
	connId := buildConnectionId(ctx.Ctx.ChainID(), req.Id, ctx)
	db, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "postgresql connection not found"
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
	commandTag, err := db.OpenSavepointTx.Tx().Exec(ctx.Ctx, req.Query, qparams...)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	prepareExecutionResponse(commandTag, response)

	return prepareResponse(rnh, response)
}

func prepareExecutionResponse(commandTag pgconn.CommandTag, response *SqlExecuteResponse) {
	// PostgreSQL doesn't provide LastInsertId like other databases
	// Use RETURNING clause in your SQL if you need the inserted ID
	response.RowsAffected = commandTag.RowsAffected()
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
	connId := buildConnectionId(ctx.Ctx.ChainID(), req.Id, ctx)
	db, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "postgresql connection not found"
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
	_, err = db.OpenSavepointTx.Tx().Exec(ctx.Ctx, fmt.Sprintf("SAVEPOINT %s", savepoint))
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
		commandTag, err := db.OpenSavepointTx.Tx().Exec(ctx.Ctx, cmd.Query, qparams...)
		if err != nil {
			txerr = err
			break
		}
		execResp := SqlExecuteResponse{}
		prepareExecutionResponse(commandTag, &execResp)
		response.Responses = append(response.Responses, execResp)
	}
	if txerr != nil {
		_, err = db.OpenSavepointTx.Tx().Exec(ctx.Ctx, fmt.Sprintf("ROLLBACK TO %s", savepoint))
		if err != nil {
			ctx.Logger(ctx.Ctx).Error(fmt.Sprintf("cannot rollback to savepoint: %s, %s", savepoint, err.Error()))
			return nil, err
		}
		response.Error = txerr.Error()
		return prepareResponse(rnh, response)
	}

	_, err = db.OpenSavepointTx.Tx().Exec(ctx.Ctx, fmt.Sprintf("RELEASE %s", savepoint))
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
	connId := buildConnectionId(ctx.Ctx.ChainID(), req.Id, ctx)
	db, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "postgresql connection not found"
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
	rows, err := db.OpenSavepointTx.Tx().Query(ctx.Ctx, req.Query, qparams...)
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

func CreateDatabase(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	keyptr, _ := memc.GetPointerFromParams(rnh, params, 0)
	requestbz, err := rnh.ReadMemFromPtr(keyptr)
	if err != nil {
		return nil, err
	}
	var req SqlCreateDatabase
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	// CREATE DATABASE cannot run inside a transaction block (SQLSTATE 25001)
	response := &SqlCreateDatabaseResponse{Error: ""}
	db, err := dbm.NewPostgreSQLDbWithCtx(ctx.GoContextParent, req.DbName, req.Connection, nil)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	defer db.Close()
	_, err = db.Pool().Exec(ctx.Ctx, fmt.Sprintf(`CREATE DATABASE %s;`, req.DbName))
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	return prepareResponse(rnh, response)
}

func prepareResponse(rnh memc.RuntimeHandler, response interface{}) ([]interface{}, error) {
	responsebz, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return rnh.AllocateWriteMem(responsebz)
}

func beginDbTx(db *SqlOpenConnection, ctx *Context) error {
	batch, err := db.Db.NewBatchWithError()
	if err != nil {
		return fmt.Errorf("cannot begin atomic db transaction: %s", err.Error())
	}
	db.OpenSavepointTx = batch
	db.setSavePoint("sp0")
	_, err = batch.Tx().Exec(ctx.Ctx, "SAVEPOINT sp0")
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
		case "bytea":
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

func buildConnectionId(chainId string, id string, ctx *Context) string {
	return fmt.Sprintf("%s_%s_%s", chainId, ctx.Env.Contract.Address.String(), id)
}
