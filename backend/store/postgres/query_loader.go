package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// queryLoader wraps a database connection and query store to provide
// convenient methods for executing queries
type queryLoader struct {
	db      dbConn
	querier *queryStore
}

// dbConn is an interface that abstracts the database connection
// It can be either a *pgxpool.Pool or pgx.Tx
type dbConn interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// exec loads and executes a query that doesn't return any rows
func (l *queryLoader) exec(ctx context.Context, name QueryName, args ...interface{}) error {
	query, err := l.querier.get(name)
	if err != nil {
		return fmt.Errorf("getting %s query: %w", name, err)
	}

	_, err = l.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing %s: %w", name, err)
	}

	return nil
}

// queryRow loads and executes a query that returns a single row
func (l *queryLoader) queryRow(ctx context.Context, name QueryName, scanner func(pgx.Row) error, args ...interface{}) error {
	query, err := l.querier.get(name)
	if err != nil {
		return fmt.Errorf("getting %s query: %w", name, err)
	}

	row := l.db.QueryRow(ctx, query, args...)
	if err := scanner(row); err != nil {
		return fmt.Errorf("scanning %s result: %w", name, err)
	}

	return nil
}

// queryRows loads and executes a query that returns multiple rows
func (l *queryLoader) queryRows(ctx context.Context, name QueryName, scanner func(pgx.Rows) error, args ...interface{}) error {
	query, err := l.querier.get(name)
	if err != nil {
		return fmt.Errorf("getting %s query: %w", name, err)
	}

	rows, err := l.db.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing %s query: %w", name, err)
	}
	defer rows.Close()

	if err := scanner(rows); err != nil {
		return fmt.Errorf("scanning %s results: %w", name, err)
	}

	return rows.Err()
}
