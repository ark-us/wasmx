package keeper_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	anypb "google.golang.org/protobuf/types/known/anypb"

	sdkmath "cosmossdk.io/math"
	txsigning "cosmossdk.io/x/tx/signing"
	abci "github.com/cometbft/cometbft/abci/types"
	client "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/loredanacirstea/wasmx/app"
	mcodec "github.com/loredanacirstea/wasmx/codec"
	mcfg "github.com/loredanacirstea/wasmx/config"
	menc "github.com/loredanacirstea/wasmx/encoding"
	multichain "github.com/loredanacirstea/wasmx/multichain"
	ibctesting "github.com/loredanacirstea/wasmx/testutil/ibc"
	wasmxtesting "github.com/loredanacirstea/wasmx/testutil/wasmx"
	cosmosmodtypes "github.com/loredanacirstea/wasmx/x/cosmosmod/types"

	// networkserver "github.com/loredanacirstea/wasmx/x/network/server"
	testdata "github.com/loredanacirstea/mythos-tests/network/testdata/wasmx"
	"github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
	precompiles "github.com/loredanacirstea/wasmx/x/wasmx/vm/precompiles"
)

func (suite *KeeperTestSuite) TestMultiChainExecMythos() {
	SkipFixmeTests(suite.T(), "TestMultiChainExecMythos")
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	config, err := mcfg.GetChainConfig(chainId)
	s.Require().NoError(err)
	sender := suite.GetRandomAccount()
	newacc := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)
	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(denom, initBalance))
	suite.Commit()

	bankAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_BANK)
	bankAddressStr := appA.MustAccAddressToString(bankAddress)
	newaccStr := appA.MustAccAddressToString(newacc.Address)

	msg := fmt.Sprintf(`{"SendCoins":{"from_address":"%s","to_address":"%s","amount":[{"denom":"%s","amount":"0x1000"}]}}`, appA.MustAccAddressToString(sender.Address), newaccStr, config.BaseDenom)
	suite.broadcastMultiChainExec([]byte(msg), sender, bankAddress, chainId)

	qmsg := fmt.Sprintf(`{"GetBalance":{"address":"%s","denom":"%s"}}`, newaccStr, config.BaseDenom)
	res := suite.queryMultiChainCall(appA.App, []byte(qmsg), sender, bankAddressStr, chainId)

	balance := &banktypes.QueryBalanceResponse{}
	err = json.Unmarshal(res, balance)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(denom, sdkmath.NewInt(0x1000)), *balance.Balance)
	// TODO try again query client - this time with conn.defer() in the test
}

func (suite *KeeperTestSuite) TestMultiChainExecLevel0() {
	SkipFixmeTests(suite.T(), "TestMultiChainExecLevel0")
	chainId := mcfg.LEVEL0_CHAIN_ID
	config, err := mcfg.GetChainConfig(chainId)
	s.Require().NoError(err)
	suite.SetCurrentChain(chainId)

	sender := suite.GetRandomAccount()
	newacc := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(denom, initBalance))
	suite.Commit()

	bankAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_BANK)
	bankAddressStr := appA.MustAccAddressToString(bankAddress)
	newaccStr := appA.MustAccAddressToString(newacc.Address)

	msg := fmt.Sprintf(`{"SendCoins":{"from_address":"%s","to_address":"%s","amount":[{"denom":"%s","amount":"0x1000"}]}}`, appA.MustAccAddressToString(sender.Address), newaccStr, config.BaseDenom)
	suite.broadcastMultiChainExec([]byte(msg), sender, bankAddress, chainId)

	qmsg := fmt.Sprintf(`{"GetBalance":{"address":"%s","denom":"%s"}}`, newaccStr, config.BaseDenom)
	res := suite.queryMultiChainCall(appA.App, []byte(qmsg), sender, bankAddressStr, chainId)

	balance := &banktypes.QueryBalanceResponse{}
	err = json.Unmarshal(res, balance)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(denom, sdkmath.NewInt(0x1000)), *balance.Balance)
	// TODO try again query client - this time with conn.defer() in the test
}

