package keeper_test

import (
	"flag"
	"fmt"
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"

	//nolint

	wt "github.com/loredanacirstea/wasmx/testutil/wasmx"

	// wasmedge "github.com/loredanacirstea/wasmx-wasmedge"
	wazero "github.com/loredanacirstea/wasmx-wazero"

	ut "github.com/loredanacirstea/mythos-tests/utils"
	vmimap "github.com/loredanacirstea/wasmx-vmimap"
	vmsmtp "github.com/loredanacirstea/wasmx-vmsmtp"
	"github.com/loredanacirstea/wasmx/x/vmhttpclient"
	"github.com/loredanacirstea/wasmx/x/vmhttpserver"
	"github.com/loredanacirstea/wasmx/x/vmoauth2client"
)

func init() {
	vmimap.Setup()
	vmsmtp.Setup()
	vmhttpclient.Setup()
	vmhttpserver.Setup()
	vmoauth2client.Setup()
}

var (
	wasmRuntime       string
	emailUsername     string
	emailPassword     string
	runListen         bool
	CLIENT_ID_WEB     string
	CLIENT_SECRET_WEB string
	provider          string = "google"
	runOAuth2         bool
	isOAuth2          bool
	runEmailServer    bool
	runKnownFixme     bool
)

// TestMain is the main entry point for the tests.
func TestMain(m *testing.M) {
	flag.StringVar(&wasmRuntime, "wasm-runtime", "default", "Set the wasm runtime (e.g. wasmedge, wazero)")
	flag.StringVar(&emailUsername, "email-username", "", "Set the email account address for tests")
	flag.StringVar(&emailPassword, "email-password", "", "Set the email account password for tests")
	flag.BoolVar(&runListen, "run-listen", false, "Run email listen test")
	flag.BoolVar(&runOAuth2, "run-oauth2", false, "Run email oauth2 test")
	flag.BoolVar(&runEmailServer, "run-email-server", false, "Run email server")
	flag.StringVar(&CLIENT_ID_WEB, "client-id", "", "Set the client ID")
	flag.StringVar(&CLIENT_SECRET_WEB, "client-secret", "", "Set the client secret")
	flag.StringVar(&provider, "provider", "", "Set the provider for the client ID")
	flag.BoolVar(&isOAuth2, "is-oauth2", false, "password is access token")
	flag.BoolVar(&runKnownFixme, "run-fixme", false, "Run skipped fixme tests")

	flag.Parse()

	os.Exit(m.Run())
}

// KeeperTestSuite is a testing suite to test keeper functions
type KeeperTestSuite struct {
	wt.KeeperTestSuite
	emailUsername     string
	emailPassword     string
	runListen         bool
	runOAuth2         bool
	runEmailServer    bool
	CLIENT_ID_WEB     string
	CLIENT_SECRET_WEB string
	provider          string
	isOAuth2          bool
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

func (suite *KeeperTestSuite) TearDownSuite() {}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	flag.Parse() // <- force parse here, in case it wasn't called
	fmt.Println("runListen: ", runListen, emailUsername, emailPassword)

	s = new(KeeperTestSuite)
	s.emailUsername = emailUsername
	s.emailPassword = emailPassword
	s.runListen = runListen
	s.runOAuth2 = runOAuth2
	s.runEmailServer = runEmailServer
	s.CLIENT_ID_WEB = CLIENT_ID_WEB
	s.CLIENT_SECRET_WEB = CLIENT_SECRET_WEB
	s.provider = provider
	s.isOAuth2 = isOAuth2
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

func SkipNoPasswordTests(t *testing.T, name string) {
	if s.emailPassword == "" {
		t.Skipf("Skipping (no password) %s", name)
	}
}
