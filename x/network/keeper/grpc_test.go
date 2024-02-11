package keeper_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	// ibctesting "mythos/v1/testutil/ibc"

	"mythos/v1/x/network/types"
	wasmxkeeper "mythos/v1/x/wasmx/keeper"
	wasmxtypes "mythos/v1/x/wasmx/types"
	precompiles "mythos/v1/x/wasmx/vm/precompiles"
)

var tstoreprefix = []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 40}
var bzkey = []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 40, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

var DefaultTarget = "bufnet"

type AppendEntry struct {
	TermId       int32            `json:"termId"`
	LeaderId     int32            `json:"leaderId"`
	PrevLogIndex int64            `json:"prevLogIndex"`
	PrevLogTerm  int32            `json:"prevLogTerm"`
	Entries      []types.LogEntry `json:"entries"`
	LeaderCommit int64            `json:"leaderCommit"`
}

func (suite *KeeperTestSuite) TestRAFTLogReplicationOneNode() {
	sender := suite.GetAccountFromMnemonic("work reward tooth zero chimney employ list grass priority pudding laundry crystal spend response bomb skill engine head science onion spider upper slab volume")
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	_, err := appA.App.AccountKeeper.GetSequence(appA.Context(), sender.Address)
	suite.Require().NoError(err)

	// goctx1 := context.Background()
	// client1, conn1 := suite.GrpcClient(goctx1, "bufnet1", mapp)
	// defer conn1.Close()

	consensusContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_CONSENSUS_RAFT)
	consensusBech32 := consensusContract.String()

	// storageContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_STORAGE_CHAIN)
	validatorAddr := sdk.AccAddress(suite.Chain().Vals.Validators[0].PubKey.Bytes()).String()
	initChainSetup := []byte(fmt.Sprintf(`{"chain_id":"mythos_7000-14","consensus_params":{"block":{"max_bytes":22020096,"max_gas":-1},"evidence":{"max_age_num_blocks":100000,"max_age_duration":172800000000000,"max_bytes":1048576},"validator":{"pub_key_types":["ed25519"]},"version":{"app":0},"abci":{"vote_extensions_enable_height":0}},"validators":[{"address":"467F6127246A6E40B59899258DF08F857145B9CB","pub_key":"shBx7GuXCf7T+HwGwffE93xWOCkIwzPpp/oKkMq3hqw=","voting_power":100000000000000,"proposer_priority":0}],"app_hash":"47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=","last_results_hash":"47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=","version":{"consensus":{"block":0,"app":0},"software":""},"validator_address":"467F6127246A6E40B59899258DF08F857145B9CB","validator_privkey":"LdBVBItkqjNrSqwDaFgxZaO7n8rN01dJ6I3BQ/9LTTyyEHHsa5cJ/tP4fAbB98T3fFY4KQjDM+mn+gqQyreGrA==","validator_pubkey":"shBx7GuXCf7T+HwGwffE93xWOCkIwzPpp/oKkMq3hqw=","peers":["%s@0.0.0.0:8090"]}`, validatorAddr))

	vals, err := suite.App().CosmosmodKeeper.GetAllValidators(appA.Context())
	suite.Require().NoError(err)
	s.Require().Equal(1, len(vals))

	msg1 := []byte(fmt.Sprintf(`{"run":{"event":{"type":"setupNode","params":[{"key":"currentNodeId","value":"0"},{"key":"nodeIPs","value":"[\"%s@0.0.0.0:8090\"]"},{"key":"initChainSetup","value":"%s"}]}}}`, validatorAddr, base64.StdEncoding.EncodeToString(initChainSetup)))
	resp, err := suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	_, err = appA.App.AccountKeeper.GetSequence(appA.Context(), sender.Address)
	suite.Require().NoError(err)

	// Check each simulated node has the correct context:
	msg1 = []byte(`{"getContextValue":{"key":"nodeIPs"}}`)
	qresp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz := appA.QueryDecode(qresp.Data)
	suite.Require().Equal(fmt.Sprintf(`[{"address":"%s","ip":"0.0.0.0:8090"}]`, validatorAddr), string(qrespbz))

	msg1 = []byte(`{"getContextValue":{"key":"currentNodeId"}}`)
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(string(qrespbz), `0`)

	msg1 = []byte(`{"getCurrentState":{}}`)
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(`#RAFT-FULL-1.initialized.unstarted`, string(qrespbz))

	// Start Leader
	msg1 = []byte(`{"run":{"event": {"type": "start", "params": []}}}`)
	resp, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	msg1 = []byte(`{"getCurrentState":{}}`)
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(`#RAFT-FULL-1.initialized.Follower`, string(qrespbz))

	_, err = appA.App.AccountKeeper.GetSequence(appA.Context(), sender.Address)
	suite.Require().NoError(err)

	_, err = appA.App.AccountKeeper.GetSequence(appA.Context(), sender.Address)
	suite.Require().NoError(err)

	msg1 = []byte(`{"delay":"electionTimeout","state":"#RAFT-FULL-1.initialized.Follower","intervalId":"1"}`)

	respbz, err := appA.App.NetworkKeeper.ExecuteEventual(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", string(respbz.Data))

	// resp, err = client1.ExecuteContract(goctx1, &types.MsgExecuteContract{
	// 	Sender:   consensusBech32,
	// 	Contract: consensusBech32,
	// 	Msg:      msg1,
	// })
	// suite.Require().NoError(err)
	// log.Printf("Response: %+v", resp)

	_, err = appA.App.AccountKeeper.GetSequence(appA.Context(), sender.Address)
	suite.Require().NoError(err)

	msg1 = []byte(`{"getCurrentState":{}}`)
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(`#RAFT-FULL-1.initialized.Leader.active`, string(qrespbz))

	// send tx
	contractAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_IDENTITY)
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
	_, err = appA.App.AccountKeeper.GetSequence(appA.Context(), sender.Address)
	suite.Require().NoError(err)

	tx := appA.PrepareCosmosTx(sender, []sdk.Msg{msg}, nil, nil)
	txstr := base64.StdEncoding.EncodeToString(tx)

	msg1 = []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, txstr))
	resp, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
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
	resp, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	msg1 = []byte(`{"run":{"event": {"type": "start", "params": []}}}`)
	resp, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	msg1 = []byte(`{"getContextValue":{"key":"logs_count"}}`)
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	// TODO commit the txs
	suite.Require().Equal(``, string(qrespbz))
	log.Print("ending test")
	time.Sleep(10 * time.Second)
}

