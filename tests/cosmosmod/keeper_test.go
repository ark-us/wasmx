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

	wazero "github.com/loredanacirstea/wasmx-wazero"
	wt "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"

	ut "github.com/loredanacirstea/mythos-tests/utils"
)

var (
	wasmRuntime string
)

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

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) getPropExtended(appA wt.AppContext) *types.ProposalExtended {
	msg := []byte(`{"GetProposalExtended":{"proposal_id":1}}`)
	resp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msg,
	})
	suite.Require().NoError(err)
	var qresp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp.Data, &qresp)
	suite.Require().NoError(err)
	var propext types.QueryProposalExtendedResponse
	err = appA.App.AppCodec().UnmarshalJSON(qresp.Data, &propext)
	suite.Require().NoError(err)
	return propext.Proposal
}
