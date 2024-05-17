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

	"mythos/v1/app"
	mcodec "mythos/v1/codec"
	mcfg "mythos/v1/config"
	menc "mythos/v1/encoding"
	ibctesting "mythos/v1/testutil/ibc"
	cosmosmodtypes "mythos/v1/x/cosmosmod/types"

	// networkserver "mythos/v1/x/network/server"
	"mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

func (suite *KeeperTestSuite) TestMultiChainExecMythos() {
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

	newaccStr := appA.MustAccAddressToString(newacc.Address)

	msg := fmt.Sprintf(`{"SendCoins":{"from_address":"%s","to_address":"%s","amount":[{"denom":"%s","amount":"0x1000"}]}}`, appA.MustAccAddressToString(sender.Address), newaccStr, config.BaseDenom)
	suite.broadcastMultiChainExec([]byte(msg), sender, bankAddress, chainId)

	qmsg := fmt.Sprintf(`{"GetBalance":{"address":"%s","denom":"%s"}}`, newaccStr, config.BaseDenom)
	res := suite.queryMultiChainCall(appA.App, []byte(qmsg), sender, bankAddress, chainId)

	balance := &banktypes.QueryBalanceResponse{}
	err = json.Unmarshal(res, balance)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(denom, sdkmath.NewInt(0x1000)), *balance.Balance)
	// TODO try again query client - this time with conn.defer() in the test
}

func (suite *KeeperTestSuite) TestMultiChainExecLevel0() {
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
	newaccStr := appA.MustAccAddressToString(newacc.Address)

	msg := fmt.Sprintf(`{"SendCoins":{"from_address":"%s","to_address":"%s","amount":[{"denom":"%s","amount":"0x1000"}]}}`, appA.MustAccAddressToString(sender.Address), newaccStr, config.BaseDenom)
	suite.broadcastMultiChainExec([]byte(msg), sender, bankAddress, chainId)

	qmsg := fmt.Sprintf(`{"GetBalance":{"address":"%s","denom":"%s"}}`, newaccStr, config.BaseDenom)
	res := suite.queryMultiChainCall(appA.App, []byte(qmsg), sender, bankAddress, chainId)

	balance := &banktypes.QueryBalanceResponse{}
	err = json.Unmarshal(res, balance)
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(denom, sdkmath.NewInt(0x1000)), *balance.Balance)
	// TODO try again query client - this time with conn.defer() in the test
}

func (suite *KeeperTestSuite) TestMultiChainInit() {
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
	encoding := menc.MakeEncodingConfig(&subChainConfig)
	addrCodec := mcodec.MustUnwrapAccBech32Codec(encoding.InterfaceRegistry.SigningContext().AddressCodec())
	valAddrCodec := mcodec.MustUnwrapValBech32Codec(encoding.InterfaceRegistry.SigningContext().ValidatorAddressCodec())
	valTokens := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)

	genesisAccs := []cosmosmodtypes.GenesisAccount{}
	balances := []banktypes.Balance{}
	_, genesisState, err := ibctesting.BuildGenesisData(&tmtypes.ValidatorSet{}, genesisAccs, subChainId, subChainConfig, 10, balances)
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

	valAddr := addrCodec.BytesToAccAddressPrefixed(sdk.ValAddress(chain.Vals.Validators[0].Address))
	valAddr = addrCodec.BytesToAccAddressPrefixed(sender.Address.Bytes())
	valStr, err := valAddrCodec.BytesToString(sdk.ValAddress(valAddr.Bytes()))

	peer := fmt.Sprintf("%s@/ip4/127.0.0.1/tcp/5001/p2p/12D3KooWJdKwTq9QcARdPuk4QBibP8MxBV7Q8xC7JRMSXWuvZBtD", valAddr.String())

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

	valTxBuilder := appA.PrepareCosmosSdkTxBuilder(sender, []sdk.Msg{validMsg}, nil, &subchainGasPrices, peer)
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
		GenTxs:   [][]byte{txjsonbz},
		Balances: sdk.Coins{sdk.NewCoin(subChainConfig.BaseDenom, initBalance)},
	}
	regreqBz, err := json.Marshal(regreq)
	suite.Require().NoError(err)
	msg := fmt.Sprintf(`{"RegisterSubChain":%s}`, string(regreqBz))

	res, err := suite.broadcastMultiChainExec([]byte(msg), sender, registryAddress, chainId)
	suite.Require().NoError(err)

	msg = fmt.Sprintf(`{"InitSubChain":{"chainId":"%s"}}`, subChainId)
	res, err = suite.broadcastMultiChainExec([]byte(msg), sender, registryAddress, chainId)
	suite.Require().NoError(err)
	evs := appA.GetSdkEventsByType(res.Events, "init_subchain")
	suite.Require().Equal(1, len(evs))

	time.Sleep(time.Second * 3)

	// test restarting the node by starting the parent chain
	// err = networkserver.StartNode(appA.App, appA.App.Logger(), appA.App.GetNetworkKeeper())
	// suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) queryMultiChainCall(mapp *app.App, msg []byte, sender simulation.Account, contractAddress sdk.AccAddress, chainId string) []byte {
	msgwrap := &wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgbz, err := json.Marshal(msgwrap)
	suite.Require().NoError(err)
	appA := suite.AppContext()
	multimsg := &types.QueryContractCallRequest{
		MultiChainId: chainId,
		Sender:       appA.MustAccAddressToString(sender.Address),
		Address:      appA.MustAccAddressToString(contractAddress),
		QueryData:    msgbz,
	}
	res, err := mapp.NetworkKeeper.ContractCall(appA.Context(), multimsg)
	suite.Require().NoError(err)

	wres := &wasmxtypes.WasmxExecutionMessage{}
	err = json.Unmarshal(res.Data, wres)
	suite.Require().NoError(err)
	return wres.Data
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

func (suite *KeeperTestSuite) broadcastMultiChainExec(msg []byte, sender simulation.Account, contractAddress sdk.AccAddress, chainId string) (*abci.ExecTxResult, error) {
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
	resp, err := appA.BroadcastTxAsync(sender, multimsg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
