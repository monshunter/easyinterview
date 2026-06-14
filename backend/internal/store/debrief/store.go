package debrief

import (
	"context"
	"database/sql"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	domain "github.com/monshunter/easyinterview/backend/internal/debrief"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) checkDB() error {
	if r == nil || r.db == nil {
		return fmt.Errorf("debrief store db is nil")
	}
	return nil
}

func (r *Repository) CreateDebrief(ctx context.Context, in domain.CreateDebriefStoreInput) (domain.CreateDebriefResult, error) {
	if err := r.checkDB(); err != nil {
		return domain.CreateDebriefResult{}, err
	}
	rawQuestions, err := json.Marshal(rawQuestionPayloads(in.Questions))
	if err != nil {
		return domain.CreateDebriefResult{}, fmt.Errorf("marshal debrief raw questions: %w", err)
	}
	jobPayload, err := json.Marshal(map[string]any{
		"debriefId":     in.DebriefID,
		"targetJobId":   in.TargetJobID,
		"language":      in.Language,
		"questionCount": len(in.Questions),
	})
	if err != nil {
		return domain.CreateDebriefResult{}, fmt.Errorf("marshal debrief job payload: %w", err)
	}
	outboxPayload, err := json.Marshal(buildDebriefCreatedPayload(in))
	if err != nil {
		return domain.CreateDebriefResult{}, fmt.Errorf("marshal debrief.created payload: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.CreateDebriefResult{}, fmt.Errorf("begin create debrief: %w", err)
	}
	defer tx.Rollback()

	var debriefID string
	err = tx.QueryRowContext(ctx, `
insert into debriefs (
  id, user_id, target_job_id, status, round_type, interviewer_role,
  language, raw_questions, notes, created_at, updated_at
)
select $1, $2, tj.id, 'draft', $4, nullif($5, ''), $6, $7, nullif($8, ''), $9, $9
from target_jobs tj
where tj.id = $3
  and tj.user_id = $2
  and tj.deleted_at is null
returning id`,
		in.DebriefID,
		in.UserID,
		in.TargetJobID,
		string(in.RoundType),
		string(in.InterviewerRole),
		in.Language,
		rawQuestions,
		in.Notes,
		in.Now,
	).Scan(&debriefID)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.CreateDebriefResult{}, domain.ErrDebriefPrerequisite
	}
	if err != nil {
		return domain.CreateDebriefResult{}, fmt.Errorf("insert debrief draft: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status,
  payload, available_at, created_at, updated_at
) values ($1,$2,'debrief',$3,$4,$5,$6,$7,$7,$7)`,
		in.JobID,
		string(sharedjobs.JobTypeDebriefGenerate),
		in.DebriefID,
		in.DebriefID,
		string(sharedtypes.JobStatusQueued),
		jobPayload,
		in.Now,
	); err != nil {
		return domain.CreateDebriefResult{}, fmt.Errorf("insert debrief_generate async job: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload, publish_status, created_at
) values ($1,$2,1,'debrief',$3,$4,'pending',$5)`,
		in.OutboxEventID,
		string(sharedevents.EventNameDebriefCreated),
		in.DebriefID,
		outboxPayload,
		in.Now,
	); err != nil {
		return domain.CreateDebriefResult{}, fmt.Errorf("insert debrief.created outbox event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.CreateDebriefResult{}, fmt.Errorf("commit create debrief: %w", err)
	}
	return domain.CreateDebriefResult{
		DebriefID: debriefID,
		Job: domain.JobRecord{
			ID:           in.JobID,
			JobType:      api.JobTypeDebriefGenerate,
			ResourceType: api.ResourceTypeDebrief,
			ResourceID:   in.DebriefID,
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    in.Now.UTC(),
			UpdatedAt:    in.Now.UTC(),
		},
	}, nil
}

func (r *Repository) LoadGenerateContext(ctx context.Context, payload domain.GenerateJobPayload) (domain.GenerateContext, error) {
	if err := r.checkDB(); err != nil {
		return domain.GenerateContext{}, err
	}

	var out domain.GenerateContext
	var targetSummary []byte
	var rawQuestions []byte
	err := r.db.QueryRowContext(ctx, `
select d.user_id::text,
       d.id::text,
       d.target_job_id::text,
       d.language,
       coalesce(tj.title, '') as target_title,
       coalesce(tj.summary, '{}'::jsonb) as target_summary,
       d.raw_questions
from debriefs d
join target_jobs tj
  on tj.id = d.target_job_id
 and tj.deleted_at is null
where d.id = $1
  and d.target_job_id = $2
  and d.status = 'draft'`,
		payload.DebriefID,
		payload.TargetJobID,
	).Scan(
		&out.UserID,
		&out.DebriefID,
		&out.TargetJobID,
		&out.Language,
		&out.TargetTitle,
		&targetSummary,
		&rawQuestions,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.GenerateContext{}, domain.ErrDebriefNotFound
	}
	if err != nil {
		return domain.GenerateContext{}, fmt.Errorf("load debrief generate context: %w", err)
	}
	if out.Language == "" {
		out.Language = payload.Language
	}
	if len(targetSummary) == 0 {
		targetSummary = []byte(`{}`)
	}
	out.TargetSummary = string(targetSummary)
	if err := json.Unmarshal(rawQuestions, &out.Questions); err != nil {
		return domain.GenerateContext{}, fmt.Errorf("decode debrief raw questions: %w", err)
	}
	return out, nil
}

func (r *Repository) GetDebrief(ctx context.Context, userID, debriefID string) (domain.DebriefRecord, error) {
	if err := r.checkDB(); err != nil {
		return domain.DebriefRecord{}, err
	}

	var out domain.DebriefRecord
	var status string
	var roundType string
	var language string
	var rawQuestions []byte
	var riskItems []byte
	var nextRoundChecklist []byte
	var promptVersion string
	var rubricVersion string
	var modelID string
	err := r.db.QueryRowContext(ctx, `
select d.id::text,
       d.target_job_id::text,
       d.status,
       d.round_type,
       coalesce(d.interviewer_role, '') as interviewer_role,
       d.language,
       d.raw_questions,
       d.risk_items,
       d.next_round_checklist,
       coalesce(d.thank_you_draft, '') as thank_you_draft,
       coalesce(d.prompt_version, '') as prompt_version,
       coalesce(d.rubric_version, '') as rubric_version,
       coalesce(d.model_id, '') as model_id,
       d.created_at,
       d.updated_at
from debriefs d
where d.id = $1
  and d.user_id = $2`,
		debriefID,
		userID,
	).Scan(
		&out.ID,
		&out.TargetJobID,
		&status,
		&roundType,
		&out.InterviewerRole,
		&language,
		&rawQuestions,
		&riskItems,
		&nextRoundChecklist,
		&out.ThankYouDraft,
		&promptVersion,
		&rubricVersion,
		&modelID,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.DebriefRecord{}, domain.ErrDebriefNotFound
	}
	if err != nil {
		return domain.DebriefRecord{}, fmt.Errorf("get debrief: %w", err)
	}

	out.Status = sharedtypes.DebriefStatus(status)
	out.RoundType = sharedtypes.DebriefRoundType(roundType)
	if err := json.Unmarshal(rawQuestions, &out.Questions); err != nil {
		return domain.DebriefRecord{}, fmt.Errorf("decode debrief questions: %w", err)
	}
	if out.Status != sharedtypes.DebriefStatusCompleted {
		out.RiskItems = nil
		out.NextRoundChecklist = nil
		out.ThankYouDraft = ""
		out.Provenance = nil
		return out, nil
	}
	if err := json.Unmarshal(riskItems, &out.RiskItems); err != nil {
		return domain.DebriefRecord{}, fmt.Errorf("decode debrief risk items: %w", err)
	}
	if err := json.Unmarshal(nextRoundChecklist, &out.NextRoundChecklist); err != nil {
		return domain.DebriefRecord{}, fmt.Errorf("decode debrief next-round checklist: %w", err)
	}
	out.Provenance = &domain.Provenance{
		PromptVersion: promptVersion,
		RubricVersion: rubricVersion,
		ModelID:       modelID,
		Language:      language,
	}
	return out, nil
}

func (r *Repository) GetSuggestionContext(ctx context.Context, in domain.SuggestionContextRequest) (domain.SuggestionContext, error) {
	if err := r.checkDB(); err != nil {
		return domain.SuggestionContext{}, err
	}
	var out domain.SuggestionContext
	var summary []byte
	err := r.db.QueryRowContext(ctx, `
select tj.id::text,
       coalesce(tj.title, '') as title,
       coalesce(tj.company_name, '') as company_name,
       coalesce(tj.summary, '{}'::jsonb) as summary
from target_jobs tj
where tj.id = $1
  and tj.user_id = $2
  and tj.deleted_at is null`,
		in.TargetJobID,
		in.UserID,
	).Scan(&out.TargetJobID, &out.Title, &out.CompanyName, &summary)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.SuggestionContext{}, domain.ErrDebriefPrerequisite
	}
	if err != nil {
		return domain.SuggestionContext{}, fmt.Errorf("get debrief suggestion context: %w", err)
	}
	if len(summary) == 0 {
		summary = []byte(`{}`)
	}
	out.Summary = string(summary)
	sessionID := strings.TrimSpace(in.SessionID)
	if sessionID != "" {
		var sessionSummary []byte
		err := r.db.QueryRowContext(ctx, `
select jsonb_build_object(
  'sessionId', ps.id::text,
  'status', ps.status,
  'language', ps.language,
  'turnCount', ps.turn_count,
  'turns', coalesce((
    select jsonb_agg(jsonb_build_object(
      'turnIndex', pt.turn_index,
      'questionText', pt.question_text,
      'questionIntent', coalesce(pt.question_intent, ''),
      'answerSummary', coalesce(pt.answer_summary, '')
    ) order by pt.turn_index)
    from practice_turns pt
    where pt.session_id = ps.id
  ), '[]'::jsonb),
  'report', coalesce((
    select jsonb_build_object(
      'preparednessLevel', coalesce(fr.preparedness_level, ''),
      'highlights', fr.highlights,
      'issues', fr.issues,
      'nextActions', fr.next_actions
    )
    from feedback_reports fr
    where fr.session_id = ps.id
      and fr.user_id = ps.user_id
      and fr.status = 'ready'
    order by fr.created_at desc
    limit 1
  ), '{}'::jsonb)
)
from practice_sessions ps
where ps.id = $1
  and ps.user_id = $2
  and ps.target_job_id = $3
  and ps.status = 'completed'`,
			sessionID,
			in.UserID,
			in.TargetJobID,
		).Scan(&sessionSummary)
		if stderrs.Is(err, sql.ErrNoRows) {
			return domain.SuggestionContext{}, domain.ErrDebriefPrerequisite
		}
		if err != nil {
			return domain.SuggestionContext{}, fmt.Errorf("get debrief practice session suggestion context: %w", err)
		}
		if len(sessionSummary) == 0 {
			sessionSummary = []byte(`{}`)
		}
		out.SessionSummary = string(sessionSummary)
	}
	resumeID := strings.TrimSpace(in.ResumeID)
	if resumeID != "" {
		var resumeSummary []byte
		err := r.db.QueryRowContext(ctx, `
select coalesce(structured_profile, '{}'::jsonb)
from resumes
where id = $1
  and user_id = $2
  and deleted_at is null`,
			resumeID,
			in.UserID,
		).Scan(&resumeSummary)
		if stderrs.Is(err, sql.ErrNoRows) {
			return domain.SuggestionContext{}, domain.ErrDebriefPrerequisite
		}
		if err != nil {
			return domain.SuggestionContext{}, fmt.Errorf("get debrief resume suggestion context: %w", err)
		}
		if len(resumeSummary) == 0 {
			resumeSummary = []byte(`{}`)
		}
		out.ResumeSummary = string(resumeSummary)
	}
	return out, nil
}