func (suite *KeeperTestSuite) TestMultiChainInit() {
	SkipFixmeTests(suite.T(), "TestMultiChainInit")
	chainId := mcfg.LEVEL0_CHAIN_ID
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)

	initBalance := sdkmath.NewInt(10_000_000_000)
	sender := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderAccount.GetPubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	appA.Faucet.Fund(appA.Context(), appA.BytesToAccAddressPrefixed(sender.Address), sdk.NewCoin(denom, initBalance))
	suite.Commit()

	registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)

	subChainConfig := menc.ChainConfig{
		Bech32PrefixAccAddr:  "ttt",
		Bech32PrefixAccPub:   "ttt",
		Bech32PrefixValAddr:  "ttt",
		Bech32PrefixValPub:   "ttt",
		Bech32PrefixConsAddr: "ttt",
		Bech32PrefixConsPub:  "ttt",
		Name:                 "ttt",
		HumanCoinUnit:        "ttt",
		BaseDenom:            "attt",
		DenomUnit:            "ttt",
		BaseDenomUnit:        18,
		BondBaseDenom:        "asttt",
		BondDenom:            "sttt",
	}

	subChainId := "tttest_1000-1"
	encoding := menc.MakeEncodingConfig(&subChainConfig, []txsigning.CustomGetSigner{types.ProvideExecuteAtomicTxGetSigners()})
	addrCodec := mcodec.MustUnwrapAccBech32Codec(encoding.InterfaceRegistry.SigningContext().AddressCodec())
	valAddrCodec := mcodec.MustUnwrapValBech32Codec(encoding.InterfaceRegistry.SigningContext().ValidatorAddressCodec())
	valTokens := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)

	genesisAccs := []cosmosmodtypes.GenesisAccount{}
	balances := []banktypes.Balance{}
	_, genesisState, err := ibctesting.BuildGenesisData(suite.App().WasmxKeeper.WasmRuntime, &tmtypes.ValidatorSet{}, genesisAccs, subChainId, subChainConfig, 10, balances, suite.CompiledCacheDir, suite.GetDB)
	s.Require().NoError(err)

	genesisStateWasmx := map[string][]byte{}
	for key, value := range genesisState {
		genesisStateWasmx[key] = value
	}

	stateBytes, err := json.Marshal(genesisStateWasmx)
	s.Require().NoError(err)

	req := abci.RequestInitChain{
		ChainId:         subChainId,
		InitialHeight:   1,
		Time:            time.Now().UTC(),
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: app.DefaultTestingConsensusParams,
		AppStateBytes:   stateBytes,
	}

	valAddr := addrCodec.BytesToAccAddressPrefixed(sender.Address.Bytes())
	valStr, err := valAddrCodec.BytesToString(sdk.ValAddress(valAddr.Bytes()))
	suite.Require().NoError(err)

	memo := fmt.Sprintf("%s@/ip4/127.0.0.1/tcp/5001/p2p/12D3KooWJdKwTq9QcARdPuk4QBibP8MxBV7Q8xC7JRMSXWuvZBtD", valAddr.String())

	valPubKey, err := cryptocodec.FromCmtPubKeyInterface(chain.Vals.Validators[0].PubKey)
	suite.Require().NoError(err)

	validMsg, err := stakingtypes.NewMsgCreateValidator(
		valStr,
		valPubKey, // validator consensus key
		sdk.NewCoin(subChainConfig.BaseDenom, valTokens),
		stakingtypes.NewDescription("moniker1", "", "", "", ""),
		stakingtypes.NewCommissionRates(sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec()),
		sdkmath.OneInt(),
	)
	s.Require().NoError(err)

	subchainGasPrices := "10attt"

	multichainapp, err := mcfg.GetMultiChainApp(appA.App.GetGoContextParent())
	suite.Require().NoError(err)
	subchainapp := multichainapp.NewApp(subChainId, &subChainConfig)

	valTxBuilder := appA.PrepareCosmosSdkTxBuilder([]sdk.Msg{validMsg}, nil, &subchainGasPrices, memo)
	accSeq := uint64(0)
	accNo := uint64(0)
	accAddress, err := subchainapp.InterfaceRegistry().SigningContext().AddressCodec().BytesToString(sender.Address.Bytes())
	suite.Require().NoError(err)
	sigV2 := signing.SignatureV2{
		PubKey: sender.PubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode(subchainapp.TxConfig().SignModeHandler().DefaultMode()),
			Signature: nil,
		},
		Sequence: accSeq,
	}
	err = valTxBuilder.SetSignatures(sigV2)
	suite.Require().NoError(err)

	signerData := authsigning.SignerData{
		ChainID:       subChainId,
		AccountNumber: accNo,
		Sequence:      accSeq,
		PubKey:        sender.PubKey,
		Address:       accAddress,
	}
	sigV2, err = tx.SignWithPrivKey(
		appA.Context().Context(),
		signing.SignMode(subchainapp.TxConfig().SignModeHandler().DefaultMode()), signerData,
		valTxBuilder, sender.PrivKey, subchainapp.TxConfig(),
		accSeq,
	)
	suite.Require().NoError(err)

	err = valTxBuilder.SetSignatures(sigV2)
	suite.Require().NoError(err)

	valSdkTx := valTxBuilder.GetTx()

	txjsonbz, err := subchainapp.TxConfig().TxJSONEncoder()(valSdkTx)
	s.Require().NoError(err)

	// test verif
	anyPk, _ := codectypes.NewAnyWithValue(sender.PubKey)
	signerData2 := txsigning.SignerData{
		Address:       signerData.Address,
		ChainID:       signerData.ChainID,
		AccountNumber: signerData.AccountNumber,
		Sequence:      signerData.Sequence,
		PubKey: &anypb.Any{
			TypeUrl: anyPk.TypeUrl,
			Value:   anyPk.Value,
		},
	}
	adaptableTx, ok := valSdkTx.(authsigning.V2AdaptableTx)
	s.Require().True(ok)
	txData := adaptableTx.GetSigningTxData()

	sigs, err := valSdkTx.GetSignaturesV2()
	s.Require().NoError(err)
	err = authsigning.VerifySignature(appA.Context().Context(), sender.PubKey, signerData2, sigs[0].Data, subchainapp.TxConfig().SignModeHandler(), txData)
	s.Require().NoError(err)
	// test verif END

	regreq := wasmxtypes.RegisterSubChainRequest{
		Data: wasmxtypes.InitSubChainDeterministicRequest{
			InitChainRequest: req,
			ChainConfig:      subChainConfig,
			Peers:            []string{},
		},
		GenTxs:         [][]byte{txjsonbz},
		InitialBalance: initBalance.BigInt(),
	}
	regreqBz, err := json.Marshal(regreq)
	suite.Require().NoError(err)
	msg := fmt.Sprintf(`{"RegisterSubChain":%s}`, string(regreqBz))

	_, err = suite.broadcastMultiChainExec([]byte(msg), sender, registryAddress, chainId)
	suite.Require().NoError(err)

	msg = fmt.Sprintf(`{"InitSubChain":{"chainId":"%s"}}`, subChainId)
	res, err := suite.broadcastMultiChainExec([]byte(msg), sender, registryAddress, chainId)
	suite.Require().NoError(err)
	evs := appA.GetSdkEventsByType(res.Events, "init_subchain")
	suite.Require().Equal(1, len(evs))

	// time.Sleep(time.Second * 3)

	// test restarting the node by starting the parent chain
	// err = networkserver.StartNode(appA.App, appA.App.Logger(), appA.App.GetNetworkKeeper())
	// suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestMultiChainDefaultInit() {
	SkipFixmeTests(suite.T(), "TestMultiChainDefaultInit")
	chainId := mcfg.LEVEL0_CHAIN_ID
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)

	initBalance := sdkmath.NewInt(10_000_000_000)
	sender := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderAccount.GetPubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(denom, initBalance))
	suite.Commit()

	registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
	registryAddressStr := appA.MustAccAddressToString(registryAddress)

	// create new subchain genesis registry
	initialBalance, ok := sdkmath.NewIntFromString("10000000000100000000")
	suite.Require().True(ok)
	regreq, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{RegisterDefaultSubChain: &wasmxtypes.RegisterDefaultSubChainRequest{
		ChainBaseName:  "ptestp",
		DenomUnit:      "ppp",
		Decimals:       18,
		LevelIndex:     1,
		InitialBalance: initialBalance.BigInt(),
	}})
	suite.Require().NoError(err)

	_, err = suite.broadcastMultiChainExec(regreq, sender, registryAddress, chainId)
	suite.Require().NoError(err)

	// TODO expected subchainId
	subChainId := "ptestp_1_1001-1"

	// create genTx data to sign - call to level0
	// buildGenTx query
	// valTokens := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
	valTokens, ok := sdkmath.NewIntFromString("10000000000000000000")
	suite.Require().True(ok)
	validMsg := stakingtypes.MsgCreateValidator{
		// TODO fix as-json.parse when description contains empty strings
		Description:       stakingtypes.NewDescription("moniker1", "id", "website", "security", "details"),
		Commission:        stakingtypes.NewCommissionRates(sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec()),
		MinSelfDelegation: sdkmath.OneInt(),
		Value:             sdk.NewCoin(denom, valTokens), // denom will be changed by level0 anyway
	}

	genTxInfo, err := json.Marshal(&wasmxtypes.QueryBuildGenTxRequest{
		ChainId: subChainId,
		Msg:     validMsg,
	})
	suite.Require().NoError(err)
	paramBz, err := json.Marshal(&wasmxtypes.ActionParam{Key: "message", Value: string(genTxInfo)})
	suite.Require().NoError(err)

	msg := []byte(fmt.Sprintf(`{"execute":{"action": {"type": "buildGenTx", "params": [%s],"event":null}}}`, string(paramBz)))
	txbz := suite.queryMultiChainCall(appA.App, msg, sender, wasmxtypes.ROLE_CONSENSUS, chainId)
	msg = []byte(fmt.Sprintf(`{"GetSubChainConfigById":{"chainId":"%s"}}`, subChainId))
	chaincfgbz := suite.queryMultiChainCall(appA.App, msg, sender, registryAddressStr, chainId)
	suite.Require().NoError(err)

	var subchainConfig menc.ChainConfig
	err = json.Unmarshal(chaincfgbz, &subchainConfig)
	suite.Require().NoError(err)

	multichainapp, err := mcfg.GetMultiChainApp(appA.App.GetGoContextParent())
	suite.Require().NoError(err)
	subchainapp := multichainapp.NewApp(subChainId, &subchainConfig)
	subtxconfig := subchainapp.TxConfig()

	sigV2 := signing.SignatureV2{
		PubKey: sender.PubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode(subtxconfig.SignModeHandler().DefaultMode()),
			Signature: nil,
		},
		Sequence: 0,
	}

	sdktx, err := subtxconfig.TxJSONDecoder()(txbz)
	suite.Require().NoError(err)
	txbuilder, err := subtxconfig.WrapTxBuilder(sdktx)
	suite.Require().NoError(err)

	subchainSender, err := subtxconfig.SigningContext().AddressCodec().BytesToString(sender.Address)
	suite.Require().NoError(err)

	err = txbuilder.SetSignatures(sigV2)
	suite.Require().NoError(err)

	signerData := authsigning.SignerData{
		ChainID:       subChainId,
		AccountNumber: 0,
		Sequence:      0,
		PubKey:        sender.PubKey,
		Address:       subchainSender,
	}

	sigV2, err = tx.SignWithPrivKey(
		appA.Context().Context(),
		signing.SignMode(subtxconfig.SignModeHandler().DefaultMode()), signerData,
		txbuilder, sender.PrivKey, subtxconfig,
		0,
	)
	suite.Require().NoError(err)

	err = txbuilder.SetSignatures(sigV2)
	suite.Require().NoError(err)

	valSdkTx := txbuilder.GetTx()
	txjsonbz, err := subtxconfig.TxJSONEncoder()(valSdkTx)
	s.Require().NoError(err)

	regreq, err = json.Marshal(&wasmxtypes.MultiChainRegistryCallData{RegisterSubChainValidator: &wasmxtypes.RegisterSubChainValidatorRequest{
		ChainId: subChainId,
		GenTx:   txjsonbz,
	}})
	s.Require().NoError(err)

	_, err = suite.broadcastMultiChainExec([]byte(regreq), sender, registryAddress, chainId)
	s.Require().NoError(err)
	regreq, err = json.Marshal(&wasmxtypes.MultiChainRegistryCallData{InitSubChain: &wasmxtypes.InitSubChainRequest{
		ChainId: subChainId,
	}})
	s.Require().NoError(err)
	res, err := suite.broadcastMultiChainExec([]byte(regreq), sender, registryAddress, chainId)
	suite.Require().NoError(err)
	evs := appA.GetSdkEventsByType(res.Events, "init_subchain")
	suite.Require().Equal(1, len(evs))

	time.Sleep(time.Second * 3)
}

