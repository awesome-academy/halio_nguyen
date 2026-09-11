package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// pgUniqueViolation is SQLSTATE 23505.
const pgUniqueViolation = "23505"

// UniqueViolationField reports which request field a unique-constraint
// violation belongs to, using a constraint-name → field map the caller
// owns (e.g. {"categories_slug_key": "slug"}). Mapping the constraint
// after the fact — instead of pre-checking with a SELECT — is deliberate:
// a pre-check is a TOCTOU race under concurrent creates, whereas the
// constraint is authoritative (phase-03 step 3 / R3). ok is false for any
// other error, or for a 23505 on a constraint the caller did not map.
func UniqueViolationField(err error, constraints map[string]string) (field string, ok bool) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgUniqueViolation {
		return "", false
	}
	field, ok = constraints[pgErr.ConstraintName]
	return field, ok
}