func (suite *KeeperTestSuite) TestRAFTMigration() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	appA := s.AppContext()
	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}

	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), valAccount.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	consensusContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_CONSENSUS_RAFT)
	consensusBech32 := consensusContract.String()

	storageContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_STORAGE_CHAIN)
	initChainSetup := []byte(fmt.Sprintf(`{"chain_id":"mythos_7000-14","consensus_params":{"block":{"max_bytes":22020096,"max_gas":-1},"evidence":{"max_age_num_blocks":100000,"max_age_duration":172800000000000,"max_bytes":1048576},"validator":{"pub_key_types":["ed25519"]},"version":{"app":0},"abci":{"vote_extensions_enable_height":0}},"validators":[{"address":"467F6127246A6E40B59899258DF08F857145B9CB","pub_key":"shBx7GuXCf7T+HwGwffE93xWOCkIwzPpp/oKkMq3hqw=","voting_power":100000000000000,"proposer_priority":0}],"app_hash":"47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=","last_results_hash":"47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=","version":{"consensus":{"block":0,"app":0},"software":""},"validator_address":"467F6127246A6E40B59899258DF08F857145B9CB","validator_privkey":"LdBVBItkqjNrSqwDaFgxZaO7n8rN01dJ6I3BQ/9LTTyyEHHsa5cJ/tP4fAbB98T3fFY4KQjDM+mn+gqQyreGrA==","validator_pubkey":"shBx7GuXCf7T+HwGwffE93xWOCkIwzPpp/oKkMq3hqw=","wasmx_blocks_contract":"%s"}`, storageContract.String()))

	msg1 := []byte(fmt.Sprintf(`{"run":{"event":{"type":"setupNode","params":[{"key":"currentNodeId","value":"0"},{"key":"nodeIPs","value":"[\"0.0.0.0:8090\"]"},{"key":"initChainSetup","value":"%s"}]}}}`, base64.StdEncoding.EncodeToString(initChainSetup)))
	resp, err := suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	// migrate contract
	wasmbin := precompiles.GetPrecompileByLabel(wasmxtypes.CONSENSUS_RAFT)
	raftInitMsg := `{"instantiate":{"context":[{"key":"log","value":""},{"key":"nodeIPs","value":"[]"},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"matchIndex","value":"[]"},{"key":"commitIndex","value":"0"},{"key":"currentTerm","value":"0"},{"key":"lastApplied","value":"0"},{"key":"max_tx_bytes","value":"65536"},{"key":"prevLogIndex","value":"0"},{"key":"currentNodeId","value":"0"},{"key":"electionReset","value":"0"},{"key":"max_block_gas","value":"20000000"},{"key":"electionTimeout","value":"0"},{"key":"maxElectionTime","value":"20000"},{"key":"minElectionTime","value":"10000"},{"key":"heartbeatTimeout","value":"5000"}],"initialState":"uninitialized"}}`
	codeId := appA.StoreCode(sender, wasmbin, []string{wasmxtypes.INTERPRETER_FSM})
	newConsensus := appA.InstantiateCode(sender, codeId, wasmxtypes.WasmxExecutionMessage{Data: []byte(raftInitMsg)}, "newconsensus", nil)

	msgServer := wasmxkeeper.NewMsgServerImpl(&appA.App.WasmxKeeper)
	_, err = msgServer.RegisterRole(appA.Context(), &wasmxtypes.MsgRegisterRole{
		Role:            wasmxtypes.ROLE_CONSENSUS,
		Label:           "consensus_raft_0.0.2",
		ContractAddress: newConsensus.String(),
		Authority:       appA.App.WasmxKeeper.GetAuthority(),
		Title:           "title",
		Description:     "description",
	})
	suite.Require().NoError(err)

	// call setup()
	msg1 = []byte(fmt.Sprintf(`{"run":{"event":{"type":"setup","params":[{"key":"address","value":"%s"}]}}}`, consensusBech32))
	resp, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: newConsensus.String(),
		Msg:      msg1,
	})
	suite.Require().NoError(err)

	// Check each simulated node has the correct context:
	msg1 = []byte(`{"getContextValue":{"key":"nodeIPs"}}`)
	qresp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz := appA.QueryDecode(qresp.Data)
	suite.Require().Equal(string(qrespbz), "[\"0.0.0.0:8090\"]")

	msg1 = []byte(`{"getContextValue":{"key":"currentNodeId"}}`)
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(string(qrespbz), `0`)
}

