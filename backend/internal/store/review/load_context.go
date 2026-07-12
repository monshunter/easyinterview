package review

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
)

func (r *Repository) LoadReportContext(ctx context.Context, reportID string) (reviewdomain.ReportContext, error) {
	if err := r.checkDB(); err != nil {
		return reviewdomain.ReportContext{}, err
	}
	var out reviewdomain.ReportContext
	err := r.db.QueryRowContext(ctx, `
select fr.user_id, fr.id, fr.session_id, ps.plan_id, fr.target_job_id, ps.language,
       pp.id, pp.goal, pp.interviewer_persona
from feedback_reports fr
join practice_sessions ps on ps.id = fr.session_id
join practice_plans pp on pp.id = ps.plan_id
where fr.id = $1`, reportID).Scan(
		&out.Session.UserID,
		&out.Session.ReportID,
		&out.Session.SessionID,
		&out.Session.PlanID,
		&out.Session.TargetJobID,
		&out.Session.Language,
		&out.Plan.ID,
		&out.Plan.Goal,
		&out.Plan.InterviewerPersona,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return reviewdomain.ReportContext{}, fmt.Errorf("review report context not found: %s", reportID)
	}
	if err != nil {
		return reviewdomain.ReportContext{}, fmt.Errorf("load report context: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `
select role, content, seq_no
from practice_messages
where session_id = $1
order by seq_no asc`, out.Session.SessionID)
	if err != nil {
		return reviewdomain.ReportContext{}, fmt.Errorf("load report turns: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var message reviewdomain.MessageSnapshot
		if err := rows.Scan(&message.Role, &message.Content, &message.SeqNo); err != nil {
			return reviewdomain.ReportContext{}, fmt.Errorf("scan report message: %w", err)
		}
		out.Messages = append(out.Messages, message)
	}
	if err := rows.Err(); err != nil {
		return reviewdomain.ReportContext{}, fmt.Errorf("iterate report messages: %w", err)
	}
	out.Rubric = registry.RubricSchema{Dimensions: []registry.RubricDimension{{Name: "overall", Weight: 1}}}
	out.ReportPromptVersion = "v0.1.0"
	out.ReportRubricVersion = "v0.1.0"
	out.ModelID = "model-profile:report.generate.default"
	out.FeatureFlag = "none"
	out.DataSourceVersion = "registry.v1"
	return out, nil
}
