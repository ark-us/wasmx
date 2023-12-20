package keeper_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// ibctesting "mythos/v1/testutil/ibc"

	app "mythos/v1/app"
	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

var tstoreprefix = []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 40}
var bzkey = []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 40, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

var DefaultTarget = "bufnet"

type LogEntry struct {
	Index    int64  `json:"index"`
	TermId   int32  `json:"termId"`
	LeaderId int32  `json:"leaderId"`
	Data     string `json:"data"`
	Result   string `json:"result"`
}

type AppendEntry struct {
	TermId       int32      `json:"termId"`
	LeaderId     int32      `json:"leaderId"`
	PrevLogIndex int64      `json:"prevLogIndex"`
	PrevLogTerm  int32      `json:"prevLogTerm"`
	Entries      []LogEntry `json:"entries"`
	LeaderCommit int64      `json:"leaderCommit"`
}

func (suite *KeeperTestSuite) TestRAFTLogReplicationOneNode() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	appA := s.GetAppContext(suite.chainA)
	mapp, ok := suite.chainA.App.(*app.App)
	suite.Require().True(ok)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	// goctx1 := context.Background()
	// client1, conn1 := suite.GrpcClient(goctx1, "bufnet1", mapp)
	// defer conn1.Close()
	consensusBech32 := "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy"
	consensusAddr := sdk.MustAccAddressFromBech32(consensusBech32)

	initChainSetupStr := `{"run":{"event":{"type":"setupNode","params":[{"key":"currentNodeId","value":"0"},{"key":"nodeIPs","value":"[\"0.0.0.0:8090\"]"},{"key":"initChainSetup","value":"eyJjaGFpbl9pZCI6Im15dGhvc183MDAwLTE0IiwiY29uc2Vuc3VzX3BhcmFtcyI6eyJibG9jayI6eyJtYXhfYnl0ZXMiOjIyMDIwMDk2LCJtYXhfZ2FzIjotMX0sImV2aWRlbmNlIjp7Im1heF9hZ2VfbnVtX2Jsb2NrcyI6MTAwMDAwLCJtYXhfYWdlX2R1cmF0aW9uIjoxNzI4MDAwMDAwMDAwMDAsIm1heF9ieXRlcyI6MTA0ODU3Nn0sInZhbGlkYXRvciI6eyJwdWJfa2V5X3R5cGVzIjpbImVkMjU1MTkiXX0sInZlcnNpb24iOnsiYXBwIjowfSwiYWJjaSI6eyJ2b3RlX2V4dGVuc2lvbnNfZW5hYmxlX2hlaWdodCI6MH19LCJ2YWxpZGF0b3JzIjpbeyJhZGRyZXNzIjoiM0M2MUFCNDI5NkEyMUM1REZFRDRFNURBQTM0MTkxMDM3OUYxODUxNiIsInB1Yl9rZXkiOiJJKzNDZ05RZTNvQzE0RGd2TkdNbGcrR1Z4L3lhMWQ4dEhnTlBEUTQxRCtBPSIsInZvdGluZ19wb3dlciI6MTAwMDAwMDAwMDAwMDAwLCJwcm9wb3Nlcl9wcmlvcml0eSI6MH0seyJhZGRyZXNzIjoiMjZDQTYxQjM4MEMyOTc5M0NGQjNCQTFFN0I3ODg3NDMyMjJCRUY1OSIsInB1Yl9rZXkiOiIyQzd5RzJvTkdHM3liMTQzYXpmNGRodWJGTTBGYXhURjlpTElDNXdVYmZ3PSIsInZvdGluZ19wb3dlciI6MTAwMDAwMDAwMDAwMDAwLCJwcm9wb3Nlcl9wcmlvcml0eSI6MH1dLCJhcHBfaGFzaCI6IjQ3REVRcGo4SEJTYSsvVEltVys1SkNldVFlUmttNU5NcEpXWkczaFN1RlU9IiwibGFzdF9yZXN1bHRzX2hhc2giOiI0N0RFUXBqOEhCU2ErL1RJbVcrNUpDZXVRZVJrbTVOTXBKV1pHM2hTdUZVPSIsImN1cnJlbnRfdmFsaWRhdG9yIjoiM0M2MUFCNDI5NkEyMUM1REZFRDRFNURBQTM0MTkxMDM3OUYxODUxNiIsInZlcnNpb24iOnsiY29uc2Vuc3VzIjp7ImJsb2NrIjowLCJhcHAiOjB9LCJzb2Z0d2FyZSI6IiJ9fQ=="}]}}}`
	msg1 := []byte(initChainSetupStr)

	resp, err := mapp.NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	// Check each simulated node has the correct context:
	msg1 = []byte(`{"getContextValue":{"key":"nodeIPs"}}`)
	qresp, err := mapp.NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz := appA.QueryDecode(qresp.Data)
	suite.Require().Equal(string(qrespbz), "[\"0.0.0.0:8090\"]")

	msg1 = []byte(`{"getContextValue":{"key":"currentNodeId"}}`)
	qresp, err = mapp.NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(string(qrespbz), `0`)

	msg1 = []byte(`{"getCurrentState":{}}`)
	qresp, err = mapp.NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(`#RAFT-FULL-1.initialized.unstarted`, string(qrespbz))

	// Start Leader
	msg1 = []byte(`{"run":{"event": {"type": "start", "params": []}}}`)
	resp, err = mapp.NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	msg1 = []byte(`{"getCurrentState":{}}`)
	qresp, err = mapp.NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(`#RAFT-FULL-1.initialized.Follower`, string(qrespbz))

	msg1 = []byte(`{"delay":"electionTimeout","state":"#RAFT-FULL-1.initialized.Follower","intervalId":1}`)
	respbz, err := appA.App.WasmxKeeper.ExecuteEventual(appA.Context(), consensusAddr, consensusAddr, msg1, nil)
	suite.Require().NoError(err)
	log.Printf("Response: %+v", string(respbz))

	// resp, err = client1.ExecuteContract(goctx1, &types.MsgExecuteContract{
	// 	Sender:   consensusBech32,
	// 	Contract: consensusBech32,
	// 	Msg:      msg1,
	// })
	// suite.Require().NoError(err)
	// log.Printf("Response: %+v", resp)

	msg1 = []byte(`{"getCurrentState":{}}`)
	qresp, err = mapp.NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(`#RAFT-FULL-1.initialized.Candidate`, string(qrespbz))

	// send tx
	contractAddress := wasmxtypes.AccAddressFromHex("0x0000000000000000000000000000000000000004")
	internalmsg := wasmxtypes.WasmxExecutionMessage{Data: appA.Hex2bz("aa0000000000000000000000000000000000000000000000000000000077")}
	msgbz, err := json.Marshal(internalmsg)
	suite.Require().NoError(err)
	msg := &wasmxtypes.MsgExecuteContract{
		Sender:       sender.Address.String(),
		Contract:     contractAddress.String(),
		Msg:          msgbz,
		Funds:        nil,
		Dependencies: nil,
	}
	tx := appA.PrepareCosmosTx(sender, []sdk.Msg{msg}, nil, nil)
	txstr := base64.StdEncoding.EncodeToString(tx)

	msg1 = []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, txstr))
	resp, err = mapp.NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	// send a second tx!
	msg = &wasmxtypes.MsgExecuteContract{
		Sender:       sender2.Address.String(),
		Contract:     contractAddress.String(),
		Msg:          msgbz,
		Funds:        nil,
		Dependencies: nil,
	}
	tx = appA.PrepareCosmosTx(sender2, []sdk.Msg{msg}, nil, nil)
	txstr = base64.StdEncoding.EncodeToString(tx)

	msg1 = []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, txstr))
	resp, err = mapp.NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	msg1 = []byte(`{"run":{"event": {"type": "start", "params": []}}}`)
	resp, err = mapp.NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	msg1 = []byte(`{"getContextValue":{"key":"logs_count"}}`)
	qresp, err = mapp.NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(`2`, string(qrespbz))

	time.Sleep(10 * time.Second)
}

