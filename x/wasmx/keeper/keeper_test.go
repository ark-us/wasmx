package keeper_test

import (
	"encoding/json"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	//nolint

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	icahosttypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v6/testing"

	"wasmx/app"
	ibctesting "wasmx/testutil/ibc"
	wasmxkeeper "wasmx/x/wasmx/keeper"
	"wasmx/x/wasmx/types"
)

var (
	// TestAccAddress defines a resuable bech32 address for testing purposes
	// TODO: update crypto.AddressHash() when sdk uses address.Module()
	// TestAccAddress = icatypes.GenerateAddress(sdk.AccAddress(crypto.AddressHash([]byte(icatypes.ModuleName))), ibcgotesting.FirstConnectionID, TestPortID)
	// TestOwnerAddress defines a reusable bech32 address for testing purposes
	TestOwnerAddress = "cosmos1fjx8p8uzx3h5qszqnwvelulzd659j8ua5qvaep"
	// TestPortID defines a reusable port identifier for testing purposes
	TestPortID, _ = icatypes.NewControllerPortID(TestOwnerAddress)
	// TestVersion defines a reusable interchainaccounts version string for testing purposes
	TestVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ibcgotesting.FirstConnectionID,
		HostConnectionId:       ibcgotesting.FirstConnectionID,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
)

// KeeperTestSuite is a testing suite to test keeper functions
type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibcgotesting.Coordinator
	chainIds    []string

	// testing chains used for convenience and readability
	chainA *ibcgotesting.TestChain
	chainB *ibcgotesting.TestChain
}

type AppContext struct {
	s *KeeperTestSuite

	app   *app.App
	chain *ibcgotesting.TestChain

	// for generate test tx
	clientCtx client.Context

	appCodec codec.Codec
	signer   keyring.Signer
	denom    string
	faucet   *wasmxkeeper.TestFaucet
}

func (suite *KeeperTestSuite) GetApp(chain *ibcgotesting.TestChain) *app.App {
	app, ok := chain.App.(*app.App)
	if !ok {
		panic("not app")
	}
	return app
}

func (suite *KeeperTestSuite) GetAppContext(chain *ibcgotesting.TestChain) AppContext {
	mapp, ok := chain.App.(*app.App)
	if !ok {
		panic("not app")
	}
	appContext := AppContext{
		s:     suite,
		app:   mapp,
		chain: chain,
	}
	encodingConfig := app.MakeEncodingConfig()
	appContext.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	appContext.denom = "amyt"

	t := suite.T()
	appContext.faucet = wasmxkeeper.NewTestFaucet(t, appContext.Context(), mapp.BankKeeper, types.ModuleName, sdk.NewCoin(appContext.denom, sdk.NewInt(100_000_000_000)))

	return appContext
}

var s *KeeperTestSuite

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

// SetupTest creates a coordinator with 2 test chains.
func (suite *KeeperTestSuite) SetupTest() {
	suite.chainIds = []string{"mythos_7001-1", "mythos_7002-1"}
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), suite.chainIds)
	suite.chainA = suite.coordinator.GetChain(suite.chainIds[0])
	suite.chainB = suite.coordinator.GetChain(suite.chainIds[1])

	// ICA setup

	allowedMsgs := []string{"/cosmos.bank.v1beta1.MsgSend", "/cosmos.staking.v1beta1.MsgDelegate"}

	// both chains can be hosts
	params := icahosttypes.NewParams(true, allowedMsgs)
	suite.GetApp(suite.chainA).ICAHostKeeper.SetParams(suite.chainA.GetContext(), params)

	params = icahosttypes.NewParams(true, allowedMsgs)
	suite.GetApp(suite.chainB).ICAHostKeeper.SetParams(suite.chainB.GetContext(), params)
}

func NewICAPath(chainA, chainB *ibcgotesting.TestChain) *ibcgotesting.Path {
	path := ibcgotesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointB.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointA.ChannelConfig.Version = TestVersion
	path.EndpointB.ChannelConfig.Version = TestVersion

	return path
}

