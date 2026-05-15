package review

import (
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) checkDB() error {
	if r == nil || r.db == nil {
		return fmt.Errorf("review store db is nil")
	}
	return nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
