package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	ExecTx(ctx context.Context, fn func(*Queries) error) error
}

type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

func NewStore(db *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: db,
		Queries:  New(db),
	}
}

// ExecTx executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func (s *SQLStore) ExecTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.connPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx error: %w, rollback error: %w", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