// func (suite *KeeperTestSuite) TestRAFTLogReplication() {
// 	chainIds := []string{"mythos_7001-1"}
// 	coordinator1 := ibctesting.NewCoordinator(suite.T(), chainIds, 0)
// 	chain1 := coordinator1.GetChain(chainIds[0])
// 	app1, ok := chain1.App.(*app.App)
// 	suite.Require().True(ok)

// 	coordinator2 := ibctesting.NewCoordinator(suite.T(), chainIds, 1)
// 	chain2 := coordinator2.GetChain(chainIds[0])
// 	app2, ok := chain2.App.(*app.App)
// 	suite.Require().True(ok)

// 	sender := suite.GetRandomAccount()
// 	sender2 := suite.GetRandomAccount()
// 	initBalance := sdkmath.NewInt(1000_000_000)
// 	appA := s.GetAppContext(chain1)
// 	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
// 	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
// 	suite.Commit()

// 	// ip1 := "tcp://localhost:8090"
// 	// ip2 := "tcp://localhost:8091"
// 	ip1 := "0.0.0.0:8090"
// 	ip2 := "0.0.0.0:8091"
// 	// ip1 := "bufnet1"
// 	// ip2 := "bufnet2"
// 	goctx1 := context.Background()
// 	goctx2 := context.Background()
// 	client1, conn1 := suite.GrpcClient(goctx1, ip1, app1)
// 	client2, conn2 := suite.GrpcClient(goctx2, ip2, app2)
// 	defer conn1.Close()
// 	defer conn2.Close()
// 	consensusBech32 := "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpfqnvljy"
// 	// contractAddress := sdk.MustAccAddressFromBech32(consensusBech32)

