package keeper_test

import (
	"flag"
	"os"
	"path"
	"testing"
	"time"

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
	benchmarkMode bool
	wasmRuntime   string
)

// TestMain is the main entry point for the tests.
func TestMain(m *testing.M) {
	// Add custom flags
	flag.BoolVar(&benchmarkMode, "benchmark", false, "Enable benchmark timing in tests")
	flag.StringVar(&wasmRuntime, "wasm-runtime", "default", "Set the wasm runtime (e.g. wasmedge, wazero)")

	// Parse the flags. Only flags after `--` in `go test` command line will be passed here.
	flag.Parse()

	os.Exit(m.Run())
}

// KeeperTestSuite is a testing suite to test keeper functions
type KeeperTestSuite struct {
	wt.KeeperTestSuite
	testStart  time.Time
	suiteStart time.Time
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
		// suite.CompiledCacheDir = path.Join(mydir, "../", "codes_compiled", "wasmedge")
		panic("wasmedge not activated")
	case "wazero":
		suite.WasmVmMeta = wazero.WazeroVmMeta{}
		suite.CompiledCacheDir = path.Join(mydir, "../", "codes_compiled", "wazero")
	default:
		// default runtime
		suite.WasmVmMeta = wazero.WazeroVmMeta{}
		suite.CompiledCacheDir = path.Join(mydir, "../", "codes_compiled", "wazero")
	}

	suite.SetupChains()
	if benchmarkMode {
		s.suiteStart = time.Now()
	}
}

// Called before each test
func (suite *KeeperTestSuite) SetupTest() {
	// suite.SetupChains()
	if benchmarkMode {
		s.testStart = time.Now()
	}
}

// Called once after all tests in the suite
func (suite *KeeperTestSuite) TearDownSuite() {
	if benchmarkMode {
		elapsed := time.Since(s.suiteStart)
		// Print or log the elapsed time for the entire suite
		println("Suite time:", elapsed.String())
	}
	suite.TearDownChains()
}

// Called after each test
func (suite *KeeperTestSuite) TearDownTest() {
	// suite.TearDownChains()
	if benchmarkMode {
		elapsed := time.Since(s.testStart)
		// Print or log the elapsed time for each individual test
		println("Test time:", elapsed.String(), "Runtime:", wasmRuntime)
	}
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
