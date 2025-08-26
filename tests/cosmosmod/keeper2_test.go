package keeper_test

import (
	"os"
	"path"

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
}

func (suite *KeeperTestSuite2) TearDownSuite() {
}

func (suite *KeeperTestSuite2) SetupTest() {
	suite.SetupChains()
}

func (suite *KeeperTestSuite2) TearDownTest() {
	suite.TearDownChains()
}
