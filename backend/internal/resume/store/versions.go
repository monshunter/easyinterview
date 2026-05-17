package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

// CreateStructuredMasterFromAsset creates the first active structured master
// version for a ready resume asset. The asset row lock serializes competing
// confirmations for the same resume asset; the partial unique index remains the
// final concurrency backstop.
func (r *Repository) CreateStructuredMasterFromAsset(ctx context.Context, in CreateStructuredMasterInput) (VersionRecord, error) {
	if r == nil || r.db == nil {
		return VersionRecord{}, fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	profile := in.StructuredProfile
	if len(profile) == 0 {
		profile = json.RawMessage(`{}`)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return VersionRecord{}, fmt.Errorf("begin structured master create: %w", err)
	}
	defer tx.Rollback()

	var parseStatus string
	err = tx.QueryRowContext(ctx, `
select parse_status
from resume_assets
where id = $1 and user_id = $2 and deleted_at is null
for update`,
		in.ResumeAssetID,
		in.UserID,
	).Scan(&parseStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return VersionRecord{}, ErrAssetNotFound
	}
	if err != nil {
		return VersionRecord{}, fmt.Errorf("lock resume asset for structured master: %w", err)
	}
	if sharedtypes.TargetJobParseStatus(parseStatus) != sharedtypes.TargetJobParseStatusReady {
		return VersionRecord{}, ErrAssetParseNotReady
	}

	rec, err := scanVersion(tx.QueryRowContext(ctx, `
insert into resume_versions (
  id, user_id, resume_asset_id, parent_version_id, version_type, target_job_id,
  display_name, seed_strategy, focus_angle, structured_profile, match_score,
  prompt_version, rubric_version, model_id, provider, created_at, updated_at
) values ($1,$2,$3,null,$4,null,$5,null,null,$6,null,$7,$8,$9,$10,$11,$11)
returning id, user_id, resume_asset_id, parent_version_id, version_type, target_job_id,
          display_name, seed_strategy, focus_angle, structured_profile, match_score,
          prompt_version, rubric_version, model_id, provider, created_at, updated_at, deleted_at`,
		in.VersionID,
		in.UserID,
		in.ResumeAssetID,
		string(sharedtypes.ResumeVersionTypeStructuredMaster),
		in.DisplayName,
		profile,
		nullableString(in.Provenance.PromptVersion),
		nullableString(in.Provenance.RubricVersion),
		nullableString(in.Provenance.ModelID),
		nullableString(in.Provenance.Provider),
		now,
	))
	if isStructuredMasterUniqueViolation(err) {
		return VersionRecord{}, ErrStructuredMasterAlreadyExists
	}
	if err != nil {
		return VersionRecord{}, fmt.Errorf("insert structured master resume version: %w", err)
	}
	rec.Provenance = in.Provenance
	if err := tx.Commit(); err != nil {
		return VersionRecord{}, fmt.Errorf("commit structured master create: %w", err)
	}
	return rec, nil
}

func (r *Repository) GetVersionByID(ctx context.Context, userID string, versionID string) (VersionRecord, error) {
	if r == nil || r.db == nil {
		return VersionRecord{}, fmt.Errorf("resume store db is nil")
	}
	rec, err := scanVersion(r.db.QueryRowContext(ctx, `
select id, user_id, resume_asset_id, parent_version_id, version_type, target_job_id,
       display_name, seed_strategy, focus_angle, structured_profile, match_score,
       prompt_version, rubric_version, model_id, provider, created_at, updated_at, deleted_at
from resume_versions
where id = $1 and user_id = $2 and deleted_at is null`,
		versionID,
		userID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return VersionRecord{}, ErrVersionNotFound
	}
	if err != nil {
		return VersionRecord{}, err
	}
	return rec, nil
}

func (r *Repository) ListVersionsByAsset(ctx context.Context, userID string, assetID string, filter VersionListFilter) (VersionListResult, error) {
	if r == nil || r.db == nil {
		return VersionListResult{}, fmt.Errorf("resume store db is nil")
	}
	var exists int
	if err := r.db.QueryRowContext(ctx, `
select 1 from resume_assets
where id = $1 and user_id = $2 and deleted_at is null`,
		assetID,
		userID,
	).Scan(&exists); errors.Is(err, sql.ErrNoRows) {
		return VersionListResult{}, ErrAssetNotFound
	} else if err != nil {
		return VersionListResult{}, err
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = sharedtypes.DefaultPageSize
	}
	if pageSize > sharedtypes.MaxPageSize {
		pageSize = sharedtypes.MaxPageSize
	}
	limit := pageSize + 1
	args := []any{userID, assetID, limit}
	query := `
select id, user_id, resume_asset_id, parent_version_id, version_type, target_job_id,
       display_name, seed_strategy, focus_angle, structured_profile, match_score,
       prompt_version, rubric_version, model_id, provider, created_at, updated_at, deleted_at
from resume_versions
where user_id = $1 and resume_asset_id = $2 and deleted_at is null`
	if filter.Cursor != "" {
		updatedAt, id, err := decodeCursor(filter.Cursor)
		if err != nil {
			return VersionListResult{}, ErrInvalidCursor
		}
		args = []any{userID, assetID, updatedAt, id, limit}
		query += ` and (updated_at, id) < ($3, $4)`
	}
	query += ` order by updated_at desc, id desc limit $` + fmt.Sprint(len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return VersionListResult{}, err
	}
	defer rows.Close()
	items := make([]VersionRecord, 0, pageSize)
	for rows.Next() {
		rec, err := scanVersion(rows)
		if err != nil {
			return VersionListResult{}, err
		}
		items = append(items, rec)
	}
	if err := rows.Err(); err != nil {
		return VersionListResult{}, err
	}
	hasMore := len(items) > pageSize
	if hasMore {
		items = items[:pageSize]
	}
	nextCursor := ""
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = encodeCursor(last.UpdatedAt, last.ID)
	}
	return VersionListResult{Items: items, NextCursor: nextCursor, HasMore: hasMore, PageSize: pageSize}, nil
}

func (r *Repository) UpdateVersionPatch(ctx context.Context, in VersionUpdateInput) (VersionRecord, error) {
	if r == nil || r.db == nil {
		return VersionRecord{}, fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return VersionRecord{}, fmt.Errorf("begin resume version update: %w", err)
	}
	defer tx.Rollback()

	current, err := scanVersion(tx.QueryRowContext(ctx, `
select id, user_id, resume_asset_id, parent_version_id, version_type, target_job_id,
       display_name, seed_strategy, focus_angle, structured_profile, match_score,
       prompt_version, rubric_version, model_id, provider, created_at, updated_at, deleted_at
from resume_versions
where id = $1 and user_id = $2 and deleted_at is null
for update`,
		in.VersionID,
		in.UserID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return VersionRecord{}, ErrVersionNotFound
	}
	if err != nil {
		return VersionRecord{}, fmt.Errorf("lock resume version for update: %w", err)
	}

	displayName := current.DisplayName
	if in.DisplayNameSet && in.DisplayName != nil {
		displayName = *in.DisplayName
	}
	focusAngle := current.FocusAngle
	if in.FocusAngleSet {
		focusAngle = in.FocusAngle
	}
	matchScore := current.MatchScore
	if in.MatchScoreSet {
		matchScore = in.MatchScore
	}
	structuredProfile := append(json.RawMessage(nil), current.StructuredProfile...)
	if in.StructuredProfileSet {
		merged, err := mergeStructuredProfile(current.StructuredProfile, in.StructuredProfilePatch, in.StructuredProfile)
		if err != nil {
			return VersionRecord{}, err
		}
		structuredProfile = merged
	}

	updated, err := scanVersion(tx.QueryRowContext(ctx, `
update resume_versions
set display_name = $1,
    focus_angle = $2,
    match_score = $3,
    structured_profile = $4,
    updated_at = $5
where id = $6 and user_id = $7 and deleted_at is null
returning id, user_id, resume_asset_id, parent_version_id, version_type, target_job_id,
          display_name, seed_strategy, focus_angle, structured_profile, match_score,
          prompt_version, rubric_version, model_id, provider, created_at, updated_at, deleted_at`,
		displayName,
		nullableStringPtr(focusAngle),
		nullableFloatPtr(matchScore),
		[]byte(structuredProfile),
		now,
		in.VersionID,
		in.UserID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return VersionRecord{}, ErrVersionNotFound
	}
	if err != nil {
		return VersionRecord{}, fmt.Errorf("update resume version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return VersionRecord{}, fmt.Errorf("commit resume version update: %w", err)
	}
	return updated, nil
}

func mergeStructuredProfile(existing json.RawMessage, patch map[string]any, patchRaw json.RawMessage) (json.RawMessage, error) {
	base := map[string]any{}
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &base); err != nil {
			return nil, fmt.Errorf("decode existing structured profile: %w", err)
		}
	}
	if patch == nil && len(patchRaw) > 0 {
		if err := json.Unmarshal(patchRaw, &patch); err != nil {
			return nil, fmt.Errorf("decode structured profile patch: %w", err)
		}
	}
	patch = cloneAnyMap(patch)
	delete(patch, "provenance")
	deepMergeMap(base, patch)
	raw, err := json.Marshal(base)
	if err != nil {
		return nil, fmt.Errorf("encode merged structured profile: %w", err)
	}
	return raw, nil
}

func deepMergeMap(dst map[string]any, src map[string]any) {
	for key, value := range src {
		srcMap, srcOK := value.(map[string]any)
		dstMap, dstOK := dst[key].(map[string]any)
		if srcOK && dstOK {
			deepMergeMap(dstMap, srcMap)
			continue
		}
		dst[key] = cloneAnyValue(value)
	}
}

func cloneAnyMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = cloneAnyValue(value)
	}
	return out
}

