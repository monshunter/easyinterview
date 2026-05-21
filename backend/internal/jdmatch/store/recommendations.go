package store

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
)

// ListRecommendationsFilter scopes a cursor-paginated list query.
type ListRecommendationsFilter struct {
	PageSize int
	Cursor   string
}

// ListRecommendationsResult is the cursor-paginated read projection.
type ListRecommendationsResult struct {
	Items      []jdmatch.RecommendationRecord
	NextCursor string
	HasMore    bool
	PageSize   int
}

// ListRecommendationsByUser returns active (not-dismissed, not-deleted)
// rows for the supplied user, ordered by score DESC, recommended_at
// DESC, id DESC. Cursor encodes the last (score, recommendedAt, id)
// triple so pagination is stable across repeated calls.
func (r *Repository) ListRecommendationsByUser(ctx context.Context, userID string, filter ListRecommendationsFilter) (ListRecommendationsResult, error) {
	if r == nil || r.db == nil {
		return ListRecommendationsResult{}, fmt.Errorf("jdmatch store: db is nil")
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return ListRecommendationsResult{}, jdmatch.ErrUserIDRequired
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	cur, err := decodeRecommendationCursor(filter.Cursor)
	if err != nil {
		return ListRecommendationsResult{}, err
	}
	var (
		rows *sql.Rows
	)
	if cur == nil {
		rows, err = r.db.QueryContext(
			ctx,
			recommendationsListSQL(false),
			uid, pageSize+1,
		)
	} else {
		rows, err = r.db.QueryContext(
			ctx,
			recommendationsListSQL(true),
			uid, cur.Score, cur.RecommendedAt, cur.ID, pageSize+1,
		)
	}
	if err != nil {
		return ListRecommendationsResult{}, fmt.Errorf("jdmatch store: list recommendations: %w", err)
	}
	defer rows.Close()
	items := make([]jdmatch.RecommendationRecord, 0, pageSize)
	for rows.Next() {
		rec, err := scanRecommendationRow(rows)
		if err != nil {
			return ListRecommendationsResult{}, fmt.Errorf("jdmatch store: scan recommendation: %w", err)
		}
		items = append(items, rec)
	}
	if err := rows.Err(); err != nil {
		return ListRecommendationsResult{}, fmt.Errorf("jdmatch store: rows err: %w", err)
	}
	hasMore := len(items) > pageSize
	if hasMore {
		items = items[:pageSize]
	}
	next := ""
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		next = encodeRecommendationCursor(recommendationCursor{
			Score:         last.Score,
			RecommendedAt: last.RecommendedAt,
			ID:            last.ID,
		})
	}
	return ListRecommendationsResult{
		Items:      items,
		NextCursor: next,
		HasMore:    hasMore,
		PageSize:   pageSize,
	}, nil
}

// GetRecommendationByIDForUser returns a single row scoped to the
// supplied user. Cross-user lookups return ErrNotFound so handlers can
// project a canonical 404 RESOURCE_NOT_FOUND envelope.
func (r *Repository) GetRecommendationByIDForUser(ctx context.Context, userID, id string) (jdmatch.RecommendationRecord, error) {
	if r == nil || r.db == nil {
		return jdmatch.RecommendationRecord{}, fmt.Errorf("jdmatch store: db is nil")
	}
	uid := strings.TrimSpace(userID)
	rid := strings.TrimSpace(id)
	if uid == "" || rid == "" {
		return jdmatch.RecommendationRecord{}, jdmatch.ErrUserIDRequired
	}
	row := r.db.QueryRowContext(ctx, recommendationsSelectSQL()+" WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL", rid, uid)
	rec, err := scanRecommendationRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return jdmatch.RecommendationRecord{}, jdmatch.ErrNotFound
	}
	if err != nil {
		return jdmatch.RecommendationRecord{}, fmt.Errorf("jdmatch store: get recommendation: %w", err)
	}
	return rec, nil
}

// MarkRecommendationDismissedInput captures the dismiss request. The
// free note is persisted on the row but never written to log / audit /
// outbox (spec D-7 privacy red line).
type MarkRecommendationDismissedInput struct {
	ID       string
	UserID   string
	Reason   string
	FreeNote string
}