func (suite *KeeperTestSuite) TestMultiChainAtomicTx() {
	SkipFixmeTests(suite.T(), "TestMultiChainAtomicTx")
	chainId := mcfg.LEVEL0_CHAIN_ID
	config, err := mcfg.GetChainConfig(chainId)
	s.Require().NoError(err)
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)

	initBalance := sdkmath.NewInt(10_000_000_000)
	sender := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderAccount.GetPubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(denom, initBalance))
	suite.Commit()

	registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
	registryAddressStr := appA.MustAccAddressToString(registryAddress)

	initialBalance, ok := sdkmath.NewIntFromString("10000000100000000000")
	suite.Require().True(ok)
	reqlevel1 := &wasmxtypes.RegisterDefaultSubChainRequest{
		ChainBaseName:  "ptestp",
		DenomUnit:      "ppp",
		Decimals:       18,
		LevelIndex:     1,
		InitialBalance: initialBalance.BigInt(),
	}

	// create level1
	subChainId2, _ := suite.createLevel1(mcfg.LEVEL0_CHAIN_ID, reqlevel1)
	suite.Require().Equal("ptestp_1_1001-1", subChainId2)

	// get config
	qmsg := []byte(fmt.Sprintf(`{"GetSubChainConfigById":{"chainId":"%s"}}`, subChainId2))
	subChainCfgBz2 := suite.queryMultiChainCall(appA.App, qmsg, sender, registryAddressStr, subChainId2)
	suite.Require().NoError(err)
	var subChainCfg2 menc.ChainConfig
	err = json.Unmarshal(subChainCfgBz2, &subChainCfg2)
	suite.Require().NoError(err)

	// get created app
	suite.SetupSubChainApp(mcfg.LEVEL0_CHAIN_ID, subChainId2, &subChainCfg2, 3)

	newacc := suite.GetRandomAccount()
	bankAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_BANK)
	bankAddressStr := appA.MustAccAddressToString(bankAddress)

	// compose tx on level0
	suite.SetCurrentChain(chainId)
	msg := fmt.Sprintf(`{"SendCoins":{"from_address":"%s","to_address":"%s","amount":[{"denom":"%s","amount":"0x1000"}]}}`, appA.MustAccAddressToString(sender.Address), appA.MustAccAddressToString(newacc.Address), config.BaseDenom)
	txbuilder1 := suite.prepareMultiChainSubExec(appA, []byte(msg), sender, bankAddress, chainId, 0, 2)
	txbz1, err := appA.App.TxConfig().TxEncoder()(txbuilder1.GetTx())
	s.Require().NoError(err)

	// compose tx on level1
	suite.SetCurrentChain(subChainId2)
	subchainapp := suite.AppContext()
	subchain2 := suite.GetChain(subChainId2)

	msg = fmt.Sprintf(`{"SendCoins":{"from_address":"%s","to_address":"%s","amount":[{"denom":"%s","amount":"0x1000"}]}}`, subchainapp.MustAccAddressToString(sender.Address), subchainapp.MustAccAddressToString(newacc.Address), subChainCfg2.BaseDenom)
	txbuilder2 := suite.prepareMultiChainSubExec(subchainapp, []byte(msg), sender, bankAddress, subChainId2, 1, 2)
	txbz2, err := subchainapp.App.TxConfig().TxEncoder()(txbuilder2.GetTx())
	s.Require().NoError(err)

	atomictx := suite.prepareMultiChainAtomicExec(subchainapp, sender, [][]byte{txbz1, txbz2}, subChainId2, []string{subChainId2, chainId})

	txbz, err := subchain2.TxConfig.TxEncoder()(atomictx.GetTx())
	s.Require().NoError(err)

	go func() {
		time.Sleep(time.Second * 3)
		suite.SetCurrentChain(chainId)

		// we use the same atomic txbz on any chain
		res, err := appA.DeliverTxRaw(txbz)
		s.Require().NoError(err)
		s.Require().True(res.IsOK(), res.GetLog(), res.GetEvents())

		suite.SetCurrentChain(chainId)
		qmsg = []byte(fmt.Sprintf(`{"GetBalance":{"address":"%s","denom":"%s"}}`, appA.MustAccAddressToString(newacc.Address), config.BaseDenom))
		qres := suite.queryMultiChainCall(appA.App, qmsg, sender, bankAddressStr, chainId)
		balance := &banktypes.QueryBalanceResponse{}
		err = json.Unmarshal(qres, balance)
		s.Require().NoError(err)
		s.Require().Equal(sdk.NewCoin(config.BaseDenom, sdkmath.NewInt(0x1000)), *balance.Balance)
	}()

	res, err := subchainapp.DeliverTxRaw(txbz)
	s.Require().NoError(err)
	s.Require().True(res.IsOK(), res.GetLog(), res.GetEvents())

	suite.SetCurrentChain(subChainId2)
	qmsg = []byte(fmt.Sprintf(`{"GetBalance":{"address":"%s","denom":"%s"}}`, subchainapp.MustAccAddressToString(newacc.Address), subChainCfg2.BaseDenom))
	qres := suite.queryMultiChainCall(subchainapp.App, qmsg, sender, bankAddressStr, chainId)
	balance := &banktypes.QueryBalanceResponse{}
	err = json.Unmarshal(qres, balance)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(subChainCfg2.BaseDenom, sdkmath.NewInt(0x1000)), *balance.Balance)
}

