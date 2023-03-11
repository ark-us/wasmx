package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type KeeperTestSuite interface {
	Assert() *assert.Assertions
	Require() *require.Assertions
	Run(name string, subtest func()) bool
	SetT(t *testing.T)
	T() *testing.T

	Commit()
}
