package keeper_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/mythos-tests/vmsql/testdata"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	vmsql "github.com/loredanacirstea/wasmx/x/vmsql"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

type Calldata struct {
	Connect *vmsql.SqlConnectionRequest `json:"Connect,omitempty"`
	Close   *vmsql.SqlCloseRequest      `json:"Close,omitempty"`
	Ping    *vmsql.SqlPingRequest       `json:"Ping,omitempty"`
	Execute *vmsql.SqlExecuteRequest    `json:"Execute,omitempty"`
	Query   *vmsql.SqlQueryRequest      `json:"Query,omitempty"`
}

type KV struct {
	Value string `json:"value"`
}

func (suite *KeeperTestSuite) TestSqlite() {
	wasmbin := testdata.WasmxTestSql
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "sqltest", nil)

	// connect
	cmdConn := &Calldata{Connect: &vmsql.SqlConnectionRequest{
		Driver:     "sqlite3",
		Connection: "test.db",
		Id:         "conn1",
	}}
	data, err := json.Marshal(cmdConn)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmsql.SqlConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	defer os.Remove("test.db")

	// create tables
	cmdExec := &Calldata{Execute: &vmsql.SqlExecuteRequest{
		Id:     "conn1",
		Query:  `CREATE TABLE IF NOT EXISTS kvstore (key BLOB PRIMARY KEY, value BLOB)`,
		Params: []byte{},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex := &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(0), resssex.LastInsertId)
	suite.Require().Equal("", resssex.LastInsertIdError)
	suite.Require().Equal(int64(0), resssex.RowsAffected)
	suite.Require().Equal("", resssex.RowsAffectedError)

	// create indexes
	cmdExec = &Calldata{Execute: &vmsql.SqlExecuteRequest{
		Id:     "conn1",
		Query:  `CREATE INDEX IF NOT EXISTS idx_kvstore_key ON kvstore(key)`,
		Params: []byte{},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex = &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(0), resssex.LastInsertId)
	suite.Require().Equal("", resssex.LastInsertIdError)
	suite.Require().Equal(int64(0), resssex.RowsAffected)
	suite.Require().Equal("", resssex.RowsAffectedError)

	key := []byte{2, 3}
	value := []byte{4, 5}

	// insert
	cmdExec = &Calldata{Execute: &vmsql.SqlExecuteRequest{
		Id: "conn1",
		Query: fmt.Sprintf(
			`INSERT OR REPLACE INTO kvstore(key, value) VALUES (X'%X', X'%X')`,
			key,
			value,
		),
		Params: []byte{},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex = &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(1), resssex.LastInsertId)
	suite.Require().Equal("", resssex.LastInsertIdError)
	suite.Require().Equal(int64(1), resssex.RowsAffected)
	suite.Require().Equal("", resssex.RowsAffectedError)

	// query
	cmdQuery := &Calldata{Query: &vmsql.SqlQueryRequest{
		Id:     "conn1",
		Query:  fmt.Sprintf(`SELECT value FROM kvstore WHERE key = X'%X'`, key),
		Params: []byte{},
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp := &vmsql.SqlQueryResponse{}
	err = json.Unmarshal(qres, qresp)
	suite.Require().NoError(err)
	suite.Require().Equal(qresp.Error, "")
	rows := []KV{}
	err = json.Unmarshal(qresp.Data, &rows)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(rows))
	suite.Require().Equal("\u0004\u0005", rows[0].Value)
	suite.Require().True(bytes.Equal(value, []byte(rows[0].Value)))

	// insert2
	key = []byte{1, 1, 1, 1, 1}
	value = []byte{2, 2, 2, 2, 2}
	paramsbz, err := json.Marshal(&vmsql.SqlQueryParams{Params: []vmsql.SqlQueryParam{{Type: "blob", Value: key}, {Type: "blob", Value: value}}})
	suite.Require().NoError(err)
	cmdExec = &Calldata{Execute: &vmsql.SqlExecuteRequest{
		Id:     "conn1",
		Query:  `INSERT OR REPLACE INTO kvstore(key, value) VALUES (?,?)`,
		Params: paramsbz,
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex = &vmsql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(2), resssex.LastInsertId)
	suite.Require().Equal("", resssex.LastInsertIdError)
	suite.Require().Equal(int64(1), resssex.RowsAffected)
	suite.Require().Equal("", resssex.RowsAffectedError)

	// query2
	paramsbz, err = json.Marshal(&vmsql.SqlQueryParams{Params: []vmsql.SqlQueryParam{{Type: "blob", Value: key}}})
	suite.Require().NoError(err)
	cmdQuery = &Calldata{Query: &vmsql.SqlQueryRequest{
		Id:     "conn1",
		Query:  `SELECT value FROM kvstore WHERE key = ?`,
		Params: paramsbz,
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qresp = &vmsql.SqlQueryResponse{}
	err = json.Unmarshal(qres, qresp)
	suite.Require().NoError(err)
	suite.Require().Equal(qresp.Error, "")
	rows = []KV{}
	err = json.Unmarshal(qresp.Data, &rows)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(rows))
	suite.Require().True(bytes.Equal(value, []byte(rows[0].Value)))

	// close connection
	cmdExec = &Calldata{Close: &vmsql.SqlCloseRequest{
		Id: "conn1",
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose := &vmsql.SqlCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)
}