func (suite *KeeperTestSuite) TestMultiChainCrossChainTx() {
	SkipFixmeTests(suite.T(), "TestMultiChainCrossChainTx")
	wasmbinFrom := testdata.WasmxCrossChain
	wasmbinTo := testdata.WasmxSimpleStorage
	chainId := mcfg.LEVEL0_CHAIN_ID
	// config, err := mcfg.GetChainConfig(chainId)
	// s.Require().NoError(err)
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)

	initBalance := sdkmath.NewInt(10_000_000_000)
	sender := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderAccount.GetPubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	senderPrefixedLevel0 := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixedLevel0, sdk.NewCoin(denom, initBalance))
	suite.Commit()

	registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
	registryAddressStr := appA.MustAccAddressToString(registryAddress)

	// create level1
	initialBalance, ok := sdkmath.NewIntFromString("10000000100000000000")
	suite.Require().True(ok)
	reqlevel1 := &wasmxtypes.RegisterDefaultSubChainRequest{
		ChainBaseName:  "ptestp",
		DenomUnit:      "ppp",
		Decimals:       18,
		LevelIndex:     1,
		InitialBalance: initialBalance.BigInt(),
	}
	subChainId2, _ := suite.createLevel1(mcfg.LEVEL0_CHAIN_ID, reqlevel1)
	suite.Require().Equal("ptestp_1_1001-1", subChainId2)

	// get config
	qmsg := []byte(fmt.Sprintf(`{"GetSubChainConfigById":{"chainId":"%s"}}`, subChainId2))
	subChainCfgBz2 := suite.queryMultiChainCall(appA.App, qmsg, sender, registryAddressStr, subChainId2)
	var subChainCfg2 menc.ChainConfig
	err := json.Unmarshal(subChainCfgBz2, &subChainCfg2)
	suite.Require().NoError(err)

	// get created app
	suite.SetupSubChainApp(mcfg.LEVEL0_CHAIN_ID, subChainId2, &subChainCfg2, 3)

	// run actual test

	// deploy from contract on level0
	suite.SetCurrentChain(chainId)
	codeIdFrom := appA.StoreCode(sender, wasmbinFrom, nil)
	contractAddressFrom := appA.InstantiateCode(sender, codeIdFrom, wasmxtypes.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"crosschain_contract":"%s"}`, wasmxtypes.ROLE_MULTICHAIN_REGISTRY))}, "wasmbinFrom", nil)

	// deploy to contract on level1
	suite.SetCurrentChain(subChainId2)
	subchainapp := suite.AppContext()
	subchain2 := suite.GetChain(subChainId2)

	codeIdTo := subchainapp.StoreCode(sender, wasmbinTo, nil)
	contractAddressTo := subchainapp.InstantiateCode(sender, codeIdTo, wasmxtypes.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"crosschain_contract":"%s"}`, wasmxtypes.ROLE_MULTICHAIN_REGISTRY))}, "wasmbinTo", nil)

	// execute cross chain transaction
	// contract message
	data := []byte(`{"set":{"key":"hello","value":"sammy"}}`)
	executeMsg := wasmxtypes.WasmxExecutionMessage{Data: data}
	msgbz, err := json.Marshal(executeMsg)
	suite.Require().NoError(err)

	// create level0 contract input
	crossreq := &types.MsgExecuteCrossChainCallRequest{
		To:           contractAddressTo.String(),
		Msg:          msgbz,
		ToChainId:    subChainId2,
		Dependencies: make([]string, 0),
		// From:            contractAddressFrom.String(),
		// FromChainId:     chainId,
	}
	crossreqbz, err := appA.App.AppCodec().MarshalJSON(crossreq)
	suite.Require().NoError(err)
	data2 := []byte(fmt.Sprintf(`{"CrossChain":%s}`, string(crossreqbz)))

	// we send this message on level0
	suite.SetCurrentChain(chainId)
	txbuilder1 := suite.prepareMultiChainSubExec(appA, data2, sender, contractAddressFrom.Bytes(), chainId, 0, 2)
	txbz1, err := appA.App.TxConfig().TxEncoder()(txbuilder1.GetTx())
	s.Require().NoError(err)

	// we send it first on level1
	suite.SetCurrentChain(subChainId2)
	atomictx := suite.prepareMultiChainAtomicExec(subchainapp, sender, [][]byte{txbz1}, subChainId2, []string{subChainId2, chainId})

	txbz, err := subchain2.TxConfig.TxEncoder()(atomictx.GetTx())
	s.Require().NoError(err)

	go func() {
		time.Sleep(time.Second * 3)
		suite.SetCurrentChain(chainId)

		// we use the same atomic txbz on any chain
		res, err := appA.DeliverTxRaw(txbz)
		s.Require().NoError(err)
		s.Require().True(res.IsOK(), res.GetLog(), res.GetEvents())
	}()

	suite.SetCurrentChain(subChainId2)
	res, err := subchainapp.DeliverTxRaw(txbz)
	s.Require().NoError(err)
	s.Require().True(res.IsOK(), res.GetLog(), res.GetEvents())

	suite.SetCurrentChain(subChainId2)
	qmsg = []byte(`{"get":{"key":"hello"}}`)
	qres := suite.queryMultiChainCall(subchainapp.App, qmsg, sender, contractAddressTo.String(), subChainId2)
	suite.Require().Equal("sammy", string(qres))
}

func (suite *KeeperTestSuite) TestMultiChainCrossChainQueryDeterministic() {
	SkipFixmeTests(suite.T(), "TestMultiChainCrossChainQueryDeterministic")
	wasmbinFrom := testdata.WasmxCrossChain
	wasmbinTo := testdata.WasmxSimpleStorage
	chainId := mcfg.LEVEL0_CHAIN_ID
	// config, err := mcfg.GetChainConfig(chainId)
	// s.Require().NoError(err)
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)

	initBalance := sdkmath.NewInt(10_000_000_000)
	sender := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderAccount.GetPubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	senderPrefixedLevel0 := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixedLevel0, sdk.NewCoin(denom, initBalance))
	suite.Commit()

	registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
	registryAddressStr := appA.MustAccAddressToString(registryAddress)

	// create level1
	initialBalance, ok := sdkmath.NewIntFromString("10000000100000000000")
	suite.Require().True(ok)
	reqlevel1 := &wasmxtypes.RegisterDefaultSubChainRequest{
		ChainBaseName:  "ptestp",
		DenomUnit:      "ppp",
		Decimals:       18,
		LevelIndex:     1,
		InitialBalance: initialBalance.BigInt(),
	}
	subChainId2, _ := suite.createLevel1(mcfg.LEVEL0_CHAIN_ID, reqlevel1)
	suite.Require().Equal("ptestp_1_1001-1", subChainId2)

	// get config
	qmsg := []byte(fmt.Sprintf(`{"GetSubChainConfigById":{"chainId":"%s"}}`, subChainId2))
	subChainCfgBz2 := suite.queryMultiChainCall(appA.App, qmsg, sender, registryAddressStr, subChainId2)
	var subChainCfg2 menc.ChainConfig
	err := json.Unmarshal(subChainCfgBz2, &subChainCfg2)
	suite.Require().NoError(err)

	// get created app
	suite.SetupSubChainApp(mcfg.LEVEL0_CHAIN_ID, subChainId2, &subChainCfg2, 3)

	// run actual test

	// deploy from contract on level0
	suite.SetCurrentChain(chainId)
	codeIdFrom := appA.StoreCode(sender, wasmbinFrom, nil)
	contractAddressFrom := appA.InstantiateCode(sender, codeIdFrom, wasmxtypes.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"crosschain_contract":"%s"}`, wasmxtypes.ROLE_MULTICHAIN_REGISTRY))}, "wasmbinFrom", nil)

	// deploy to contract on level1
	suite.SetCurrentChain(subChainId2)
	subchainapp := suite.AppContext()
	subchain2 := suite.GetChain(subChainId2)

	codeIdTo := subchainapp.StoreCode(sender, wasmbinTo, nil)
	contractAddressTo := subchainapp.InstantiateCode(sender, codeIdTo, wasmxtypes.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"crosschain_contract":"%s"}`, wasmxtypes.ROLE_MULTICHAIN_REGISTRY))}, "wasmbinTo", nil)

	// store an initial value
	data := []byte(`{"set":{"key":"hello","value":"brian"}}`)
	subchainapp.ExecuteContract(sender, contractAddressTo, wasmxtypes.WasmxExecutionMessage{Data: data}, nil, nil)

	// execute cross chain deterministic query
	// contract message
	data = []byte(`{"testGetWithStore":{"key":"hello"}}`)
	executeMsg := wasmxtypes.WasmxExecutionMessage{Data: data}
	msgbz, err := json.Marshal(executeMsg)
	suite.Require().NoError(err)

	// create level0 contract input
	crossreq := &types.MsgExecuteCrossChainCallRequest{
		To:           contractAddressTo.String(),
		Msg:          msgbz,
		ToChainId:    subChainId2,
		Dependencies: make([]string, 0),
		// From:            contractAddressFrom.String(),
		// FromChainId:     chainId,
	}
	crossreqbz, err := appA.App.AppCodec().MarshalJSON(crossreq)
	suite.Require().NoError(err)
	data2 := []byte(fmt.Sprintf(`{"CrossChainQuery":%s}`, string(crossreqbz)))

	// we send this message on level0
	suite.SetCurrentChain(chainId)
	txbuilder1 := suite.prepareMultiChainSubExec(appA, data2, sender, contractAddressFrom.Bytes(), chainId, 0, 2)
	txbz1, err := appA.App.TxConfig().TxEncoder()(txbuilder1.GetTx())
	s.Require().NoError(err)

	// we send it first on level1
	suite.SetCurrentChain(subChainId2)
	atomictx := suite.prepareMultiChainAtomicExec(subchainapp, sender, [][]byte{txbz1}, subChainId2, []string{subChainId2, chainId})

	txbz, err := subchain2.TxConfig.TxEncoder()(atomictx.GetTx())
	s.Require().NoError(err)

	go func() {
		time.Sleep(time.Second * 3)
		suite.SetCurrentChain(chainId)

		// we use the same atomic txbz on any chain
		res, err := appA.DeliverTxRaw(txbz)
		s.Require().NoError(err)
		s.Require().True(res.IsOK(), res.GetLog(), res.GetEvents())
	}()

	suite.SetCurrentChain(subChainId2)
	res, err := subchainapp.DeliverTxRaw(txbz)
	s.Require().NoError(err)
	s.Require().True(res.IsOK(), res.GetLog(), res.GetEvents())

	txmsgdata, err := mcodec.TxMsgDataFromBz(res.Data)
	s.Require().NoError(err)
	s.Require().Equal(1, len(txmsgdata.MsgResponses))

	var sdkmsg sdk.Msg
	err = subchainapp.Chain.Codec.UnpackAny(txmsgdata.MsgResponses[0], &sdkmsg)
	s.Require().NoError(err)
	atomicres := sdkmsg.(*types.MsgExecuteAtomicTxResponse)
	s.Require().Equal(1, len(atomicres.Results))
	s.Require().Equal(abci.CodeTypeOK, atomicres.Results[0].Code)

	suite.SetCurrentChain(subChainId2)
	qmsg = []byte(`{"get":{"key":"hello"}}`)
	qres := suite.queryMultiChainCall(subchainapp.App, qmsg, sender, contractAddressTo.String(), subChainId2)
	suite.Require().Equal("brian", string(qres))
}

