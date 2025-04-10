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
	driverName        = "sqlite3"
	dbName            = "ss.db?cache=shared&mode=rwc&_journal_mode=WAL"
	reservedStoreKey  = "_RESERVED_"
	keyLatestHeight   = "latest_height"
	keyPruneHeight    = "prune_height"
	valueRemovedStore = "removed_store"

	reservedUpsertStmt = `
	INSERT INTO state_storage(store_key, key, value, version)
    VALUES(?, ?, ?, ?)
  ON CONFLICT(store_key, key, version) DO UPDATE SET
    value = ?;
	`
	upsertStmt = `
	INSERT INTO state_storage(store_key, key, value, version)
    VALUES(?, ?, ?, ?)
  ON CONFLICT(store_key, key, version) DO UPDATE SET
    value = ?;
	`
	delStmt = `
	UPDATE state_storage SET tombstone = ?
	WHERE id = (
		SELECT id FROM state_storage WHERE store_key = ? AND key = ? AND version <= ? ORDER BY version DESC LIMIT 1
	) AND tombstone = 0;
	`
	defaultVersion = uint64(0)
)

type KeyValue struct {
	Key   string `db:"key" json:"key"`
	Value []byte `db:"value" json:"value"`
}

type SqliteChainDb struct {
	db *sql.DB

	// earliestVersion defines the earliest version set in the database, which is
	// only updated when the database is pruned.
	earliestVersion uint64
}

