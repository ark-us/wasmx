package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	// keeper "mythos/v1/x/cosmosmod/keeper"
)

func (suite *KeeperTestSuite) TestStakingCreateValidator() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(10_000_000_000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	valAccount := suite.GetRandomAccount()
	valAddr := sdk.ValAddress(valAccount.Address)

	valAddrStr, err := appA.ValidatorAddressCodec().BytesToString(valAddr)
	suite.Require().NoError(err)

	valAccountAddrStr, err := appA.AddressCodec().BytesToString(valAccount.Address)
	suite.Require().NoError(err)

	valFunds := sdkmath.NewInt(1000_000_000)
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(valAccount.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	createValMsg, err := stakingtypes.NewMsgCreateValidator(
		valAddrStr,
		valAccount.PubKey,
		sdk.NewCoin(appA.Chain.Config.BaseDenom, valFunds),
		stakingtypes.NewDescription("", "", "", "", ""),
		stakingtypes.NewCommissionRates(sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec()),
		sdkmath.OneInt(),
	)
	suite.Require().NoError(err)

	res, err := appA.DeliverTx(valAccount, createValMsg)
	suite.Require().NoError(err)

	evs := appA.GetSdkEventsByType(res.GetEvents(), "message")
	suite.Require().GreaterOrEqual(len(evs), 1, "missing message events")
	msgname := "/cosmos.staking.v1beta1.MsgCreateValidator"
	found := false
	validAddr := ""
	for _, ev := range evs {
		for _, attr := range ev.Attributes {
			if attr.Key == "action" && attr.Value == msgname {
				found = true
			}
			if attr.Key == "sender" {
				validAddr = attr.Value
			}
		}
	}
	suite.Require().True(found)
	suite.Require().Equal(valAccountAddrStr, validAddr)

	evs = appA.GetSdkEventsByType(res.GetEvents(), "create_validator")
	suite.Require().Equal(1, len(evs), "missing create_validator events")
	validAddr = ""
	for _, attr := range evs[0].Attributes {
		if attr.Key == "validator" {
			validAddr = attr.Value
		}
	}
	suite.Require().Equal(valAccountAddrStr, validAddr)
}