// MarkRecommendationDismissed sets dismissed_at = now plus reason +
// free_note. Returns the updated record. Returns ErrNotFound when the
// (id, user_id) tuple does not match a row, ErrAlreadyDismissed when
// the row was already dismissed.
func (r *Repository) MarkRecommendationDismissed(ctx context.Context, in MarkRecommendationDismissedInput) (jdmatch.RecommendationRecord, error) {
	if r == nil || r.db == nil {
		return jdmatch.RecommendationRecord{}, fmt.Errorf("jdmatch store: db is nil")
	}
	if strings.TrimSpace(in.ID) == "" || strings.TrimSpace(in.UserID) == "" {
		return jdmatch.RecommendationRecord{}, jdmatch.ErrUserIDRequired
	}
	now := r.now()
	reason := strings.TrimSpace(in.Reason)
	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}
	note := in.FreeNote
	var notePtr *string
	if note != "" {
		notePtr = &note
	}
	row := r.db.QueryRowContext(
		ctx,
		`UPDATE jd_match_recommendations
		SET dismissed_at = $3, dismiss_reason = COALESCE($4, dismiss_reason), dismiss_free_note = COALESCE($5, dismiss_free_note), updated_at = $3
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL AND dismissed_at IS NULL
		RETURNING `+recommendationsColumnList(),
		in.ID, in.UserID, now, reasonPtr, notePtr,
	)
	rec, err := scanRecommendationRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		// Either the row does not exist (ErrNotFound) or it was
		// already dismissed (ErrAlreadyDismissed). Disambiguate via
		// a fresh read.
		probe := r.db.QueryRowContext(ctx,
			`SELECT dismissed_at IS NOT NULL FROM jd_match_recommendations WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
			in.ID, in.UserID,
		)
		var alreadyDismissed bool
		if probeErr := probe.Scan(&alreadyDismissed); probeErr != nil {
			if errors.Is(probeErr, sql.ErrNoRows) {
				return jdmatch.RecommendationRecord{}, jdmatch.ErrNotFound
			}
			return jdmatch.RecommendationRecord{}, fmt.Errorf("jdmatch store: dismiss probe: %w", probeErr)
		}
		if alreadyDismissed {
			return jdmatch.RecommendationRecord{}, jdmatch.ErrAlreadyDismissed
		}
		return jdmatch.RecommendationRecord{}, jdmatch.ErrNotFound
	}
	if err != nil {
		return jdmatch.RecommendationRecord{}, fmt.Errorf("jdmatch store: dismiss recommendation: %w", err)
	}
	return rec, nil
}

// UpsertRecommendationInput is the agent_scan generator hand-off shape.
type UpsertRecommendationInput struct {
	ID                  string
	UserID              string
	Title               string
	Company             string
	CompanyTag          *string
	Level               *string
	Location            string
	Comp                *string
	PostedLabel         *string
	Score               int
	FitMust             int
	FitTotal            int
	FitPlus             int
	FitTotalPlus        int
	Reasons             []string
	Risks               []string
	Highlights          []string
	SourceURL           *string
	SourceLabel         *string
	NetworkNote         *string
	SimilarInterviewers *int
	InterviewHypotheses []string
	PromptVersion       string
	RubricVersion       string
	ModelID             string
	Language            string
	FeatureFlag         string
	DataSourceVersion   string
}

// UpsertRecommendation inserts a new recommendation row or updates the
// existing one (id = primary key). All AI provenance fields land in the
// row's typed columns so handlers can project them into the
// GenerationProvenance DTO without a JSONB scan.
func (r *Repository) UpsertRecommendation(ctx context.Context, in UpsertRecommendationInput) (jdmatch.RecommendationRecord, error) {
	if r == nil || r.db == nil {
		return jdmatch.RecommendationRecord{}, fmt.Errorf("jdmatch store: db is nil")
	}
	if strings.TrimSpace(in.ID) == "" || strings.TrimSpace(in.UserID) == "" {
		return jdmatch.RecommendationRecord{}, jdmatch.ErrUserIDRequired
	}
	now := r.now()
	fit, err := json.Marshal(map[string]int{
		"must":      in.FitMust,
		"total":     in.FitTotal,
		"plus":      in.FitPlus,
		"totalPlus": in.FitTotalPlus,
	})
	if err != nil {
		return jdmatch.RecommendationRecord{}, fmt.Errorf("jdmatch store: marshal fit: %w", err)
	}
	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO jd_match_recommendations (
			id, user_id, title, company, company_tag, level, location, comp, posted_label,
			score, fit, reasons, risks, highlights, source_url, source_label, network_note,
			similar_interviewers, interview_hypotheses, prompt_version, rubric_version,
			model_id, language, feature_flag, data_source_version, recommended_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $26)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			company = EXCLUDED.company,
			company_tag = EXCLUDED.company_tag,
			level = EXCLUDED.level,
			location = EXCLUDED.location,
			comp = EXCLUDED.comp,
			posted_label = EXCLUDED.posted_label,
			score = EXCLUDED.score,
			fit = EXCLUDED.fit,
			reasons = EXCLUDED.reasons,
			risks = EXCLUDED.risks,
			highlights = EXCLUDED.highlights,
			source_url = EXCLUDED.source_url,
			source_label = EXCLUDED.source_label,
			network_note = EXCLUDED.network_note,
			similar_interviewers = EXCLUDED.similar_interviewers,
			interview_hypotheses = EXCLUDED.interview_hypotheses,
			prompt_version = EXCLUDED.prompt_version,
			rubric_version = EXCLUDED.rubric_version,
			model_id = EXCLUDED.model_id,
			language = EXCLUDED.language,
			feature_flag = EXCLUDED.feature_flag,
			data_source_version = EXCLUDED.data_source_version,
			recommended_at = EXCLUDED.recommended_at,
			updated_at = EXCLUDED.updated_at
		RETURNING `+recommendationsColumnList(),
		in.ID, in.UserID, in.Title, in.Company, in.CompanyTag, in.Level, in.Location, in.Comp, in.PostedLabel,
		in.Score, fit, pq.Array(in.Reasons), pq.Array(in.Risks), pq.Array(in.Highlights),
		in.SourceURL, in.SourceLabel, in.NetworkNote, in.SimilarInterviewers,
		pq.Array(in.InterviewHypotheses), in.PromptVersion, in.RubricVersion,
		in.ModelID, in.Language, in.FeatureFlag, in.DataSourceVersion, now,
	)
	rec, err := scanRecommendationRow(row)
	if err != nil {
		return jdmatch.RecommendationRecord{}, fmt.Errorf("jdmatch store: upsert recommendation: %w", err)
	}
	return rec, nil
}

