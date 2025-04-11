package db_sqlite

import (
	"database/sql"
	"fmt"

	dbm "github.com/cosmos/cosmos-db"
)

var _ dbm.Batch = (*Batch)(nil)

type batchAction int

const (
	batchActionSet batchAction = 0
	batchActionDel batchAction = 1
)

type batchOp struct {
	action     batchAction
	key, value []byte
}

type Batch struct {
	db   *sql.DB
	tx   *sql.Tx
	ops  []batchOp
	size int
}

func NewBatch(db *sql.DB) (*Batch, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL transaction: %w", err)
	}

	return &Batch{
		db:  db,
		tx:  tx,
		ops: make([]batchOp, 0),
	}, nil
}

func (b *Batch) Size() int {
	return b.size
}

func (b *Batch) Reset() error {
	b.ops = nil
	b.ops = make([]batchOp, 0)
	b.size = 0

	tx, err := b.db.Begin()
	if err != nil {
		return err
	}

	b.tx = tx
	return nil
}

func (b *Batch) Set(key, value []byte) error {
	b.size += len(key) + len(value)
	b.ops = append(b.ops, batchOp{action: batchActionSet, key: key, value: value})
	return nil
}

func (b *Batch) Delete(key []byte) error {
	b.size += len(key)
	b.ops = append(b.ops, batchOp{action: batchActionDel, key: key})
	return nil
}

func (b *Batch) Write() error {
	_, err := b.tx.Exec(reservedUpsertStmt, keyLatestHeight, 0, 0)

	if err != nil {
		return fmt.Errorf("failed to exec reserved upsert SQL statement: %w", err)
	}

	for _, op := range b.ops {
		switch op.action {
		case batchActionSet:
			_, err := b.tx.Exec(upsertStmt, op.key, op.value, op.value)
			if err != nil {
				return fmt.Errorf("failed to exec batch set SQL statement: %w", err)
			}

		case batchActionDel:
			_, err := b.tx.Exec(delStmt, op.key)
			if err != nil {
				return fmt.Errorf("failed to exec batch del SQL statement: %w", err)
			}
		}
	}

	if err := b.tx.Commit(); err != nil {
		return fmt.Errorf("failed to write SQL transaction: %w", err)
	}

	return nil
}

// Close implements Batch.
func (b *Batch) Close() error {
	// if b.batch != nil {
	// 	b.batch.Destroy()
	// 	b.batch = nil
	// }
	return nil
}

// GetByteSize implements Batch
func (b *Batch) GetByteSize() (int, error) {
	return b.size, nil
}

// WriteSync implements Batch.
func (b *Batch) WriteSync() error {
	if b.tx == nil {
		return errBatchClosed
	}
	// err := b.db.db.Write(b.db.woSync, b.batch)
	err := b.Write()
	if err != nil {
		return err
	}
	// Make sure batch cannot be used afterwards. Callers should still call Close(), for errors.
	return b.Close()
}
