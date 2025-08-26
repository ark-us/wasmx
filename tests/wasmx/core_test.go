package keeper_test

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	mcfg "github.com/loredanacirstea/wasmx/config"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// we need these tests to run first, hence the name "TestAaa"

func (suite *KeeperTestSuite) TestAaaCoreIterateCodeInfos() {
	suite.SetCurrentChain(mcfg.MYTHOS_CHAIN_ID_TEST)
	appA := s.AppContext()
	count := 0
	appA.App.WasmxKeeper.IterateCodeInfos(appA.Context(), func(id uint64, info wasmxtypes.CodeInfo) bool {
		count += 1
		return false
	})
	syscontracts := wasmxtypes.DefaultSystemContracts(appA.AccBech32Codec(), "", "", 2, false, "", mcfg.BondBaseDenom)
	suite.Require().Equal(len(syscontracts), count)
}

func (suite *KeeperTestSuite) TestAaaCoreIterateContractInfos() {
	suite.SetCurrentChain(mcfg.MYTHOS_CHAIN_ID_TEST)
	appA := s.AppContext()

	syscontracts := wasmxtypes.DefaultSystemContracts(appA.AccBech32Codec(), "", "", 2, false, "", mcfg.BondBaseDenom)
	syslen := len(syscontracts)
	count := 0

	mapsys := make(map[string]wasmxtypes.SystemContract, 0)
	for _, c := range syscontracts {
		mapsys[c.Address] = c
	}

	appA.App.WasmxKeeper.IterateContractInfos(appA.Context(), func(addr sdk.AccAddress, info wasmxtypes.ContractInfo) bool {
		hexaddr := "0x" + hex.EncodeToString(addr)
		c, found := mapsys[hexaddr]
		if found {
			suite.Require().Equal(c.Label, info.Label, hex.EncodeToString(addr))
			count += 1
		}
		return false
	})

	// - erc20, derc20
	suite.Require().Equal(syslen-2, count)
}
