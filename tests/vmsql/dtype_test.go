package keeper_test

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/loredanacirstea/mythos-tests/vmsql/testdata"
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

type ReadFieldRequest struct {
	Identifier TableIdentifier `json:"identifier"`
	FieldId    int64           `json:"fieldId"`
	FieldName  string          `json:"fieldName"`
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

type InstantiateTokens struct{}

type CalldataDType struct {
	Initialize       *InstantiateDType        `json:"Initialize,omitempty"`
	InitializeTokens *InstantiateTokens       `json:"InitializeTokens,omitempty"`
	CreateTable      *CreateTableDTypeRequest `json:"CreateTable,omitempty"`
	Connect          *ConnectRequest          `json:"Connect,omitempty"`
	Close            *ConnectRequest          `json:"Close,omitempty"`
	Insert           *InsertDTypeRequest      `json:"Insert,omitempty"`
	InsertOrReplace  *InsertDTypeRequest      `json:"InsertOrReplace,omitempty"`
	Update           *UpdateDTypeRequest      `json:"Update,omitempty"`
	Delete           *DeleteDTypeRequest      `json:"Delete,omitempty"`
	Read             *ReadDTypeRequest        `json:"Read,omitempty"`
	ReadField        *ReadFieldRequest        `json:"ReadField,omitempty"`
	BuildSchema      *BuildSchemaRequest      `json:"BuildSchema,omitempty"`
}

func (suite *KeeperTestSuite) TestDTypeContract() {
	connFile := "newdb.db"
	connDriver := "sqlite3"
	connName := "newdbconn"
	dbname := "newdb"
	tablename1 := "newtable1"
	defer os.Remove(connFile)
	defer os.Remove(connFile + "-shm")
	defer os.Remove(connFile + "-wal")

	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	contractAddress := suite.deployDType(sender)

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
	cmd := &CalldataDType{BuildSchema: &BuildSchemaRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId1,
	},
	}}
	data, err := json.Marshal(cmd)
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
	resss := &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	row1_1 := resss.Responses[0].LastInsertId

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
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	// row1_2 := resss.Responses[0].LastInsertId

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
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))

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
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))

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
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))

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
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))

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

	// readField
	cmd = &CalldataDType{ReadField: &ReadFieldRequest{Identifier: TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId2,
	},
		FieldName: "table1_id",
		Data:      []byte(`{"id":1}`),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp = suite.parseQueryResponse(qres)
	suite.Require().Equal("", qresp.Error)
	v, err := strconv.Atoi(string(qresp.Data))
	suite.Require().NoError(err)
	suite.Require().Equal(table2rows[0].Table1ID, int64(v))

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
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))

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

	suite.testGraph(sender, contractAddress, tableId1)

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

