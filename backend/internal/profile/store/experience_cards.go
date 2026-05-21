package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/monshunter/easyinterview/backend/internal/profile"
)

// ListExperienceCardsByUser returns one page of experience_cards rows for
// userID using `updated_at DESC, id DESC` stable ordering. Cursor semantics:
// rows whose (updated_at, id) tuple is strictly less than the cursor.
func (r *Repository) ListExperienceCardsByUser(ctx context.Context, userID string, cursor *profile.ListCardsCursor, pageSize int32) (profile.ListCardsResult, error) {
	if r == nil || r.db == nil {
		return profile.ListCardsResult{}, fmt.Errorf("profile store db is nil")
	}
	if pageSize <= 0 {
		pageSize = profile.DefaultExperienceCardSize
	}
	if pageSize > profile.MaxExperienceCardSize {
		pageSize = profile.MaxExperienceCardSize
	}
	limit := int(pageSize) + 1
	var rows *sql.Rows
	var err error
	if cursor == nil {
		rows, err = r.db.QueryContext(ctx, `
select id, user_id, profile_id, title, company_name, situation, task, action, result,
       metrics, skills, language, source_type, source_ref_id, confidence, created_at, updated_at
  from experience_cards
 where user_id = $1 and archived_at is null
 order by updated_at desc, id desc
 limit $2`,
			userID, limit)
	} else {
		rows, err = r.db.QueryContext(ctx, `
select id, user_id, profile_id, title, company_name, situation, task, action, result,
       metrics, skills, language, source_type, source_ref_id, confidence, created_at, updated_at
  from experience_cards
 where user_id = $1 and archived_at is null
   and (updated_at, id) < ($2, $3)
 order by updated_at desc, id desc
 limit $4`,
			userID, cursor.UpdatedAt, cursor.ID, limit)
	}
	if err != nil {
		return profile.ListCardsResult{}, fmt.Errorf("list experience cards: %w", err)
	}
	defer rows.Close()

	out := profile.ListCardsResult{
		Items:    make([]profile.ExperienceCardRecord, 0, pageSize),
		PageSize: pageSize,
	}
	for rows.Next() {
		rec, err := scanExperienceCard(rows)
		if err != nil {
			return profile.ListCardsResult{}, fmt.Errorf("scan experience card: %w", err)
		}
		out.Items = append(out.Items, *rec)
	}
	if err := rows.Err(); err != nil {
		return profile.ListCardsResult{}, fmt.Errorf("iterate experience cards: %w", err)
	}
	if int32(len(out.Items)) > pageSize {
		// Drop the sentinel "has more" row; record the cursor to the LAST row
		// of the current page (not the sentinel).
		out.Items = out.Items[:pageSize]
		out.HasMore = true
	}
	if out.HasMore && len(out.Items) > 0 {
		last := out.Items[len(out.Items)-1]
		out.NextCursor = encodeCardCursor(last)
	}
	return out, nil
}

// CreateExperienceCard inserts a new experience_cards row and returns the
// stored record. profile_id is resolved from candidate_profiles by user_id;
// caller (handler.ensureProfileExists) guarantees a profile row exists.
func (r *Repository) CreateExperienceCard(ctx context.Context, id string, userID string, attrs profile.ExperienceCardAttrs, source profile.ExperienceCardSource) (*profile.ExperienceCardRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("profile store db is nil")
	}
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("profile store create: id is required")
	}
	row := r.db.QueryRowContext(ctx, `
with prof as (
  select id from candidate_profiles where user_id = $2 and deleted_at is null
)
insert into experience_cards (
  id, user_id, profile_id, title, company_name, situation, task, action, result,
  metrics, skills, language, source_type, source_ref_id, confidence,
  created_at, updated_at
) values (
  $1, $2,
  (select id from prof),
  $3, $4, $5, $6, $7, $8,
  '{}'::jsonb, $9, $10, $11, $12, $13,
  now(), now()
)
returning id, user_id, profile_id, title, company_name, situation, task, action, result,
          metrics, skills, language, source_type, source_ref_id, confidence,
          created_at, updated_at`,
		id,
		userID,
		attrs.Title,
		attrs.CompanyName,
		attrs.Situation,
		attrs.Task,
		attrs.Action,
		attrs.Result,
		pq.Array(attrs.Skills),
		attrs.Language,
		source.SourceType,
		nilUUID(source.SourceRefID),
		source.Confidence,
	)
	rec, err := scanExperienceCard(row)
	if isUniqueViolation(err) {
		return nil, profile.ErrValidationFailed
	}
	if err != nil {
		return nil, fmt.Errorf("create experience card: %w", err)
	}
	return rec, nil
}

