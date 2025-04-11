package db_sqlite

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	dbm "github.com/cosmos/cosmos-db"
)

var _ dbm.DB = (*SqliteChainDb)(nil)

const (
	driverName      = "sqlite3"
	dbName          = "ss.db?cache=shared&mode=rwc&_journal_mode=WAL"
	keyLatestHeight = "_RESERVED_latest_height"
	keyPruneHeight  = "_RESERVED_prune_height"

	reservedUpsertStmt = `
	INSERT INTO state_storage(key, value)
    VALUES(?, ?)
  ON CONFLICT(key) DO UPDATE SET
    value = ?;
	`
	upsertStmt = `
	INSERT INTO state_storage(key, value)
    VALUES(?, ?)
  ON CONFLICT(key) DO UPDATE SET
    value = ?;
	`
	delStmt = `
	UPDATE state_storage SET tombstone = ?
	WHERE id = (
		SELECT id FROM state_storage WHERE key = ?
	) AND tombstone = 0;
	`
)

type KeyValue struct {
	Key   string `db:"key" json:"key"`
	Value []byte `db:"value" json:"value"`
}

type SqliteChainDb struct {
	db *sql.DB
}

func NewSqliteChainDb(dbpath string) (*SqliteChainDb, error) {
	db, err := sql.Open(driverName, dbpath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite DB: %w", err)
	}

	stmt := `
	CREATE TABLE IF NOT EXISTS state_storage (
		id integer not null primary key,
		key varchar not null,
		value varchar not null,
		tombstone integer unsigned default 0,
		unique (key)
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_key ON state_storage (key);
	`
	_, err = db.Exec(stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to exec SQL statement: %w", err)
	}
	return &SqliteChainDb{db: db}, nil
}

func (s *SqliteChainDb) Close() error {
	var err error
	if s.db != nil {
		err = s.db.Close()
	}
	s.db = nil
	return err
}
func (s *SqliteChainDb) Delete(key []byte) error {
	s.Set(key, nil)
	return nil
}

// Get([]byte) ([]byte, error)
func (s *SqliteChainDb) Get(key []byte) ([]byte, error) {
	stmt, err := s.db.Prepare(`
	SELECT value, tombstone FROM state_storage
	WHERE key = ?
	LIMIT 1;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var (
		value []byte
		tomb  uint64
	)
	if err := stmt.QueryRow(key).Scan(&value, &tomb); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to query row: %w", err)
	}

	return value, nil
}

// Has(key []byte) (bool, error)
func (s *SqliteChainDb) Has(key []byte) (bool, error) {
	value, err := s.Get(key)
	if err != nil {
		return false, err
	}
	return len(value) > 0, nil
}
func (s *SqliteChainDb) Set(key []byte, value []byte) error {
	_, err := s.db.Exec(upsertStmt, key, value, value)
	if err != nil {
		return err
	}
	return nil
}

func (s *SqliteChainDb) SetSync(key []byte, value []byte) error {
	return s.Set(key, value)
}

func (s *SqliteChainDb) DeleteSync(key []byte) error {
	return s.Delete(key)
}

func (s *SqliteChainDb) Iterator(start, end []byte) (dbm.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, fmt.Errorf("key empty")
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, fmt.Errorf("start key after end key")
	}

	return newIterator(s, start, end, false)
}

func (s *SqliteChainDb) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, fmt.Errorf("key empty")
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, fmt.Errorf("start key after end key")
	}

	return newIterator(s, start, end, true)
}

func (s *SqliteChainDb) NewBatch() dbm.Batch {
	batch, err := NewBatch(s.db)
	if err != nil {
		panic(err)
	}
	return batch
}

func (s *SqliteChainDb) GetLatestVersion() (uint64, error) {
	stmt, err := s.db.Prepare(`
	SELECT value
	FROM state_storage
	WHERE key = ?
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var latestHeight uint64
	if err := stmt.QueryRow(keyLatestHeight).Scan(&latestHeight); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// in case of a fresh database
			return 0, nil
		}

		return 0, fmt.Errorf("failed to query row: %w", err)
	}
	return latestHeight, nil
}

func (s *SqliteChainDb) NewBatchWithSize(size int) dbm.Batch {
	return s.NewBatch()
}

func (s *SqliteChainDb) Print() error {
	itr, err := s.Iterator(nil, nil)
	if err != nil {
		return err
	}
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		value := itr.Value()
		fmt.Printf("[%X]:\t[%X]\n", key, value)
	}
	return nil
}

func (s *SqliteChainDb) Stats() map[string]string {
	// _stats := s.db.Stats()
	stats := make(map[string]string, 0)
	// for _, key := range keys {
	// 	stats[key] = s.db.Stats() // s.db.GetProperty(key)
	// }
	return stats
}
