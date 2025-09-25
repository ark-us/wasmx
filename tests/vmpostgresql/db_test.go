package keeper_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/loredanacirstea/mythos-tests/vmpostgresql/testdata"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	vmpostgresql "github.com/loredanacirstea/wasmx-vmpostgresql"
	"github.com/loredanacirstea/wasmx/codec"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

type MsgNestedCall struct {
	Execute        []*vmpostgresql.SqlExecuteRequest `json:"execute"`
	Query          []*vmpostgresql.SqlQueryRequest   `json:"query"`
	IterationIndex uint32                            `json:"iteration_index"`
	RevertArray    []bool                            `json:"revert_array"`
	IsQueryArray   []bool                            `json:"isquery_array"`
}

type Calldata struct {
	Connect        *vmpostgresql.SqlConnectionRequest   `json:"Connect,omitempty"`
	Close          *vmpostgresql.SqlCloseRequest        `json:"Close,omitempty"`
	Ping           *vmpostgresql.SqlPingRequest         `json:"Ping,omitempty"`
	Execute        *vmpostgresql.SqlExecuteRequest      `json:"Execute,omitempty"`
	BatchAtomic    *vmpostgresql.SqlExecuteBatchRequest `json:"BatchAtomic,omitempty"`
	Query          *vmpostgresql.SqlQueryRequest        `json:"Query,omitempty"`
	NestedCall     *MsgNestedCall                       `json:"NestedCall,omitempty"`
	CreateDatabase *vmpostgresql.SqlCreateDatabase      `json:"CreateDatabase,omitempty"`
}

var localConn = `postgresql://localhost:5432/postgres`

// var conn0db = "postgres"
var dbname = "testpostgres"

