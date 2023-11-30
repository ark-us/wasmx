package ibctesting

import (
	"testing"
	"time"

	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"
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
		chains[chainID] = NewTestChain(t, coord, chainID, index)
	}

	coord.Chains = chains

	return coord
}