func (suite *KeeperTestSuite) TestMultiChainLevelsTx() {
	SkipFixmeTests(suite.T(), "TestMultiChainLevelsTx")
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)

	initBalance := sdkmath.NewInt(10_000_000_000)
	sender := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderAccount.GetPubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	senderPrefixedLevel0 := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixedLevel0, sdk.NewCoin(denom, initBalance))
	suite.Commit()

	registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
	registryAddressStr := appA.MustAccAddressToString(registryAddress)
	initialBalance, ok := sdkmath.NewIntFromString("10000000100000000000")
	suite.Require().True(ok)
	reqlevel1 := &wasmxtypes.RegisterDefaultSubChainRequest{
		ChainBaseName:  "ptestp",
		DenomUnit:      "ppp",
		Decimals:       18,
		LevelIndex:     1,
		InitialBalance: initialBalance.BigInt(),
	}
	subChainId1, res := suite.createLevel1(chainId, reqlevel1)
	suite.Require().Equal("ptestp_1_1001-1", subChainId1)

	// we expect 1 level2 was created
	evs := appA.GetSdkEventsByType(res.Events, "init_subchain")
	suite.Require().Equal(2, len(evs))
	level2ChainId := appA.GetAttributeValueFromEvent(evs[1], "chain_id")
	suite.Require().Equal("leveln_2_2001-1", level2ChainId)

	// check level1 has 3 chain
	qmsg := []byte(`{"GetSubChainIdsByLevel":{"level":1}}`)
	respbz := suite.queryMultiChainCall(appA.App, qmsg, sender, registryAddressStr, chainId)
	var chainIds []string
	err := json.Unmarshal(respbz, &chainIds)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(chainIds))

	// check level2 has 1 chain
	qmsg = []byte(`{"GetSubChainIdsByLevel":{"level":2}}`)
	respbz = suite.queryMultiChainCall(appA.App, qmsg, sender, registryAddressStr, chainId)
	var chainIds2 []string
	err = json.Unmarshal(respbz, &chainIds2)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(chainIds2))
	suite.Require().Equal(level2ChainId, chainIds2[0])

	reqlevel2 := &wasmxtypes.RegisterDefaultSubChainRequest{
		ChainBaseName:  "qtestq",
		DenomUnit:      "qqq",
		Decimals:       18,
		LevelIndex:     1,
		InitialBalance: initialBalance.BigInt(),
	}
	subChainId2, res := suite.createLevel1(chainId, reqlevel2)
	suite.Require().Equal("qtestq_10002-1", subChainId2)

	// we expect 1 level2 & level3 were created
	evs = appA.GetSdkEventsByType(res.Events, "init_subchain")
	suite.Require().Equal(3, len(evs))
	suite.Require().Equal(subChainId2, appA.GetAttributeValueFromEvent(evs[0], "chain_id"))
	level2ChainId = appA.GetAttributeValueFromEvent(evs[1], "chain_id")
	suite.Require().Equal("leveln_20002-1", level2ChainId)
	level3ChainId := appA.GetAttributeValueFromEvent(evs[2], "chain_id")
	suite.Require().Equal("leveln_30001-1", level3ChainId)

	// check level1 has 3 chain
	qmsg = []byte(`{"GetSubChainIdsByLevel":{"level":1}}`)
	respbz = suite.queryMultiChainCall(appA.App, qmsg, sender, registryAddressStr, chainId)
	err = json.Unmarshal(respbz, &chainIds)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(chainIds))

	// check level2 has 2 chain
	qmsg = []byte(`{"GetSubChainIdsByLevel":{"level":2}}`)
	respbz = suite.queryMultiChainCall(appA.App, qmsg, sender, registryAddressStr, chainId)
	err = json.Unmarshal(respbz, &chainIds2)
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(chainIds2))
	suite.Require().Equal(level2ChainId, chainIds2[1])

	// check level3 has 1 chain
	qmsg = []byte(`{"GetSubChainIdsByLevel":{"level":3}}`)
	respbz = suite.queryMultiChainCall(appA.App, qmsg, sender, registryAddressStr, chainId)
	err = json.Unmarshal(respbz, &chainIds2)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(chainIds2))
	suite.Require().Equal(level3ChainId, chainIds2[0])

	var level wasmxtypes.QueryGetCurrentLevelResponse

	// verify that level2 chain contains level1 chain ids
	suite.createSubChainApp(sender, registryAddressStr, chainId, level2ChainId, 4)
	suite.SetCurrentChain(level2ChainId)
	subchainapp := suite.AppContext()
	registryAddressLevel2Str := subchainapp.MustAccAddressToString(registryAddress)
	qmsg = []byte(`{"GetCurrentLevel":{}}`)
	respbz = suite.queryMultiChainCall(subchainapp.App, qmsg, sender, registryAddressLevel2Str, level2ChainId)
	err = json.Unmarshal(respbz, &level)
	suite.Require().NoError(err)
	suite.Require().Equal(int32(2), level.Level)

	qmsg = []byte(`{"GetSubChainIdsByLevel":{"level":1}}`)
	respbz = suite.queryMultiChainCall(subchainapp.App, qmsg, sender, registryAddressLevel2Str, level2ChainId)
	err = json.Unmarshal(respbz, &chainIds2)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(chainIds2))
	suite.Require().Equal(subChainId2, chainIds2[0])

	// verify that level3 chain contains level2 chain id
	suite.SetCurrentChain(chainId) // for the below query
	suite.createSubChainApp(sender, registryAddressStr, chainId, level3ChainId, 5)
	suite.SetCurrentChain(level3ChainId)
	subchainapp = suite.AppContext()
	registryAddressLevel3Str := subchainapp.MustAccAddressToString(registryAddress)
	qmsg = []byte(`{"GetCurrentLevel":{}}`)
	respbz = suite.queryMultiChainCall(subchainapp.App, qmsg, sender, registryAddressLevel3Str, level3ChainId)
	err = json.Unmarshal(respbz, &level)
	suite.Require().NoError(err)
	suite.Require().Equal(int32(3), level.Level)

	qmsg = []byte(`{"GetSubChainIdsByLevel":{"level":2}}`)
	respbz = suite.queryMultiChainCall(subchainapp.App, qmsg, sender, registryAddressLevel3Str, level3ChainId)
	err = json.Unmarshal(respbz, &chainIds2)
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(chainIds2))
	suite.Require().Equal(level2ChainId, chainIds2[0])
}

