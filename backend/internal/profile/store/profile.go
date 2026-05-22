package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/monshunter/easyinterview/backend/internal/profile"
)

// GetCandidateProfileByUser returns the candidate_profiles row for userID, or
// (nil, profile.ErrNotFound) if no row exists. This is the read-only path
// consumed by GetCandidateProfileForUser cross-owner internal API (spec D-13)
// — it must NOT seed.
func (r *Repository) GetCandidateProfileByUser(ctx context.Context, userID string) (*profile.CandidateProfileRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("profile store db is nil")
	}
	row := r.db.QueryRowContext(ctx, `
select user_id, headline, years_of_experience, "current_role",
       preferred_practice_language, ui_language, region,
       profile_version, created_at, updated_at
  from candidate_profiles
 where user_id = $1 and deleted_at is null`,
		userID,
	)
	rec, err := scanCandidateProfile(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, profile.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get candidate profile: %w", err)
	}
	return rec, nil
}

// SeedCandidateProfile inserts a Lite candidate_profiles row using the
// supplied user_settings defaults. Unique-violation indicates concurrent seed
// race and surfaces as profile.ErrValidationFailed so callers can re-read.
func (r *Repository) SeedCandidateProfile(ctx context.Context, userID string, defaults profile.UserSettings) (*profile.CandidateProfileRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("profile store db is nil")
	}
	if r.newID == nil {
		return nil, fmt.Errorf("profile store newID is nil")
	}
	row := r.db.QueryRowContext(ctx, `
insert into candidate_profiles (
  id, user_id, headline, summary_md, "current_role", years_of_experience,
  seniority_level, preferred_practice_language, ui_language, region,
  profile_version, created_at, updated_at
) values (
  $1, $2, null, null, null, null,
  null, $3, $4, $5,
  1, now(), now()
)
returning user_id, headline, years_of_experience, "current_role",
          preferred_practice_language, ui_language, region,
          profile_version, created_at, updated_at`,
		r.newID(),
		userID,
		defaults.PreferredPracticeLanguage,
		defaults.UILanguage,
		defaults.Region,
	)
	rec, err := scanCandidateProfile(row)
	if isUniqueViolation(err) {
		return nil, profile.ErrValidationFailed
	}
	if err != nil {
		return nil, fmt.Errorf("seed candidate profile: %w", err)
	}
	return rec, nil
}

// UpsertLite applies the supplied patch and returns the post-write row.
// profile_version is incremented atomically; updated_at is set to now().
// First-time callers (no existing row) trigger a seed using the supplied
// defaults so PATCH semantics also produce a usable row.
func (r *Repository) UpsertLite(ctx context.Context, userID string, patch profile.ProfilePatch, defaults profile.UserSettings) (*profile.CandidateProfileRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("profile store db is nil")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin upsert lite: %w", err)
	}
	defer tx.Rollback()

	var existsID string
	err = tx.QueryRowContext(ctx, `
select id from candidate_profiles
 where user_id = $1 and deleted_at is null
 for update`,
		userID,
	).Scan(&existsID)
	if errors.Is(err, sql.ErrNoRows) {
		// Insert with seed defaults, then proceed to apply patch via UPDATE
		// inside the same transaction so profile_version increments on this
		// write (matching D-12 semantics).
		if _, err := tx.ExecContext(ctx, `
insert into candidate_profiles (
  id, user_id, headline, summary_md, "current_role", years_of_experience,
  seniority_level, preferred_practice_language, ui_language, region,
  profile_version, created_at, updated_at
) values (
  $1, $2, null, null, null, null,
  null, $3, $4, $5,
  1, now(), now()
)`,
			r.newID(),
			userID,
			defaults.PreferredPracticeLanguage,
			defaults.UILanguage,
			defaults.Region,
		); err != nil {
			if isUniqueViolation(err) {
				return nil, profile.ErrValidationFailed
			}
			return nil, fmt.Errorf("upsert lite seed insert: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("upsert lite for update: %w", err)
	}

	row := tx.QueryRowContext(ctx, `
update candidate_profiles set
  headline = case when $2::bool then $3 else headline end,
  years_of_experience = case when $4::bool then $5 else years_of_experience end,
  "current_role" = case when $6::bool then $7 else "current_role" end,
  preferred_practice_language = case when $8::bool then $9 else preferred_practice_language end,
  ui_language = case when $10::bool then $11 else ui_language end,
  region = case when $12::bool then $13 else region end,
  profile_version = profile_version + 1,
  updated_at = now()
 where user_id = $1 and deleted_at is null
returning user_id, headline, years_of_experience, "current_role",
          preferred_practice_language, ui_language, region,
          profile_version, created_at, updated_at`,
		userID,
		patch.Headline != nil, ptrStringValue(patch.Headline),
		patch.YearsOfExperience != nil, ptrInt32Value(patch.YearsOfExperience),
		patch.CurrentRole != nil, ptrStringValue(patch.CurrentRole),
		patch.PreferredPracticeLanguage != nil, ptrStringValue(patch.PreferredPracticeLanguage),
		patch.UILanguage != nil, ptrStringValue(patch.UILanguage),
		patch.Region != nil, ptrStringValue(patch.Region),
	)
	rec, err := scanCandidateProfile(row)
	if err != nil {
		return nil, fmt.Errorf("upsert lite update: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit upsert lite: %w", err)
	}
	return rec, nil
}

// DeleteCandidateProfileForUser hard-deletes the candidate_profiles row for
// userID. Returns the number of rows removed (0 or 1).
func (r *Repository) DeleteCandidateProfileForUser(ctx context.Context, userID string) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("profile store db is nil")
	}
	res, err := r.db.ExecContext(ctx, `delete from candidate_profiles where user_id = $1`, userID)
	if err != nil {
		return 0, fmt.Errorf("delete candidate profile: %w", err)
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCandidateProfile(row rowScanner) (*profile.CandidateProfileRecord, error) {
	var (
		rec       profile.CandidateProfileRecord
		headline  sql.NullString
		yoe       sql.NullInt32
		currentRl sql.NullString
		region    sql.NullString
	)
	if err := row.Scan(
		&rec.UserID,
		&headline,
		&yoe,
		&currentRl,
		&rec.PreferredPracticeLanguage,
		&rec.UILanguage,
		&region,
		&rec.ProfileVersion,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if headline.Valid {
		v := headline.String
		rec.Headline = &v
	}
	if yoe.Valid {
		v := yoe.Int32
		rec.YearsOfExperience = &v
	}
	if currentRl.Valid {
		v := currentRl.String
		rec.CurrentRole = &v
	}
	if region.Valid {
		v := region.String
		rec.Region = &v
	}
	return &rec, nil
}

func ptrStringValue(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

func ptrInt32Value(v *int32) any {
	if v == nil {
		return nil
	}
	return *v
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return pqErr.Code == "23505"
}