func (suite *KeeperTestSuite) Commit() {
	suite.coordinator.CommitBlock(suite.chainA, suite.chainB)
}

func (s *KeeperTestSuite) GetRandomAccount() simulation.Account {
	pk := ed25519.GenPrivKey()
	privKey := secp256k1.GenPrivKeyFromSecret(pk.GetKey().Seed())
	pubKey := privKey.PubKey()
	address := sdk.AccAddress(pubKey.Address())
	account := simulation.Account{
		PrivKey: privKey,
		PubKey:  pubKey,
		Address: address,
	}
	return account
}

func (suite *KeeperTestSuite) RegisterInterchainAccount(endpoint *ibcgotesting.Endpoint, owner string, version string) error {
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return err
	}

	channelSequence := endpoint.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(endpoint.Chain.GetContext())

	if err := suite.GetApp(endpoint.Chain).ICAControllerKeeper.RegisterInterchainAccount(endpoint.Chain.GetContext(), endpoint.ConnectionID, owner, version); err != nil {
		return err
	}

	// commit state changes for proof verification
	endpoint.Chain.NextBlock()

	// update port/channel ids
	endpoint.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	endpoint.ChannelConfig.PortID = portID
	endpoint.ChannelConfig.Version = TestVersion

	return nil
}

// SetupICAPath invokes the InterchainAccounts entrypoint and subsequent channel handshake handlers
func (suite *KeeperTestSuite) SetupICAPath(path *ibcgotesting.Path, owner string, version string) error {
	if err := suite.RegisterInterchainAccount(path.EndpointA, owner, version); err != nil {
		return err
	}

	if err := path.EndpointB.ChanOpenTry(); err != nil {
		return err
	}

	if err := path.EndpointA.ChanOpenAck(); err != nil {
		return err
	}

	if err := path.EndpointB.ChanOpenConfirm(); err != nil {
		return err
	}

	return nil
}

func (s AppContext) Context() sdk.Context {
	return s.chain.GetContext()
}

func (suite *AppContext) RegisterInterTxAccount(endpoint *ibcgotesting.Endpoint, owner string) error {
	// types.New
	return nil
}

var DEFAULT_GAS_PRICE = "0.05amyt"
var DEFAULT_GAS_LIMIT = uint64(15_000_000)

func (s AppContext) prepareCosmosTx(account simulation.Account, msgs []sdk.Msg, gasLimit *uint64, gasPrice *string) []byte {
	encodingConfig := app.MakeEncodingConfig()
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	var parsedGasPrices sdk.DecCoins
	var err error

	if gasLimit != nil {
		txBuilder.SetGasLimit(*gasLimit)
	} else {
		txBuilder.SetGasLimit(DEFAULT_GAS_LIMIT)
	}

	if gasPrice != nil {
		parsedGasPrices, err = sdk.ParseDecCoins(*gasPrice)
	} else {
		parsedGasPrices, err = sdk.ParseDecCoins(DEFAULT_GAS_PRICE)
	}
	s.s.Require().NoError(err)
	feeAmount := parsedGasPrices.AmountOf("amyt").MulInt64(int64(DEFAULT_GAS_LIMIT)).RoundInt()

	fees := &sdk.Coins{{Denom: s.denom, Amount: feeAmount}}
	txBuilder.SetFeeAmount(*fees)
	err = txBuilder.SetMsgs(msgs...)
	s.s.Require().NoError(err)

	seq, err := s.app.AccountKeeper.GetSequence(s.Context(), account.Address)
	s.s.Require().NoError(err)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: account.PubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	err = txBuilder.SetSignatures(sigV2)
	s.s.Require().NoError(err)

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := s.app.AccountKeeper.GetAccount(s.Context(), account.Address).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       s.Context().ChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, account.PrivKey, encodingConfig.TxConfig,
		seq,
	)
	s.s.Require().NoError(err)

	err = txBuilder.SetSignatures(sigV2)
	s.s.Require().NoError(err)

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.s.Require().NoError(err)
	return bz
}