// DeleteRecommendationsForUser hard-deletes every recommendation owned
// by the supplied user (privacy delete cascade).
func (r *Repository) DeleteRecommendationsForUser(ctx context.Context, userID string) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("jdmatch store: db is nil")
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return 0, jdmatch.ErrUserIDRequired
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM jd_match_recommendations WHERE user_id = $1`, uid)
	if err != nil {
		return 0, fmt.Errorf("jdmatch store: delete recommendations: %w", err)
	}
	return res.RowsAffected()
}

// CountActiveRecommendationsByUser returns the count of recommendations
// for a user excluding dismissed/deleted rows.
func (r *Repository) CountActiveRecommendationsByUser(ctx context.Context, userID string) (int, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("jdmatch store: db is nil")
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return 0, jdmatch.ErrUserIDRequired
	}
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM jd_match_recommendations WHERE user_id = $1 AND dismissed_at IS NULL AND deleted_at IS NULL`,
		uid,
	).Scan(&n)
	return n, err
}

func recommendationsColumnList() string {
	return `id, user_id, title, company, company_tag, level, location, comp, posted_label,
		score, fit, reasons, risks, highlights, seen, dismissed_at, dismiss_reason,
		dismiss_free_note, source_url, source_label, network_note, similar_interviewers,
		interview_hypotheses, prompt_version, rubric_version, model_id, language,
		feature_flag, data_source_version, recommended_at, updated_at, deleted_at`
}

func recommendationsSelectSQL() string {
	return `SELECT ` + recommendationsColumnList() + ` FROM jd_match_recommendations`
}

func recommendationsListSQL(withCursor bool) string {
	clause := `WHERE user_id = $1 AND dismissed_at IS NULL AND deleted_at IS NULL`
	if withCursor {
		clause = `WHERE user_id = $1 AND dismissed_at IS NULL AND deleted_at IS NULL
		AND (score, recommended_at, id) < ($2, $3, $4)`
	}
	limit := "$2"
	if withCursor {
		limit = "$5"
	}
	return recommendationsSelectSQL() + " " + clause + " ORDER BY score DESC, recommended_at DESC, id DESC LIMIT " + limit
}

