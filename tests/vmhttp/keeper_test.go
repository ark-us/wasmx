package keeper_test

import (
	"encoding/json"
	"flag"
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"

	//nolint

	wt "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/vmhttpclient"
	"github.com/loredanacirstea/wasmx/x/vmhttpserver"

	// wasmedge "github.com/loredanacirstea/wasmx-wasmedge"
	wazero "github.com/loredanacirstea/wasmx-wazero"

	ut "github.com/loredanacirstea/mythos-tests/utils"
)

var (
	wasmRuntime string
)

func init() {
	vmhttpclient.Setup()
	vmhttpserver.Setup()
}

// TestMain is the main entry point for the tests.
func TestMain(m *testing.M) {
	flag.StringVar(&wasmRuntime, "wasm-runtime", "default", "Set the wasm runtime (e.g. wasmedge, wazero)")

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
}

func (suite *KeeperTestSuite) TearDownTest() {
	defer os.Remove("dtype.db")
	defer os.Remove("dtype.db-shm")
	defer os.Remove("dtype.db-wal")
}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) parseQueryResponse(qres []byte, qresp interface{}) {
	err := json.Unmarshal(qres, qresp)
	suite.Require().NoError(err)
}