func (suite *KeeperTestSuite) TestPostgreSqlWrapContract() {
	wasmbin := testdata.WasmxTestPostgreSql
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "sqltest", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "somerole", contractAddress, sender)

	// suite.CreateDb(appA, sender, contractAddress)

	cmdConn := &Calldata{Connect: &vmpostgresql.SqlConnectionRequest{
		Connection: localConn,
		DbName:     dbname,
		Id:         "conn1",
	}}
	data, err := json.Marshal(cmdConn)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmpostgresql.SqlConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	defer CleanupDatabase(localConn, dbname, appA.Context())

	// create tables
	cmdExec := &Calldata{Execute: &vmpostgresql.SqlExecuteRequest{
		Id:     "conn1",
		Query:  `CREATE TABLE IF NOT EXISTS kvstore (key BYTEA PRIMARY KEY, value BYTEA)`,
		Params: vmpostgresql.Params{},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex := &vmpostgresql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(0), resssex.RowsAffected)

	// create indexes
	cmdExec = &Calldata{Execute: &vmpostgresql.SqlExecuteRequest{
		Id:     "conn1",
		Query:  `CREATE INDEX IF NOT EXISTS idx_kvstore_key ON kvstore(key)`,
		Params: vmpostgresql.Params{},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex = &vmpostgresql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(0), resssex.RowsAffected)

	key := []byte{2, 3}
	value := []byte{4, 5}

	// insert
	cmdExec = &Calldata{Execute: &vmpostgresql.SqlExecuteRequest{
		Id: "conn1",
		Query: fmt.Sprintf(
			`INSERT INTO kvstore(key, value) VALUES (decode('%x', 'hex'), decode('%x', 'hex')) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`,
			key,
			value,
		),
		Params: vmpostgresql.Params{},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex = &vmpostgresql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(1), resssex.RowsAffected)

	// query
	cmdQuery := &Calldata{Query: &vmpostgresql.SqlQueryRequest{
		Id:     "conn1",
		Query:  fmt.Sprintf(`SELECT value FROM kvstore WHERE key = decode('%x', 'hex')`, key),
		Params: vmpostgresql.Params{},
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows, err := utils.ParseQueryToRows(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(rows))
	suite.Require().Equal(value, rows[0].Value)
	suite.Require().True(bytes.Equal(value, rows[0].Value))

	// insert2
	key = []byte{1, 1, 1, 1, 1}
	value = []byte{2, 2, 2, 2, 2}
	paramsbz, err := ParamsMarshal([]vmpostgresql.SqlQueryParam{{Type: "BYTEA", Value: key}, {Type: "BYTEA", Value: value}})
	suite.Require().NoError(err)
	cmdExec = &Calldata{Execute: &vmpostgresql.SqlExecuteRequest{
		Id:     "conn1",
		Query:  `INSERT INTO kvstore(key, value) VALUES ($1,$2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`,
		Params: paramsbz,
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex = &vmpostgresql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(1), resssex.RowsAffected)

	// query2
	paramsbz, err = ParamsMarshal([]vmpostgresql.SqlQueryParam{{Type: "BYTEA", Value: key}})
	suite.Require().NoError(err)
	cmdQuery = &Calldata{Query: &vmpostgresql.SqlQueryRequest{
		Id:     "conn1",
		Query:  `SELECT value FROM kvstore WHERE key = $1`,
		Params: paramsbz,
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows, err = utils.ParseQueryToRows(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(rows))
	suite.Require().True(bytes.Equal(value, rows[0].Value))

	// batch atomic
	cmdExec = &Calldata{BatchAtomic: &vmpostgresql.SqlExecuteBatchRequest{
		Id: "conn1",
		Commands: []vmpostgresql.SqlExecuteCommand{
			{
				Query:  fmt.Sprintf(`INSERT INTO kvstore(key, value) VALUES (decode('%x', 'hex'),decode('%x', 'hex')) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`, []byte{2, 2, 2}, []byte{2, 2, 3}),
				Params: [][]byte{},
			},
			{
				Query:  fmt.Sprintf(`INSERT INTO kvstore(key, value) VALUES (decode('%x', 'hex')) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`, []byte{2, 2, 2}),
				Params: [][]byte{},
			},
		},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssexb := &vmpostgresql.SqlExecuteBatchResponse{}
	err = appA.DecodeExecuteResponse(res, resssexb)
	suite.Require().NoError(err)
	suite.Require().Contains(resssexb.Error, "has more target columns than expressions")

	cmdQuery = &Calldata{Query: &vmpostgresql.SqlQueryRequest{
		Id:    "conn1",
		Query: `SELECT * FROM kvstore;`,
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows, err = utils.ParseQueryToRows(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(rows))

	// close connection
	cmdExec = &Calldata{Close: &vmpostgresql.SqlCloseRequest{
		Id: "conn1",
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose := &vmpostgresql.SqlCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)
}

func (suite *KeeperTestSuite) TestRolledBackDbCalls() {
	wasmbin := testdata.WasmxTestPostgreSql
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "sqltest", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "somerole", contractAddress, sender)

	// connect
	// suite.CreateDb(appA, sender, contractAddress)

	cmdConn := &Calldata{Connect: &vmpostgresql.SqlConnectionRequest{
		Connection: localConn,
		DbName:     dbname,
		Id:         "conn2",
	}}
	data, err := json.Marshal(cmdConn)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmpostgresql.SqlConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	defer CleanupDatabase(localConn, dbname, appA.Context())

	// create tables
	cmdExec := &Calldata{Execute: &vmpostgresql.SqlExecuteRequest{
		Id:     "conn2",
		Query:  `CREATE TABLE IF NOT EXISTS kvstore (key VARCHAR PRIMARY KEY, value VARCHAR)`,
		Params: vmpostgresql.Params{},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex := &vmpostgresql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(0), resssex.RowsAffected)

	// create indexes
	cmdExec = &Calldata{Execute: &vmpostgresql.SqlExecuteRequest{
		Id:     "conn2",
		Query:  `CREATE INDEX IF NOT EXISTS idx_kvstore_key ON kvstore(key)`,
		Params: vmpostgresql.Params{},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex = &vmpostgresql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(0), resssex.RowsAffected)

	// simple reverted call
	cmdExec = &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 0,
			RevertArray:    []bool{true},
			IsQueryArray:   []bool{},
			Execute: []*vmpostgresql.SqlExecuteRequest{
				{
					Id:     "conn2",
					Query:  `INSERT INTO kvstore(key, value) VALUES ('hello', 'alice') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`,
					Params: vmpostgresql.Params{},
				},
			},
			Query: []*vmpostgresql.SqlQueryRequest{
				{
					Id:     "conn2",
					Query:  `SELECT value FROM kvstore WHERE key = 'hello'`,
					Params: vmpostgresql.Params{},
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
	rows, err := utils.ParseQueryToRowsStr(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(rows))

	// simple query call -> rolled back db changes
	cmdExec = &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 1,
			RevertArray:    []bool{false, false},
			IsQueryArray:   []bool{true},
			Execute: []*vmpostgresql.SqlExecuteRequest{
				{
					Id:     "conn2",
					Query:  `INSERT INTO kvstore(key, value) VALUES ('hello', 'alice') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`,
					Params: vmpostgresql.Params{},
				},
				{
					Id:     "conn2",
					Query:  `INSERT INTO kvstore(key, value) VALUES ('hello2', 'alice2') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`,
					Params: vmpostgresql.Params{},
				},
			},
			Query: []*vmpostgresql.SqlQueryRequest{
				{
					Id:     "conn2",
					Query:  `SELECT value FROM kvstore WHERE key = 'hello'`,
					Params: vmpostgresql.Params{},
				},
				{
					Id:     "conn2",
					Query:  `SELECT value FROM kvstore WHERE key = 'hello2'`,
					Params: vmpostgresql.Params{},
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
	rows, err = utils.ParseQueryToRowsStr(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(rows))
	suite.Require().Equal("alice", rows[0].Value)

	// alice2 was rolled back
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[1]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows, err = utils.ParseQueryToRowsStr(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(rows))

	// close connection
	cmdExec = &Calldata{Close: &vmpostgresql.SqlCloseRequest{
		Id: "conn2",
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose := &vmpostgresql.SqlCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)
}

func (suite *KeeperTestSuite) TestNestedCalls() {
	wasmbin := testdata.WasmxTestPostgreSql
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "sqltest", nil)

	// set a role to have access to protected APIs
	utils.RegisterRole(suite, appA, "somerole", contractAddress, sender)

	// connect
	// suite.CreateDb(appA, sender, contractAddress)

	cmdConn := &Calldata{Connect: &vmpostgresql.SqlConnectionRequest{
		Connection: localConn,
		DbName:     dbname,
		Id:         "conn3",
	}}
	data, err := json.Marshal(cmdConn)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmpostgresql.SqlConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	defer CleanupDatabase(localConn, dbname, appA.Context())

	// create tables
	cmdExec := &Calldata{Execute: &vmpostgresql.SqlExecuteRequest{
		Id:     "conn3",
		Query:  `CREATE TABLE IF NOT EXISTS kvstore (key VARCHAR PRIMARY KEY, value VARCHAR)`,
		Params: vmpostgresql.Params{},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex := &vmpostgresql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(0), resssex.RowsAffected)

	// create indexes
	cmdExec = &Calldata{Execute: &vmpostgresql.SqlExecuteRequest{
		Id:     "conn3",
		Query:  `CREATE INDEX IF NOT EXISTS idx_kvstore_key ON kvstore(key)`,
		Params: vmpostgresql.Params{},
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex = &vmpostgresql.SqlExecuteResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)
	suite.Require().Equal(int64(0), resssex.RowsAffected)

	// nested call with reverted transaction
	cmdExec = &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 2,
			RevertArray:    []bool{false, true, false},
			IsQueryArray:   []bool{false, false},
			Execute: []*vmpostgresql.SqlExecuteRequest{
				{
					Id:     "conn3",
					Query:  `INSERT INTO kvstore(key, value) VALUES ('hello', 'alice') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`,
					Params: vmpostgresql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `INSERT INTO kvstore(key, value) VALUES ('mykey', 'myvalue') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`,
					Params: vmpostgresql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `INSERT INTO kvstore(key, value) VALUES ('mykey2', 'myvalue2') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`,
					Params: vmpostgresql.Params{},
				},
			},
			Query: []*vmpostgresql.SqlQueryRequest{
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = 'hello'`,
					Params: vmpostgresql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = 'mykey'`,
					Params: vmpostgresql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = 'mykey2'`,
					Params: vmpostgresql.Params{},
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
	rows, err := utils.ParseQueryToRowsStr([]byte(nestedresp[0]))
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(rows))
	suite.Require().Equal("alice", rows[0].Value)

	// alice passed
	cmdQuery := &Calldata{Query: cmdExec.NestedCall.Query[0]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows, err = utils.ParseQueryToRowsStr(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(rows))
	suite.Require().Equal("alice", rows[0].Value)

	// myvalue was rolled back
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[1]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows, err = utils.ParseQueryToRowsStr(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(rows))

	// myvalue2 was rolled back
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[2]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows, err = utils.ParseQueryToRowsStr(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(rows))

	// test nested query
	cmdExec = &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 2,
			RevertArray:    []bool{false, false, false},
			IsQueryArray:   []bool{true, false},
			Execute: []*vmpostgresql.SqlExecuteRequest{
				{
					Id:     "conn3",
					Query:  `INSERT INTO kvstore(key, value) VALUES ('hello', 'alice2') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`,
					Params: vmpostgresql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `INSERT INTO kvstore(key, value) VALUES ('mykey', 'myvalue') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`,
					Params: vmpostgresql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `INSERT INTO kvstore(key, value) VALUES ('mykey2', 'myvalue2') ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;`,
					Params: vmpostgresql.Params{},
				},
			},
			Query: []*vmpostgresql.SqlQueryRequest{
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = 'hello'`,
					Params: vmpostgresql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = 'mykey'`,
					Params: vmpostgresql.Params{},
				},
				{
					Id:     "conn3",
					Query:  `SELECT value FROM kvstore WHERE key = 'mykey2'`,
					Params: vmpostgresql.Params{},
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
	rows, err = utils.ParseQueryToRowsStr(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(rows))
	suite.Require().Equal("alice2", rows[0].Value)

	// myvalue was rolled back (query)
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[1]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows, err = utils.ParseQueryToRowsStr(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(rows))

	// myvalue2 was rolled back (query)
	cmdQuery = &Calldata{Query: cmdExec.NestedCall.Query[2]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	rows, err = utils.ParseQueryToRowsStr(qres)
	suite.Require().NoError(err)
	suite.Require().Equal(0, len(rows))

	// close connection
	cmdExec = &Calldata{Close: &vmpostgresql.SqlCloseRequest{
		Id: "conn3",
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose := &vmpostgresql.SqlCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)
}

func ParamsMarshal(params []vmpostgresql.SqlQueryParam) ([][]byte, error) {
	res := vmpostgresql.Params{}
	for _, param := range params {
		paramsbz, err := json.Marshal(&param)
		if err != nil {
			return nil, err
		}
		res = append(res, paramsbz)
	}
	return res, nil
}

func CleanupDatabase(conn string, dbname string, ctx context.Context) error {
	db, err := dbm.NewPostgreSQLDbWithCtx(ctx, "", conn, nil)
	if err != nil {
		return err
	}
	_, err = db.Pool().Exec(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS %s;`, dbname))
	if err != nil {
		return err
	}
	return db.Close()
}

// now database creation is done by the postgresql host api
func (suite *KeeperTestSuite) CreateDb(appA ut.AppContext, sender simulation.Account, contractAddress codec.AccAddressPrefixed) {
	cmdExec := &Calldata{CreateDatabase: &vmpostgresql.SqlCreateDatabase{
		Connection: localConn,
		DbName:     dbname,
	}}
	data, err := json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmpostgresql.SqlConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
}
