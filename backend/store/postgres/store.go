package postgres

import (
	"context"
	"embed"
	"fmt"

	"chat-room/store"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed queries/*.sql
var queries embed.FS

//go:embed migrations/*.sql
var migrations embed.FS

type Store struct {
	db      *pgxpool.Pool
	querier *queryStore
	loader  *queryLoader
}

type Tx struct {
	*Store
	tx     pgx.Tx
	loader *queryLoader
}

func New(ctx context.Context, connString string) (*Store, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parsing connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	qs, err := newQueryStore(queries)
	if err != nil {
		return nil, fmt.Errorf("loading queries: %w", err)
	}

	s := &Store{
		db:      pool,
		querier: qs,
	}
	s.loader = &queryLoader{db: pool, querier: qs}

	return s, nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) BeginTx(ctx context.Context) (store.Transaction, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}

	t := &Tx{
		Store: s,
		tx:    tx,
	}
	t.loader = &queryLoader{db: tx, querier: s.querier}

	return t, nil
}

func (t *Tx) Commit() error {
	return t.tx.Commit(context.Background())
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback(context.Background())
}
