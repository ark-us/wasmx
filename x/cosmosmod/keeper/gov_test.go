package keeper_test

import (
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
	precompiles "mythos/v1/x/wasmx/vm/precompiles"
)

func (suite *KeeperTestSuite) TestContinuousVoting() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	appA := s.AppContext()

	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender2.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	senderAddrStr, err := appA.AddressCodec().BytesToString(sender.Address)
	suite.Require().NoError(err)

	msg := []byte(fmt.Sprintf(`{"SubmitProposal":{"messages":[],"proposer":"%s","initial_deposit":[{"denom":"amyt","amount":"0x100000"}],"metadata":"metadata","title":"title","summary":"summary","expedited":false}}`, senderAddrStr))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msg,
	})
	suite.Require().NoError(err)
	propext := suite.getPropExtended(appA)
	suite.Require().Equal(true, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(2), propext.Winner)

	senderAddrStr, err = appA.AddressCodec().BytesToString(sender.Address)
	suite.Require().NoError(err)

	msg = []byte(fmt.Sprintf(`{"DepositVote":{"proposal_id":1,"option_id":2,"voter":"%s","amount":"0x100000","arbitrationAmount":"0x0","metadata":"metadata"}}`, senderAddrStr))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msg,
	})
	suite.Require().NoError(err)

	propext = suite.getPropExtended(appA)
	suite.Require().Equal(false, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(2), propext.Winner)

	proposal, err := suite.App().GovKeeper.Proposal(appA.Context(), &govtypes.QueryProposalRequest{ProposalId: 1})
	suite.Require().NoError(err)
	s.Require().Equal(uint64(1), proposal.Proposal.Id)
	s.Require().Equal(govtypes.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD, proposal.Proposal.Status)
	s.Require().Equal(strconv.Itoa(0x200000), proposal.Proposal.FinalTallyResult.YesCount)
	s.Require().Equal("0", proposal.Proposal.FinalTallyResult.NoCount)
	s.Require().Equal("0", proposal.Proposal.FinalTallyResult.AbstainCount)
	s.Require().Equal("0", proposal.Proposal.FinalTallyResult.NoWithVetoCount)

	tally, err := suite.App().GovKeeper.TallyResult(appA.Context(), &govtypes.QueryTallyResultRequest{ProposalId: 1})
	suite.Require().NoError(err)
	s.Require().Equal(strconv.Itoa(0x200000), tally.Tally.YesCount)
	s.Require().Equal("0", tally.Tally.NoCount)
	s.Require().Equal("0", tally.Tally.AbstainCount)
	s.Require().Equal("0", tally.Tally.NoWithVetoCount)

	senderAddrStr, err = appA.AddressCodec().BytesToString(sender.Address)
	suite.Require().NoError(err)

	msg = []byte(fmt.Sprintf(`{"AddProposalOption":{"proposal_id":1,"option":{"messages":[],"proposer":"%s","amount":"0x100000","arbitrationAmount":"0x0","metadata":"metadata","title":"title","summary":"summary"}}}`, senderAddrStr))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msg,
	})
	suite.Require().NoError(err)

	propext = suite.getPropExtended(appA)
	suite.Require().Equal(false, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(2), propext.Winner)

	senderAddrStr, err = appA.AddressCodec().BytesToString(sender.Address)
	suite.Require().NoError(err)

	msg = []byte(fmt.Sprintf(`{"DepositVote":{"proposal_id":1,"option_id":3,"voter":"%s","amount":"0x200000","arbitrationAmount":"0x0","metadata":"metadata"}}`, senderAddrStr))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msg,
	})
	suite.Require().NoError(err)

	propext = suite.getPropExtended(appA)
	suite.Require().Equal(true, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(3), propext.Winner)

	tally, err = suite.App().GovKeeper.TallyResult(appA.Context(), &govtypes.QueryTallyResultRequest{ProposalId: 1})
	suite.Require().NoError(err)
	s.Require().Equal(strconv.Itoa(0x300000), tally.Tally.YesCount)
	s.Require().Equal(strconv.Itoa(0x200000), tally.Tally.NoCount)
	s.Require().Equal("0", tally.Tally.AbstainCount)
	s.Require().Equal("0", tally.Tally.NoWithVetoCount)
}