func (r *Repository) UpdateDebriefCompleted(ctx context.Context, in domain.UpdateDebriefCompletedInput) (domain.DebriefRecord, error) {
	if err := r.checkDB(); err != nil {
		return domain.DebriefRecord{}, err
	}
	questions, err := json.Marshal(in.Questions)
	if err != nil {
		return domain.DebriefRecord{}, fmt.Errorf("marshal completed debrief questions: %w", err)
	}
	riskItems, err := json.Marshal(in.RiskItems)
	if err != nil {
		return domain.DebriefRecord{}, fmt.Errorf("marshal completed debrief risk items: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DebriefRecord{}, fmt.Errorf("begin complete debrief: %w", err)
	}
	defer tx.Rollback()

	var targetJobID string
	err = tx.QueryRowContext(ctx, `
update debriefs
set status = 'completed',
    raw_questions = $1,
    risk_items = $2,
    prompt_version = $3,
    rubric_version = $4,
    model_id = $5,
    updated_at = $6
where id = $7
  and user_id = $8
  and status = 'draft'
returning target_job_id::text`,
		questions,
		riskItems,
		in.Provenance.PromptVersion,
		in.Provenance.RubricVersion,
		in.Provenance.ModelID,
		in.Now,
		in.DebriefID,
		in.UserID,
	).Scan(&targetJobID)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.DebriefRecord{}, domain.ErrDebriefIllegalState
	}
	if err != nil {
		return domain.DebriefRecord{}, fmt.Errorf("complete debrief: %w", err)
	}

	outboxPayload, err := json.Marshal(buildDebriefCompletedPayload(in, targetJobID))
	if err != nil {
		return domain.DebriefRecord{}, fmt.Errorf("marshal debrief.completed payload: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload, publish_status, created_at
) values ($1,$2,1,'debrief',$3,$4,'pending',$5)`,
		in.OutboxEventID,
		string(sharedevents.EventNameDebriefCompleted),
		in.DebriefID,
		outboxPayload,
		in.Now,
	); err != nil {
		return domain.DebriefRecord{}, fmt.Errorf("insert debrief.completed outbox event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.DebriefRecord{}, fmt.Errorf("commit complete debrief: %w", err)
	}

	provenance := in.Provenance
	return domain.DebriefRecord{
		ID:          in.DebriefID,
		TargetJobID: targetJobID,
		Status:      sharedtypes.DebriefStatusCompleted,
		Questions:   append([]domain.QuestionRecord(nil), in.Questions...),
		RiskItems:   append([]domain.RiskItem(nil), in.RiskItems...),
		Provenance:  &provenance,
		UpdatedAt:   in.Now.UTC(),
	}, nil
}

func (r *Repository) RecordDebriefAuditEvent(ctx context.Context, event domain.DebriefAuditEvent) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("marshal debrief audit metadata: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `
insert into audit_events (
  id, user_id, actor_type, actor_id, action, resource_type,
  resource_id, result, metadata, created_at
) values ($1,$2,'user',$3,$4,$5,$6,$7,$8,$9)`,
		event.AuditEventID,
		event.UserID,
		event.UserID,
		event.Action,
		event.ResourceType,
		event.ResourceID,
		event.Result,
		metadata,
		event.CreatedAt,
	); err != nil {
		return fmt.Errorf("insert debrief audit event: %w", err)
	}
	return nil
}

type rawQuestionPayload struct {
	QuestionText        string `json:"questionText"`
	MyAnswerSummary     string `json:"myAnswerSummary"`
	InterviewerReaction string `json:"interviewerReaction,omitempty"`
}

func rawQuestionPayloads(in []domain.QuestionInput) []rawQuestionPayload {
	out := make([]rawQuestionPayload, 0, len(in))
	for _, q := range in {
		out = append(out, rawQuestionPayload{
			QuestionText:        q.QuestionText,
			MyAnswerSummary:     q.MyAnswerSummary,
			InterviewerReaction: q.InterviewerReaction,
		})
	}
	return out
}

func buildDebriefCreatedPayload(in domain.CreateDebriefStoreInput) sharedevents.DebriefCreatedPayload {
	return sharedevents.DebriefCreatedPayload{
		DebriefID:     in.DebriefID,
		TargetJobID:   in.TargetJobID,
		RoundType:     in.RoundType,
		QuestionCount: len(in.Questions),
	}
}

func buildDebriefCompletedPayload(in domain.UpdateDebriefCompletedInput, targetJobID string) sharedevents.DebriefCompletedPayload {
	return sharedevents.DebriefCompletedPayload{
		DebriefID:          in.DebriefID,
		TargetJobID:        targetJobID,
		RiskItemCount:      len(in.RiskItems),
		PracticeFocusCount: len(in.RiskItems),
	}
}
