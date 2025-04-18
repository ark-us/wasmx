package vmsql

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// TODO this API is only for priviledged contracts
// have a way of running as consensusless contract
// and as part of the deterministic engine, but without waiting for a side effect/error
// TODO reverting database executions
// TODO tx vs. queries
func Connect(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
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
	// TODO req.Connection - should we restrict this path and make it relative to our DataDirectory? or introduce a list of allowed directories that WASMX can modify and make sure the path is within these directories.
	db, err := sql.Open(req.Driver, req.Connection)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	err = vctx.SetConnection(req.Id, db)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func Close(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
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
	db, found := vctx.GetConnection(req.Id)
	if !found {
		response.Error = "sql connection not found"
		return prepareResponse(rnh, response)
	}
	err = db.Close()
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

func Ping(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
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
	db, found := vctx.GetConnection(req.Id)
	if !found {
		response.Error = "sql connection not found"
		return prepareResponse(rnh, response)
	}
	err = db.Ping()
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	return prepareResponse(rnh, response)
}

// TODO to have flexibility, we need to allow contracts to create the full sql query
// but this has security issues that should be addressed on the contract side
// or use map[string]interface{} to provide JSON-encoded arguments that the host
// can use to construct the query
// Embedding values directly in a SQL query is dangerous and should be avoided
// unless you're doing it safely and are sure the data is not user-controlled.
// Always prefer parameter binding (?) to avoid SQL injection.
func Execute(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
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
	db, found := vctx.GetConnection(req.Id)
	if !found {
		response.Error = "sql connection not found"
		return prepareResponse(rnh, response)
	}

	reqparams := &SqlQueryParams{}
	if len(req.Params) > 0 {
		err = json.Unmarshal(req.Params, reqparams)
		if err != nil {
			response.Error = fmt.Sprintf("invalid query params: %s", err.Error())
			return prepareResponse(rnh, response)
		}
	}

	qparams := parseSqlQueryParams(reqparams.Params)
	res, err := db.Exec(req.Query, qparams...)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

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

	return prepareResponse(rnh, response)
}

func parseSqlQueryParams(params []SqlQueryParam) []interface{} {
	qparams := []interface{}{}
	for _, param := range params {
		switch param.Type {
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

func Query(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
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
	db, found := vctx.GetConnection(req.Id)
	if !found {
		response.Error = "sql connection not found"
		return prepareResponse(rnh, response)
	}

	reqparams := &SqlQueryParams{}
	if len(req.Params) > 0 {
		err = json.Unmarshal(req.Params, reqparams)
		if err != nil {
			response.Error = fmt.Sprintf("invalid query params: %s", err.Error())
			return prepareResponse(rnh, response)
		}
	}

	qparams := parseSqlQueryParams(reqparams.Params)
	rows, err := db.Query(req.Query, qparams...)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}
	defer rows.Close()

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
	ptr, err := rnh.AllocateWriteMem(responsebz)
	if err != nil {
		return nil, err
	}
	returns := make([]interface{}, 1)
	returns[0] = ptr
	return returns, nil
}

func BuildWasmxSqlVM(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	fndefs := []memc.IFn{
		// Connect(req) -> resp
		vm.BuildFn("Connect", Connect, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Close", Close, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		// TODO
		// vm.BuildFn("SetOptions", SetOptions, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Ping", Ping, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Execute", Execute, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Query", Query, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		// vm.BuildFn("QueryRow", QueryRow, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		// vm.BuildFn("Stats", Stats, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "sql", context, fndefs)
}
