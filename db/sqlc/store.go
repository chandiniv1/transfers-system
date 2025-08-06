package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store defines all the methods your application will call, including custom transactions.
type Store interface {
	Querier

	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// SQLStore provides all functions to execute db queries and transactions.
type SQLStore struct {
	db *pgxpool.Pool
	*Queries
}

// NewStore creates a new SQLStore instance with the given database connection.
func NewStore(db *pgxpool.Pool) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx runs a function within a database transaction.
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := New(tx)

	if err := fn(q); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