func (suite *KeeperTestSuite) TestMultiChainLevelsQuery() {
	SkipFixmeTests(suite.T(), "TestMultiChainLevelsQuery")
	chainId := mcfg.MYTHOS_CHAIN_ID_TEST
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)

	initBalance := sdkmath.NewInt(10_000_000_000)
	sender := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderAccount.GetPubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	senderPrefixedLevel0 := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixedLevel0, sdk.NewCoin(denom, initBalance))
	suite.Commit()

	registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
	registryAddressStr := appA.MustAccAddressToString(registryAddress)
	initialBalance, ok := sdkmath.NewIntFromString("10000000100000000000")
	suite.Require().True(ok)
	reqlevel1 := &wasmxtypes.RegisterDefaultSubChainRequest{
		ChainBaseName:  "ptestp",
		DenomUnit:      "ppp",
		Decimals:       18,
		LevelIndex:     1,
		InitialBalance: initialBalance.BigInt(),
	}
	subChainId1, res := suite.createLevel1(chainId, reqlevel1)
	suite.Require().Equal("ptestp_1_1001-1", subChainId1)

	// we expect 1 level2 was created
	evs := appA.GetSdkEventsByType(res.Events, "init_subchain")
	level2ChainId := appA.GetAttributeValueFromEvent(evs[1], "chain_id")

	suite.SetCurrentChain(chainId)
	suite.createSubChainApp(sender, registryAddressStr, chainId, subChainId1, 4)
	suite.SetCurrentChain(chainId)
	suite.createSubChainApp(sender, registryAddressStr, chainId, level2ChainId, 5)

	// deploy erc20 rollup
	newacc := suite.GetRandomAccount()
	erc20rbin := precompiles.GetPrecompileByLabel(nil, wasmxtypes.ERC20_ROLLUP_v001)
	erc20data := &wasmxtypes.Erc20RollupTokenInstantiate{
		Admins:      []string{},
		Minters:     []string{},
		Name:        "erc20rollup",
		Symbol:      "erc20symbol",
		Decimals:    18,
		BaseDenom:   "appp",
		SubChainIds: []string{},
	}

	// level 1 deploy erc20
	suite.SetCurrentChain(subChainId1)
	subchainapp1 := suite.AppContext()
	codeIdErc20R1 := subchainapp1.StoreCode(sender, erc20rbin, nil)
	senderAddr1 := subchainapp1.MustAccAddressToString(sender.Address)
	erc20data.Minters = []string{senderAddr1}
	erc20databz, err := json.Marshal(erc20data)
	suite.Require().NoError(err)
	addressErc20R1 := subchainapp1.InstantiateCode(sender, codeIdErc20R1, wasmxtypes.WasmxExecutionMessage{Data: erc20databz}, "addressErc20R1", nil)
	// mint tokens on level1
	newaccAddr1 := subchainapp1.MustAccAddressToString(newacc.Address)
	// subchainapp1.ExecuteContract(sender, addressErc20R1, wasmxtypes.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"mint":{"to":"%s","value":"1000000000"}}`, newaccAddr1))}, nil, nil)
	msg := []byte(fmt.Sprintf(`{"mint":{"to":"%s","value":"1000000000"}}`, newaccAddr1))
	suite.broadcastMultiChainExec(msg, sender, addressErc20R1.Bytes(), subChainId1)
	// get totalsupply on level1
	qmsg := `{"totalSupply":{}}`
	qres := suite.queryMultiChainCall(subchainapp1.App, []byte(qmsg), sender, addressErc20R1.String(), subChainId1)
	supply := &wasmxtypes.MsgTotalSupplyResponse{}
	err = json.Unmarshal(qres, supply)
	s.Require().NoError(err)
	s.Require().Equal(erc20data.Symbol, supply.Supply.Denom)
	s.Require().Equal(sdkmath.NewInt(1000000000), supply.Supply.Amount)

	// get balance on level1
	qmsg = fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, newaccAddr1)
	qres = suite.queryMultiChainCall(subchainapp1.App, []byte(qmsg), sender, addressErc20R1.String(), subChainId1)
	balance := &wasmxtypes.MsgBalanceOfResponse{}
	err = json.Unmarshal(qres, balance)
	s.Require().NoError(err)
	s.Require().Equal(erc20data.Symbol, balance.Balance.Denom)
	s.Require().Equal(sdkmath.NewInt(1000000000), balance.Balance.Amount)

	// level 2 deploy erc20
	suite.SetCurrentChain(level2ChainId)
	subchainapp2 := suite.AppContext()
	senderPrefixed := subchainapp2.AccBech32Codec().BytesToAccAddressPrefixed(sender.Address)

	subchainapp2.Faucet.Fund(subchainapp2.Context(), senderPrefixed, sdk.NewCoin(subchainapp2.Chain.Config.BaseDenom, initBalance))
	suite.Commit()
	codeIdErc20R2 := subchainapp2.StoreCode(sender, erc20rbin, nil)
	senderAddr2 := subchainapp2.MustAccAddressToString(sender.Address)
	erc20data.Minters = []string{senderAddr2}
	erc20data.BaseDenom = "alvl2"
	erc20data.SubChainIds = []string{subChainId1}
	erc20databz, err = json.Marshal(erc20data)
	suite.Require().NoError(err)
	addressErc20R2 := subchainapp2.InstantiateCode(sender, codeIdErc20R2, wasmxtypes.WasmxExecutionMessage{Data: erc20databz}, "addressErc20R2", nil)
	// mint tokens on level2
	newaccAddr2 := subchainapp2.MustAccAddressToString(newacc.Address)
	// subchainapp1.ExecuteContract(sender, addressErc20R2, wasmxtypes.WasmxExecutionMessage{Data: []byte(fmt.Sprintf(`{"mint":{"to":"%s","value":"2000000000"}}`, newaccAddr2))}, nil, nil)
	msg = []byte(fmt.Sprintf(`{"mint":{"to":"%s","value":"2000000000"}}`, newaccAddr2))
	suite.broadcastMultiChainExec(msg, sender, addressErc20R2.Bytes(), level2ChainId)
	// get totalsupply on level2
	qmsg = `{"totalSupply":{}}`
	qres = suite.queryMultiChainCall(subchainapp2.App, []byte(qmsg), sender, addressErc20R2.String(), level2ChainId)
	err = json.Unmarshal(qres, supply)
	s.Require().NoError(err)
	s.Require().Equal(erc20data.Symbol, supply.Supply.Denom)
	s.Require().Equal(sdkmath.NewInt(2000000000), supply.Supply.Amount)

	// get balance on level2
	qmsg = fmt.Sprintf(`{"balanceOf":{"owner":"%s"}}`, newaccAddr2)
	qres = suite.queryMultiChainCall(subchainapp2.App, []byte(qmsg), sender, addressErc20R2.String(), level2ChainId)
	err = json.Unmarshal(qres, balance)
	s.Require().NoError(err)
	s.Require().Equal(erc20data.Symbol, balance.Balance.Denom)
	s.Require().Equal(sdkmath.NewInt(2000000000), balance.Balance.Amount)

	// non-deterministic queries!

	// get cross chain supply on level2
	qmsg = `{"totalSupplyCrossChainNonDeterministic":{}}`
	qres = suite.queryMultiChainCall(subchainapp2.App, []byte(qmsg), sender, addressErc20R2.String(), level2ChainId)
	supplyCrossChain := &wasmxtypes.MsgTotalSupplyCrossChainResponse{}
	err = json.Unmarshal(qres, supplyCrossChain)
	s.Require().NoError(err)
	s.Require().Equal(erc20data.Symbol, supplyCrossChain.Supply.Denom)
	s.Require().Equal(sdkmath.NewInt(3000000000), supplyCrossChain.Supply.Amount)
	s.Require().Equal(2, len(supplyCrossChain.Chains))
	s.Require().Equal(sdkmath.NewInt(2000000000), supplyCrossChain.Chains[0].Value)
	s.Require().Equal(level2ChainId, supplyCrossChain.Chains[0].ChainId)
	s.Require().Equal(sdkmath.NewInt(1000000000), supplyCrossChain.Chains[1].Value)
	s.Require().Equal(subChainId1, supplyCrossChain.Chains[1].ChainId)

	// get cross chain balance on level2
	qmsg = fmt.Sprintf(`{"balanceOfCrossChainNonDeterministic":{"owner":"%s"}}`, newaccAddr2)
	qres = suite.queryMultiChainCall(subchainapp2.App, []byte(qmsg), sender, addressErc20R2.String(), level2ChainId)
	balanceCrossChain := &wasmxtypes.MsgBalanceOfCrossChainResponse{}
	err = json.Unmarshal(qres, balanceCrossChain)
	s.Require().NoError(err)
	s.Require().Equal(erc20data.Symbol, balanceCrossChain.Balance.Denom)
	s.Require().Equal(sdkmath.NewInt(3000000000), balanceCrossChain.Balance.Amount)

	// ensure we do not run nondeterministic crosschain queries as part of a tx
	_, err = suite.queryMultiChainCallWithWrongModeExec(subchainapp2.App, []byte(qmsg), sender, addressErc20R2.String(), level2ChainId)
	s.Require().Error(err, "")
}

