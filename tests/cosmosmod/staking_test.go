package keeper_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// keeper "github.com/loredanacirstea/wasmx/x/cosmosmod/keeper"
	mcfg "github.com/loredanacirstea/wasmx/config"
	testutil "github.com/loredanacirstea/wasmx/testutil/wasmx"
	cosmosmodtypes "github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestStakingCreateValidatorFailDuplicate() {
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)
	// this is the existing validator account
	sender := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderAccount.GetPubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}
	initBalance := sdkmath.NewInt(10_000_000_000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))

	// TODO we should create another validator consensus key
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

	// we get the validator key from the chain (existing validator)
	valPubKey, err := cryptocodec.FromCmtPubKeyInterface(chain.Vals.Validators[0].PubKey)
	suite.Require().NoError(err)

	// the operator is a new address
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

func (suite *KeeperTestSuite) TestConsensusUnauthorized() {
	// TODO call from EOA to consensusless contracts
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)

	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(10_000_000_000)
	appA := s.AppContext()
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(appA.Chain.Config.BaseDenom, initBalance))
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)

	consensusAddress, err := appA.App.WasmxKeeper.GetAddressOrRole(appA.Context(), wasmxtypes.ROLE_CONSENSUS)
	s.Require().NoError(err)

	msg1 := []byte(`{"run":{"event": {"type": "newTransaction", "params": [{"key": "transaction", "value":""}]}}}`)
	res, err := appA.ExecuteContractNoCheck(sender, consensusAddress, wasmxtypes.WasmxExecutionMessage{Data: msg1}, nil, nil, 0, nil)
	s.Require().NoError(err)
	s.Require().True(res.IsErr(), "should have failed authorization")

	_, err = suite.App().NetworkKeeper.ExecuteContract(appA.Context(), &networktypes.MsgExecuteContract{
		Sender:   senderPrefixed.String(),
		Contract: consensusAddress.String(),
		Msg:      msg1,
	})
	s.Require().Error(err)
	// expectedErr := fmt.Sprintf(`unauthorized caller: %s: Unjail`, senderPrefixed.String())
	// s.Require().Contains(res.Log, hex.EncodeToString([]byte(expectedErr)))
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
