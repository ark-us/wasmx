package keeper_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/vmsql"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"
)

type InstantiateDType struct {
	Dir    string `json:"dir"`
	Driver string `json:"driver"`
}

type CreateTableDTypeRequest struct {
	TableId int64 `json:"table_id"`
}

type CreateTableDTypeResponse struct {
}

type TableIdentifier struct {
	DbConnectionId   int64  `json:"db_connection_id"`
	DbId             int64  `json:"db_id"`
	TableId          int64  `json:"table_id"`
	DbConnectionName string `json:"db_connection_name"`
	DbName           string `json:"db_name"`
	TableName        string `json:"table_name"`
}

type InsertDTypeRequest struct {
	Identifier TableIdentifier `json:"identifier"`
	Data       []byte          `json:"data"`
}

type UpdateDTypeRequest struct {
	Identifier TableIdentifier `json:"identifier"`
	Condition  []byte          `json:"condition"`
	Data       []byte          `json:"data"`
}

type DeleteDTypeRequest struct {
	Identifier TableIdentifier `json:"identifier"`
	Condition  []byte          `json:"condition"`
}

type ReadDTypeRequest struct {
	Identifier TableIdentifier `json:"identifier"`
	Data       []byte          `json:"data"`
}

type BuildSchemaRequest struct {
	Identifier TableIdentifier `json:"identifier"`
}

type BuildSchemaResponse struct {
	Data []byte `json:"data"`
}

type InsertDTypeResponse struct {
}

type ConnectRequest struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type CalldataDType struct {
	Initialize      *InstantiateDType        `json:"Initialize,omitempty"`
	CreateTable     *CreateTableDTypeRequest `json:"CreateTable,omitempty"`
	Connect         *ConnectRequest          `json:"Connect,omitempty"`
	Close           *ConnectRequest          `json:"Close,omitempty"`
	Insert          *InsertDTypeRequest      `json:"Insert,omitempty"`
	InsertOrReplace *InsertDTypeRequest      `json:"InsertOrReplace,omitempty"`
	Update          *UpdateDTypeRequest      `json:"Update,omitempty"`
	Delete          *DeleteDTypeRequest      `json:"Delete,omitempty"`
	Read            *ReadDTypeRequest        `json:"Read,omitempty"`
	BuildSchema     *BuildSchemaRequest      `json:"BuildSchema,omitempty"`
}

