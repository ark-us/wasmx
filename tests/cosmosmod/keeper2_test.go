package keeper_test

import (
	"os"

	//nolint

	wazero "github.com/loredanacirstea/wasmx-wazero"
	wt "github.com/loredanacirstea/wasmx/testutil/wasmx"

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
		// suite.CompiledCacheDir = ut.GetCompiledCacheDir(mydir, "wasmedge")
		panic("wasmedge not activated")
	default:
		// default runtime
		suite.WasmVmMeta = wazero.WazeroVmMeta{}
		suite.CompiledCacheDir = ut.GetCompiledCacheDir(mydir, "wazero")
	}
}

func (suite *KeeperTestSuite2) TearDownSuite() {
}

func (suite *KeeperTestSuite2) SetupTest() {
	suite.SetupChains()
}

func (suite *KeeperTestSuite2) TearDownTest() {
	suite.TearDownChains()
}
