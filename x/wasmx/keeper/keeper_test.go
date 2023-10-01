package keeper_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	//nolint

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"

	"mythos/v1/app"
	ibctesting "mythos/v1/testutil/ibc"
	wasmxkeeper "mythos/v1/x/wasmx/keeper"
	"mythos/v1/x/wasmx/keeper/testutil"
	"mythos/v1/x/wasmx/types"
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

func (suite *KeeperTestSuite) GetApp(chain *ibcgotesting.TestChain) *app.App {
	app, ok := chain.App.(*app.App)
	if !ok {
		panic("not app")
	}
	return app
}

func (suite *KeeperTestSuite) GetAppContext(chain *ibcgotesting.TestChain) testutil.AppContext {
	mapp, ok := chain.App.(*app.App)
	if !ok {
		panic("not app")
	}
	appContext := testutil.AppContext{
		S:     suite,
		App:   mapp,
		Chain: chain,
	}
	encodingConfig := app.MakeEncodingConfig()
	appContext.ClientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	appContext.Denom = app.BaseDenom

	t := suite.T()
	appContext.Faucet = wasmxkeeper.NewTestFaucet(t, appContext.Context(), mapp.BankKeeper, types.ModuleName, sdk.NewCoin(appContext.Denom, sdk.NewInt(100_000_000_000)))

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

func (suite *KeeperTestSuite) CommitNBlocks(chain *ibcgotesting.TestChain, n uint64) {
	suite.coordinator.CommitNBlocks(chain, n)
}

func (suite *KeeperTestSuite) Coordinator() *ibcgotesting.Coordinator {
	return suite.coordinator
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

// -short
func SkipCIExpensiveTests(t *testing.T, name string) {
	if testing.Short() {
		t.Skipf("Skipping expensive test: %s", name)
	}
}
