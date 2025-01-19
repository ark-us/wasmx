package keeper_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// keeper "github.com/loredanacirstea/wasmx/x/cosmosmod/keeper"
	mcfg "github.com/loredanacirstea/wasmx/config"
	testutil "github.com/loredanacirstea/wasmx/testutil/wasmx"
	cosmosmodtypes "github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
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

	fmt.Println("--valAddrStr--", valAddrStr, valAddr.String())
	memo := fmt.Sprintf("%s@/ip4/127.0.0.1/tcp/5001/p2p/12D3KooWPQ1Y8AwXx5xm8bjh6pfqZAsf3fFas3a2pVCaSuB4iHBg", valAddr.String())
	res, err := appA.DeliverTxWithOpts(sender, createValMsg, memo, 0, nil)
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

func (suite *KeeperTestSuite) TestStakingCreateValidatorFailDuplicateConsAddress() {
	// TODO
	suite.T().Skip("TODO: TestStakingCreateValidatorFailDuplicateConsAddress")
}

func (suite *KeeperTestSuite) TestStakingJailValidator() {
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)
	// chain := suite.GetChain(chainId)
	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(10_000_000_000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	newval := suite.NewValidator()

	valPubKey, err := cryptocodec.FromCmtPubKeyInterface(newval.PubKey)
	suite.Require().NoError(err)

	valAddr := appA.BytesToAccAddressPrefixed(sender.Address.Bytes())
	valAddrStr, err := appA.ValidatorAddressCodec().BytesToString(sdk.ValAddress(valAddr.Bytes()))
	suite.Require().NoError(err)

	valFunds := sdkmath.NewInt(100) // small amount, so we can validate blocks without this validator
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

	memo := fmt.Sprintf("%s@/ip4/127.0.0.1/tcp/5001/p2p/12D3KooWPQ1Y8AwXx5xm8bjh6pfqZAsf3fFas3a2pVCaSuB4iHBg", valAddr.String())
	res, err := appA.DeliverTxWithOpts(sender, createValMsg, memo, 0, nil)
	suite.Require().NoError(err)
	lastBlockHeight := suite.App().LastBlockHeight()

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

	infos, err := appA.App.SlashingKeeper.SigningInfos(appA.Context(), &slashingtypes.QuerySigningInfosRequest{})
	s.Require().NoError(err)
	s.Require().Equal(2, len(infos.Info))

	consAddress := infos.Info[len(infos.Info)-1].Address
	for _, v := range infos.Info {
		if v.StartHeight == lastBlockHeight {
			consAddress = v.Address
		}
	}

	info, err := appA.App.SlashingKeeper.SigningInfo(appA.Context(), &slashingtypes.QuerySigningInfoRequest{ConsAddress: consAddress})
	s.Require().NoError(err)
	s.Require().Equal(int64(0), info.MissedBlocksCounter)
	s.Require().Equal(int64(0), info.IndexOffset)

	allvals, err := appA.App.StakingKeeper.GetAllValidators(appA.Context())
	s.Require().NoError(err)
	s.Require().Equal(2, len(allvals))
	s.Require().False(allvals[1].Jailed)
	s.Require().True(allvals[1].IsBonded())

	// previous validator set
	valset, err := appA.ABCIClient().Validators(appA.Context(), &lastBlockHeight, nil, nil)
	s.Require().NoError(err)
	s.Require().Equal(1, len(valset.Validators))

	bitmap := getBlockBitMap(consAddress, appA)
	s.Require().Equal(1, len(bitmap.Chunks))
	s.Require().Equal(uint8(0), bitmap.Chunks[0][0]) // base2: 00000...

	contractAddress := appA.BytesToAccAddressPrefixed(wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_IDENTITY))
	internalmsg := wasmxtypes.WasmxExecutionMessage{Data: appA.Hex2bz("aa0000000000000000000000000000000000000000000000000000000077")}

	// first block with new validator
	appA.ExecuteContract(sender, contractAddress, internalmsg, nil, nil)
	lastBlockHeight = suite.App().LastBlockHeight()
	valset, err = appA.ABCIClient().Validators(appA.Context(), &lastBlockHeight, nil, nil)
	s.Require().NoError(err)
	s.Require().Equal(2, len(valset.Validators))

	info, err = appA.App.SlashingKeeper.SigningInfo(appA.Context(), &slashingtypes.QuerySigningInfoRequest{ConsAddress: consAddress})
	s.Require().NoError(err)
	s.Require().Equal(int64(0), info.MissedBlocksCounter)
	s.Require().Equal(int64(0), info.IndexOffset)

	bitmap = getBlockBitMap(consAddress, appA)
	s.Require().Equal(1, len(bitmap.Chunks))
	s.Require().Equal(uint8(0), bitmap.Chunks[0][0]) // base2: 00 00 0...

	// first missed block
	appA.ExecuteContract(sender, contractAddress, internalmsg, nil, nil)

	info, err = appA.App.SlashingKeeper.SigningInfo(appA.Context(), &slashingtypes.QuerySigningInfoRequest{ConsAddress: consAddress})
	s.Require().NoError(err)
	s.Require().Equal(int64(1), info.MissedBlocksCounter)
	s.Require().Equal(int64(1), info.IndexOffset)

	bitmap = getBlockBitMap(consAddress, appA)
	s.Require().Equal(1, len(bitmap.Chunks))
	s.Require().Equal(uint8(1), bitmap.Chunks[0][0]) // base2: 01 00 0...

	// second missed block
	appA.ExecuteContract(sender, contractAddress, internalmsg, nil, nil)
	info, err = appA.App.SlashingKeeper.SigningInfo(appA.Context(), &slashingtypes.QuerySigningInfoRequest{ConsAddress: consAddress})
	s.Require().NoError(err)
	s.Require().Equal(int64(2), info.MissedBlocksCounter)
	s.Require().Equal(int64(2), info.IndexOffset)

	bitmap = getBlockBitMap(consAddress, appA)
	s.Require().Equal(1, len(bitmap.Chunks))
	s.Require().Equal(uint8(3), bitmap.Chunks[0][0]) // base2: 11 00 00...

	// 3rd missed block
	appA.ExecuteContract(sender, contractAddress, internalmsg, nil, nil)
	info, err = appA.App.SlashingKeeper.SigningInfo(appA.Context(), &slashingtypes.QuerySigningInfoRequest{ConsAddress: consAddress})
	s.Require().NoError(err)
	s.Require().Equal(int64(3), info.MissedBlocksCounter)
	s.Require().Equal(int64(3), info.IndexOffset)

	bitmap = getBlockBitMap(consAddress, appA)
	s.Require().Equal(1, len(bitmap.Chunks))
	s.Require().Equal(uint8(7), bitmap.Chunks[0][0]) // base2: 11 10 00...
	s.Require().Equal(uint8(0), bitmap.Chunks[0][1])

	// 4th missed block (normally jailed at 3 missed blocks, but it waits until startHeight + signedBlocksWindow)
	appA.ExecuteContract(sender, contractAddress, internalmsg, nil, nil)
	allvals, err = appA.App.StakingKeeper.GetAllValidators(appA.Context())
	s.Require().NoError(err)
	s.Require().Equal(2, len(allvals))
	s.Require().True(allvals[1].Jailed)
	s.Require().True(allvals[1].IsBonded())

	lastBlockHeight = suite.App().LastBlockHeight()
	_, header, _, err := suite.GetBlock(appA.Context(), lastBlockHeight)
	s.Require().NoError(err)

	info, err = appA.App.SlashingKeeper.SigningInfo(appA.Context(), &slashingtypes.QuerySigningInfoRequest{ConsAddress: consAddress})
	s.Require().NoError(err)
	s.Require().Equal(int64(0), info.MissedBlocksCounter)
	s.Require().Equal(int64(0), info.IndexOffset)
	params, err := appA.App.SlashingKeeper.Params(appA.Context())
	s.Require().NoError(err)

	jailedUntil := header.Time.Add(params.DowntimeJailDuration)
	s.Require().Equal(jailedUntil, info.JailedUntil)

	// bitmap reset
	bitmap = getBlockBitMap(consAddress, appA)
	s.Require().Equal(1, len(bitmap.Chunks))
	s.Require().Equal(uint8(0), bitmap.Chunks[0][0]) // base2: 00 00...

	// 5th missed block, but not counted anymore
	appA.ExecuteContract(sender, contractAddress, internalmsg, nil, nil)
	info, err = appA.App.SlashingKeeper.SigningInfo(appA.Context(), &slashingtypes.QuerySigningInfoRequest{ConsAddress: consAddress})
	s.Require().NoError(err)
	s.Require().Equal(int64(0), info.MissedBlocksCounter)
	s.Require().Equal(int64(0), info.IndexOffset)

	// Unjail validator
	unjailMsg := &slashingtypes.MsgUnjail{ValidatorAddr: validAddr} // consAddress
	_, err = appA.DeliverTxWithOpts(sender, unjailMsg, memo, 0, nil)
	suite.Require().NoError(err)

	allvals, err = appA.App.StakingKeeper.GetAllValidators(appA.Context())
	s.Require().NoError(err)
	s.Require().Equal(2, len(allvals))
	s.Require().False(allvals[1].Jailed)
	s.Require().True(allvals[1].IsBonded())

	info, err = appA.App.SlashingKeeper.SigningInfo(appA.Context(), &slashingtypes.QuerySigningInfoRequest{ConsAddress: consAddress})
	s.Require().NoError(err)
	s.Require().Equal(int64(0), info.MissedBlocksCounter)
	s.Require().Equal(int64(0), info.IndexOffset)
	s.Require().Equal(time.Unix(0, 0).UTC(), info.JailedUntil)
}

