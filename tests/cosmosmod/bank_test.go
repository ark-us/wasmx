package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	// keeper "github.com/loredanacirstea/wasmx/x/cosmosmod/keeper"
	mcfg "github.com/loredanacirstea/wasmx/config"
	multichain "github.com/loredanacirstea/wasmx/multichain"
)

func (suite *KeeperTestSuite) TestBankSend1() {
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)

	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(2200000000000000000).MulRaw(1000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)

	msg := &banktypes.MsgSend{
		FromAddress: senderPrefixed.String(),
		ToAddress:   "mythos10qhagn6kzj5yamwp7ryyhmemseyau0wkvzm9ce",
		Amount:      sdk.Coins{sdk.NewCoin("amyt", sdkmath.NewInt(1200000000000000000))}, // 1200000000000000000amyt
	}
	// fees 200000000000amyt
	// gas 9000000
	fees := "200000000000amyt"

	msgMultiChain, err := multichain.MultiChainWrap(appA.ClientCtx, msg, senderPrefixed)
	s.Require().NoError(err)

	res, err := appA.DeliverTxWithOpts(sender, msgMultiChain, "", uint64(9000000), &fees)
	s.Require().NoError(err)
	fmt.Println("--res--", res)
	fmt.Println("--res.GasUsed--", res.GasUsed)
	s.Require().True(res.IsOK(), res.Log, res.GetEvents())

	// --res.GasUsed-- 3371280
}

func (suite *KeeperTestSuite) TestBankSend2() {
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)

	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(2200000000000000000).MulRaw(1000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)

	msg := &banktypes.MsgSend{
		FromAddress: senderPrefixed.String(),
		ToAddress:   "mythos10qhagn6kzj5yamwp7ryyhmemseyau0wkvzm9ce",
		Amount:      sdk.Coins{sdk.NewCoin("amyt", sdkmath.NewInt(1200000000000000000))}, // 1200000000000000000amyt
	}
	// fees 200000000000amyt
	// gas 9000000
	fees := "200000000000amyt"

	res, err := appA.DeliverTxWithOpts(sender, msg, "", uint64(9000000), &fees)
	s.Require().NoError(err)
	fmt.Println("--res--", res)
	fmt.Println("--res.GasUsed--", res.GasUsed)
	s.Require().True(res.IsOK(), res.Log, res.GetEvents())

	// -res.GasUsed-- 3369220

}