func (s AppContext) DeliverTx(account simulation.Account, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	bz := s.prepareCosmosTx(account, msgs, nil, nil)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}

// func (s AppContext) SimulateTx(account simulation.Account, msgs ...sdk.Msg) (sdk.GasInfo, *sdk.Result) {
// 	bz := s.prepareCosmosTx(account, msgs, nil, nil)
// 	gasInfo, res, err := s.app.Simulate(bz)
// 	s.s.Require().NoError(err)
// 	return gasInfo, res
// }

// func (s AppContext) Query(account simulation.Account, contract sdk.AccAddress, queryData []byte, queryPath string) abci.ResponseQuery {
// 	req := abci.RequestQuery{Data: queryData, Path: queryPath}
// 	return s.app.BaseApp.Query(req)
// }

// // SmartQuery This will serialize the query message and submit it to the contract.
// // The response is parsed into the provided interface.
// // Usage: SmartQuery(addr, QueryMsg{Foo: 1}, &response)
// func (s AppContext) SmartQuery(contractAddr string, queryMsg interface{}, response interface{}) error {
// 	msg, err := json.Marshal(queryMsg)
// 	if err != nil {
// 		return err
// 	}

// 	req := wasmtypes.QuerySmartContractStateRequest{
// 		Address:   contractAddr,
// 		QueryData: msg,
// 	}
// 	reqBin, err := proto.Marshal(&req)
// 	if err != nil {
// 		return err
// 	}

// 	// TODO: what is the query?
// 	res := s.app.Query(abci.RequestQuery{
// 		Path: "/cosmwasm.wasm.v1.Query/SmartContractState",
// 		Data: reqBin,
// 	})

// 	if res.Code != 0 {
// 		return fmt.Errorf("query failed: (%d) %s", res.Code, res.Log)
// 	}

// 	// unpack protobuf
// 	var resp wasmtypes.QuerySmartContractStateResponse
// 	err = proto.Unmarshal(res.Value, &resp)
// 	if err != nil {
// 		return err
// 	}
// 	// unpack json content
// 	return json.Unmarshal(resp.Data, response)
// }

// func (s AppContext) EwasmQuery(account simulation.Account, contract sdk.AccAddress, queryData []byte, funds sdk.Coins) string {
// 	query := wasmtypes.QuerySmartContractCallRequest{
// 		Sender:    account.Address.String(),
// 		Address:   contract.String(),
// 		QueryData: queryData,
// 		Funds:     funds,
// 	}
// 	bz, err := query.Marshal()
// 	s.s.Require().NoError(err)

// 	req := abci.RequestQuery{Data: bz, Path: "/cosmwasm.wasm.v1.Query/SmartContractCall"}
// 	abcires := s.app.BaseApp.Query(req)
// 	var resp wasmtypes.QuerySmartContractCallResponse
// 	err = resp.Unmarshal(abcires.Value)
// 	s.s.Require().NoError(err)

// 	var data wasmkeeper.EwasmQueryResponse
// 	err = json.Unmarshal(resp.Data, &data)
// 	s.s.Require().NoError(err)
// 	return hex.EncodeToString(data.Data)
// }

func (s AppContext) DeliverTxWithOpts(account simulation.Account, msg sdk.Msg, gasLimit uint64, gasPrice *string) abci.ResponseDeliverTx {
	bz := s.prepareCosmosTx(account, []sdk.Msg{msg}, &gasLimit, gasPrice)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}