// 	initChainSetup := "eyJjaGFpbl9pZCI6Im15dGhvc183MDAwLTE0IiwiY29uc2Vuc3VzX3BhcmFtcyI6eyJibG9jayI6eyJtYXhfYnl0ZXMiOjIyMDIwMDk2LCJtYXhfZ2FzIjotMX0sImV2aWRlbmNlIjp7Im1heF9hZ2VfbnVtX2Jsb2NrcyI6MTAwMDAwLCJtYXhfYWdlX2R1cmF0aW9uIjoxNzI4MDAwMDAwMDAwMDAsIm1heF9ieXRlcyI6MTA0ODU3Nn0sInZhbGlkYXRvciI6eyJwdWJfa2V5X3R5cGVzIjpbImVkMjU1MTkiXX0sInZlcnNpb24iOnsiYXBwIjowfSwiYWJjaSI6eyJ2b3RlX2V4dGVuc2lvbnNfZW5hYmxlX2hlaWdodCI6MH19LCJ2YWxpZGF0b3JzIjpbeyJhZGRyZXNzIjoiODRFMjhCRThGODk4QzE5NTMxOEU1QTg0RDMxRERBMzU5NzlFMUMwQSIsInB1Yl9rZXkiOiI4aERqZkJqc1c4d3k3Z2MzVlNSZk9LMnVoTlljaWJIdDdHbkI3a2p3KytjPSIsInZvdGluZ19wb3dlciI6MTAwMDAwMDAwMDAwMDAwLCJwcm9wb3Nlcl9wcmlvcml0eSI6MH1dLCJhcHBfaGFzaCI6IjQ3REVRcGo4SEJTYSsvVEltVys1SkNldVFlUmttNU5NcEpXWkczaFN1RlU9IiwibGFzdF9yZXN1bHRzX2hhc2giOiI0N0RFUXBqOEhCU2ErL1RJbVcrNUpDZXVRZVJrbTVOTXBKV1pHM2hTdUZVPSIsImN1cnJlbnRfdmFsaWRhdG9yIjoiODRFMjhCRThGODk4QzE5NTMxOEU1QTg0RDMxRERBMzU5NzlFMUMwQSIsInZlcnNpb24iOnsiY29uc2Vuc3VzIjp7ImJsb2NrIjowLCJhcHAiOjB9LCJzb2Z0d2FyZSI6IiJ9fQ"
// 	// await runFnOwner("run", machine, {event: {type: "setupNode", params: [{key: "currentNodeId", value: "0"},{key: "nodeIPs", value: "[\"0.0.0.0:8090\"]"},{key: "initChainSetup", value: initChainSetup2}]}});
// 	msg1 := []byte(fmt.Sprintf(`{"run":{"event":{"type":"setupNode","params":[{"key":"currentNodeId","value":"0"},{"key":"nodeIPs","value":"[\"%s\",\"%s\"]"},{"key":"initChainSetup","value":"%s"}]}}}`, ip1, ip2, initChainSetup))
// 	resp, err := mapp.NetworkKeeper.ExecuteContract(goctx1, &types.MsgExecuteContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	log.Printf("Response: %+v", resp)

// 	msg2 := []byte(fmt.Sprintf(`{"run":{"event":{"type":"setupNode","params":[{"key":"currentNodeId","value":"1"},{"key":"nodeIPs","value":"[\"%s\",\"%s\"]"}]}}}`, ip1, ip2))
// 	resp, err = mapp.NetworkKeeper.ExecuteContract(goctx2, &types.MsgExecuteContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg2,
// 	})
// 	suite.Require().NoError(err)
// 	log.Printf("Response: %+v", resp)

