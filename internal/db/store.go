package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	sqlcdb "github.com/nabhag8848/orchex/internal/db/sqlc"
)

// Store is the app's DB access: sqlc queries plus transactions.
// The handler never begins a connection itself.
type Store struct {
	pool *pgxpool.Pool
	*sqlcdb.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		Queries: sqlcdb.New(pool),
	}
}

func (s *Store) InTx(ctx context.Context, fn func(*sqlcdb.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(s.Queries.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