func (suite *KeeperTestSuite) TestTendermintMigration() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	appA := s.AppContext()
	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}

	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), valAccount.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	consensusContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_CONSENSUS_TENDERMINT)
	consensusBech32 := consensusContract.String()

	storageContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_STORAGE_CHAIN)
	initChainSetup := []byte(fmt.Sprintf(`{"chain_id":"mythos_7000-14","consensus_params":{"block":{"max_bytes":22020096,"max_gas":-1},"evidence":{"max_age_num_blocks":100000,"max_age_duration":172800000000000,"max_bytes":1048576},"validator":{"pub_key_types":["ed25519"]},"version":{"app":0},"abci":{"vote_extensions_enable_height":0}},"validators":[{"address":"467F6127246A6E40B59899258DF08F857145B9CB","pub_key":"shBx7GuXCf7T+HwGwffE93xWOCkIwzPpp/oKkMq3hqw=","voting_power":100000000000000,"proposer_priority":0}],"app_hash":"47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=","last_results_hash":"47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=","version":{"consensus":{"block":0,"app":0},"software":""},"validator_address":"467F6127246A6E40B59899258DF08F857145B9CB","validator_privkey":"LdBVBItkqjNrSqwDaFgxZaO7n8rN01dJ6I3BQ/9LTTyyEHHsa5cJ/tP4fAbB98T3fFY4KQjDM+mn+gqQyreGrA==","validator_pubkey":"shBx7GuXCf7T+HwGwffE93xWOCkIwzPpp/oKkMq3hqw=","wasmx_blocks_contract":"%s"}`, storageContract.String()))

	msg1 := []byte(fmt.Sprintf(`{"run":{"event":{"type":"setupNode","params":[{"key":"currentNodeId","value":"0"},{"key":"nodeIPs","value":"[\"0.0.0.0:8090\"]"},{"key":"initChainSetup","value":"%s"}]}}}`, base64.StdEncoding.EncodeToString(initChainSetup)))
	resp, err := suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp)

	// migrate contract
	wasmbin := precompiles.GetPrecompileByLabel(wasmxtypes.CONSENSUS_TENDERMINT)
	initMsg := `{"instantiate":{"context":[{"key":"log","value":""},{"key":"nodeIPs","value":"[]"},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"matchIndex","value":"[]"},{"key":"commitIndex","value":"0"},{"key":"currentTerm","value":"0"},{"key":"lastApplied","value":"0"},{"key":"max_tx_bytes","value":"65536"},{"key":"prevLogIndex","value":"0"},{"key":"currentNodeId","value":"0"},{"key":"electionReset","value":"0"},{"key":"max_block_gas","value":"20000000"},{"key":"electionTimeout","value":"0"},{"key":"maxElectionTime","value":"20000"},{"key":"minElectionTime","value":"10000"},{"key":"roundTimeout","value":"10000"}],"initialState":"uninitialized"}}`
	codeId := appA.StoreCode(sender, wasmbin, []string{wasmxtypes.INTERPRETER_FSM})
	newConsensus := appA.InstantiateCode(sender, codeId, wasmxtypes.WasmxExecutionMessage{Data: []byte(initMsg)}, "newconsensus", nil)

	msgServer := wasmxkeeper.NewMsgServerImpl(&appA.App.WasmxKeeper)
	_, err = msgServer.RegisterRole(appA.Context(), &wasmxtypes.MsgRegisterRole{
		Role:            wasmxtypes.ROLE_CONSENSUS,
		Label:           "consensus_tendermint_0.0.2",
		ContractAddress: newConsensus.String(),
		Authority:       appA.App.WasmxKeeper.GetAuthority(),
		Title:           "title",
		Description:     "description",
	})
	suite.Require().NoError(err)

	// call setup()
	msg1 = []byte(fmt.Sprintf(`{"run":{"event":{"type":"setup","params":[{"key":"address","value":"%s"}]}}}`, consensusBech32))
	resp, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   consensusBech32,
		Contract: newConsensus.String(),
		Msg:      msg1,
	})
	suite.Require().NoError(err)

	// Check each simulated node has the correct context:
	msg1 = []byte(`{"getContextValue":{"key":"nodeIPs"}}`)
	qresp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz := appA.QueryDecode(qresp.Data)
	suite.Require().Equal(string(qrespbz), "[\"0.0.0.0:8090\"]")

	msg1 = []byte(`{"getContextValue":{"key":"currentNodeId"}}`)
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   consensusBech32,
		Contract: consensusBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(string(qrespbz), `0`)
}

