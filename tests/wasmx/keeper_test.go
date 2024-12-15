package keeper_test

import (
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"

	//nolint

	wt "github.com/loredanacirstea/wasmx/testutil/wasmx"

	wasmedge "github.com/loredanacirstea/wasmx-wasmedge"
	wazero "github.com/loredanacirstea/wasmx-wazero"
)

// KeeperTestSuite is a testing suite to test keeper functions
type KeeperTestSuite struct {
	wt.KeeperTestSuite
}

var s *KeeperTestSuite

func (suite *KeeperTestSuite) SetupSuite() {
	mydir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	suite.WasmVmMeta = wazero.WazeroVmMeta{}
	suite.CompiledCacheDir = path.Join(mydir, "../", "codes_compiled", "wazero")

	suite.WasmVmMeta = wasmedge.WasmEdgeVmMeta{}
	suite.CompiledCacheDir = path.Join(mydir, "../", "codes_compiled", "wasmedge")

	suite.WasmVmMeta = wazero.WazeroVmMeta{}
	suite.CompiledCacheDir = path.Join(mydir, "../", "codes_compiled", "wazero")

	suite.SetupChains()
}

// Called before each test
func (suite *KeeperTestSuite) SetupTest() {
	// suite.SetupChains()
}

// Called once after all tests in the suite
func (suite *KeeperTestSuite) TearDownSuite() {
	suite.TearDownChains()
}

// Called after each test
func (suite *KeeperTestSuite) TearDownTest() {
	// suite.TearDownChains()
}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

// -short
func SkipCIExpensiveTests(t *testing.T, name string) {
	if testing.Short() {
		t.Skipf("Skipping expensive test: %s", name)
	}
}
