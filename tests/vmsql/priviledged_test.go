package keeper_test

import (
	_ "embed"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/mythos-tests/vmsql/testdata"
	"github.com/loredanacirstea/mythos-tests/vmsql/utils"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	vmsql "github.com/loredanacirstea/wasmx/x/vmsql"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestPriviledgedAPI() {
	// test that contracts without a role requiring a protected host API, only use the mocked host API
	wasmbin := testdata.WasmxTestSql
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	// test that a contract with a priviledged API, without the proper role is executed using the Mocked API
	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "sqltest", nil)

	cmdQuery := &Calldata{Query: &vmsql.SqlQueryRequest{
		Id:     "conn1",
		Query:  `SELECT value FROM kvstore WHERE key = "hello"`,
		Params: vmsql.Params{},
	}}
	data, err := json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	// response from mocked host
	suite.Require().Equal(`{"error":"","data":"null"}`, string(qres))

	// now set a role for this priviledged contract
	utils.RegisterRole(suite, appA, "somerole", contractAddress, sender)

	cmdQuery = &Calldata{Query: &vmsql.SqlQueryRequest{
		Id:     "conn1",
		Query:  `SELECT value FROM kvstore WHERE key = "hello"`,
		Params: vmsql.Params{},
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(`{"error":"sql connection not found","data":"null"}`, string(qres))
}
