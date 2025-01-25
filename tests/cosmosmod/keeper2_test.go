package keeper_test

import (
	"encoding/json"
	"os"
	"path"

	//nolint

	wazero "github.com/loredanacirstea/wasmx-wazero"
	wt "github.com/loredanacirstea/wasmx/testutil/wasmx"
	"github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"

	ut "github.com/loredanacirstea/mythos-tests/utils"
)

// KeeperTestSuite2 is a testing suite that runs each test on a separate chain
type KeeperTestSuite2 struct {
	wt.KeeperTestSuite
}

var s2 *KeeperTestSuite2

func (suite *KeeperTestSuite2) SetupSuite() {
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

func (suite *KeeperTestSuite2) TearDownSuite() {
	suite.TearDownChains()
}

func (suite *KeeperTestSuite2) SetupTest() {
}

func (suite *KeeperTestSuite2) TearDownTest() {
}

func (suite *KeeperTestSuite2) getPropExtended(appA wt.AppContext) *types.ProposalExtended {
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
