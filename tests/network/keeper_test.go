package keeper_test

import (
	"flag"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"

	//nolint

	wt "github.com/loredanacirstea/wasmx/testutil/wasmx"

	// wasmedge "github.com/loredanacirstea/wasmx-wasmedge"
	wazero "github.com/loredanacirstea/wasmx-wazero"

	ut "github.com/loredanacirstea/mythos-tests/utils"
)

var (
	wasmRuntime   string
	runKnownFixme bool
)

// TestMain is the main entry point for the tests.
func TestMain(m *testing.M) {
	flag.StringVar(&wasmRuntime, "wasm-runtime", "default", "Set the wasm runtime (e.g. wasmedge, wazero)")
	flag.BoolVar(&runKnownFixme, "run-fixme", false, "Run skipped fixme tests")

	// Parse the flags. Only flags after `--` in `go test` command line will be passed here.
	flag.Parse()

	os.Exit(m.Run())
}

// KeeperTestSuite is a testing suite to test keeper functions
type KeeperTestSuite struct {
	wt.KeeperTestSuite
}

var s *KeeperTestSuite

func (suite *KeeperTestSuite) SetupSuite() {
	suite.MaxBlockGas = 100_000_000_000
	suite.SystemContractsModify = ut.SystemContractsModify(wasmRuntime)
	mydir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	switch wasmRuntime {
	case "wasmedge":
		// suite.WasmVmMeta = wasmedge.WasmEdgeVmMeta{}
		// suite.CompiledCacheDir = ut.GetCompiledCacheDir(mydir, "wasmedge")
		panic("wasmedge not activated")
	default:
		// default runtime
		suite.WasmVmMeta = wazero.WazeroVmMeta{}
		suite.CompiledCacheDir = ut.GetCompiledCacheDir(mydir, "wazero")
	}
}

func (suite *KeeperTestSuite) TearDownSuite() {
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.SetupChains()
}

func (suite *KeeperTestSuite) TearDownTest() {
	suite.TearDownChains()
}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func SkipFixmeTests(t *testing.T, name string) {
	if !runKnownFixme {
		t.Skipf("TODO: fixme %s", name)
	}
}