func (s AppContext) StoreCode(sender simulation.Account, wasmbin []byte) uint64 {
	storeCodeMsg := &types.MsgStoreCode{
		Sender:       sender.Address.String(),
		WasmByteCode: wasmbin,
	}

	res := s.DeliverTx(sender, storeCodeMsg)
	s.s.Require().True(res.IsOK(), res.GetLog())
	s.s.Commit()

	codeId := s.s.GetCodeIdFromLog(res.GetLog())

	bytecode, err := s.app.WasmxKeeper.GetByteCode(s.Context(), codeId)
	s.s.Require().NoError(err)
	s.s.Require().Equal(bytecode, wasmbin)
	return codeId
}

func (s AppContext) InstantiateCode(sender simulation.Account, codeId uint64, instantiateMsgStr string) sdk.AccAddress {
	instantiateMsg := []byte(instantiateMsgStr)
	instantiateCodeMsg := &types.MsgInstantiateContract{
		Sender: sender.Address.String(),
		CodeId: codeId,
		Label:  "test",
		Msg:    instantiateMsg,
	}
	res := s.DeliverTxWithOpts(sender, instantiateCodeMsg, 1000000, nil) // 135690
	s.s.Require().True(res.IsOK(), res.GetLog())
	s.s.Commit()
	contractAddressStr := s.s.GetContractAddressFromLog(res.GetLog())
	contractAddress := sdk.MustAccAddressFromBech32(contractAddressStr)
	return contractAddress
}

func (s AppContext) ExecuteContract(sender simulation.Account, contractAddress sdk.AccAddress, executeMsgStr string, funds sdk.Coins) abci.ResponseDeliverTx {
	executeMsgBz := []byte(executeMsgStr)
	executeMsg := &types.MsgExecuteContract{
		Sender:   sender.Address.String(),
		Contract: contractAddress.String(),
		Msg:      executeMsgBz,
		Funds:    funds,
	}
	res := s.DeliverTxWithOpts(sender, executeMsg, 1500000, nil) // 135690
	s.s.Require().True(res.IsOK(), res.GetLog())
	s.s.Require().NotContains(res.GetLog(), "failed to execute message", res.GetLog())
	s.s.Commit()
	return res
}

type Attribute struct {
	Key   string
	Value string
}

type Event struct {
	Type       string
	Attributes *[]Attribute
}

type Log struct {
	MsgIndex uint64
	Events   []Event
}

func (s *KeeperTestSuite) GetFromLog(logstr string, logtype string) *[]Attribute {
	var logs []Log
	err := json.Unmarshal([]byte(logstr), &logs)
	s.Require().NoError(err)
	for _, log := range logs {
		for _, ev := range log.Events {
			if ev.Type == logtype {
				return ev.Attributes
			}
		}
	}
	return nil
}

func (s *KeeperTestSuite) GetCodeIdFromLog(logstr string) uint64 {
	attrs := s.GetFromLog(logstr, "store_code")
	if attrs == nil {
		return 0
	}
	for _, attr := range *attrs {
		if attr.Key == "code_id" {
			ui64, err := strconv.ParseUint(attr.Value, 10, 64)
			s.Require().NoError(err)
			return ui64
		}
	}
	return 0
}

func (s *KeeperTestSuite) GetContractAddressFromLog(logstr string) string {
	attrs := s.GetFromLog(logstr, "instantiate")
	s.Require().NotNil(attrs)
	for _, attr := range *attrs {
		if attr.Key == "contract_address" {
			return attr.Value
		}
	}
	s.Require().True(false, "no contract address found in log")
	return ""
}

func GetSdkEvents(events []abci.Event, evtype string) []sdk.Event {
	sdkEvents := make([]sdk.Event, len(events))
	for _, ev := range events {
		if ev.GetType() != evtype {
			continue
		}

		attributes := make([]sdk.Attribute, len(ev.Attributes))
		for _, attr := range ev.Attributes {
			attributes = append(attributes, sdk.NewAttribute(string(attr.Key), string(attr.Value)))
		}

		sdkev := sdk.NewEvent(ev.Type, attributes...)
		sdkEvents = append(sdkEvents, sdkev)
	}
	return sdkEvents
}
