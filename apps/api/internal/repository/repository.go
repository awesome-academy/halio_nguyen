package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is the narrow subset of *pgxpool.Pool and pgx.Tx that every repository
// method depends on. Accepting DB instead of a concrete pool type lets a
// service open a transaction and hand the same repository methods an open
// pgx.Tx transparently — no second set of *Tx-flavored methods is needed to
// support a create-in-one-transaction or cancel-in-one-transaction flow.
type DB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Compile-time proof that both the pool and an open transaction satisfy DB,
// so a service can pass either interchangeably to a repository method.
var (
	_ DB = (*pgxpool.Pool)(nil)
	_ DB = (pgx.Tx)(nil)
)