func scanRecommendationRow(row interface{ Scan(dest ...any) error }) (jdmatch.RecommendationRecord, error) {
	var (
		rec                jdmatch.RecommendationRecord
		fitRaw             []byte
		reasons            []string
		risks              []string
		highlights         []string
		interviewHypos     []string
		companyTag         sql.NullString
		level              sql.NullString
		comp               sql.NullString
		postedLabel        sql.NullString
		dismissedAt        sql.NullTime
		dismissReason      sql.NullString
		dismissFreeNote    sql.NullString
		sourceURL          sql.NullString
		sourceLabel        sql.NullString
		networkNote        sql.NullString
		similarInterviewer sql.NullInt32
		promptVersion      sql.NullString
		rubricVersion      sql.NullString
		modelID            sql.NullString
		deletedAt          sql.NullTime
	)
	if err := row.Scan(
		&rec.ID, &rec.UserID, &rec.Title, &rec.Company, &companyTag, &level, &rec.Location, &comp, &postedLabel,
		&rec.Score, &fitRaw, pq.Array(&reasons), pq.Array(&risks), pq.Array(&highlights),
		&rec.Seen, &dismissedAt, &dismissReason, &dismissFreeNote,
		&sourceURL, &sourceLabel, &networkNote, &similarInterviewer,
		pq.Array(&interviewHypos), &promptVersion, &rubricVersion, &modelID,
		&rec.Language, &rec.FeatureFlag, &rec.DataSourceVersion, &rec.RecommendedAt, &rec.UpdatedAt, &deletedAt,
	); err != nil {
		return jdmatch.RecommendationRecord{}, err
	}
	if len(fitRaw) > 0 {
		var fit map[string]int
		if err := json.Unmarshal(fitRaw, &fit); err == nil {
			rec.FitMust = fit["must"]
			rec.FitTotal = fit["total"]
			rec.FitPlus = fit["plus"]
			rec.FitTotalPlus = fit["totalPlus"]
		}
	}
	rec.Reasons = reasons
	rec.Risks = risks
	rec.Highlights = highlights
	rec.InterviewHypotheses = interviewHypos
	if companyTag.Valid {
		v := companyTag.String
		rec.CompanyTag = &v
	}
	if level.Valid {
		v := level.String
		rec.Level = &v
	}
	if comp.Valid {
		v := comp.String
		rec.Comp = &v
	}
	if postedLabel.Valid {
		v := postedLabel.String
		rec.PostedLabel = &v
	}
	if dismissedAt.Valid {
		v := dismissedAt.Time
		rec.DismissedAt = &v
	}
	if dismissReason.Valid {
		v := dismissReason.String
		rec.DismissReason = &v
	}
	if dismissFreeNote.Valid {
		v := dismissFreeNote.String
		rec.DismissFreeNote = &v
	}
	if sourceURL.Valid {
		v := sourceURL.String
		rec.SourceURL = &v
	}
	if sourceLabel.Valid {
		v := sourceLabel.String
		rec.SourceLabel = &v
	}
	if networkNote.Valid {
		v := networkNote.String
		rec.NetworkNote = &v
	}
	if similarInterviewer.Valid {
		v := int(similarInterviewer.Int32)
		rec.SimilarInterviewers = &v
	}
	if promptVersion.Valid {
		v := promptVersion.String
		rec.PromptVersion = &v
	}
	if rubricVersion.Valid {
		v := rubricVersion.String
		rec.RubricVersion = &v
	}
	if modelID.Valid {
		v := modelID.String
		rec.ModelID = &v
	}
	if deletedAt.Valid {
		v := deletedAt.Time
		rec.DeletedAt = &v
	}
	return rec, nil
}

type recommendationCursor struct {
	Score         int       `json:"s"`
	RecommendedAt time.Time `json:"r"`
	ID            string    `json:"i"`
}

func encodeRecommendationCursor(c recommendationCursor) string {
	raw, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeRecommendationCursor(s string) (*recommendationCursor, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("jdmatch store: invalid cursor: %w", err)
	}
	var c recommendationCursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("jdmatch store: invalid cursor payload: %w", err)
	}
	return &c, nil
}
