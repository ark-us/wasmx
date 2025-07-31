package keeper_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/mythos-tests/vmkv/testdata"
	ut "github.com/loredanacirstea/wasmx/testutil/wasmx"
	vmkv "github.com/loredanacirstea/wasmx/x/vmkv"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

type MsgNestedCall struct {
	Execute        []*vmkv.KvSetRequest `json:"execute"`
	Query          []*vmkv.KvGetRequest `json:"query"`
	IterationIndex uint32               `json:"iteration_index"`
	RevertArray    []bool               `json:"revert_array"`
	IsQueryArray   []bool               `json:"isquery_array"`
}

type Calldata struct {
	Connect    *vmkv.KvConnectionRequest `json:"Connect,omitempty"`
	Close      *vmkv.KvCloseRequest      `json:"Close,omitempty"`
	Get        *vmkv.KvGetRequest        `json:"Get,omitempty"`
	Has        *vmkv.KvHasRequest        `json:"Has,omitempty"`
	Set        *vmkv.KvSetRequest        `json:"Set,omitempty"`
	NestedCall *MsgNestedCall            `json:"NestedCall,omitempty"`
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (suite *KeeperTestSuite) TestKvWrapContract() {
	wasmbin := testdata.WasmxTestKvDB
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "kvtest", nil)

	// set a role to have access to protected APIs
	suite.registerRole("somerole", contractAddress, sender)

	// connect
	cmdConn := &Calldata{Connect: &vmkv.KvConnectionRequest{
		Driver: "goleveldb",
		Dir:    "testkv.db",
		Id:     "conn1",
	}}
	data, err := json.Marshal(cmdConn)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmkv.KvConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	defer os.RemoveAll("testkv.db")

	key := []byte{2, 3}
	value := []byte{4, 5}
	cmdExec := &Calldata{Set: &vmkv.KvSetRequest{
		Id:    "conn1",
		Key:   key,
		Value: value,
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssex := &vmkv.KvSetResponse{}
	err = appA.DecodeExecuteResponse(res, resssex)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssex.Error)

	// query
	cmdQuery := &Calldata{Get: &vmkv.KvGetRequest{
		Id:  "conn1",
		Key: key,
	}}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qress := suite.parseQueryResponse(qres)
	suite.Require().True(bytes.Equal(value, qress.Value))

	// close connection
	cmdExec = &Calldata{Close: &vmkv.KvCloseRequest{
		Id: "conn1",
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose := &vmkv.KvCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)
}

func (suite *KeeperTestSuite) TestRolledBackDbCalls() {
	SkipFixmeTests(suite.T(), "TestRolledBackDbCalls")
	wasmbin := testdata.WasmxTestKvDB
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(ut.DEFAULT_BALANCE).MulRaw(5000)

	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	suite.Commit()

	codeId := appA.StoreCode(sender, wasmbin, nil)
	contractAddress := appA.InstantiateCode(sender, codeId, types.WasmxExecutionMessage{Data: []byte{}}, "kvtest", nil)

	// set a role to have access to protected APIs
	suite.registerRole("somerole", contractAddress, sender)

	// connect
	cmdConn := &Calldata{Connect: &vmkv.KvConnectionRequest{
		Driver: "goleveldb",
		Dir:    "testkv.db",
		Id:     "conn2",
	}}
	data, err := json.Marshal(cmdConn)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmkv.KvConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	defer os.RemoveAll("testkv.db")

	key := []byte{2, 3}
	value := []byte{4, 5}

	key2 := []byte{6, 8}
	value2 := []byte{8, 8}

	// simple reverted call
	cmdExec := &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 0,
			RevertArray:    []bool{true},
			IsQueryArray:   []bool{},
			Execute: []*vmkv.KvSetRequest{
				{
					Id:    "conn2",
					Key:   key,
					Value: value,
				},
			},
			Query: []*vmkv.KvGetRequest{
				{
					Id:  "conn2",
					Key: key,
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
	cmdQuery := &Calldata{Get: cmdExec.NestedCall.Query[0]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(`{"error":"","value":"null"}`, string(qres))

	// simple query call -> rolled back db changes
	cmdExec = &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 1,
			RevertArray:    []bool{false, false},
			IsQueryArray:   []bool{true},
			Execute: []*vmkv.KvSetRequest{
				{
					Id:    "conn2",
					Key:   key,
					Value: value,
				},
				{
					Id:    "conn2",
					Key:   key2,
					Value: value2,
				},
			},
			Query: []*vmkv.KvGetRequest{
				{
					Id:  "conn2",
					Key: key,
				},
				{
					Id:  "conn2",
					Key: key2,
				},
			},
		},
	}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)

	// alice was committed
	cmdQuery = &Calldata{Get: cmdExec.NestedCall.Query[0]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qress := suite.parseQueryResponse(qres)
	suite.Require().True(bytes.Equal(value, qress.Value))

	// alice2 was rolled back
	cmdQuery = &Calldata{Get: cmdExec.NestedCall.Query[1]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(`{"error":"","value":"null"}`, string(qres))

	// close connection
	cmdExec = &Calldata{Close: &vmkv.KvCloseRequest{
		Id: "conn2",
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose := &vmkv.KvCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)
}

func (suite *KeeperTestSuite) TestNestedCalls() {
	SkipFixmeTests(suite.T(), "TestNestedCalls")
	wasmbin := testdata.WasmxTestKvDB
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
	cmdConn := &Calldata{Connect: &vmkv.KvConnectionRequest{
		Driver: "goleveldb",
		Dir:    "testkv.db",
		Id:     "conn2",
	}}
	data, err := json.Marshal(cmdConn)
	suite.Require().NoError(err)
	res := appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resss := &vmkv.KvConnectionResponse{}
	err = appA.DecodeExecuteResponse(res, resss)
	suite.Require().NoError(err)
	suite.Require().Equal("", resss.Error)
	defer os.RemoveAll("testkv.db")

	key := []byte{2, 3}
	value := []byte{4, 5}

	key2 := []byte{6, 8}
	value2 := []byte{8, 8}

	key3 := []byte{1, 1}
	value3 := []byte{9, 9}

	// nested call with reverted transaction
	cmdExec := &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 2,
			RevertArray:    []bool{false, true, false},
			IsQueryArray:   []bool{false, false},
			Execute: []*vmkv.KvSetRequest{
				{
					Id:    "conn2",
					Key:   key,
					Value: value,
				},
				{
					Id:    "conn2",
					Key:   key2,
					Value: value2,
				},
				{
					Id:    "conn2",
					Key:   key3,
					Value: value3,
				},
			},
			Query: []*vmkv.KvGetRequest{
				{
					Id:  "conn2",
					Key: key,
				},
				{
					Id:  "conn2",
					Key: key2,
				},
				{
					Id:  "conn2",
					Key: key3,
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

	// rows := suite.parseQueryToRows([]byte(nestedresp[0]))
	// suite.Require().Equal(1, len(rows))
	// suite.Require().Equal("alice", rows[0].Value)

	// alice passed
	cmdQuery := &Calldata{Get: cmdExec.NestedCall.Query[0]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres := appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qress := suite.parseQueryResponse(qres)
	suite.Require().True(bytes.Equal(value, qress.Value))

	// myvalue was rolled back
	cmdQuery = &Calldata{Get: cmdExec.NestedCall.Query[1]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(`{"error":"","value":"null"}`, string(qres))

	// myvalue2 was rolled back
	cmdQuery = &Calldata{Get: cmdExec.NestedCall.Query[2]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(`{"error":"","value":"null"}`, string(qres))

	// test nested query
	value = []byte{5, 5, 5, 5}
	cmdExec = &Calldata{
		NestedCall: &MsgNestedCall{
			IterationIndex: 2,
			RevertArray:    []bool{false, false, false},
			IsQueryArray:   []bool{true, false},
			Execute: []*vmkv.KvSetRequest{
				{
					Id:    "conn2",
					Key:   key,
					Value: value,
				},
				{
					Id:    "conn2",
					Key:   key2,
					Value: value2,
				},
				{
					Id:    "conn2",
					Key:   key3,
					Value: value3,
				},
			},
			Query: []*vmkv.KvGetRequest{
				{
					Id:  "conn2",
					Key: key,
				},
				{
					Id:  "conn2",
					Key: key2,
				},
				{
					Id:  "conn2",
					Key: key3,
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
	cmdQuery = &Calldata{Get: cmdExec.NestedCall.Query[0]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	qress = suite.parseQueryResponse(qres)
	suite.Require().True(bytes.Equal(value, qress.Value))

	// myvalue was rolled back (query)
	cmdQuery = &Calldata{Get: cmdExec.NestedCall.Query[1]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(`{"error":"","value":"null"}`, string(qres))

	// myvalue2 was rolled back (query)
	cmdQuery = &Calldata{Get: cmdExec.NestedCall.Query[2]}
	data, err = json.Marshal(cmdQuery)
	suite.Require().NoError(err)
	qres = appA.WasmxQueryRaw(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	suite.Require().Equal(`{"error":"","value":"null"}`, string(qres))

	// close connection
	cmdExec = &Calldata{Close: &vmkv.KvCloseRequest{
		Id: "conn2",
	}}
	data, err = json.Marshal(cmdExec)
	suite.Require().NoError(err)
	res = appA.ExecuteContract(sender, contractAddress, types.WasmxExecutionMessage{Data: data}, nil, nil)
	resssclose := &vmkv.KvCloseResponse{}
	err = appA.DecodeExecuteResponse(res, resssclose)
	suite.Require().NoError(err)
	suite.Require().Equal("", resssclose.Error)
}

func (suite *KeeperTestSuite) parseQueryResponse(qres []byte) *vmkv.KvGetResponse {
	qresp := &vmkv.KvGetResponse{}
	err := json.Unmarshal(qres, qresp)
	suite.Require().NoError(err)
	suite.Require().Equal(qresp.Error, "")
	return qresp
}
