package keeper_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	client "github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simulation "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"mythos/v1/app"
	mcodec "mythos/v1/codec"
	mcfg "mythos/v1/config"
	menc "mythos/v1/encoding"
	ibctesting "mythos/v1/testutil/ibc"
	cosmosmodtypes "mythos/v1/x/cosmosmod/types"
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
	// config, err := mcfg.GetChainConfig(chainId)
	// s.Require().NoError(err)
	suite.SetCurrentChain(chainId)
	chain := suite.GetChain(chainId)

	sender := suite.GetRandomAccount()
	initBalance := sdkmath.NewInt(1000_000_000)

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
	senderPrivKey := chain.SenderPrivKey
	encoding := menc.MakeEncodingConfig(&subChainConfig)
	addrCodec := mcodec.MustUnwrapAccBech32Codec(encoding.InterfaceRegistry.SigningContext().AddressCodec())
	acc := cosmosmodtypes.NewBaseAccount(addrCodec.BytesToAccAddressPrefixed(senderPrivKey.PubKey().Address().Bytes()), senderPrivKey.PubKey(), 0, 0)
	amount := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
	balances := []banktypes.Balance{{
		Address: acc.GetAddressPrefixed().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(subChainConfig.BaseDenom, amount)),
	}}
	_, genesisState, err := ibctesting.BuildGenesisData(chain.Vals, []cosmosmodtypes.GenesisAccount{acc}, subChainId, subChainConfig, 10, balances)
	s.Require().NoError(err)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	s.Require().NoError(err)

	req := &abci.RequestInitChain{
		ChainId:         subChainId,
		InitialHeight:   1,
		Time:            time.Now().UTC(),
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: app.DefaultTestingConsensusParams,
		AppStateBytes:   stateBytes,
	}

	initChainBz, err := json.Marshal(req)
	suite.Require().NoError(err)
	chainConfigBz, err := json.Marshal(subChainConfig)
	suite.Require().NoError(err)

	valAddr, err := addrCodec.BytesToString(sdk.ValAddress(chain.Vals.Validators[0].Address))

	peer := fmt.Sprintf("%s@/ip4/127.0.0.1/tcp/5001/p2p/12D3KooWJdKwTq9QcARdPuk4QBibP8MxBV7Q8xC7JRMSXWuvZBtD", valAddr)
	peers := []string{peer}
	peersbz, _ := json.Marshal(peers)

	msg := fmt.Sprintf(`{"InitSubChain":{"init_chain_request":%s,"chain_config":%s,"peers":%s}}`, string(initChainBz), string(chainConfigBz), string(peersbz))
	res, err := suite.broadcastMultiChainExec([]byte(msg), sender, registryAddress, chainId)
	suite.Require().NoError(err)
	evs := appA.GetSdkEventsByType(res.Events, "init_subchain")
	suite.Require().Equal(1, len(evs))
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
