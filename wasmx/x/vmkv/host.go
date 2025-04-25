package vmkv

import (
	"encoding/json"
	"fmt"

	consensusmeta "cosmossdk.io/store/consensusmeta"
	dbm "github.com/cosmos/cosmos-db"

	vmtypes "github.com/loredanacirstea/wasmx/x/wasmx/vm"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// TODO this API is only for priviledged contracts
func Connect(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req KvConnectionRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetKvDbContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &KvConnectionResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	connstr := req.Dir + req.Name

	conn, found := vctx.GetConnection(connId)
	if found {
		if conn.Connection == connstr {
			// TODO maybe test connection
			return prepareResponse(rnh, response)
		} else {
			response.Error = "connection id already in use"
			return prepareResponse(rnh, response)
		}
	}

	// TODO req.Connection - should we restrict this path and make it relative to our DataDirectory? or introduce a list of allowed directories that WASMX can modify and make sure the path is within these directories.
	db, err := dbm.NewDBwithOptions(req.Name, dbm.BackendType(req.Driver), req.Dir, nil) // TODO opts
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

	err = vctx.SetConnection(connId, connstr, db, closedChannel)
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
	var req KvCloseRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	vctx, err := GetKvDbContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}

	response := &KvCloseResponse{Error: ""}
	connId := buildConnectionId(req.Id, ctx)
	db, found := vctx.GetConnection(connId)
	if !found {
		response.Error = "kv db connection not found"
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

func Get(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvGetResponse{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req KvGetRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	conn, err := getConnectionFromCtx(ctx, req.Id)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	setTempStoreIfNotSet(conn)

	response.Value = conn.getCurrentStore().Get(req.Key)
	return prepareResponse(rnh, response)
}

func Has(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvHasResponse{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req KvHasRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}
	conn, err := getConnectionFromCtx(ctx, req.Id)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	setTempStoreIfNotSet(conn)

	response.Found = conn.getCurrentStore().Has(req.Key)
	return prepareResponse(rnh, response)
}

func Set(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvSetResponse{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req KvSetRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	conn, err := getConnectionFromCtx(ctx, req.Id)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	setTempStoreIfNotSet(conn)

	conn.getCurrentStore().Set(req.Key, req.Value)
	return prepareResponse(rnh, response)
}

func Delete(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvDeleteResponse{Error: ""}
	ctx := _context.(*Context)
	requestbz, err := rnh.ReadMemFromPtr(params[0])
	if err != nil {
		return nil, err
	}
	var req KvSetRequest
	err = json.Unmarshal(requestbz, &req)
	if err != nil {
		return nil, err
	}

	conn, err := getConnectionFromCtx(ctx, req.Id)
	if err != nil {
		response.Error = err.Error()
		return prepareResponse(rnh, response)
	}

	setTempStoreIfNotSet(conn)

	conn.getCurrentStore().Delete(req.Key)
	return prepareResponse(rnh, response)
}

func Iterator(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	response := &KvIteratorResponse{Error: ""}
	return prepareResponse(rnh, response)
}

// func NewBatch(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
// 	response := &KvNewBatchResponse{Error: ""}
// 	return prepareResponse(rnh, response)
// }

func Stats(_context interface{}, rnh memc.RuntimeHandler, params []interface{}) ([]interface{}, error) {
	// TODO implement
	response := map[string]string{}
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

func buildConnectionId(id string, ctx *Context) string {
	return fmt.Sprintf("%s_%s", ctx.Env.Contract.Address.String(), id)
}

func getConnectionFromCtx(ctx *Context, id string) (*KvOpenConnection, error) {
	vctx, err := GetKvDbContext(ctx.Context.GoContextParent)
	if err != nil {
		return nil, err
	}
	connId := buildConnectionId(id, ctx)
	conn, found := vctx.GetConnection(connId)
	if !found {
		return nil, fmt.Errorf("kv db connection not found")
	}
	return conn, nil
}

func setTempStoreIfNotSet(conn *KvOpenConnection) error {
	if conn.Store != nil {
		return nil
	}
	conn.Store = consensusmeta.NewStoreWithDB(&conn.Db)
	key := "sp0"
	err := conn.newCurrentTempStore(key)
	if err != nil {
		return err
	}
	return nil
}

func BuildWasmxKvVM(ctx_ *vmtypes.Context, rnh memc.RuntimeHandler) (interface{}, error) {
	context := &Context{Context: ctx_}
	vm := rnh.GetVm()
	// follow cosmos-db interface
	fndefs := []memc.IFn{
		vm.BuildFn("Connect", Connect, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Close", Close, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Get", Get, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Has", Has, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Set", Set, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Delete", Delete, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		vm.BuildFn("Iterator", Iterator, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		// vm.BuildFn("NewBatch", NewBatch, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		// vm.BuildFn("Print", Print, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
		// vm.BuildFn("Stats", Stats, []interface{}{vm.ValType_I32()}, []interface{}{vm.ValType_I32()}, 0),
	}

	return vm.BuildModule(rnh, "kvdb", context, fndefs)
}