// 	// Check each simulated node has the correct context:

// 	// Node 1
// 	msg1 = []byte(`{"getContextValue":{"key":"nodeIPs"}}`)
// 	qresp, err := mapp.NetworkKeeper.QueryContract(goctx1, &types.MsgQueryContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	mapp := s.GetAppContext(chain1)
// 	qrespbz := mapp.QueryDecode(qresp.Data)
// 	suite.Require().Equal(string(qrespbz), fmt.Sprintf(`["%s","%s"]`, ip1, ip2))

// 	msg1 = []byte(`{"getContextValue":{"key":"currentNodeId"}}`)
// 	qresp, err = mapp.NetworkKeeper.QueryContract(goctx1, &types.MsgQueryContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	qrespbz = mapp.QueryDecode(qresp.Data)
// 	suite.Require().Equal(string(qrespbz), `0`)

// 	msg1 = []byte(`{"getCurrentState":{}}`)
// 	qresp, err = mapp.NetworkKeeper.QueryContract(goctx1, &types.MsgQueryContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	mapp = s.GetAppContext(chain1)
// 	qrespbz = mapp.QueryDecode(qresp.Data)
// 	suite.Require().Equal(`Follower`, string(qrespbz))

// 	// Node 2
// 	msg2 = []byte(`{"getContextValue":{"key":"nodeIPs"}}`)
// 	qresp, err = client2.QueryContract(goctx2, &types.MsgQueryContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg2,
// 	})
// 	suite.Require().NoError(err)
// 	qrespbz = mapp.QueryDecode(qresp.Data)
// 	suite.Require().Equal(string(qrespbz), fmt.Sprintf(`["%s","%s"]`, ip1, ip2))

// 	msg2 = []byte(`{"getContextValue":{"key":"currentNodeId"}}`)
// 	qresp, err = client2.QueryContract(goctx2, &types.MsgQueryContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg2,
// 	})
// 	suite.Require().NoError(err)
// 	qrespbz = mapp.QueryDecode(qresp.Data)
// 	suite.Require().Equal(string(qrespbz), `1`)

// 	msg2 = []byte(`{"getCurrentState":{}}`)
// 	qresp, err = client2.QueryContract(goctx2, &types.MsgQueryContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg2,
// 	})
// 	suite.Require().NoError(err)
// 	qrespbz = mapp.QueryDecode(qresp.Data)
// 	suite.Require().Equal(`Follower`, string(qrespbz))

// 	// Start Leader

// 	// Node 1
// 	msg1 = []byte(`{"run":{"event": {"type": "change", "params": []}}}`)
// 	resp, err = client1.ExecuteContract(goctx1, &types.MsgExecuteContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	log.Printf("Response: %+v", resp)

// 	msg1 = []byte(`{"getCurrentState":{}}`)
// 	qresp, err = client1.QueryContract(goctx1, &types.MsgQueryContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	mapp = s.GetAppContext(chain1)
// 	qrespbz = mapp.QueryDecode(qresp.Data)
// 	suite.Require().Equal(`Candidate`, string(qrespbz))

// 	msg1 = []byte(`{"run":{"event": {"type": "change", "params": []}}}`)
// 	resp, err = client1.ExecuteContract(goctx1, &types.MsgExecuteContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	log.Printf("Response: %+v", resp)

// 	msg1 = []byte(`{"getCurrentState":{}}`)
// 	qresp, err = client1.QueryContract(goctx1, &types.MsgQueryContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	mapp = s.GetAppContext(chain1)
// 	qrespbz = mapp.QueryDecode(qresp.Data)
// 	suite.Require().Equal(`active`, string(qrespbz))

// 	// send tx
// 	contractAddress := wasmxtypes.AccAddressFromHex("0x0000000000000000000000000000000000000004")
// 	internalmsg := wasmxtypes.WasmxExecutionMessage{Data: appA.Hex2bz("aa0000000000000000000000000000000000000000000000000000000077")}
// 	msgbz, err := json.Marshal(internalmsg)
// 	suite.Require().NoError(err)
// 	msg := &wasmxtypes.MsgExecuteContract{
// 		Sender:       sender.Address.String(),
// 		Contract:     contractAddress.String(),
// 		Msg:          msgbz,
// 		Funds:        nil,
// 		Dependencies: nil,
// 	}
// 	tx := mapp.PrepareCosmosTx(sender, []sdk.Msg{msg}, nil, nil)
// 	txstr := base64.StdEncoding.EncodeToString(tx)

