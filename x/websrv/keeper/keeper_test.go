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

	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"

	app "mythos/v1/app"
	ibctesting "mythos/v1/testutil/ibc"
	wasmxkeeper "mythos/v1/x/wasmx/keeper"
	"mythos/v1/x/wasmx/keeper/testutil"
	"mythos/v1/x/wasmx/types"
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
	appContext.ClientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig).WithChainID(chain.ChainID)
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
	suite.chainIds = []string{"mythos_7001-1"}
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), suite.chainIds)
	suite.chainA = suite.coordinator.GetChain(suite.chainIds[0])
}

func (suite *KeeperTestSuite) Commit() {
	suite.coordinator.CommitBlock(suite.chainA)
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
