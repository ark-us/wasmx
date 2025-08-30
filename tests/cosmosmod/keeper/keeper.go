package keeper

import (
	"encoding/json"
	"fmt"

	//nolint

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	wt "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// KeeperTestSuite is a testing suite to run tests on the same chain
type KeeperTestSuite struct {
	wt.KeeperTestSuite
}

func (suite *KeeperTestSuite) GetPropExtended(appA wt.AppContext, proposalId uint64, contract string) *types.ProposalExtended {
	msg := []byte(fmt.Sprintf(`{"GetProposalExtended":{"proposal_id":%d}}`, proposalId))
	resp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: contract,
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

func (suite *KeeperTestSuite) GetProp(appA wt.AppContext, proposalId uint64, contract string) *govtypes.Proposal {
	msg := []byte(fmt.Sprintf(`{"GetProposal":{"proposal_id":%d}}`, proposalId))
	resp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: contract,
		Msg:      msg,
	})
	suite.Require().NoError(err)
	var qresp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp.Data, &qresp)
	suite.Require().NoError(err)
	var propext types.QueryProposalResponse
	err = appA.App.AppCodec().UnmarshalJSON(qresp.Data, &propext)
	suite.Require().NoError(err)
	if propext.Proposal == nil {
		return nil
	}
	proposal, err := types.CosmosProposalFromInternal(appA.App.AppCodec(), *propext.Proposal)
	suite.Require().NoError(err)
	return proposal
}

func (suite *KeeperTestSuite) GetTallyResult(appA wt.AppContext, proposalId uint64, contract string) *govtypes.TallyResult {
	msg := []byte(fmt.Sprintf(`{"GetTallyResult":{"proposal_id":%d}}`, proposalId))
	resp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: contract,
		Msg:      msg,
	})
	suite.Require().NoError(err)
	var qresp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp.Data, &qresp)
	suite.Require().NoError(err)
	var propext govtypes.QueryTallyResultResponse
	err = appA.App.AppCodec().UnmarshalJSON(qresp.Data, &propext)
	suite.Require().NoError(err)
	if propext.Tally == nil {
		return nil
	}
	tally, err := types.CosmosTallyFromInternal(propext.Tally)
	suite.Require().NoError(err)
	return tally
}
