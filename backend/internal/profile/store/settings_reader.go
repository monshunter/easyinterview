package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/monshunter/easyinterview/backend/internal/profile"
)

// SettingsReader resolves backend-profile's UserSettings shape from the
// user_settings table maintained by backend-auth. backend-profile only reads
// from user_settings (spec D-10 — no write-back); missing rows fall back to
// B4 baseline defaults to keep first GetMyProfile call functional even before
// backend-auth has provisioned a row.
type SettingsReader struct {
	db *sql.DB
}

// NewSettingsReader wires a user_settings reader against the supplied *sql.DB.
func NewSettingsReader(db *sql.DB) *SettingsReader {
	return &SettingsReader{db: db}
}

// GetUserSettings returns the supplied user's preferred_practice_language /
// ui_language / region. Missing row returns B4 defaults (en / zh-CN / nil).
func (r *SettingsReader) GetUserSettings(ctx context.Context, userID string) (profile.UserSettings, error) {
	if r == nil || r.db == nil {
		return profile.UserSettings{}, fmt.Errorf("settings reader db is nil")
	}
	var (
		out    profile.UserSettings
		region sql.NullString
	)
	err := r.db.QueryRowContext(ctx, `
select preferred_practice_language, ui_language, region
  from user_settings
 where user_id = $1`,
		userID,
	).Scan(&out.PreferredPracticeLanguage, &out.UiLanguage, &region)
	if errors.Is(err, sql.ErrNoRows) {
		return profile.UserSettings{PreferredPracticeLanguage: "en", UiLanguage: "zh-CN"}, nil
	}
	if err != nil {
		return profile.UserSettings{}, fmt.Errorf("read user_settings: %w", err)
	}
	if region.Valid {
		v := region.String
		out.Region = &v
	}
	return out, nil
}
