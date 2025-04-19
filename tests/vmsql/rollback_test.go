package keeper_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func TestSqliteSavepointRelease(t *testing.T) {
	db, err := sql.Open("sqlite3", "test.db")
	require.NoError(t, err)
	defer os.Remove("test.db")

	ctx := context.Background()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, name) VALUES (?, ?)`, 1, "Alice")
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	_, err = tx.Exec("SAVEPOINT sp1")
	require.NoError(t, err)

	// Perform reads and writes
	row := tx.QueryRow(`SELECT name FROM users WHERE id = ?`, 1)
	var name string
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Alice", name)

	_, err = tx.Exec(`INSERT INTO users (id, name) VALUES (?, ?)`, 2, "Bobby")
	require.NoError(t, err)

	_, err = tx.Exec(`REPLACE INTO users (id, name) VALUES (?, ?)`, 1, "Sam")
	require.NoError(t, err)

	row = tx.QueryRow(`SELECT name FROM users WHERE id = ?`, 1)
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Sam", name)

	// Finalize the savepoint and commit transaction
	_, err = tx.Exec(`RELEASE sp1`)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	row = db.QueryRow(`SELECT name FROM users WHERE id = ?`, 1)
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Sam", name)

	row = db.QueryRow(`SELECT name FROM users WHERE id = ?`, 2)
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Bobby", name)
}

func TestSqliteSavepointRollback(t *testing.T) {
	db, err := sql.Open("sqlite3", "test.db")
	require.NoError(t, err)
	defer os.Remove("test.db")

	ctx := context.Background()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, name) VALUES (?, ?)`, 1, "Alice")
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	_, err = tx.Exec("SAVEPOINT sp1")
	require.NoError(t, err)

	// Perform reads and writes
	row := tx.QueryRow(`SELECT name FROM users WHERE id = ?`, 1)
	var name string
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Alice", name)

	_, err = tx.Exec(`INSERT INTO users (id, name) VALUES (?, ?)`, 2, "Bobby")
	require.NoError(t, err)

	_, err = tx.Exec(`REPLACE INTO users (id, name) VALUES (?, ?)`, 1, "Sam")
	require.NoError(t, err)

	row = tx.QueryRow(`SELECT name FROM users WHERE id = ?`, 1)
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Sam", name)

	// rollback to savepoint
	_, err = tx.Exec(`ROLLBACK TO sp1`)
	require.NoError(t, err)

	row = db.QueryRow(`SELECT name FROM users WHERE id = ?`, 1)
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Alice", name)

	row = db.QueryRow(`SELECT name FROM users WHERE id = ?`, 2)
	err = row.Scan(&name)
	require.Error(t, err)
}

func TestSqliteSavepointRollbackNested(t *testing.T) {
	db, err := sql.Open("sqlite3", "test.db")
	require.NoError(t, err)
	defer os.Remove("test.db")

	ctx := context.Background()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (id, name) VALUES (?, ?)`, 1, "Alice")
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	_, err = tx.Exec("SAVEPOINT sp1")
	require.NoError(t, err)

	// Perform reads and writes
	row := tx.QueryRow(`SELECT name FROM users WHERE id = ?`, 1)
	var name string
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Alice", name)

	_, err = tx.Exec(`INSERT INTO users (id, name) VALUES (?, ?)`, 2, "Bobby")
	require.NoError(t, err)

	_, err = tx.Exec("SAVEPOINT sp2")
	require.NoError(t, err)

	_, err = tx.Exec(`REPLACE INTO users (id, name) VALUES (?, ?)`, 1, "Sam")
	require.NoError(t, err)

	row = tx.QueryRow(`SELECT name FROM users WHERE id = ?`, 1)
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Sam", name)

	// Rollback to savepoint sp2, but commit sp1
	_, err = tx.Exec(`ROLLBACK TO sp2`)
	require.NoError(t, err)

	_, err = tx.Exec(`RELEASE sp1`)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// sp2 was rolled back, so id 1 remains Alice
	row = db.QueryRow(`SELECT name FROM users WHERE id = ?`, 1)
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Alice", name)

	// sp1 was commited, so id2 should remain inserted
	row = db.QueryRow(`SELECT name FROM users WHERE id = ?`, 2)
	err = row.Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "Bobby", name)
}