func (suite *KeeperTestSuite) TestRAFTP2PMigration() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	appA := s.AppContext()
	valAccount := simulation.Account{
		PrivKey: s.Chain().SenderPrivKey,
		PubKey:  s.Chain().SenderPrivKey.PubKey(),
		Address: s.Chain().SenderAccount.GetAddress(),
	}

	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender2.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	msg1 := []byte(`{"getContextValue":{"key":"validatorNodesInfo"}}`)
	qresp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_CONSENSUS,
		Contract: wasmxtypes.ROLE_CONSENSUS,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	nodesInfo := appA.QueryDecode(qresp.Data)

	// migrate contract
	wasmbin := precompiles.GetPrecompileByLabel(appA.AddressCodec(), wasmxtypes.CONSENSUS_RAFTP2P)
	raftInitMsg := `{"instantiate":{"context":[{"key":"log","value":""},{"key":"validatorNodesInfo","value":"[]"},{"key":"votedFor","value":"0"},{"key":"nextIndex","value":"[]"},{"key":"matchIndex","value":"[]"},{"key":"commitIndex","value":"0"},{"key":"currentTerm","value":"0"},{"key":"lastApplied","value":"0"},{"key":"max_tx_bytes","value":"65536"},{"key":"prevLogIndex","value":"0"},{"key":"currentNodeId","value":"0"},{"key":"electionReset","value":"0"},{"key":"max_block_gas","value":"20000000"},{"key":"electionTimeout","value":"0"},{"key":"maxElectionTime","value":"20000"},{"key":"minElectionTime","value":"10000"},{"key":"heartbeatTimeout","value":"5000"}],"initialState":"uninitialized"}}`
	codeId := appA.StoreCode(sender, wasmbin, []string{wasmxtypes.INTERPRETER_FSM})
	newConsensus := appA.InstantiateCode(sender, codeId, wasmxtypes.WasmxExecutionMessage{Data: []byte(raftInitMsg)}, "newconsensus", nil)

	// Register contract role proposal
	newlabel := wasmxtypes.CONSENSUS_RAFTP2P + "2"
	title := "Register consensus"
	description := "Register consensus"

	authority, err := appA.AddressCodec().BytesToString(authtypes.NewModuleAddress(wasmxtypes.ROLE_GOVERNANCE))
	suite.Require().NoError(err)

	newConsensusStr := newConsensus.String()

	proposal := &wasmxtypes.MsgRegisterRole{Authority: authority, Title: title, Description: description, Role: "consensus", Label: newlabel, ContractAddress: newConsensusStr}
	appA.PassGovProposal(valAccount, sender, []sdk.Msg{proposal}, "", title, description, false)

	resp := appA.App.WasmxKeeper.GetRoleLabelByContract(appA.Context(), newConsensus.Bytes())
	s.Require().Equal(newlabel, resp)

	role := appA.App.WasmxKeeper.GetRoleByLabel(appA.Context(), newlabel)
	s.Require().Equal(newConsensusStr, role.ContractAddress)
	s.Require().Equal(newlabel, role.Label)
	s.Require().Equal("consensus", role.Role)

	// check that the setup was done on the new contract

	// Check each simulated node has the correct context:
	msg1 = []byte(`{"getContextValue":{"key":"validatorNodesInfo"}}`)
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   newConsensusStr,
		Contract: newConsensusStr,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz := appA.QueryDecode(qresp.Data)
	suite.Require().Equal(string(qrespbz), string(nodesInfo))

	msg1 = []byte(`{"getContextValue":{"key":"currentNodeId"}}`)
	qresp, err = suite.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   newConsensusStr,
		Contract: newConsensusStr,
		Msg:      msg1,
	})
	suite.Require().NoError(err)
	qrespbz = appA.QueryDecode(qresp.Data)
	suite.Require().Equal(string(qrespbz), `0`)
}
