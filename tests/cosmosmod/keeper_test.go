package keeper_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"

	//nolint

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	wazero "github.com/loredanacirstea/wasmx-wazero"
	wt "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"

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

// KeeperTestSuite is a testing suite to run tests on the same chain
type KeeperTestSuite struct {
	wt.KeeperTestSuite
}

var s *KeeperTestSuite

func (suite *KeeperTestSuite) SetupSuite() {
	suite.MaxBlockGas = 100_000_000_000
	suite.SystemContractsModify = ut.SystemContractsModify(wasmRuntime)
	suite.GenesisModify = ut.GenesisModify
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

func (suite *KeeperTestSuite) TearDownSuite() {
	suite.TearDownChains()
}

func (suite *KeeperTestSuite) SetupTest() {
}

func (suite *KeeperTestSuite) TearDownTest() {
}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	s2 = new(KeeperTestSuite2)
	suite.Run(t, s)
	suite.Run(t, s2)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) getPropExtended(appA wt.AppContext, proposalId uint64, contract string) *types.ProposalExtended {
	msg := []byte(fmt.Sprintf(`{"GetProposalExtended":{"proposal_id":%d}}`, proposalId))
	resp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: contract,
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

func (suite *KeeperTestSuite) getProp(appA wt.AppContext, proposalId uint64, contract string) *govtypes.Proposal {
	msg := []byte(fmt.Sprintf(`{"GetProposal":{"proposal_id":%d}}`, proposalId))
	resp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: contract,
		Msg:      msg,
	})
	suite.Require().NoError(err)
	var qresp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp.Data, &qresp)
	suite.Require().NoError(err)
	var propext types.QueryProposalResponse
	err = appA.App.AppCodec().UnmarshalJSON(qresp.Data, &propext)
	suite.Require().NoError(err)
	if propext.Proposal == nil {
		return nil
	}
	proposal, err := types.CosmosProposalFromInternal(appA.App.AppCodec(), *propext.Proposal)
	suite.Require().NoError(err)
	return proposal
}

func (suite *KeeperTestSuite) getTallyResult(appA wt.AppContext, proposalId uint64, contract string) *govtypes.TallyResult {
	msg := []byte(fmt.Sprintf(`{"GetTallyResult":{"proposal_id":%d}}`, proposalId))
	resp, err := suite.App().NetworkKeeper.QueryContract(appA.Context(), &networktypes.MsgQueryContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: contract,
		Msg:      msg,
	})
	suite.Require().NoError(err)
	var qresp wasmxtypes.ContractResponse
	err = json.Unmarshal(resp.Data, &qresp)
	suite.Require().NoError(err)
	var propext govtypes.QueryTallyResultResponse
	err = appA.App.AppCodec().UnmarshalJSON(qresp.Data, &propext)
	suite.Require().NoError(err)
	if propext.Tally == nil {
		return nil
	}
	tally, err := types.CosmosTallyFromInternal(propext.Tally)
	suite.Require().NoError(err)
	return tally
}

func SkipFixmeTests(t *testing.T, name string) {
	if !runKnownFixme {
		t.Skipf("TODO: fixme %s", name)
	}
}
