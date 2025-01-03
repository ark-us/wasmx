package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// keeper "github.com/loredanacirstea/wasmx/x/cosmosmod/keeper"
	mcfg "github.com/loredanacirstea/wasmx/config"
)

func (suite *KeeperTestSuite) TestStakingCreateValidator() {
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(10_000_000_000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	valPubKey, err := cryptocodec.FromCmtPubKeyInterface(chain.Vals.Validators[0].PubKey)
	suite.Require().NoError(err)

	valAddr := appA.BytesToAccAddressPrefixed(sender.Address.Bytes())
	valAddrStr, err := appA.ValidatorAddressCodec().BytesToString(sdk.ValAddress(valAddr.Bytes()))
	suite.Require().NoError(err)

	valFunds := sdkmath.NewInt(1000_000_000)
	appA.Faucet.Fund(appA.Context(), valAddr, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	createValMsg, err := stakingtypes.NewMsgCreateValidator(
		valAddrStr,
		valPubKey,
		sdk.NewCoin(appA.Chain.Config.BaseDenom, valFunds),
		stakingtypes.NewDescription("", "", "", "", ""),
		stakingtypes.NewCommissionRates(sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec()),
		sdkmath.OneInt(),
	)
	suite.Require().NoError(err)

	res, err := appA.DeliverTx(sender, createValMsg)
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
	suite.Require().Equal(valAddr.String(), validAddr)

	evs = appA.GetSdkEventsByType(res.GetEvents(), "create_validator")
	suite.Require().Equal(1, len(evs), "missing create_validator events")
	validAddr = ""
	for _, attr := range evs[0].Attributes {
		if attr.Key == "validator" {
			validAddr = attr.Value
		}
	}
	suite.Require().Equal(valAddr.String(), validAddr)
}

func (suite *KeeperTestSuite) TestStakingCreateValidatorFailDuplicate() {
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)
	sender := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderAccount.GetPubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}
	initBalance := sdkmath.NewInt(10_000_000_000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	valPubKey, err := cryptocodec.FromCmtPubKeyInterface(chain.Vals.Validators[0].PubKey)
	suite.Require().NoError(err)

	valAddr := appA.BytesToAccAddressPrefixed(sender.Address.Bytes())
	valAddrStr, err := appA.ValidatorAddressCodec().BytesToString(sdk.ValAddress(valAddr.Bytes()))
	suite.Require().NoError(err)

	valFunds := sdkmath.NewInt(1000_000_000)
	appA.Faucet.Fund(appA.Context(), valAddr, sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	createValMsg, err := stakingtypes.NewMsgCreateValidator(
		valAddrStr,
		valPubKey,
		sdk.NewCoin(appA.Chain.Config.BaseDenom, valFunds),
		stakingtypes.NewDescription("", "", "", "", ""),
		stakingtypes.NewCommissionRates(sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec()),
		sdkmath.OneInt(),
	)
	suite.Require().NoError(err)

	res, err := appA.DeliverTx(sender, createValMsg)
	suite.Require().NoError(err)

	evs := appA.GetSdkEventsByType(res.GetEvents(), "message")
	suite.Require().GreaterOrEqual(len(evs), 0, "should not be events message events")
	evs = appA.GetSdkEventsByType(res.GetEvents(), "create_validator")
	suite.Require().Equal(0, len(evs), "should not be create_validator events")
}