func (suite *KeeperTestSuite) createLevel1(chainId string, req *wasmxtypes.RegisterDefaultSubChainRequest) (string, *abci.ExecTxResult) {
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)

	initBalance := sdkmath.NewInt(10_000_000_000)
	sender := simulation.Account{
		PrivKey: chain.SenderPrivKey,
		PubKey:  chain.SenderAccount.GetPubKey(),
		Address: chain.SenderAccount.GetAddress(),
	}

	appA := s.AppContext()
	denom := appA.Chain.Config.BaseDenom
	senderPrefixed := appA.BytesToAccAddressPrefixed(sender.Address)
	appA.Faucet.Fund(appA.Context(), senderPrefixed, sdk.NewCoin(denom, initBalance))
	suite.Commit()

	registryAddress := wasmxtypes.AccAddressFromHex(wasmxtypes.ADDR_MULTICHAIN_REGISTRY)
	registryAddressStr := appA.MustAccAddressToString(registryAddress)

	// create new subchain genesis registry
	regreq, err := json.Marshal(&wasmxtypes.MultiChainRegistryCallData{RegisterDefaultSubChain: req})
	suite.Require().NoError(err)

	res, err := suite.broadcastMultiChainExec(regreq, sender, registryAddress, chainId)
	suite.Require().NoError(err)
	evs := appA.GetSdkEventsByType(res.Events, "register_subchain")
	suite.Require().Equal(1, len(evs))
	subChainId := appA.GetAttributeValueFromEvent(evs[0], "chain_id")

	// create genTx data to sign - call to level0
	// buildGenTx query
	// valTokens := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
	valTokens, ok := sdkmath.NewIntFromString("10000000000000000000")
	suite.Require().True(ok)
	validMsg := stakingtypes.MsgCreateValidator{
		// TODO fix as-json.parse when description contains empty strings
		Description:       stakingtypes.NewDescription("moniker1", "id", "website", "security", "details"),
		Commission:        stakingtypes.NewCommissionRates(sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec()),
		MinSelfDelegation: sdkmath.OneInt(),
		Value:             sdk.NewCoin(denom, valTokens), // denom will be changed by level0 anyway
	}

	genTxInfo, err := json.Marshal(&wasmxtypes.QueryBuildGenTxRequest{
		ChainId: subChainId,
		Msg:     validMsg,
	})
	suite.Require().NoError(err)
	paramBz, err := json.Marshal(&wasmxtypes.ActionParam{Key: "message", Value: string(genTxInfo)})
	suite.Require().NoError(err)

	msg := []byte(fmt.Sprintf(`{"execute":{"action": {"type": "buildGenTx", "params": [%s],"event":null}}}`, string(paramBz)))
	txbz := suite.queryMultiChainCall(appA.App, msg, sender, wasmxtypes.ROLE_CONSENSUS, chainId)
	msg = []byte(fmt.Sprintf(`{"GetSubChainConfigById":{"chainId":"%s"}}`, subChainId))
	chaincfgbz := suite.queryMultiChainCall(appA.App, msg, sender, registryAddressStr, chainId)

	var subchainConfig menc.ChainConfig
	err = json.Unmarshal(chaincfgbz, &subchainConfig)
	suite.Require().NoError(err)

	// create a temporary app, to sign the transaction
	// must be in a different directory than when the subchain is instantiated later
	_, appCreator := multichain.CreateMockAppCreator(suite.WasmVmMeta, app.NewAppCreator, app.DefaultNodeHome+"temp", nil)
	iapp := appCreator(subChainId, &subchainConfig)
	subchainapp := iapp.(*app.App)
	subtxconfig := subchainapp.TxConfig()

	sigV2 := signing.SignatureV2{
		PubKey: sender.PubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode(subtxconfig.SignModeHandler().DefaultMode()),
			Signature: nil,
		},
		Sequence: 0,
	}

	sdktx, err := subtxconfig.TxJSONDecoder()(txbz)
	suite.Require().NoError(err)
	txbuilder, err := subtxconfig.WrapTxBuilder(sdktx)
	suite.Require().NoError(err)

	subchainSender, err := subtxconfig.SigningContext().AddressCodec().BytesToString(sender.Address)
	suite.Require().NoError(err)

	err = txbuilder.SetSignatures(sigV2)
	suite.Require().NoError(err)

	signerData := authsigning.SignerData{
		ChainID:       subChainId,
		AccountNumber: 0,
		Sequence:      0,
		PubKey:        sender.PubKey,
		Address:       subchainSender,
	}

	sigV2, err = tx.SignWithPrivKey(
		appA.Context().Context(),
		signing.SignMode(subtxconfig.SignModeHandler().DefaultMode()), signerData,
		txbuilder, sender.PrivKey, subtxconfig,
		0,
	)
	suite.Require().NoError(err)

	err = txbuilder.SetSignatures(sigV2)
	suite.Require().NoError(err)

	valSdkTx := txbuilder.GetTx()
	txjsonbz, err := subtxconfig.TxJSONEncoder()(valSdkTx)
	s.Require().NoError(err)

	regreq, err = json.Marshal(&wasmxtypes.MultiChainRegistryCallData{RegisterSubChainValidator: &wasmxtypes.RegisterSubChainValidatorRequest{
		ChainId: subChainId,
		GenTx:   txjsonbz,
	}})
	s.Require().NoError(err)

	_, err = suite.broadcastMultiChainExec([]byte(regreq), sender, registryAddress, chainId)
	s.Require().NoError(err)
	regreq, err = json.Marshal(&wasmxtypes.MultiChainRegistryCallData{InitSubChain: &wasmxtypes.InitSubChainRequest{
		ChainId: subChainId,
	}})
	s.Require().NoError(err)
	res, err = suite.broadcastMultiChainExec([]byte(regreq), sender, registryAddress, chainId)
	suite.Require().NoError(err)
	evs = appA.GetSdkEventsByType(res.Events, "init_subchain")
	suite.Require().GreaterOrEqual(len(evs), 1)
	subChainIdInit := appA.GetAttributeValueFromEvent(evs[0], "chain_id")
	suite.Require().Equal(subChainId, subChainIdInit)
	return subChainId, res
}

