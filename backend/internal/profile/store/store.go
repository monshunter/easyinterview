// Package store implements the candidate_profiles + experience_cards
// persistence boundary defined by backend-profile spec §2.1 / §4.3.
package store

import (
	"database/sql"

	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
)

// Repository is the canonical SQL implementation of profile.Store. It is
// constructed by cmd/api via NewRepository(db).
type Repository struct {
	db    *sql.DB
	newID func() string
}

// NewRepository wires the SQL implementation backing profile.Store. Defaults:
// idx.NewID for primary keys; tests may override via NewRepositoryOptions.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db, newID: idx.NewID}
}

// NewRepositoryOptions exposes injection points for tests (e.g. deterministic
// id generator).
type NewRepositoryOptions struct {
	NewID func() string
}

// NewRepositoryWith builds a Repository with optional overrides.
func NewRepositoryWith(db *sql.DB, opts NewRepositoryOptions) *Repository {
	r := &Repository{db: db, newID: idx.NewID}
	if opts.NewID != nil {
		r.newID = opts.NewID
	}
	return r
}
