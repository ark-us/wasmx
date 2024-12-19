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
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"

	wasmedge "github.com/loredanacirstea/wasmx-wasmedge"
	wazero "github.com/loredanacirstea/wasmx-wazero"
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
	suite.SystemContractsModify = systemContractsModify
	mydir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if wasmRuntime == "wasmedge" {
		suite.WasmVmMeta = wasmedge.WasmEdgeVmMeta{}
		suite.CompiledCacheDir = path.Join(mydir, "../", "codes_compiled", "wasmedge")
		// panic("wasmedge not supported")
	} else {
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

func systemContractsModify(contracts []wasmxtypes.SystemContract) []wasmxtypes.SystemContract {
	var compiledMap map[string]bool
	if wasmRuntime == "wasmedge" {
		compiledMap = wasmedgeCompiled
	} else {
		// compiledMap = wazeroCompiled
		return contracts
	}
	for i := range contracts {
		pinned, ok := compiledMap[contracts[i].Label]
		if ok && pinned {
			contracts[i].Pinned = true
		} else {
			contracts[i].Pinned = false
		}

	}
	return contracts
}

var wazeroCompiled = map[string]bool{}

var wasmedgeCompiled = map[string]bool{
	// wasmxtypes.AUTH_v001:                true,
	// wasmxtypes.ROLES_v001:               true,
	// "ecrecovereth":                      true,
	// "sha2-256":                          true,
	// "ripmd160":                          true,
	// "modexp":                            true,
	// "ecadd":                             true,
	// "ecmul":                             true,
	// "ecpairings":                        true,
	// "blake2f":                           true,
	// wasmxtypes.INTERPRETER_EVM_SHANGHAI: true,
	// // wasmxtypes.INTERPRETER_PYTHON: true,
	// // wasmxtypes.INTERPRETER_JS: true,
	// // wasmxtypes.INTERPRETER_FSM: true,
	// // wasmxtypes.INTERPRETER_TAY: true,
	// "secp384r1":                               true,
	// "secp384r1_registry":                      true,
	// wasmxtypes.STAKING_v001:                   true,
	// wasmxtypes.BANK_v001:                      true,
	// wasmxtypes.ERC20_v001:                     true,
	// wasmxtypes.DERC20_v001:                    true,
	// wasmxtypes.SLASHING_v001:                  true,
	// wasmxtypes.DISTRIBUTION_v001:              true,
	// wasmxtypes.GOV_v001:                       true,
	// wasmxtypes.GOV_CONT_v001:                  true,
	// "raft_library":                            true,
	// "raftp2p_library":                         true,
	// "tendermint_library":                      true,
	// "tendermintp2p_library":                   true,
	// "ava_snowman_library":                     true,
	// wasmxtypes.TIME_v001:                      true,
	// "level0_library":                          true,
	// wasmxtypes.MULTICHAIN_REGISTRY_LOCAL_v001: true,
	// "lobby_library":                           true,
	// wasmxtypes.METAREGISTRY_v001:              true,
	// "level0_ondemand_library":                 true,
	// wasmxtypes.MULTICHAIN_REGISTRY_v001:       true,
	// wasmxtypes.CHAT_v001:                      true,
	// wasmxtypes.CHAT_VERIFIER_v001:             true,
	// wasmxtypes.HOOKS_v001:                     true,
	// wasmxtypes.HOOKS_v001:                     true,
}
