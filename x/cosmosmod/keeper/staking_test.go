package keeper_test

import (
	math "cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	keeper "mythos/v1/x/cosmosmod/keeper"
)

func (suite *KeeperTestSuite) TestStakingCreateValidator() {
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), sender.Address, sdk.NewCoin(appA.Denom, initBalance))
	suite.Commit()

	// generate validator private/public key
	// privVal := mock.NewPV()
	// pubKey, err := privVal.GetPubKey()
	// suite.Require().NoError(err)

	privVal := secp256k1.GenPrivKey()
	pubKey := privVal.PubKey()

	addr1 := sdk.AccAddress(pubKey.Address())
	valAddr1 := sdk.ValAddress(addr1)
	valFunds := sdkmath.NewInt(1000_000_000)
	appA.Faucet.Fund(appA.Context(), addr1, sdk.NewCoin(appA.Denom, valFunds))

	// // send funds from module to addr to perform DepositValidatorRewardsPool
	// err = f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tokens)))
	// require.NoError(t, err)
	// tstaking := stakingtestutil.NewHelper(t, f.sdkCtx, f.stakingKeeper)
	// tstaking.Commission = stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	// tstaking.CreateValidator(valAddr1, valConsPk0, math.NewInt(100), true)

	stakingServer := keeper.NewMsgServerImpl(appA.App.CosmosmodKeeper)

	createValMsg, err := stakingtypes.NewMsgCreateValidator(
		valAddr1.String(),
		pubKey,
		sdk.NewCoin(appA.Denom, valFunds),
		stakingtypes.NewDescription("", "", "", "", ""),
		stakingtypes.NewCommissionRates(math.LegacyOneDec(), math.LegacyOneDec(), math.LegacyOneDec()),
		math.OneInt(),
	)
	suite.Require().NoError(err)

	_, err = stakingServer.CreateValidator(appA.Context(), createValMsg)
	suite.Require().NoError(err)
}