func (suite *KeeperTestSuite) TestStakingUnauthorized() {
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)

	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(10_000_000_000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)

	stakingAddress, err := appA.App.WasmxKeeper.GetAddressOrRole(appA.Context(), wasmxtypes.ROLE_STAKING)
	s.Require().NoError(err)
	appA.WasmxQueryRaw(sender, stakingAddress, wasmxtypes.WasmxExecutionMessage{Data: []byte(`{"GetAllValidators":{}"}`)}, nil, nil)

	res, err := appA.ExecuteContractNoCheck(sender, stakingAddress, wasmxtypes.WasmxExecutionMessage{Data: []byte(`{"Unjail":{"address":""}}`)}, nil, nil, 0, nil)
	s.Require().NoError(err)
	s.Require().True(res.IsErr(), "should have failed authorization")
	expectedErr := fmt.Sprintf(`unauthorized caller: %s: Unjail`, senderPrefixed.String())
	s.Require().Contains(res.Log, hex.EncodeToString([]byte(expectedErr)))
}

func getBlockBitMap(consAddress string, appA testutil.AppContext) *cosmosmodtypes.MissedBlocksBitMap {
	msg := []byte(fmt.Sprintf(`{"GetMissedBlockBitmap":{"cons_address":"%s"}}`, consAddress))
	resp, err := appA.App.NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{Sender: wasmxtypes.ROLE_SLASHING, Contract: wasmxtypes.ROLE_SLASHING, Msg: msg})
	s.Require().NoError(err)

	var data wasmxtypes.WasmxExecutionMessage
	err = json.Unmarshal(resp.Data, &data)
	s.Require().NoError(err)

	var bitmap cosmosmodtypes.MissedBlocksBitMap
	err = json.Unmarshal(data.Data, &bitmap)
	s.Require().NoError(err)
	return &bitmap
}
