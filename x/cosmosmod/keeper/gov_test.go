package keeper_test

import (
	"encoding/json"
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	testwasmx "mythos/v1/testutil/wasmx"
	"mythos/v1/x/cosmosmod/types"
	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestContinuousVoting() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	appA := s.AppContext()

	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	appA.Faucet.Fund(appA.Context(), sender2.Address, sdk.NewCoin(appA.Denom, initBalance))

	msg := []byte(fmt.Sprintf(`{"SubmitProposal":{"messages":[],"proposer":"%s","initial_deposit":[{"denom":"amyt","amount":"0x100000"}],"metadata":"metadata","title":"title","summary":"summary","expedited":false}}`, sender.Address.String()))
	_, err := suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msg,
	})
	suite.Require().NoError(err)
	propext := suite.getPropExtended(appA)
	suite.Require().Equal(true, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(1), propext.Winner)

	msg = []byte(fmt.Sprintf(`{"DepositVote":{"proposal_id":1,"option_id":1,"voter":"%s","amount":"0x100000","arbitrationAmount":"0x0","metadata":"metadata"}}`, sender.Address.String()))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msg,
	})
	suite.Require().NoError(err)

	propext = suite.getPropExtended(appA)
	suite.Require().Equal(false, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(1), propext.Winner)

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

	msg = []byte(fmt.Sprintf(`{"AddProposalOption":{"proposal_id":1,"option":{"messages":[],"proposer":"%s","amount":"0x100000","arbitrationAmount":"0x0","metadata":"metadata","title":"title","summary":"summary"}}}`, sender.Address.String()))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msg,
	})
	suite.Require().NoError(err)

	propext = suite.getPropExtended(appA)
	suite.Require().Equal(false, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(1), propext.Winner)

	msg = []byte(fmt.Sprintf(`{"DepositVote":{"proposal_id":1,"option_id":2,"voter":"%s","amount":"0x200000","arbitrationAmount":"0x0","metadata":"metadata"}}`, sender.Address.String()))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msg,
	})
	suite.Require().NoError(err)

	propext = suite.getPropExtended(appA)
	suite.Require().Equal(true, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(2), propext.Winner)

	tally, err = suite.App().GovKeeper.TallyResult(appA.Context(), &govtypes.QueryTallyResultRequest{ProposalId: 1})
	suite.Require().NoError(err)
	s.Require().Equal(strconv.Itoa(0x300000), tally.Tally.YesCount)
	s.Require().Equal(strconv.Itoa(0x200000), tally.Tally.NoCount)
	s.Require().Equal("0", tally.Tally.AbstainCount)
	s.Require().Equal("0", tally.Tally.NoWithVetoCount)
}

func (suite *KeeperTestSuite) getPropExtended(appA testwasmx.AppContext) *types.ProposalExtended {
	msg := []byte(`{"GetProposalExtended":{"proposal_id":1}}`)
	resp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msg,
	})
	suite.Require().NoError(err)
	var qresp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp.Data, &qresp)
	suite.Require().NoError(err)
	var propext types.QueryProposalExtendedResponse
	err = appA.App.AppCodec().UnmarshalJSON(qresp.Data, &propext)
	suite.Require().NoError(err)
	return propext.Proposal
}
