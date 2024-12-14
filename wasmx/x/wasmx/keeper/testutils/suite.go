package testutils

import (
	"context"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/loredanacirstea/wasmx/multichain"
	testkeeper "github.com/loredanacirstea/wasmx/testutil/keeper"
	"github.com/loredanacirstea/wasmx/x/wasmx/keeper"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

var s *KeeperTestSuite

type KeeperTestSuite struct {
	suite.Suite
	GoCtx       context.Context
	CancelFn    context.CancelFunc
	G           *errgroup.Group
	WasmxKeeper *keeper.Keeper
	Ctx         sdk.Context
	Cleanup     func()
}

// SetupSuite runs once before all the tests in the suite
func (suite *KeeperTestSuite) SetupSuite() {
	logger := log.NewNopLogger()
	suite.G, suite.GoCtx, suite.CancelFn = multichain.GetTestCtx(logger, true)
	suite.WasmxKeeper, suite.Ctx, suite.Cleanup = testkeeper.WasmxKeeper(suite.T(), memc.WasmRuntimeMockVmMeta{}, suite.G, suite.GoCtx)
}

// TearDownSuite runs once after all the tests in the suite
func (suite *KeeperTestSuite) TearDownSuite() {
	// Cancel the context to clean up resources
	if suite.CancelFn != nil {
		suite.CancelFn()
	}
}

func (suite *KeeperTestSuite) TearDownTest() {
	suite.Cleanup()
}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)
}
