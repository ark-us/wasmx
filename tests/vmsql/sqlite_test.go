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
	Ping    *vmsql.SqlPingRequest       `json:"Ping,omitempty"`
	Execute *vmsql.SqlExecuteRequest    `json:"Execute,omitempty"`
	Query   *vmsql.SqlQueryRequest      `json:"Query,omitempty"`
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
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	defer os.Remove("test.db")

	// create tables
	cmdExec := &Calldata{Execute: &vmsql.SqlExecuteRequest{
		Id:    "conn1",
		Query: `CREATE TABLE IF NOT EXISTS kvstore (key BLOB PRIMARY KEY, value BLOB)`,
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// create indexes
	cmdExec = &Calldata{Execute: &vmsql.SqlExecuteRequest{
		Id:    "conn1",
		Query: `CREATE INDEX IF NOT EXISTS idx_kvstore_key ON kvstore(key)`,
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

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
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// query
	cmdQuery := &Calldata{Query: &vmsql.SqlQueryRequest{
		Id:    "conn1",
		Query: fmt.Sprintf(`SELECT value FROM kvstore WHERE key = X'%X'`, key),
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	qresp := &vmsql.SqlQueryResponse{}
	err = json.Unmarshal(qres, qresp)
	suite.Require().NoError(err)
	suite.Require().Equal(qresp.Error, "")
	suite.Require().True(bytes.Equal(qresp.Data, value))
}
