package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// UserIdentity is the cross-owner read-only projection consumed by
// backend-jobs-recommendations/001 BuildJobMatchProfile (spec D-17 / D-18).
// It carries only fields safe to expose across owners — display name,
// optional avatar URL, and a masked email. Raw email, password hash,
// session details, and audit metadata never appear here.
type UserIdentity struct {
	DisplayName string
	AvatarURL   *string
	EmailMasked string
}

// GetUserIdentityForUser is the cross-owner internal API. Read-only:
//
//   - Does not modify the users row, user_settings, sessions, or any
//     auth state.
//   - Does not write audit_events.
//   - Returns ErrUserNotFound when the userID does not match any row.
//   - emailMasked is the canonical safe projection of the raw email
//     produced by backend-auth's maskEmail helper (first + *** + last char
//     of the local part, domain preserved; e.g. "a***e@example.com");
//     raw email is never returned.
func GetUserIdentityForUser(ctx context.Context, db *sql.DB, userID string) (UserIdentity, error) {
	if db == nil {
		return UserIdentity{}, ErrIdentityDBRequired
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return UserIdentity{}, ErrIdentityUserIDRequired
	}
	var (
		email       string
		displayName sql.NullString
	)
	err := db.QueryRowContext(
		ctx,
		"SELECT email, display_name FROM users WHERE id = $1 AND deleted_at IS NULL",
		uid,
	).Scan(&email, &displayName)
	if errors.Is(err, sql.ErrNoRows) {
		return UserIdentity{}, ErrUserNotFound
	}
	if err != nil {
		return UserIdentity{}, fmt.Errorf("auth: get identity: %w", err)
	}
	id := UserIdentity{
		EmailMasked: maskEmail(email),
	}
	if displayName.Valid && displayName.String != "" {
		id.DisplayName = displayName.String
	} else {
		id.DisplayName = AnonymousDisplayName
	}
	return id, nil
}

// AnonymousDisplayName is the non-PII fallback display name used when a
// user has not yet supplied one. Cross-owner callers (backend-jobs-
// recommendations BuildJobMatchProfile) also fall back here if the
// identity lookup fails. Exported so callers can reuse the same constant
// without inventing a divergent fallback string.
const AnonymousDisplayName = "Candidate"

var (
	// ErrUserNotFound indicates the userID did not match a non-deleted
	// users row. Callers should fall back to a non-PII anonymous
	// display name rather than block the whole endpoint.
	ErrUserNotFound = errors.New("auth: user identity not found")
	// ErrIdentityDBRequired is returned when GetUserIdentityForUser is
	// called with a nil *sql.DB.
	ErrIdentityDBRequired = errors.New("auth: identity requires a non-nil *sql.DB")
	// ErrIdentityUserIDRequired is returned when the caller does not
	// supply a userID.
	ErrIdentityUserIDRequired = errors.New("auth: identity requires a non-empty userID")
)
