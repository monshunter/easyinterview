package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
)

func loadCompletionReportContext(ctx context.Context, tx *sql.Tx, userID, sessionID string) (domain.ReportContextSnapshot, error) {
	var (
		input         domain.ReportContextSnapshotInput
		focusCodes    pq.StringArray
		company       sql.NullString
		seniority     sql.NullString
		rawJD         sql.NullString
		targetSummary []byte
		structured    []byte
	)
	err := tx.QueryRowContext(ctx, `
select pp.id, pp.goal, pp.interviewer_persona, pp.difficulty, pp.language,
       pp.time_budget_minutes, pp.resume_id::text, pp.focus_dimension_codes,
       pp.round_id, pp.round_sequence,
       tj.id, coalesce(tj.title,''), tj.company_name, tj.seniority_level,
       tj.target_language, tj.raw_jd_text, tj.summary,
       r.id, coalesce(nullif(btrim(r.display_name),''), r.title), r.language,
       coalesce(nullif(btrim(r.parsed_text_snapshot),''), nullif(btrim(r.original_text),''), nullif(btrim(r.raw_text),''), ''),
       r.structured_profile, ps.language
from practice_sessions ps
join practice_plans pp
  on pp.id=ps.plan_id and pp.user_id=ps.user_id and pp.target_job_id=ps.target_job_id
join target_jobs tj
  on tj.id=ps.target_job_id and tj.user_id=ps.user_id and tj.resume_id=pp.resume_id and tj.deleted_at is null
join resumes r
  on r.id=pp.resume_id and r.user_id=ps.user_id and r.deleted_at is null
where ps.user_id=$1 and ps.id=$2
  and pp.round_id is not null and pp.round_sequence is not null
for update of pp, tj, r`, userID, sessionID).Scan(
		&input.Plan.ID,
		&input.Plan.Goal,
		&input.Plan.InterviewerPersona,
		&input.Plan.Difficulty,
		&input.Plan.Language,
		&input.Plan.TimeBudgetMinutes,
		&input.Plan.ResumeID,
		&focusCodes,
		&input.Plan.RoundID,
		&input.Plan.RoundSequence,
		&input.TargetJob.ID,
		&input.TargetJob.Title,
		&company,
		&seniority,
		&input.TargetJob.Language,
		&rawJD,
		&targetSummary,
		&input.Resume.ID,
		&input.Resume.DisplayName,
		&input.Resume.Language,
		&input.Resume.SourceSnapshot,
		&structured,
		&input.Conversation.Language,
	)
	if err == sql.ErrNoRows {
		return domain.ReportContextSnapshot{}, domain.ErrSessionConflict
	}
	if err != nil {
		return domain.ReportContextSnapshot{}, fmt.Errorf("load completion report context: %w", err)
	}
	input.TargetJob.Company = company.String
	input.TargetJob.Seniority = seniority.String
	input.TargetJob.RawJD = rawJD.String
	input.TargetJob.Summary = append(json.RawMessage(nil), targetSummary...)
	input.Resume.StructuredProfile = append(json.RawMessage(nil), structured...)
	input.Plan.FocusDimensionCodes = append([]string(nil), focusCodes...)
	input.Conversation.SessionID = sessionID

	rows, err := tx.QueryContext(ctx, `
select kind, label, coalesce(description,''), evidence_level, display_order
from target_job_requirements
where target_job_id=$1
order by display_order asc, created_at asc, id asc`, input.TargetJob.ID)
	if err != nil {
		return domain.ReportContextSnapshot{}, fmt.Errorf("load completion target requirements: %w", err)
	}
	for rows.Next() {
		var requirement domain.ReportRequirementSnapshot
		if err := rows.Scan(&requirement.Kind, &requirement.Label, &requirement.Description, &requirement.EvidenceLevel, &requirement.DisplayOrder); err != nil {
			rows.Close()
			return domain.ReportContextSnapshot{}, fmt.Errorf("scan completion target requirement: %w", err)
		}
		input.TargetJob.Requirements = append(input.TargetJob.Requirements, requirement)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return domain.ReportContextSnapshot{}, fmt.Errorf("iterate completion target requirements: %w", err)
	}
	rows.Close()

	if err := tx.QueryRowContext(ctx, `
select count(*), coalesce(max(seq_no),0)
from practice_messages
where session_id=$1`, sessionID).Scan(&input.Conversation.MessageCount, &input.Conversation.LastMessageSeqNo); err != nil {
		return domain.ReportContextSnapshot{}, fmt.Errorf("load completion message coordinate: %w", err)
	}
	snapshot, err := domain.BuildReportContextSnapshot(input)
	if err != nil {
		return domain.ReportContextSnapshot{}, fmt.Errorf("build completion report context: %w", err)
	}
	return snapshot, nil
}