// UpdateExperienceCard applies patch fields to a card owned by userID. Cross-
// user access (cardID belongs to another user) returns profile.ErrNotFound so
// the handler can map to 404 + RESOURCE_NOT_FOUND without exposing existence
// (spec D-8).
func (r *Repository) UpdateExperienceCard(ctx context.Context, cardID string, userID string, patch profile.ExperienceCardPatch) (*profile.ExperienceCardRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("profile store db is nil")
	}
	row := r.db.QueryRowContext(ctx, `
update experience_cards set
  title = case when $3::bool then $4 else title end,
  company_name = case when $5::bool then $6 else company_name end,
  situation = case when $7::bool then $8 else situation end,
  task = case when $9::bool then $10 else task end,
  action = case when $11::bool then $12 else action end,
  result = case when $13::bool then $14 else result end,
  skills = case when $15::bool then $16 else skills end,
  language = case when $17::bool then $18 else language end,
  updated_at = now()
 where id = $1 and user_id = $2 and archived_at is null
returning id, user_id, profile_id, title, company_name, situation, task, action, result,
          metrics, skills, language, source_type, source_ref_id, confidence,
          created_at, updated_at`,
		cardID,
		userID,
		patch.Title != nil, ptrStringValue(patch.Title),
		patch.CompanyName != nil, ptrStringValue(patch.CompanyName),
		patch.Situation != nil, ptrStringValue(patch.Situation),
		patch.Task != nil, ptrStringValue(patch.Task),
		patch.Action != nil, ptrStringValue(patch.Action),
		patch.Result != nil, ptrStringValue(patch.Result),
		patch.Skills != nil, ptrSkillsValue(patch.Skills),
		patch.Language != nil, ptrStringValue(patch.Language),
	)
	rec, err := scanExperienceCard(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, profile.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update experience card: %w", err)
	}
	return rec, nil
}

// DeleteExperienceCardsForUser removes all experience_cards belonging to
// userID. Returns the affected row count for privacy-delete audit tombstone.
func (r *Repository) DeleteExperienceCardsForUser(ctx context.Context, userID string) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("profile store db is nil")
	}
	res, err := r.db.ExecContext(ctx, `delete from experience_cards where user_id = $1`, userID)
	if err != nil {
		return 0, fmt.Errorf("delete experience cards: %w", err)
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

// CountExperienceCardsBySource returns a {source_type -> count} map for
// userID. Missing source_type keys default to 0 so callers always see the
// full taxonomy (spec D-11).
func (r *Repository) CountExperienceCardsBySource(ctx context.Context, userID string) (profile.SourceCounts, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("profile store db is nil")
	}
	out := make(profile.SourceCounts, len(profile.SourceTypes))
	for _, t := range profile.SourceTypes {
		out[t] = 0
	}
	rows, err := r.db.QueryContext(ctx, `
select source_type, count(*)::bigint
  from experience_cards
 where user_id = $1 and archived_at is null
 group by source_type`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("count experience cards: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sourceType string
		var count int64
		if err := rows.Scan(&sourceType, &count); err != nil {
			return nil, fmt.Errorf("scan count: %w", err)
		}
		out[sourceType] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate counts: %w", err)
	}
	return out, nil
}

func scanExperienceCard(row rowScanner) (*profile.ExperienceCardRecord, error) {
	var (
		rec         profile.ExperienceCardRecord
		companyName sql.NullString
		situation   sql.NullString
		task        sql.NullString
		action      sql.NullString
		result      sql.NullString
		metricsRaw  []byte
		skills      pq.StringArray
		sourceRefID sql.NullString
	)
	if err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&rec.ProfileID,
		&rec.Title,
		&companyName,
		&situation,
		&task,
		&action,
		&result,
		&metricsRaw,
		&skills,
		&rec.Language,
		&rec.SourceType,
		&sourceRefID,
		&rec.Confidence,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if companyName.Valid {
		rec.CompanyName = companyName.String
	}
	if situation.Valid {
		rec.Situation = situation.String
	}
	if task.Valid {
		rec.Task = task.String
	}
	if action.Valid {
		rec.Action = action.String
	}
	if result.Valid {
		rec.Result = result.String
	}
	rec.Skills = append([]string{}, skills...)
	if sourceRefID.Valid {
		v := sourceRefID.String
		rec.SourceRefID = &v
	}
	return &rec, nil
}

func ptrSkillsValue(skills *[]string) any {
	if skills == nil {
		return nil
	}
	return pq.Array(*skills)
}

func nilUUID(s *string) any {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	return *s
}

func encodeCardCursor(rec profile.ExperienceCardRecord) string {
	return cursorEncoder{UpdatedAt: rec.UpdatedAt, ID: rec.ID}.encode()
}