func (suite *KeeperTestSuite) createSubChainApp(sender simulation.Account, registryAddressStr string, chainId string, subChainId string, index int32) mcfg.MythosApp {
	appA := s.AppContext()
	msg := []byte(fmt.Sprintf(`{"GetSubChainConfigById":{"chainId":"%s"}}`, subChainId))
	chaincfgbz := suite.queryMultiChainCall(appA.App, msg, sender, registryAddressStr, chainId)

	var subchainConfig menc.ChainConfig
	err := json.Unmarshal(chaincfgbz, &subchainConfig)
	suite.Require().NoError(err)

	multichainapp, err := mcfg.GetMultiChainApp(appA.App.GetGoContextParent())
	suite.Require().NoError(err)

	var subchainapp mcfg.MythosApp
	isubchainapp, err := multichainapp.GetApp(subChainId)
	if err == nil {
		subchainapp = isubchainapp.(mcfg.MythosApp)
	} else {
		subchainapp = multichainapp.NewApp(subChainId, &subchainConfig)
	}

	// get created app
	suite.SetupSubChainApp(chainId, subChainId, &subchainConfig, index)

	return subchainapp
}

func (suite *KeeperTestSuite) queryMultiChainCall(mapp *app.App, msg []byte, sender simulation.Account, contractAddress string, chainId string) []byte {
	msgwrap := &wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgbz, err := json.Marshal(msgwrap)
	suite.Require().NoError(err)
	appA := suite.AppContext()
	multimsg := &types.QueryContractCallRequest{
		MultiChainId: chainId,
		Sender:       appA.MustAccAddressToString(sender.Address),
		Address:      contractAddress,
		QueryData:    msgbz,
	}
	// TODO we should use network.QueryMultiChain instead; baseapp.Query sets the proper ExecMode
	ctx := appA.Context().WithExecMode(sdk.ExecModeQuery)
	res, err := mapp.NetworkKeeper.ContractCall(ctx, multimsg)
	suite.Require().NoError(err)

	wres := &wasmxtypes.WasmxExecutionMessage{}
	err = json.Unmarshal(res.Data, wres)
	suite.Require().NoError(err)
	return wres.Data
}

func (suite *KeeperTestSuite) queryMultiChainCallWithWrongModeExec(mapp *app.App, msg []byte, sender simulation.Account, contractAddress string, chainId string) (*types.QueryContractCallResponse, error) {
	msgwrap := &wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgbz, err := json.Marshal(msgwrap)
	suite.Require().NoError(err)
	appA := suite.AppContext()
	multimsg := &types.QueryContractCallRequest{
		MultiChainId: chainId,
		Sender:       appA.MustAccAddressToString(sender.Address),
		Address:      contractAddress,
		QueryData:    msgbz,
	}
	// TODO we should use network.QueryMultiChain instead; baseapp.Query sets the proper ExecMode
	ctx := appA.Context().WithExecMode(sdk.ExecModeFinalize)
	return mapp.NetworkKeeper.ContractCall(ctx, multimsg)
}

func (suite *KeeperTestSuite) queryMultiChainCall__(mapp *app.App, msg []byte, sender simulation.Account, contractAddress sdk.AccAddress, chainId string) []byte {
	goctx1 := context.Background()
	_, conn1 := suite.GrpcClient(goctx1, "bufnet1", mapp)
	defer conn1.Close()
	queryClient := types.NewQueryClient(conn1)
	appA := suite.AppContext()
	msgwrap := &wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgbz, err := json.Marshal(msgwrap)
	suite.Require().NoError(err)
	multimsg := &types.QueryContractCallRequest{
		MultiChainId: chainId,
		Sender:       appA.MustAccAddressToString(sender.Address),
		Address:      appA.MustAccAddressToString(contractAddress),
		QueryData:    msgbz,
	}
	res, err := queryClient.ContractCall(
		context.Background(),
		multimsg,
	)
	suite.Require().NoError(err)
	return res.Data
}

func (suite *KeeperTestSuite) queryMultiChainCall_(msg []byte, sender simulation.Account, contractAddress sdk.AccAddress, chainId string) []byte {
	// clientCtx, err := client.GetClientQueryContext(cmd)
	// suite.Require().NoError(err)
	clientCtx := client.Context{}
	queryClient := types.NewQueryClient(clientCtx)
	appA := suite.AppContext()
	msgwrap := &wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgbz, err := json.Marshal(msgwrap)
	suite.Require().NoError(err)
	multimsg := &types.QueryContractCallRequest{
		MultiChainId: chainId,
		Sender:       appA.MustAccAddressToString(sender.Address),
		Address:      appA.MustAccAddressToString(contractAddress),
		QueryData:    msgbz,
	}
	res, err := queryClient.ContractCall(
		context.Background(),
		multimsg,
	)
	suite.Require().NoError(err)
	return res.Data
}

func (suite *KeeperTestSuite) composeMultiChainTx(msg []byte, sender simulation.Account, contractAddress sdk.AccAddress, chainId string) sdk.Msg {
	appA := s.AppContext()
	msgwrap := &wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgbz, err := json.Marshal(msgwrap)
	suite.Require().NoError(err)
	msgexec := &wasmxtypes.MsgExecuteContract{
		Sender:   appA.MustAccAddressToString(sender.Address),
		Contract: appA.MustAccAddressToString(contractAddress),
		Msg:      msgbz,
	}
	msgAny, err := codectypes.NewAnyWithValue(msgexec)
	suite.Require().NoError(err)
	multimsg := &types.MsgMultiChainWrap{
		MultiChainId: chainId,
		Sender:       appA.MustAccAddressToString(sender.Address),
		Data:         msgAny,
	}
	return multimsg
}

func (suite *KeeperTestSuite) broadcastMultiChainExec(msg []byte, sender simulation.Account, contractAddress sdk.AccAddress, chainId string) (*abci.ExecTxResult, error) {
	appA := s.AppContext()
	multimsg := suite.composeMultiChainTx(msg, sender, contractAddress, chainId)
	// initializing chains is expensive ~6mil
	gasLimit := uint64(30000000)
	resp, err := appA.BroadcastTxAsync(sender, []sdk.Msg{multimsg}, &gasLimit, nil, "")
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (suite *KeeperTestSuite) prepareMultiChainSubExec(appCtx wasmxtesting.AppContext, msg []byte, sender simulation.Account, contractAddress sdk.AccAddress, chainId string, subtxindex int32, subtxcount int32) client.TxBuilder {
	multimsg := suite.composeMultiChainTx(msg, sender, contractAddress, chainId)
	txBuilder := appCtx.PrepareCosmosSdkTxBuilder([]sdk.Msg{multimsg}, nil, nil, "")
	appCtx.SetMultiChainExtensionOptions(txBuilder, chainId, subtxindex, subtxcount)
	return appCtx.SignCosmosSdkTx(txBuilder, sender)
}

func (suite *KeeperTestSuite) prepareMultiChainAtomicExec(appCtx wasmxtesting.AppContext, sender simulation.Account, txsbz [][]byte, leaderChainId string, chainIds []string) client.TxBuilder {
	subtxmsg := &types.MsgExecuteAtomicTxRequest{
		Txs:    txsbz,
		Sender: sender.Address.Bytes(),
	}

	txBuilder := appCtx.PrepareCosmosSdkTxBuilder([]sdk.Msg{subtxmsg}, nil, nil, "")
	appCtx.SetMultiChainAtomicExtensionOptions(txBuilder, chainIds, leaderChainId)
	// we do not need to sign atomic transactions
	// return appCtx.SignCosmosSdkTx(txBuilder, sender)
	return txBuilder
}
