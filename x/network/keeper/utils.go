package keeper

import (
	"path/filepath"

	dbm "github.com/cosmos/cosmos-db"
)

func OpenDBNetwork(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("applicationnetwork", backendType, dataDir)
}