func cloneAnyValue(in any) any {
	switch v := in.(type) {
	case map[string]any:
		return cloneAnyMap(v)
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = cloneAnyValue(item)
		}
		return out
	default:
		return v
	}
}

func nullableFloatPtr(in *float64) any {
	if in == nil {
		return nil
	}
	return *in
}

func scanVersion(row rowScanner) (VersionRecord, error) {
	var rec VersionRecord
	var parentVersionID, targetJobID, seedStrategy, focusAngle sql.NullString
	var matchScore sql.NullFloat64
	var promptVersion, rubricVersion, modelID, provider sql.NullString
	var structuredProfile []byte
	var versionType string
	var deletedAt sql.NullTime
	if err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&rec.ResumeAssetID,
		&parentVersionID,
		&versionType,
		&targetJobID,
		&rec.DisplayName,
		&seedStrategy,
		&focusAngle,
		&structuredProfile,
		&matchScore,
		&promptVersion,
		&rubricVersion,
		&modelID,
		&provider,
		&rec.CreatedAt,
		&rec.UpdatedAt,
		&deletedAt,
	); err != nil {
		return VersionRecord{}, err
	}
	rec.ParentVersionID = stringPtrFromNull(parentVersionID)
	rec.VersionType = sharedtypes.ResumeVersionType(versionType)
	rec.TargetJobID = stringPtrFromNull(targetJobID)
	if seedStrategy.Valid {
		v := sharedtypes.ResumeSeedStrategy(seedStrategy.String)
		rec.SeedStrategy = &v
	}
	rec.FocusAngle = stringPtrFromNull(focusAngle)
	if len(structuredProfile) == 0 {
		structuredProfile = []byte(`{}`)
	}
	rec.StructuredProfile = append(json.RawMessage(nil), structuredProfile...)
	if matchScore.Valid {
		rec.MatchScore = &matchScore.Float64
	}
	rec.PromptVersion = stringPtrFromNull(promptVersion)
	rec.RubricVersion = stringPtrFromNull(rubricVersion)
	rec.ModelID = stringPtrFromNull(modelID)
	rec.Provider = stringPtrFromNull(provider)
	if rec.PromptVersion != nil {
		rec.Provenance.PromptVersion = *rec.PromptVersion
	}
	if rec.RubricVersion != nil {
		rec.Provenance.RubricVersion = *rec.RubricVersion
	}
	if rec.ModelID != nil {
		rec.Provenance.ModelID = *rec.ModelID
	}
	if rec.Provider != nil {
		rec.Provenance.Provider = *rec.Provider
	}
	if deletedAt.Valid {
		rec.DeletedAt = &deletedAt.Time
	}
	return rec, nil
}

func isStructuredMasterUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code) == "23505" && pqErr.Constraint == "uq_resume_versions_structured_master_per_asset"
	}
	return false
}