// 	msg1 = []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, txstr))
// 	resp, err = client1.ExecuteContract(goctx1, &types.MsgExecuteContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	log.Printf("Response: %+v", resp)

// 	// send a second tx!
// 	msg = &wasmxtypes.MsgExecuteContract{
// 		Sender:       sender2.Address.String(),
// 		Contract:     contractAddress.String(),
// 		Msg:          msgbz,
// 		Funds:        nil,
// 		Dependencies: nil,
// 	}
// 	tx = mapp.PrepareCosmosTx(sender2, []sdk.Msg{msg}, nil, nil)
// 	txstr = base64.StdEncoding.EncodeToString(tx)

// 	msg1 = []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, txstr))
// 	resp, err = client1.ExecuteContract(goctx1, &types.MsgExecuteContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	log.Printf("Response: %+v", resp)

// 	msg1 = []byte(`{"getContextValue":{"key":"logs_count"}}`)
// 	qresp, err = client1.QueryContract(goctx1, &types.MsgQueryContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	qrespbz = mapp.QueryDecode(qresp.Data)
// 	suite.Require().Equal(`2`, string(qrespbz))

// 	msg1 = []byte(fmt.Sprintf(`{"getContextValue":{"key":"logs_%s"}}`, string(qrespbz)))
// 	qresp, err = client1.QueryContract(goctx1, &types.MsgQueryContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg1,
// 	})
// 	suite.Require().NoError(err)
// 	logEntrybz := mapp.QueryDecode(qresp.Data)

// 	var logEntry LogEntry
// 	err = json.Unmarshal(logEntrybz, &logEntry)
// 	suite.Require().NoError(err)

// 	entry := AppendEntry{
// 		TermId:       1,
// 		LeaderId:     1,
// 		PrevLogIndex: 0,
// 		PrevLogTerm:  0,
// 		Entries:      []LogEntry{logEntry},
// 		LeaderCommit: 2, // we finalize 1 block // TODO this should be 1
// 	}
// 	entrybz, err := json.Marshal(entry)
// 	suite.Require().NoError(err)

// 	// receiveHeartbeat
// 	respReceive, err := client2.GrpcReceiveRequest(goctx2, &types.MsgGrpcReceiveRequest{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Data:     entrybz,
// 	})
// 	suite.Require().NoError(err)
// 	log.Printf("Response: %+v", respReceive)

// 	msg2 = []byte(`{"getContextValue":{"key":"logs_count"}}`)
// 	qresp, err = client2.QueryContract(goctx2, &types.MsgQueryContract{
// 		Sender:   consensusBech32,
// 		Contract: consensusBech32,
// 		Msg:      msg2,
// 	})
// 	suite.Require().NoError(err)
// 	qrespbz = mapp.QueryDecode(qresp.Data)
// 	suite.Require().Equal(`2`, string(qrespbz))

// 	// Test finalize block

// 	time.Sleep(10 * time.Second)
// }

func (suite *KeeperTestSuite) TestRAFTEncodeTx() {
	mnemonic := "enrich nose brisk lobster room large uniform mystery crush govern lazy vital feed dove soul emotion oblige shuffle else entry trend there dentist garbage"
	sender := suite.GetAccountFromMnemonic(mnemonic)
	initBalance := sdkmath.NewInt(1000_000_000)
	appA := s.GetAppContext(suite.chainA)
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	// send tx
	contractAddress := wasmxtypes.AccAddressFromHex("0x0000000000000000000000000000000000000004")
	internalmsg := wasmxtypes.WasmxExecutionMessage{Data: appA.Hex2bz("aa0000000000000000000000000000000000000000000000000000000077")}
	msgbz, err := json.Marshal(internalmsg)
	suite.Require().NoError(err)
	msg := &wasmxtypes.MsgExecuteContract{
		Sender:       sender.Address.String(),
		Contract:     contractAddress.String(),
		Msg:          msgbz,
		Funds:        nil,
		Dependencies: nil,
	}
	tx := appA.PrepareCosmosTx(sender, []sdk.Msg{msg}, nil, nil)
	txstr := base64.StdEncoding.EncodeToString(tx)

	msgbase64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, txstr)))
	fmt.Println("==msgbase64==", msgbase64)
}
