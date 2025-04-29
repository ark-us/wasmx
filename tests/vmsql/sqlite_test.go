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

type MsgNestedCall struct {
	Execute        []*vmsql.SqlExecuteRequest `json:"execute"`
	Query          []*vmsql.SqlQueryRequest   `json:"query"`
	IterationIndex uint32                     `json:"iteration_index"`
	RevertArray    []bool                     `json:"revert_array"`
	IsQueryArray   []bool                     `json:"isquery_array"`
}

type Calldata struct {
	Connect     *vmsql.SqlConnectionRequest   `json:"Connect,omitempty"`
	Close       *vmsql.SqlCloseRequest        `json:"Close,omitempty"`
	Ping        *vmsql.SqlPingRequest         `json:"Ping,omitempty"`
	Execute     *vmsql.SqlExecuteRequest      `json:"Execute,omitempty"`
	BatchAtomic *vmsql.SqlExecuteBatchRequest `json:"BatchAtomic,omitempty"`
	Query       *vmsql.SqlQueryRequest        `json:"Query,omitempty"`
	NestedCall  *MsgNestedCall                `json:"NestedCall,omitempty"`
}

type KV struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

type KVString struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (suite *KeeperTestSuite) TestSqliteWrapContract() {
	wasmbin := testdata.WasmxTestSql
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "sqltest", nil)

	// set a role to have access to protected APIs
	suite.registerRole("somerole", contractAddress, sender)

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
		Params: vmsql.Params{},
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
		Params: vmsql.Params{},
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
		Params: vmsql.Params{},
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
		Params: vmsql.Params{},
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows := suite.parseQueryToRows(qres)
	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(value, rows[0].Value)
	suite.Require().True(bytes.Equal(value, rows[0].Value))

	// insert2
	key = []byte{1, 1, 1, 1, 1}
	value = []byte{2, 2, 2, 2, 2}
	paramsbz, err := paramsMarshal([]vmsql.SqlQueryParam{{Type: "blob", Value: key}, {Type: "blob", Value: value}})
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
	paramsbz, err = paramsMarshal([]vmsql.SqlQueryParam{{Type: "blob", Value: key}})
	suite.Require().NoError(err)
	cmdQuery = &Calldata{Query: &vmsql.SqlQueryRequest{
		Id:     "conn1",
		Query:  `SELECT value FROM kvstore WHERE key = ?`,
		Params: paramsbz,
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows = suite.parseQueryToRows(qres)
	suite.Require().Equal(1, len(rows))
	suite.Require().True(bytes.Equal(value, rows[0].Value))

	// batch atomic
	cmdExec = &Calldata{BatchAtomic: &vmsql.SqlExecuteBatchRequest{
		Id: "conn1",
		Commands: []vmsql.SqlExecuteCommand{
			{
				Query:  fmt.Sprintf(`INSERT OR REPLACE INTO kvstore(key, value) VALUES (X'%X',X'%X')`, []byte{2, 2, 2}, []byte{2, 2, 3}),
				Params: [][]byte{},
			},
			{
				Query:  fmt.Sprintf(`INSERT OR REPLACE INTO kvstore(key, value) VALUES (X'%X')`, []byte{2, 2, 2}),
				Params: [][]byte{},
			},
		},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssexb := &vmsql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resssexb)
	suite.Require().NoError(err)
	suite.Require().Equal("1 values for 2 columns", resssexb.Error)

	cmdQuery = &Calldata{Query: &vmsql.SqlQueryRequest{
		Id:    "conn1",
		Query: `SELECT * FROM kvstore WHERE 1;`,
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows = suite.parseQueryToRows(qres)
	suite.Require().Equal(2, len(rows))

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

func (suite *KeeperTestSuite) TestRolledBackDbCalls() {
	wasmbin := testdata.WasmxTestSql
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "sqltest", nil)

	// set a role to have access to protected APIs
	suite.registerRole("somerole", contractAddress, sender)

	// connect
	cmdConn := &Calldata{Connect: &vmsql.SqlConnectionRequest{
		Driver:     "sqlite3",
		Connection: "test.db",
		Id:         "conn2",
	}}
	data, err := json.Marshal(cmdConn)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmsql.SqlConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	defer os.Remove("test.db")
	defer os.Remove("test.db-shm")
	defer os.Remove("test.db-wal")

	// create tables
	cmdExec := &Calldata{Execute: &vmsql.SqlExecuteRequest{
		Id:     "conn2",
		Query:  `CREATE TABLE IF NOT EXISTS kvstore (key VARCHAR PRIMARY KEY, value VARCHAR)`,
		Params: vmsql.Params{},
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
		Id:     "conn2",
		Query:  `CREATE INDEX IF NOT EXISTS idx_kvstore_key ON kvstore(key)`,
		Params: vmsql.Params{},
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

	// simple reverted call
	cmdExec = &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 0,
			RevertArray:    []bool{true},
			IsQueryArray:   []bool{},
			Execute: []*vmsql.SqlExecuteRequest{
				{
					Id:     "conn2",
					Query:  `INSERT OR REPLACE INTO kvstore(key, value) VALUES ("hello", "alice")`,
					Params: vmsql.Params{},
				},
			},
			Query: []*vmsql.SqlQueryRequest{
				{
					Id:     "conn2",
					Query:  `SELECT value FROM kvstore WHERE key = "hello"`,
					Params: vmsql.Params{},
				},
			},
		},
	}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res, err = appA.ExecuteContractNoCheck(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil, 3000000, nil)
	suite.Require().NoError(err)
	suite.Require().True(res.IsErr(), "tx should have reverted")

	// alice failed
	cmdQuery := &Calldata{Query: cmdExec.NestedCall.Query[0]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows := suite.parseQueryToRowsStr(qres)
	suite.Require().Equal(0, len(rows))

	// simple query call -> rolled back db changes
	cmdExec = &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 1,
			RevertArray:    []bool{false, false},
			IsQueryArray:   []bool{true},
			Execute: []*vmsql.SqlExecuteRequest{
				{
					Id:     "conn2",
					Query:  `INSERT OR REPLACE INTO kvstore(key, value) VALUES ("hello", "alice")`,
					Params: vmsql.Params{},
				},
				{
					Id:     "conn2",
					Query:  `INSERT OR REPLACE INTO kvstore(key, value) VALUES ("hello2", "alice2")`,
					Params: vmsql.Params{},
				},
			},
			Query: []*vmsql.SqlQueryRequest{
				{
					Id:     "conn2",
					Query:  `SELECT value FROM kvstore WHERE key = "hello"`,
					Params: vmsql.Params{},
				},
				{
					Id:     "conn2",
					Query:  `SELECT value FROM kvstore WHERE key = "hello2"`,
					Params: vmsql.Params{},
				},
			},
		},
	}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// alice was committed
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[0]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows = suite.parseQueryToRowsStr(qres)
	suite.Require().Equal(1, len(rows))
	suite.Require().Equal("alice", rows[0].Value)

	// alice2 was rolled back
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[1]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows = suite.parseQueryToRowsStr(qres)
	suite.Require().Equal(0, len(rows))

	// close connection
	cmdExec = &Calldata{Close: &vmsql.SqlCloseRequest{
		Id: "conn2",
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose := &vmsql.SqlCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)
}

func (suite *KeeperTestSuite) TestNestedCalls() {
	wasmbin := testdata.WasmxTestSql
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "sqltest", nil)

	// set a role to have access to protected APIs
	suite.registerRole("somerole", contractAddress, sender)

	// connect
	cmdConn := &Calldata{Connect: &vmsql.SqlConnectionRequest{
		Driver:     "sqlite3",
		Connection: "test.db",
		Id:         "conn3",
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
		Id:     "conn3",
		Query:  `CREATE TABLE IF NOT EXISTS kvstore (key VARCHAR PRIMARY KEY, value VARCHAR)`,
		Params: vmsql.Params{},
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
		Id:     "conn3",
		Query:  `CREATE INDEX IF NOT EXISTS idx_kvstore_key ON kvstore(key)`,
		Params: vmsql.Params{},
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

	// nested call with reverted transaction
	cmdExec = &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 2,
			RevertArray:    []bool{false, true, false},
			IsQueryArray:   []bool{false, false},
			Execute: []*vmsql.SqlExecuteRequest{
				{
					Id:     "conn3",
					Query:  `INSERT OR REPLACE INTO kvstore(key, value) VALUES ("hello", "alice")`,
					Params: vmsql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `INSERT OR REPLACE INTO kvstore(key, value) VALUES ("mykey", "myvalue")`,
					Params: vmsql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `INSERT OR REPLACE INTO kvstore(key, value) VALUES ("mykey2", "myvalue2")`,
					Params: vmsql.Params{},
				},
			},
			Query: []*vmsql.SqlQueryRequest{
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = "hello"`,
					Params: vmsql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = "mykey"`,
					Params: vmsql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = "mykey2"`,
					Params: vmsql.Params{},
				},
			},
		},
	}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	nestedresp := []string{}
	err = appA.DecodeExecuteResponse(res, &nestedresp)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(nestedresp))
	suite.Require().Equal("nested call must revert", nestedresp[1])
	rows := suite.parseQueryToRowsStr([]byte(nestedresp[0]))
	suite.Require().Equal(1, len(rows))
	suite.Require().Equal("alice", rows[0].Value)

	// alice passed
	cmdQuery := &Calldata{Query: cmdExec.NestedCall.Query[0]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows = suite.parseQueryToRowsStr(qres)
	suite.Require().Equal(1, len(rows))
	suite.Require().Equal("alice", rows[0].Value)

	// myvalue was rolled back
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[1]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows = suite.parseQueryToRowsStr(qres)
	suite.Require().Equal(0, len(rows))

	// myvalue2 was rolled back
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[2]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows = suite.parseQueryToRowsStr(qres)
	suite.Require().Equal(0, len(rows))

	// test nested query
	cmdExec = &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 2,
			RevertArray:    []bool{false, false, false},
			IsQueryArray:   []bool{true, false},
			Execute: []*vmsql.SqlExecuteRequest{
				{
					Id:     "conn3",
					Query:  `INSERT OR REPLACE INTO kvstore(key, value) VALUES ("hello", "alice2")`,
					Params: vmsql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `INSERT OR REPLACE INTO kvstore(key, value) VALUES ("mykey", "myvalue")`,
					Params: vmsql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `INSERT OR REPLACE INTO kvstore(key, value) VALUES ("mykey2", "myvalue2")`,
					Params: vmsql.Params{},
				},
			},
			Query: []*vmsql.SqlQueryRequest{
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = "hello"`,
					Params: vmsql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = "mykey"`,
					Params: vmsql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = "mykey2"`,
					Params: vmsql.Params{},
				},
			},
		},
	}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	err = appA.DecodeExecuteResponse(res, &nestedresp)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(nestedresp))

	// alice2 passed
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[0]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows = suite.parseQueryToRowsStr(qres)
	suite.Require().Equal(1, len(rows))
	suite.Require().Equal("alice2", rows[0].Value)

	// myvalue was rolled back (query)
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[1]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows = suite.parseQueryToRowsStr(qres)
	suite.Require().Equal(0, len(rows))

	// myvalue2 was rolled back (query)
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[2]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows = suite.parseQueryToRowsStr(qres)
	suite.Require().Equal(0, len(rows))

	// close connection
	cmdExec = &Calldata{Close: &vmsql.SqlCloseRequest{
		Id: "conn3",
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose := &vmsql.SqlCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)
}

func (suite *KeeperTestSuite) parseQueryResponse(qres []byte) *vmsql.SqlQueryResponse {
	qresp := &vmsql.SqlQueryResponse{}
	err := json.Unmarshal(qres, qresp)
	suite.Require().NoError(err)
	suite.Require().Equal(qresp.Error, "")
	return qresp
}

func (suite *KeeperTestSuite) parseQueryToRows(qres []byte) []KV {
	qresp := suite.parseQueryResponse(qres)
	rows := []KV{}
	err := json.Unmarshal(qresp.Data, &rows)
	suite.Require().NoError(err)
	return rows
}

func (suite *KeeperTestSuite) parseQueryToRowsStr(qres []byte) []KVString {
	qresp := suite.parseQueryResponse(qres)
	rows := []KVString{}
	err := json.Unmarshal(qresp.Data, &rows)
	suite.Require().NoError(err)
	return rows
}

func paramsMarshal(params []vmsql.SqlQueryParam) ([][]byte, error) {
	res := vmsql.Params{}
	for _, param := range params {
		paramsbz, err := json.Marshal(&param)
		if err != nil {
			return nil, err
		}
		res = append(res, paramsbz)
	}
	return res, nil
}