func (suite *KeeperTestSuite) TestDType() {
	defer os.Remove("dtype.db")
	defer os.Remove("dtype.db-shm")
	defer os.Remove("dtype.db-wal")
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	wasmbin := precompiles.GetPrecompileByLabel(appA.AddressCodec(), types.DTYPE_v001)
	codeId := appA.StoreCode(sender, wasmbin, nil)
	cmdi := &InstantiateDType{Dir: "", Driver: "sqlite3"}
	data, err := json.Marshal(cmdi)
	suite.Require().NoError(err)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: data}, "dtype", nil)

	// set a role to have access to protected APIs
	suite.registerRole("dtype", contractAddress, sender)

	cmd := &CalldataDType{Initialize: cmdi}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	connFile := "newdb.db"
	connDriver := "sqlite3"
	connName := "newdbconn"
	dbname := "newdb"
	tablename1 := "newtable1"

	defer os.Remove(connFile)
	defer os.Remove(connFile + "-shm")
	defer os.Remove(connFile + "-wal")

	identif := suite.createDb(
		sender,
		contractAddress,
		connFile,
		connDriver,
		connName,
		dbname,
	)
	tableId1 := suite.createTable(
		sender,
		contractAddress,
		identif.DbId,
		tablename1,
	)

	fields := []string{
		fmt.Sprintf(`{"name":"id","table_id":%d,"order_index":1,"value_type":"INTEGER","indexed":false,"sql_options":"PRIMARY KEY","permissions":""}`, tableId1),
		fmt.Sprintf(`{"name":"field1","table_id":%d,"order_index":2,"value_type":"VARCHAR","indexed":false,"permissions":""}`, tableId1),
		fmt.Sprintf(`{"name":"field2","table_id":%d,"order_index":3,"value_type":"INTEGER","indexed":false,"permissions":""}`, tableId1),
		fmt.Sprintf(`{"name":"field3","table_id":%d,"order_index":4,"value_type":"BLOB","indexed":false,"permissions":""}`, tableId1),
		fmt.Sprintf(`{"name":"field4","table_id":%d,"order_index":5,"value_type":"BOOLEAN","indexed":false,"permissions":""}`, tableId1),
	}

	suite.createFields(sender, contractAddress, fields)
	suite.instantiateTable(sender, contractAddress, identif.DbConnectionId, tableId1)

	// create table2
	tablename2 := "newtable2"
	tableId2 := suite.createTable(
		sender,
		contractAddress,
		identif.DbId,
		tablename2,
	)

	fields2 := []string{
		fmt.Sprintf(`{"name":"id","table_id":%d,"order_index":1,"value_type":"INTEGER","indexed":false,"sql_options":"PRIMARY KEY","permissions":""}`, tableId2),
		fmt.Sprintf(`{"name":"table1_id","table_id":%d,"order_index":2,"value_type":"INTEGER","indexed":true,"foreign_key_table":"newtable1","foreign_key_field":"id","foreign_key_sql_options":"ON DELETE CASCADE ON UPDATE CASCADE","permissions":""}`, tableId2),
		fmt.Sprintf(`{"name":"field1","table_id":%d,"order_index":3,"value_type":"VARCHAR","indexed":false,"permissions":""}`, tableId2),
	}

	suite.createFields(sender, contractAddress, fields2)
	suite.instantiateTable(sender, contractAddress, identif.DbConnectionId, tableId2)

	// build json schema for table1

	// create rows in table1
	cmd = &CalldataDType{BuildSchema: &BuildSchemaRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId1,
	},
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	var schemaResp BuildSchemaResponse
	err = json.Unmarshal(qres, &schemaResp)
	suite.Require().NoError(err, string(qres))

	// create rows in table1
	cmd = &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId1,
	},
		Data: []byte(fmt.Sprintf(`{"field1":"somevalue","field2":2,"field3":"%s","field4":true}`, base64.StdEncoding.EncodeToString([]byte(`someblobvalue`)))),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	row1_1 := resss.LastInsertId

	cmd = &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId1,
	},
		Data: []byte(fmt.Sprintf(`{"field1":"somevalue2","field2":3,"field3":"%s","field4":false}`, base64.StdEncoding.EncodeToString([]byte(`dgfdgdfgdf`)))),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	// row1_2 := resss.LastInsertId

	// create rows in table2
	cmd = &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId2,
	},
		Data: []byte(fmt.Sprintf(`{"field1":"somevalue","table1_id":%d}`, row1_1)),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)

	cmd = &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId2,
	},
		Data: []byte(fmt.Sprintf(`{"field1":"somevalue2","table1_id":%d}`, row1_1)),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)

	cmd = &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId2,
	},
		Data: []byte(fmt.Sprintf(`{"field1":"somevalue3","table1_id":%d}`, row1_1)),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)

	// update rows in table2
	cmd = &CalldataDType{Update: &UpdateDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId2,
	},
		Condition: []byte(`{"id":1}`),
		Data:      []byte(`{"field1":"somevalue1111"}`),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)

	// read
	cmd = &CalldataDType{Read: &ReadDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId2,
	},
		Data: []byte(`{"id":1}`),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp := suite.parseQueryResponse(qres)
	var table2rows []struct {
		Field1   string `json:"field1"`
		ID       int64  `json:"id"`
		Table1ID int64  `json:"table1_id"`
	}
	err = json.Unmarshal(qresp.Data, &table2rows)
	suite.Require().NoError(err, string(qresp.Data))
	suite.Require().Equal(1, len(table2rows))

	// count
	cmd = &CalldataDType{Read: &ReadDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId2,
	},
		Data: []byte(`{}`),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp = suite.parseQueryResponse(qres)
	err = json.Unmarshal(qresp.Data, &table2rows)
	suite.Require().NoError(err, string(qresp.Data))
	suite.Require().Equal(3, len(table2rows))

	// delete in cascade
	cmd = &CalldataDType{Delete: &DeleteDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId1,
	},
		Condition: []byte(`{"id":1}`),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)

	// count
	cmd = &CalldataDType{Read: &ReadDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId2,
	},
		Data: []byte(`{}`),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp = suite.parseQueryResponse(qres)
	err = json.Unmarshal(qresp.Data, &table2rows)
	suite.Require().NoError(err, string(qresp.Data))
	suite.Require().Equal(0, len(table2rows))

	// close connection
	cmd = &CalldataDType{Close: &ConnectRequest{
		Id: identif.DbConnectionId,
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose := &vmsql.SqlCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)

	cmd = &CalldataDType{Close: &ConnectRequest{
		Name: "dtype_connection",
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose = &vmsql.SqlCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)
}

func (suite *KeeperTestSuite) createDb(
	sender simulation.Account,
	contractAddress mcodec.AccAddressPrefixed,
	connFile string,
	connDriver string,
	connName string,
	dbname string,
) TableIdentifier {
	appA := s.AppContext()
	// insert new database with connection
	cmd := &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionName: "dtype_connection",
		DbName:           "dtype",
		TableName:        "dtype_db_connection",
	},
		Data: []byte(fmt.Sprintf(`{"connection":"%s","driver":"%s","name":"%s"}`, connFile, connDriver, connName)),
	}}
	data, err := json.Marshal(cmd)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	newconnId := resss.LastInsertId

	// new db definition
	cmd = &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionName: "dtype_connection",
		DbName:           "dtype",
		TableName:        "dtype_db",
	},
		Data: []byte(fmt.Sprintf(`{"name":"%s","connection_id":%d}`, dbname, newconnId)),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	newDbId := resss.LastInsertId

	return TableIdentifier{
		DbConnectionId:   newconnId,
		DbId:             newDbId,
		DbConnectionName: connName,
		DbName:           dbname,
	}
}

