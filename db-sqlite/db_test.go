package db_sqlite_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	sqlite "github.com/loredanacirstea/db-sqlite"
)

func TestDb(t *testing.T) {
	db, err := sqlite.NewSqliteChainDb("testdb.db")
	require.NoError(t, err)

	// Set
	err = db.Set([]byte{1, 2, 4}, []byte{1, 1, 1})
	require.NoError(t, err)
	value, err := db.Get([]byte{1, 2, 4})
	require.NoError(t, err)
	require.Equal(t, []byte{1, 1, 1}, value)

	// Batch
	batch := db.NewBatchWithSize(100000)
	err = batch.Set([]byte{1, 2, 3}, []byte{2, 2, 2})
	require.NoError(t, err)
	err = batch.Write()
	require.NoError(t, err)

	err = batch.Close()
	require.NoError(t, err)

	value, err = db.Get([]byte{1, 2, 3})
	require.NoError(t, err)
	require.Equal(t, []byte{2, 2, 2}, value)

	err = os.Remove("testdb.db")
	require.NoError(t, err)
}
