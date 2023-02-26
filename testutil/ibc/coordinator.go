package ibctesting

import (
	"testing"
	"time"

	ibcgotesting "github.com/cosmos/ibc-go/v6/testing"
)

var (
	globalStartTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
)

// NewCoordinator initializes Coordinator with N TestChain's
func NewCoordinator(t *testing.T, chainIds []string) *ibcgotesting.Coordinator {
	chains := make(map[string]*ibcgotesting.TestChain)
	coord := &ibcgotesting.Coordinator{
		T:           t,
		CurrentTime: globalStartTime,
	}

	// setup Cosmos chains
	ibcgotesting.DefaultTestingAppInit = ibcgotesting.SetupTestingApp

	for _, chainID := range chainIds {
		chains[chainID] = NewTestChain(t, coord, chainID)
	}

	coord.Chains = chains

	return coord
}
