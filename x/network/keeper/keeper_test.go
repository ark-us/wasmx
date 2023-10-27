package keeper_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/cometbft/cometbft/node"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	//nolint

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server"
	sdkconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"

	app "mythos/v1/app"
	"mythos/v1/server/config"
	ibctesting "mythos/v1/testutil/ibc"
	"mythos/v1/x/network/keeper"
	"mythos/v1/x/network/types"
	wasmxkeeper "mythos/v1/x/wasmx/keeper"
	"mythos/v1/x/wasmx/keeper/testutil"

	wasmxtypes "mythos/v1/x/wasmx/types"
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
	appContext.Faucet = wasmxkeeper.NewTestFaucet(t, appContext.Context(), mapp.BankKeeper, wasmxtypes.ModuleName, sdk.NewCoin(appContext.Denom, sdkmath.NewInt(100_000_000_000)))

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

	mapp, ok := suite.chainA.App.(*app.App)
	if !ok {
		panic("not app")
	}
	suite.GrpcSetup(suite.T(), mapp)
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

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func (suite *KeeperTestSuite) GrpcSetup(t *testing.T, app servertypes.Application) {
	serverCtx := server.NewDefaultContext()
	clientCtx := suite.GetAppContext(suite.chainA).ClientCtx
	config, err := config.GetConfig(serverCtx.Viper)
	require.NoError(t, err)

	lis = bufconn.Listen(bufSize)
	grpcServer, err := NewGRPCServer(serverCtx, clientCtx, config.GRPC, app, nil)
	require.NoError(t, err)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func grpcClient(t *testing.T, ctx context.Context) (types.MsgClient, *grpc.ClientConn) {
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return types.NewMsgClient(conn), conn
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func NewGRPCServer(
	svrCtx *server.Context,
	clientCtx client.Context,
	cfg sdkconfig.GRPCConfig,
	app servertypes.Application,
	tmNode *node.Node,
) (*grpc.Server, error) {
	maxSendMsgSize := cfg.MaxSendMsgSize
	if maxSendMsgSize == 0 {
		maxSendMsgSize = sdkconfig.DefaultGRPCMaxSendMsgSize
	}

	maxRecvMsgSize := cfg.MaxRecvMsgSize
	if maxRecvMsgSize == 0 {
		maxRecvMsgSize = sdkconfig.DefaultGRPCMaxRecvMsgSize
	}

	grpcSrv := grpc.NewServer(
		grpc.ForceServerCodec(codec.NewProtoCodec(clientCtx.InterfaceRegistry).GRPCCodec()),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
	)

	fn := func(description string, timeDelay int64, fn func() error, gracefulStop func()) (chan struct{}, error) {
		return nil, nil
	}
	err := keeper.RegisterGRPCServer(svrCtx, clientCtx, tmNode, app, grpcSrv, fn)
	if err != nil {
		return nil, fmt.Errorf("failed to register grpc server: %w", err)
	}
	return grpcSrv, nil
}
