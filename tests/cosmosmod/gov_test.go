package keeper_test

import (
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/ethereum/go-ethereum/common"

	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestContinuousVoting() {
	sender := suite.GetRandomAccount()
	sender2 := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1_000_000_000_000_000_000)
	appA := s.AppContext()

	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender2.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	receiver := common.HexToAddress(types.ADDR_GOV_CONT)
	receiverAcc := types.AccAddressFromEvm(receiver)
	govAddress := appA.BytesToAccAddressPrefixed(receiverAcc.Bytes()).String()

	senderAddrStr, err := appA.AddressCodec().BytesToString(sender.Address)
	suite.Require().NoError(err)

	// 0x100000
	msg := []byte(fmt.Sprintf(`{"SubmitProposal":{"messages":[],"proposer":"%s","initial_deposit":[{"denom":"amyt","amount":"1048576"}],"metadata":"metadata","title":"title","summary":"summary","expedited":false}}`, senderAddrStr))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: govAddress,
		Msg:      msg,
	})
	suite.Require().NoError(err)
	propext := suite.GetPropExtended(appA, 1, govAddress)
	suite.Require().Equal(true, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(2), propext.Winner)

	senderAddrStr, err = appA.AddressCodec().BytesToString(sender.Address)
	suite.Require().NoError(err)

	// 0x100000
	msg = []byte(fmt.Sprintf(`{"DepositVote":{"proposal_id":1,"option_id":2,"voter":"%s","amount":"1048576","arbitration_amount":"0","metadata":"metadata"}}`, senderAddrStr))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: govAddress,
		Msg:      msg,
	})
	suite.Require().NoError(err)

	propext = suite.GetPropExtended(appA, 1, govAddress)
	suite.Require().Equal(false, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(2), propext.Winner)

	proposal := suite.GetProp(appA, 1, govAddress)
	suite.Require().NoError(err)
	suite.Require().NotNil(proposal)
	s.Require().Equal(uint64(1), proposal.Id)
	s.Require().Equal(govtypes.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD, proposal.Status)
	s.Require().Equal(strconv.Itoa(0x200000), proposal.FinalTallyResult.YesCount)
	s.Require().Equal("0", proposal.FinalTallyResult.NoCount)
	s.Require().Equal("0", proposal.FinalTallyResult.AbstainCount)
	s.Require().Equal("0", proposal.FinalTallyResult.NoWithVetoCount)

	tally := suite.GetTallyResult(appA, 1, govAddress)
	suite.Require().NoError(err)
	s.Require().Equal(strconv.Itoa(0x200000), tally.YesCount)
	s.Require().Equal("0", tally.NoCount)
	s.Require().Equal("0", tally.AbstainCount)
	s.Require().Equal("0", tally.NoWithVetoCount)

	senderAddrStr, err = appA.AddressCodec().BytesToString(sender.Address)
	suite.Require().NoError(err)

	// 0x100000
	msg = []byte(fmt.Sprintf(`{"AddProposalOption":{"proposal_id":1,"option":{"messages":[],"proposer":"%s","amount":"1048576","arbitration_amount":"0","metadata":"metadata","title":"title","summary":"summary"}}}`, senderAddrStr))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: govAddress,
		Msg:      msg,
	})
	suite.Require().NoError(err)

	propext = suite.GetPropExtended(appA, 1, govAddress)
	suite.Require().Equal(false, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(2), propext.Winner)

	senderAddrStr, err = appA.AddressCodec().BytesToString(sender.Address)
	suite.Require().NoError(err)

	// 0x400000
	msg = []byte(fmt.Sprintf(`{"DepositVote":{"proposal_id":1,"option_id":3,"voter":"%s","amount":"4194304","arbitration_amount":"0","metadata":"metadata"}}`, senderAddrStr))
	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: govAddress,
		Msg:      msg,
	})
	suite.Require().NoError(err)

	propext = suite.GetPropExtended(appA, 1, govAddress)
	suite.Require().Equal(true, propext.VoteStatus.Changed)
	suite.Require().Equal(uint32(3), propext.Winner)

	tally = suite.GetTallyResult(appA, 1, govAddress)
	suite.Require().NoError(err)
	s.Require().Equal(strconv.Itoa(0x500000), tally.YesCount)
	s.Require().Equal(strconv.Itoa(0x200000), tally.NoCount)
	s.Require().Equal("0", tally.AbstainCount)
	s.Require().Equal("0", tally.NoWithVetoCount)
}