func NewSqliteChainDb(dbpath string) (*SqliteChainDb, error) {
	fmt.Println("--NewSqliteChainDb---", dbpath)
	// db, err := sql.Open(driverName, filepath.Join(dataDir, dbName))
	db, err := sql.Open(driverName, dbpath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite DB: %w", err)
	}

	stmt := `
	CREATE TABLE IF NOT EXISTS state_storage (
		id integer not null primary key,
		store_key varchar not null,
		key varchar not null,
		value varchar not null,
		version integer unsigned not null,
		tombstone integer unsigned default 0,
		unique (store_key, key, version)
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_store_key_version ON state_storage (store_key, key, version);
	`
	_, err = db.Exec(stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	pruneHeight, err := getPruneHeight(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get prune height: %w", err)
	}
	fmt.Println("--NewSqliteChainDb pruneHeight---", pruneHeight)
	return &SqliteChainDb{db: db, earliestVersion: pruneHeight}, nil
}

func getPruneHeight(storage *sql.DB) (uint64, error) {
	fmt.Println("--getPruneHeight")
	stmt, err := storage.Prepare(`SELECT value FROM state_storage WHERE store_key = ? AND key = ?`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var value uint64
	if err := stmt.QueryRow(reservedStoreKey, keyPruneHeight).Scan(&value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}

		return 0, fmt.Errorf("failed to query row: %w", err)
	}
	fmt.Println("--getPruneHeight", value)
	return value, nil
}

func (s *SqliteChainDb) Close() error {
	fmt.Println("--Close")
	err := s.db.Close()
	s.db = nil
	return err
}
func (s *SqliteChainDb) Delete(key []byte) error {
	fmt.Println("--Delete")
	s.Set(key, nil)
	return nil
}

// Get([]byte) ([]byte, error)
func (s *SqliteChainDb) Get(key []byte) ([]byte, error) {
	fmt.Println("--Get key: ", key, string(key))
	storeKey := []byte{}
	targetVersion := defaultVersion
	fmt.Println("--Get version: ", targetVersion, s.earliestVersion)
	if targetVersion < s.earliestVersion {
		return nil, fmt.Errorf("version mismatch: earliestVersion %d, targetVersion %d", s.earliestVersion, targetVersion)
	}

	stmt, err := s.db.Prepare(`
	SELECT value, tombstone FROM state_storage
	WHERE store_key = ? AND key = ? AND version <= ?
	ORDER BY version DESC LIMIT 1;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var (
		value []byte
		tomb  uint64
	)
	fmt.Println("--Get QueryRow: ")
	if err := stmt.QueryRow(storeKey, key, targetVersion).Scan(&value, &tomb); err != nil {
		fmt.Println("--Get QueryRow err: ", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to query row: %w", err)
	}

	fmt.Println("--Get value: ", tomb, targetVersion, value == nil, value, string(value))

	// A tombstone of zero or a target version that is less than the tombstone
	// version means the key is not deleted at the target version.
	if tomb == 0 || targetVersion < tomb {
		return value, nil
	}

	// the value is considered deleted
	return nil, nil
}

// Has(key []byte) (bool, error)
func (s *SqliteChainDb) Has(key []byte) (bool, error) {
	fmt.Println("--Has--", key, string(key))
	value, err := s.Get(key)
	if err != nil {
		return false, err
	}
	return len(value) > 0, nil
}
func (s *SqliteChainDb) Set(key []byte, value []byte) error {
	storeKey := []byte{}
	version := defaultVersion
	fmt.Println("--Set key: ", key, string(key))
	fmt.Println("--Set value: ", value, string(value))

	_, err := s.db.Exec(upsertStmt, storeKey, key, value, version, value)
	if err != nil {
		fmt.Println("--Set err: ", err)
		return err
	}
	return nil
}

func (s *SqliteChainDb) SetSync(key []byte, value []byte) error {
	fmt.Println("--SetSync--", key, string(key))
	return s.Set(key, value)
}

func (s *SqliteChainDb) DeleteSync(key []byte) error {
	fmt.Println("--DeleteSync--", key, string(key))
	return s.Delete(key)
}

func (s *SqliteChainDb) Iterator(start, end []byte) (dbm.Iterator, error) {
	fmt.Println("--Iterator start--", start, string(start))
	fmt.Println("--Iterator end--", end, string(end))
	storeKey := []byte{}
	targetVersion := defaultVersion
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, fmt.Errorf("key empty")
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, fmt.Errorf("start key after end key")
	}

	return newIterator(s, storeKey, targetVersion, start, end, false)
}

func (s *SqliteChainDb) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	fmt.Println("--ReverseIterator start--", start, string(start))
	fmt.Println("--ReverseIterator end--", end, string(end))
	storeKey := []byte{}
	targetVersion := defaultVersion
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, fmt.Errorf("key empty")
	}

	if start != nil && end != nil && bytes.Compare(start, end) > 0 {
		return nil, fmt.Errorf("start key after end key")
	}

	return newIterator(s, storeKey, targetVersion, start, end, true)
}

func (s *SqliteChainDb) NewBatch() dbm.Batch {
	fmt.Println("--NewBatch--")
	batch, err := NewBatch(s.db, 1)
	if err != nil {
		panic(err)
	}
	return batch
}

func (s *SqliteChainDb) GetLatestVersion() (uint64, error) {
	fmt.Println("--GetLatestVersion--")
	stmt, err := s.db.Prepare(`
	SELECT value
	FROM state_storage
	WHERE store_key = ? AND key = ?
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare SQL statement: %w", err)
	}

	defer stmt.Close()

	var latestHeight uint64
	if err := stmt.QueryRow(reservedStoreKey, keyLatestHeight).Scan(&latestHeight); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// in case of a fresh database
			return 0, nil
		}

		return 0, fmt.Errorf("failed to query row: %w", err)
	}
	fmt.Println("--GetLatestVersion--", latestHeight)
	return latestHeight, nil
}

func (s *SqliteChainDb) VersionExists(v uint64) (bool, error) {
	fmt.Println("--VersionExists--", v)
	latestVersion, err := s.GetLatestVersion()
	if err != nil {
		return false, err
	}
	fmt.Println("--VersionExists--", latestVersion >= v && v >= s.earliestVersion)
	return latestVersion >= v && v >= s.earliestVersion, nil
}

func (s *SqliteChainDb) SetLatestVersion(version uint64) error {
	_, err := s.db.Exec(reservedUpsertStmt, reservedStoreKey, keyLatestHeight, version, 0, version)
	if err != nil {
		return fmt.Errorf("failed to exec SQL statement: %w", err)
	}

	return nil
}

func (s *SqliteChainDb) NewBatchWithSize(size int) dbm.Batch {
	fmt.Println("--NewBatchWithSize--", size)
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