func (suite *KeeperTestSuite) TestRaftToTendermintMigration() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	appA := s.AppContext()
	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), valAccount.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	raftContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_CONSENSUS_RAFT)
	raftContractBech32 := raftContract.String()

	// storageContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_STORAGE_CHAIN)
	// call start() on RAFT consensus
	msg1 := []byte(fmt.Sprintf(`{"run":{"event":{"type":"start","params":[{"key":"address","value":"%s"}]}}}`, raftContract.String()))
	_, err := suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   raftContractBech32,
		Contract: raftContractBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)

	// Candidate -> Leader
	msg1 = []byte(`{"delay":"electionTimeout","state":"#RAFT-FULL-1.initialized.Follower","intervalId":"1"}`)
	_, err = suite.App().NetworkKeeper.ExecuteEventual(appA.Context(), &types.MsgExecuteContract{
		Sender:   raftContractBech32,
		Contract: raftContractBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)

	tendermintInitMsg := wasmxtypes.WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"log","value":""},{"key":"nodeIPs","value":"[]"},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"currentTerm","value":"0"},{"key":"max_tx_bytes","value":"65536"},{"key":"currentNodeId","value":"0"},{"key":"max_block_gas","value":"20000000"},{"key":"roundTimeout","value":10000}],"initialState":"uninitialized"}}`)}
	wasmbin := precompiles.GetPrecompileByLabel(wasmxtypes.CONSENSUS_TENDERMINT)

	codeId := appA.StoreCode(sender, wasmbin, []string{wasmxtypes.INTERPRETER_FSM, wasmxtypes.BuildDep(wasmxtypes.ADDR_CONSENSUS_TENDERMINT_LIBRARY, wasmxtypes.ROLE_LIBRARY)})
	newConsensus := appA.InstantiateCode(sender, codeId, tendermintInitMsg, "newconsensus", nil)

	// Register contract role proposal
	newlabel := wasmxtypes.CONSENSUS_TENDERMINT + "2"
	title := "Register consensus"
	description := "Register consensus"
	authority := authtypes.NewModuleAddress(wasmxtypes.ROLE_GOVERNANCE).String()
	proposal := &wasmxtypes.MsgRegisterRole{Authority: authority, Title: title, Description: description, Role: "consensus", Label: newlabel, ContractAddress: newConsensus.String()}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{proposal}, "", title, description, false)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), newConsensus)
	s.Require().Equal(newlabel, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), newlabel)
	s.Require().Equal(newConsensus.String(), role.ContractAddress)
	s.Require().Equal(newlabel, role.Label)
	s.Require().Equal("consensus", role.Role)

	// check that the setup was done on the new contract
	stateKey := types.FSM_CONTEXT_KEY + types.STATE_KEY
	state := suite.GetContextValue(appA.Context(), stateKey, newConsensus.String())
	fmt.Println("---state---", newConsensus.String(), stateKey, state)

	// execute next block

	time.Sleep(time.Second * 10)

	return
	// migrate contract
	tendermintContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_CONSENSUS_TENDERMINT)

	msgServer := wasmxkeeper.NewMsgServerImpl(&appA.App.WasmxKeeper)
	_, err = msgServer.RegisterRole(appA.Context(), &wasmxtypes.MsgRegisterRole{
		Role:            wasmxtypes.ROLE_CONSENSUS,
		Label:           "consensus_tendermint_0.0.1",
		ContractAddress: tendermintContract.String(),
		Authority:       appA.App.WasmxKeeper.GetAuthority(),
		Title:           "title",
		Description:     "description",
	})
	suite.Require().NoError(err)

	// call setup()
	msg1 = []byte(fmt.Sprintf(`{"run":{"event":{"type":"setup","params":[{"key":"address","value":"%s"}]}}}`, raftContract.String()))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   raftContractBech32,
		Contract: tendermintContract.String(),
		Msg:      msg1,
	})
	suite.Require().NoError(err)

	// Check each simulated node has the correct context:
	msg1 = []byte(`{"getContextValue":{"key":"nodeIPs"}}`)
	qresp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   tendermintContract.String(),
		Contract: tendermintContract.String(),
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz := appA.QueryDecode(qresp.Data)
	suite.Require().Equal(string(qrespbz), "[\"0.0.0.0:8090\"]")

	msg1 = []byte(`{"getContextValue":{"key":"currentNodeId"}}`)
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &types.MsgQueryContract{
		Sender:   tendermintContract.String(),
		Contract: tendermintContract.String(),
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(string(qrespbz), `0`)

	// call start()
	msg1 = []byte(fmt.Sprintf(`{"run":{"event":{"type":"start","params":[{"key":"address","value":"%s"}]}}}`, raftContract.String()))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   raftContractBech32,
		Contract: tendermintContract.String(),
		Msg:      msg1,
	})
	suite.Require().NoError(err)

	msg1 = []byte(`{"delay":"roundTimeout","state":"#Tendermint_0.initialized.prestart","intervalId":1}`)
	_, err = appA.App.NetworkKeeper.ExecuteEventual(appA.Context(), &types.MsgExecuteContract{
		Sender:   tendermintContract.String(),
		Contract: tendermintContract.String(),
		Msg:      msg1,
	})
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestRaftToAvaSnowmanMigration() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	appA := s.AppContext()
	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	appA.Faucet.Fund(appA.Context(), valAccount.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()
	raftContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_CONSENSUS_RAFT)
	raftContractBech32 := raftContract.String()

	// storageContract := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_STORAGE_CHAIN)
	// call start() on RAFT consensus
	msg1 := []byte(fmt.Sprintf(`{"run":{"event":{"type":"start","params":[{"key":"address","value":"%s"}]}}}`, raftContract.String()))
	_, err := suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   raftContractBech32,
		Contract: raftContractBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)

	// Candidate -> Leader
	msg1 = []byte(`{"delay":"electionTimeout","state":"#RAFT-FULL-1.initialized.Follower","intervalId":"1"}`)
	_, err = suite.App().NetworkKeeper.ExecuteEventual(appA.Context(), &types.MsgExecuteContract{
		Sender:   raftContractBech32,
		Contract: raftContractBech32,
		Msg:      msg1,
	})
	suite.Require().NoError(err)

	avaInitMsg := wasmxtypes.WasmxExecutionMessage{Data: []byte(`{"instantiate":{"context":[{"key":"rounds","value":"3"},{"key":"sampleSize","value":3},{"key":"betaThreshold","value":3},{"key":"roundsCounter","value":"0"},{"key":"alphaThreshold","value":"2"}],"initialState":"uninitialized"}}`)}
	wasmbin := precompiles.GetPrecompileByLabel(wasmxtypes.CONSENSUS_AVA_SNOWMAN)

	codeId := appA.StoreCode(sender, wasmbin, []string{wasmxtypes.INTERPRETER_FSM, wasmxtypes.BuildDep(wasmxtypes.ADDR_CONSENSUS_AVA_SNOWMAN_LIBRARY, wasmxtypes.ROLE_LIBRARY)})
	newConsensus := appA.InstantiateCode(sender, codeId, avaInitMsg, "newconsensus", nil)

	// Register contract role proposal
	newlabel := wasmxtypes.CONSENSUS_AVA_SNOWMAN + "2"
	title := "Register consensus"
	description := "Register consensus"
	proposal := &wasmxtypes.MsgRegisterRole{Title: title, Description: description, Role: "consensus", Label: newlabel, ContractAddress: newConsensus.String()}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{proposal}, "", title, description, false)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), newConsensus)
	s.Require().Equal(newlabel, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), newlabel)
	s.Require().Equal(newConsensus.String(), role.ContractAddress)
	s.Require().Equal(newlabel, role.Label)
	s.Require().Equal("consensus", role.Role)

	// check that the setup was done on the new contract
	stateKey := types.FSM_CONTEXT_KEY + types.STATE_KEY
	state := suite.GetContextValue(appA.Context(), stateKey, newConsensus.String())
	fmt.Println("---state---", newConsensus.String(), stateKey, state)

	time.Sleep(time.Second * 5)

	// just execute a tx
	contractAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_IDENTITY)
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
	_, err = appA.App.AccountKeeper.GetSequence(appA.Context(), sender.Address)
	suite.Require().NoError(err)

	tx := appA.PrepareCosmosTx(sender, []sdk.Msg{msg}, nil, nil)
	txstr := base64.StdEncoding.EncodeToString(tx)

	msg1 = []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, txstr))
	resp2, err := suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   "consensus",
		Contract: "consensus",
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp2)

	// execute another tx
	_, err = appA.App.AccountKeeper.GetSequence(appA.Context(), sender.Address)
	suite.Require().NoError(err)
	tx = appA.PrepareCosmosTx(sender, []sdk.Msg{msg}, nil, nil)
	txstr = base64.StdEncoding.EncodeToString(tx)
	msg1 = []byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, txstr))
	resp2, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &types.MsgExecuteContract{
		Sender:   "consensus",
		Contract: "consensus",
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	log.Printf("Response: %+v", resp2)
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
// 	appA := s.AppContext()
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
// 	mapp := s.AppContext()
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
// 	contractAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_IDENTITY)
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

// func (suite *KeeperTestSuite) TestRAFTEncodeTx() {
// 	mnemonic := "enrich nose brisk lobster room large uniform mystery crush govern lazy vital feed dove soul emotion oblige shuffle else entry trend there dentist garbage"
// 	sender := suite.GetAccountFromMnemonic(mnemonic)
// 	initBalance := sdkmath.NewInt(1000_000_000)
// 	appA := s.GetAppContext(suite.chainA)
// 	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
// 	suite.Commit()

// 	// send tx
// 	contractAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_IDENTITY)
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
// 	tx := appA.PrepareCosmosTx(sender, []sdk.Msg{msg}, nil, nil)
// 	txstr := base64.StdEncoding.EncodeToString(tx)

// 	msgbase64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":"%s"}]}}}`, txstr)))
// 	fmt.Println("==msgbase64==", msgbase64)
// }