func (suite *KeeperTestSuite) TestDTypeErc20() {
	sender := suite.GetRandomAccount()
	receiver := suite.GetRandomAccount()
	spender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	receiverPrefixed := appA.BytesToAccAddressPrefixed(receiver.Address)
	appA.Faucet.Fund(appA.Context(), receiverPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	spenderPrefixed := appA.BytesToAccAddressPrefixed(spender.Address)
	appA.Faucet.Fund(appA.Context(), spenderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	wasmbin := testdata.WasmxErc20DType
	codeId := appA.StoreCode(sender, wasmbin, nil)

	suite.deployDType(sender)

	tokenInfo := struct {
		Name        string `json:"name"`
		Symbol      string `json:"symbol"`
		Decimals    int32  `json:"decimals"`
		BaseDenom   string `json:"base_denom"`
		TotalSupply string `json:"total_supply"`
	}{
		Name:        "token",
		Symbol:      "TKN",
		Decimals:    6,
		BaseDenom:   "amyt",
		TotalSupply: "1000000000",
	}
	tokenInfoBz, err := json.Marshal(&tokenInfo)
	suite.Require().NoError(err)

	erc20Address := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: tokenInfoBz}, "erc20dtype", nil)

	// set a role to have access to protected APIs
	suite.registerRole("erc20dtype", erc20Address, sender)

	// TODO remove & just trigger activate() for a role
	appA.ExecuteContractWithGas(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"instantiate":%s}`, string(tokenInfoBz)))}, nil, nil, 50000000, nil)

	qres := appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(`{"name":{}}`)}, nil, nil)
	var respName struct {
		Name string `json:"name"`
	}
	err = json.Unmarshal(qres, &respName)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("token", respName.Name)

	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(`{"symbol":{}}`)}, nil, nil)
	var respSymbol struct {
		Symbol string `json:"symbol"`
	}
	err = json.Unmarshal(qres, &respSymbol)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("TKN", respSymbol.Symbol)

	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(`{"decimals":{}}`)}, nil, nil)
	var respDecimals struct {
		Decimals int32 `json:"decimals"`
	}
	err = json.Unmarshal(qres, &respDecimals)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal(int32(6), respDecimals.Decimals)

	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(`{"totalSupply":{}}`)}, nil, nil)
	var respSupply struct {
		Supply sdk.Coin `json:"supply"`
	}
	err = json.Unmarshal(qres, &respSupply)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("1000000000", respSupply.Supply.Amount.BigInt().String())

	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, senderPrefixed.String()))}, nil, nil)
	var respBalance struct {
		Balance sdk.Coin `json:"balance"`
	}
	err = json.Unmarshal(qres, &respBalance)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("1000000000", respBalance.Balance.Amount.BigInt().String())

	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, receiverPrefixed.String()))}, nil, nil)
	err = json.Unmarshal(qres, &respBalance)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("0", respBalance.Balance.Amount.BigInt().String())

	appA.ExecuteContractWithGas(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"transfer":{"to":"%s","value":"1000"}}`, receiverPrefixed.String()))}, nil, nil, 100000000, nil)

	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, receiverPrefixed.String()))}, nil, nil)
	err = json.Unmarshal(qres, &respBalance)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("1000", respBalance.Balance.Amount.BigInt().String())

	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, senderPrefixed.String()))}, nil, nil)
	err = json.Unmarshal(qres, &respBalance)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("999999000", respBalance.Balance.Amount.BigInt().String())

	// test allowance
	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"allowance":{"owner":"%s","spender":"%s"}}`, senderPrefixed.String(), spenderPrefixed.String()))}, nil, nil)
	var respAllowance struct {
		Remaining sdkmath.Int `json:"remaining"`
	} // *big.Int
	err = json.Unmarshal(qres, &respAllowance)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("0", respAllowance.Remaining.BigInt().String())

	appA.ExecuteContractWithGas(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"approve":{"spender":"%s","value":"10000"}}`, spenderPrefixed.String()))}, nil, nil, 100000000, nil)

	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"allowance":{"owner":"%s","spender":"%s"}}`, senderPrefixed.String(), spenderPrefixed.String()))}, nil, nil)
	err = json.Unmarshal(qres, &respAllowance)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("10000", respAllowance.Remaining.BigInt().String())

	// test transferFrom
	appA.ExecuteContractWithGas(spender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"transferFrom":{"from":"%s","to":"%s","value":"1000"}}`, senderPrefixed.String(), receiverPrefixed.String()))}, nil, nil, 100000000, nil)

	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"allowance":{"owner":"%s","spender":"%s"}}`, senderPrefixed.String(), spenderPrefixed.String()))}, nil, nil)
	err = json.Unmarshal(qres, &respAllowance)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("9000", respAllowance.Remaining.BigInt().String())

	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, receiverPrefixed.String()))}, nil, nil)
	err = json.Unmarshal(qres, &respBalance)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("2000", respBalance.Balance.Amount.BigInt().String())

	qres = appA.WasmxQueryRaw(sender, erc20Address, types.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, senderPrefixed.String()))}, nil, nil)
	err = json.Unmarshal(qres, &respBalance)
	suite.Require().NoError(err, string(qres))
	suite.Require().Equal("999998000", respBalance.Balance.Amount.BigInt().String())
}

func (suite *KeeperTestSuite) deployDType(sender simulation.Account) mcodec.AccAddressPrefixed {
	appA := s.AppContext()
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

	cmd = &CalldataDType{InitializeTokens: &InstantiateTokens{}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	appA.ExecuteContractWithGas(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 50000000, nil)
	return contractAddress
}

func (suite *KeeperTestSuite) testGraph(
	sender simulation.Account,
	contractAddress mcodec.AccAddressPrefixed,
	tableId1 int64,
) {
	appA := s.AppContext()
	dtypeConnId := int64(1)
	dtypeDbId := int64(1)
	tableNodeId := int64(5)
	tableRelationId := int64(6)
	tableRelationTypeId := int64(7)

	// node1
	cmd := &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: dtypeConnId,
		DbId:           dtypeDbId,
		TableId:        tableNodeId,
	},
		Data: []byte(fmt.Sprintf(`{"table_id":%d,"record_id":%d,"name":"%s"}`, tableId1, 1, "somevalue")),
	}}
	data, err := json.Marshal(cmd)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	node1Id := resss.Responses[0].LastInsertId

	// node 2
	cmd = &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: dtypeConnId,
		DbId:           dtypeDbId,
		TableId:        tableNodeId,
	},
		Data: []byte(fmt.Sprintf(`{"table_id":%d,"record_id":%d,"name":"%s"}`, tableId1, 2, "somevalue2")),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	node2Id := resss.Responses[0].LastInsertId

	// create relation type
	cmd = &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: dtypeConnId,
		DbId:           dtypeDbId,
		TableId:        tableRelationTypeId,
	},
		Data: []byte(`{"name":"relation1","reverse_name":"relation1reverse","reversable":true})`),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	relType1 := resss.Responses[0].LastInsertId

	// create relation
	cmd = &CalldataDType{Insert: &InsertDTypeRequest{Identifier: TableIdentifier{
		DbConnectionId: dtypeConnId,
		DbId:           dtypeDbId,
		TableId:        tableRelationId,
	},
		Data: []byte(fmt.Sprintf(`{"relation_type_id":%d,"source_node_id":%d,"target_node_id":%d,"order_index":0})`, relType1, node1Id, node2Id)),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	// relId := resss.Responses[0].LastInsertId
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
	resss := &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	newconnId := resss.Responses[0].LastInsertId

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
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	newDbId := resss.Responses[0].LastInsertId

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
	resss := &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	tableId := resss.Responses[0].LastInsertId

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
		resss := &vmsql.SqlExecuteBatchResponse{}
		err = appA.DecodeExecuteResponse(res, resss)
		suite.Require().NoError(err)
		suite.Require().Equal("", resss.Error)
		suite.Require().Equal(1, len(resss.Responses))
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
