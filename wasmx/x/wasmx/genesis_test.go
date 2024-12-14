package wasmx_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/loredanacirstea/wasmx/testutil/nullify"
	"github.com/loredanacirstea/wasmx/x/wasmx"
	testutils "github.com/loredanacirstea/wasmx/x/wasmx/keeper/testutils"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
)

var s *KeeperTestSuite

type KeeperTestSuite struct {
	testutils.KeeperTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)
}

func (suite *KeeperTestSuite) TestGenesis() {
	t := suite.T()
	keeper := suite.WasmxKeeper
	ctx := suite.Ctx
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}
	wasmx.InitGenesis(ctx, *keeper, genesisState)
	got := wasmx.ExportGenesis(ctx, *keeper)
	require.NotNil(t, got)
	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
