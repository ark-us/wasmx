package ibctesting

import (
	"testing"
	"time"

	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"

	mcfg "mythos/v1/config"
)

var (
	globalStartTime = time.Now().UTC()
)

// NewCoordinator initializes Coordinator with N TestChain's
func NewCoordinator(t *testing.T, chainIds []string, index int32) *ibcgotesting.Coordinator {
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
		chains[chainID] = NewTestChain(t, coord, chainID, *config, index)
	}

	coord.Chains = chains

	return coord
}
