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
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	mcodec "github.com/loredanacirstea/wasmx/codec"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/vmsql"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

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

	contractAddress, err := utils.DeployDType(suite, appA, sender)
	suite.Require().NoError(err)

	identif := utils.CreateDb(
		suite,
		appA,
		sender,
		contractAddress,
		connFile,
		connDriver,
		connName,
		dbname,
	)
	tableId1 := utils.CreateTable(
		suite,
		appA,
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

	utils.CreateFields(suite, appA, sender, contractAddress, fields)
	utils.InstantiateTable(suite, appA, sender, contractAddress, identif.DbConnectionId, tableId1)

	// create table2
	tablename2 := "newtable2"
	tableId2 := utils.CreateTable(
		suite,
		appA,
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

	utils.CreateFields(suite, appA, sender, contractAddress, fields2)
	utils.InstantiateTable(suite, appA, sender, contractAddress, identif.DbConnectionId, tableId2)

	// build json schema for table1

	// create rows in table1
	cmd := &vmsql.CalldataDType{BuildSchema: &vmsql.BuildSchemaRequest{Identifier: vmsql.TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId1,
	},
	}}
	data, err := json.Marshal(cmd)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	var schemaResp vmsql.BuildSchemaResponse
	err = json.Unmarshal(qres, &schemaResp)
	suite.Require().NoError(err, string(qres))

	// create rows in table1
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
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

	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
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
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
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

	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
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

	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
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
	cmd = &vmsql.CalldataDType{Update: &vmsql.UpdateDTypeRequest{Identifier: vmsql.TableIdentifier{
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
	cmd = &vmsql.CalldataDType{Read: &vmsql.ReadDTypeRequest{Identifier: vmsql.TableIdentifier{
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
	cmd = &vmsql.CalldataDType{ReadFields: &vmsql.ReadFieldsRequest{Identifier: vmsql.TableIdentifier{
		DbConnectionId: identif.DbConnectionId,
		DbId:           identif.DbId,
		TableId:        tableId2,
	},
		Fields: []string{"table1_id"},
		Data:   []byte(`{"id":1}`),
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
	cmd = &vmsql.CalldataDType{Read: &vmsql.ReadDTypeRequest{Identifier: vmsql.TableIdentifier{
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
	cmd = &vmsql.CalldataDType{Delete: &vmsql.DeleteDTypeRequest{Identifier: vmsql.TableIdentifier{
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
	cmd = &vmsql.CalldataDType{Read: &vmsql.ReadDTypeRequest{Identifier: vmsql.TableIdentifier{
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
	cmd = &vmsql.CalldataDType{Close: &vmsql.ConnectRequest{
		Id: identif.DbConnectionId,
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose := &vmsql.SqlCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)

	cmd = &vmsql.CalldataDType{Close: &vmsql.ConnectRequest{
		Name: vmsql.DTypeConnection,
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

	utils.DeployDType(suite, appA, sender)

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
	utils.RegisterRole(suite, appA, "erc20dtype", erc20Address, sender)

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

func (suite *KeeperTestSuite) TestDTypeIdentity() {
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

	dtypeAddress, err := utils.DeployDType(suite, appA, sender)
	suite.Require().NoError(err)
	identif := vmsql.TableIdentifier{
		DbConnectionName: vmsql.DTypeConnection,
		DbName:           vmsql.DbName,
	}

	// insert identity
	identif.TableName = vmsql.IdentityTable
	identif.TableId = vmsql.IdentityTableId
	cmd := &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{
		Identifier: identif,
		Data:       []byte(fmt.Sprintf(`{"name":"Reri Palma","address":"%s"}`, senderPrefixed.String())),
	}}
	data, err := json.Marshal(cmd)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, dtypeAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	identityId := resss.Responses[0].LastInsertId

	identif.TableName = vmsql.FullNameTable
	identif.TableId = vmsql.FullNameTableId
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{
		Identifier: identif,
		Data:       []byte(`{"title":"Ms","honorific_prefix":"Dr.","given_name":"Reri","middle_name":"","family_name":"Palma","suffix":"","postnominal":"","full_display_name":"","locale":"en-US","notes":""}`),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, dtypeAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	fullNameId := resss.Responses[0].LastInsertId

	identif.TableName = vmsql.EmailTable
	identif.TableId = vmsql.EmailTableId
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{
		Identifier: identif,
		Data:       []byte(`{"full_address":"reri.palma@gmail.com","username":"reri.palma","domain":"mail.com","provider":"Google","host":"smtp.gmail.com","category":"personal","notes":""}`),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, dtypeAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	emailId := resss.Responses[0].LastInsertId

	// insert identity node
	identif.TableName = vmsql.DTypeNodeName
	identif.TableId = vmsql.TableNodeId
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{
		Identifier: identif,
		Data:       []byte(fmt.Sprintf(`{"table_id":%d,"record_id":%d,"name":"%s"}`, vmsql.IdentityTableId, identityId, "Reri Palma")),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, dtypeAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	// identityNodeIdId := resss.Responses[0].LastInsertId

	// insert full name node
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{
		Identifier: identif,
		Data:       []byte(fmt.Sprintf(`{"table_id":%d,"record_id":%d,"name":"%s"}`, vmsql.FullNameTableId, fullNameId, "Reri Palma")),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, dtypeAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	// fullNameNodeIdId := resss.Responses[0].LastInsertId

	// insert email node
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{
		Identifier: identif,
		Data:       []byte(fmt.Sprintf(`{"table_id":%d,"record_id":%d,"name":"%s"}`, vmsql.EmailTableId, emailId, "Reri Palma")),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, dtypeAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	// emailNodeIdId := resss.Responses[0].LastInsertId

	// create relation type
	identif.TableName = vmsql.DTypeRelationTypeName
	identif.TableId = vmsql.TableRelationTypeId
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{
		Identifier: identif,
		Data:       []byte(`{"name":"identity shard","reverse_name":"full identity","reversable":true}`),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, dtypeAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
	relType := resss.Responses[0].LastInsertId

	// create relation for full name
	identif.TableName = vmsql.DTypeRelationName
	identif.TableId = vmsql.TableRelationId
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{
		Identifier: identif,
		Data:       []byte(fmt.Sprintf(`{"relation_type_id":%d,"source_node_id":%d,"target_node_id":%d,"order_index":0}`, relType, fullNameId, identityId)),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, dtypeAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))

	// create relation for email
	identif.TableName = vmsql.DTypeRelationName
	identif.TableId = vmsql.TableRelationId
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{
		Identifier: identif,
		Data:       []byte(fmt.Sprintf(`{"relation_type_id":%d,"source_node_id":%d,"target_node_id":%d,"order_index":0}`, relType, emailId, identityId)),
	}}
	data, err = json.Marshal(cmd)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, dtypeAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss = &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	suite.Require().Equal(1, len(resss.Responses))
	suite.Require().Equal("", resss.Responses[0].LastInsertIdError)
	suite.Require().Equal("", resss.Responses[0].RowsAffectedError)
	suite.Require().Greater(resss.Responses[0].LastInsertId, int64(0))
	suite.Require().Greater(resss.Responses[0].RowsAffected, int64(0))
}

func (suite *KeeperTestSuite) testGraph(
	sender simulation.Account,
	contractAddress mcodec.AccAddressPrefixed,
	tableId1 int64,
) {
	appA := s.AppContext()
	dtypeConnId := int64(1) // record id inside connection table
	dtypeDbId := int64(1)   // record id inside db table
	tableNodeId := int64(6)
	tableRelationId := int64(7)
	tableRelationTypeId := int64(8)

	// node1
	cmd := &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
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
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
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
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
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
	cmd = &vmsql.CalldataDType{Insert: &vmsql.InsertDTypeRequest{Identifier: vmsql.TableIdentifier{
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
