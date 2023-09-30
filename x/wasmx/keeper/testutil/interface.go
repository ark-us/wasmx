package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ibcgotesting "github.com/cosmos/ibc-go/v8/testing"
)

type KeeperTestSuite interface {
	Assert() *assert.Assertions
	Require() *require.Assertions
	Run(name string, subtest func()) bool
	SetT(t *testing.T)
	T() *testing.T

	Commit()
	Coordinator() *ibcgotesting.Coordinator
	CommitNBlocks(chain *ibcgotesting.TestChain, n uint64)
}
