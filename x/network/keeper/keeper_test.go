package keeper_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	cometdbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/node"
	cometstore "github.com/cometbft/cometbft/store"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	//nolint

	sdkmath "cosmossdk.io/math"
	dbm "github.com/cosmos/cosmos-db"
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

// func (suite *KeeperTestSuite) TearDownSuite() {
// 	s.T().Log("tearing down integration test suite")
// }

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
	fmt.Println("--grpcSetup--")
	fmt.Println("--grpcSetup-0-")
	// serverCtx := server.GetServerContextFromCmd(cmd)
	serverCtx := server.NewDefaultContext()
	fmt.Println("--grpcSetup-1-")
	// fmt.Println("--grpcSetup-1-", cmd.Context())
	// cmd.Context().Value(ClientContextKey)
	// clientCtx, err := client.GetClientQueryContext(cmd)
	// require.NoError(t, err)

	// clientCtx := client.Context{}.WithHomeDir(home).
	// 	WithViper("").
	// 	WithCodec(codec.NewProtoCodec(codectypes.NewInterfaceRegistry())).
	// 	WithChainID(chainID)

	clientCtx := suite.GetAppContext(suite.chainA).ClientCtx

	fmt.Println("--grpcSetup-2-")

	config, err := config.GetConfig(serverCtx.Viper)
	require.NoError(t, err)

	// app := appCreator(svrCtx.Logger, db, traceWriter, svrCtx.Viper)

	lis = bufconn.Listen(bufSize)

	// g, ctx := mythosserver.GetCtx(serverCtx, true)
	// ctx, _ := context.WithCancel(context.Background())
	// tmNode, _, err := mythosserver.StartCmtNode(ctx, serverCtx.Config, app, serverCtx)
	// require.NoError(t, err)
	// defer cleanupFn()

	// grpcServer, err := keeper.NewGRPCServer(serverCtx, clientCtx, config.GRPC, app, nil, nil, nil, nil)
	grpcServer, err := NewGRPCServer(serverCtx, clientCtx, config.GRPC, app, nil, nil, nil, nil)
	fmt.Println("--grpcSetup-END-")
	require.NoError(t, err)

	// s := grpc.NewServer()
	// pb.RegisterGreeterServer(s, &server{})
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func grpcClient(t *testing.T, ctx context.Context) (types.MsgClient, *grpc.ClientConn) {
	fmt.Println("--grpcClient--")
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	// defer conn.Close()
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
	blockStore *cometstore.BlockStore,
	stateDB cometdbm.DB,
	networkDB dbm.DB,
) (*grpc.Server, error) {
	fmt.Println("--NewGRPCServer--")
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

	err := keeper.RegisterGRPCServer(svrCtx, clientCtx, tmNode, app, grpcSrv, blockStore, stateDB, networkDB)
	fmt.Println("-END-RegisterGRPCServer--", err)
	if err != nil {
		return nil, fmt.Errorf("failed to register grpc server: %w", err)
	}
	fmt.Println("-END-NewGRPCServer--")
	return grpcSrv, nil
}
