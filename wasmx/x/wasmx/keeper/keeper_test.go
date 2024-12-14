package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	testutils "github.com/loredanacirstea/wasmx/x/wasmx/keeper/testutils"
)

var s *KeeperTestSuite

type KeeperTestSuite struct {
	testutils.KeeperTestSuite
}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)
}
