package ibctesting

import (
	"testing"
	"time"

	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"

	mcfg "github.com/loredanacirstea/wasmx/config"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

var (
	globalStartTime = time.Now().UTC()
)

// NewCoordinator initializes Coordinator with N TestChain's
func NewCoordinator(t *testing.T, wasmVmMeta memc.IWasmVmMeta, compiledCacheDir string, chainIds []string, index int32) *ibcgotesting.Coordinator {
	chains := make(map[string]*ibcgotesting.TestChain)
	coord := &ibcgotesting.Coordinator{
		T:           t,
		CurrentTime: globalStartTime,
	}

	// setup Cosmos chains
	ibcgotesting.DefaultTestingAppInit = ibcgotesting.SetupTestingApp

	for _, chainID := range chainIds {
		config, err := mcfg.GetChainConfig(chainID)
		if err != nil {
			panic(err)
		}
		chains[chainID] = NewTestChain(t, wasmVmMeta, compiledCacheDir, coord, chainID, *config, index)
	}

	coord.Chains = chains

	return coord
}