func (suite *KeeperTestSuite) createTable(
	sender simulation.Account,
	contractAddress mcodec.AccAddressPrefixed,
	dbId int64,
	tablename string,
) int64 {
	appA := s.AppContext()
	// insert new table definition
	cmd := &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionName: "dtype_connection",
		DbName:           "dtype",
		TableName:        "dtype_table",
	},
		Data: []byte(fmt.Sprintf(`{"name":"%s","db_id":%d}`, tablename, dbId)),
	}}
	data, err := json.Marshal(cmd)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	tableId := resss.LastInsertId

	return tableId
}

func (suite *KeeperTestSuite) createFields(
	sender simulation.Account,
	contractAddress mcodec.AccAddressPrefixed,
	fields []string,
) {
	appA := s.AppContext()
	// insert table fields
	for _, fielddef := range fields {
		cmd := &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
			DbConnectionName: "dtype_connection",
			DbName:           "dtype",
			TableName:        "dtype_field",
		},
			Data: []byte(fielddef),
		}}
		data, err := json.Marshal(cmd)
		suite.Require().NoError(err)
		res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
		resss := &vmsql.SqlExecuteResponse{}
		err = appA.DecodeExecuteResponse(res, resss)
		suite.Require().NoError(err)
		suite.Require().Equal("", resss.Error)
	}
}

func (suite *KeeperTestSuite) instantiateTable(
	sender simulation.Account,
	contractAddress mcodec.AccAddressPrefixed,
	newconnId int64,
	tableId int64,
) {
	appA := s.AppContext()
	// start database connection
	cmd := &CalldataDType{Connect: &ConnectRequest{Id: newconnId}}
	data, err := json.Marshal(cmd)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssc := &vmsql.SqlConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resssc)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssc.Error)

	// create table
	cmd = &CalldataDType{CreateTable: &CreateTableDTypeRequest{
		TableId: tableId,
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssc.Error)
}
