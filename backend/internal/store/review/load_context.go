package review

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
)

func (r *Repository) LoadReportContext(ctx context.Context, reportID string) (reviewdomain.ReportContext, error) {
	if err := r.checkDB(); err != nil {
		return reviewdomain.ReportContext{}, err
	}
	var (
		out            reviewdomain.ReportContext
		rowReportID    string
		rowSessionID   string
		rowTargetJobID string
		generationRaw  []byte
	)
	err := r.db.QueryRowContext(ctx, `
	select fr.user_id::text, fr.id::text, fr.session_id::text, fr.target_job_id::text, fr.generation_context
	from feedback_reports fr
	where fr.id = $1`, reportID).Scan(
		&out.Session.UserID,
		&rowReportID,
		&rowSessionID,
		&rowTargetJobID,
		&generationRaw,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return reviewdomain.ReportContext{}, fmt.Errorf("%w: report %s not found", reviewdomain.ErrReportContextInvalid, reportID)
	}
	if err != nil {
		return reviewdomain.ReportContext{}, fmt.Errorf("load report context: %w", err)
	}
	out.Session.ReportID = rowReportID
	out.Session.SessionID = rowSessionID
	out.Session.TargetJobID = rowTargetJobID
	frozen, err := decodeFrozenReportContext(generationRaw)
	if err != nil {
		return out, fmt.Errorf("%w: decode frozen report context: %v", reviewdomain.ErrReportContextInvalid, err)
	}
	if rowReportID != reportID || rowSessionID != frozen.Conversation.SessionID || rowTargetJobID != frozen.TargetJob.ID {
		return out, fmt.Errorf("%w: frozen report context row identity mismatch", reviewdomain.ErrReportContextInvalid)
	}
	out.FrozenContext = frozen
	out.Session.SessionID = frozen.Conversation.SessionID
	out.Session.TargetJobID = frozen.TargetJob.ID
	out.Session.Language = frozen.Conversation.Language
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
	if err := validateFrozenConversation(frozen, out.Messages); err != nil {
		return out, fmt.Errorf("%w: %v", reviewdomain.ErrReportContextInvalid, err)
	}
	return out, nil
}

func decodeFrozenReportContext(raw []byte) (practicedomain.ReportContextSnapshot, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return practicedomain.ReportContextSnapshot{}, fmt.Errorf("generation_context is empty")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var snapshot practicedomain.ReportContextSnapshot
	if err := decoder.Decode(&snapshot); err != nil {
		return practicedomain.ReportContextSnapshot{}, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return practicedomain.ReportContextSnapshot{}, fmt.Errorf("generation_context has trailing JSON")
		}
		return practicedomain.ReportContextSnapshot{}, err
	}
	if err := practicedomain.ValidateReportContextSnapshot(snapshot); err != nil {
		return practicedomain.ReportContextSnapshot{}, err
	}
	return snapshot, nil
}

func validateFrozenConversation(snapshot practicedomain.ReportContextSnapshot, messages []reviewdomain.MessageSnapshot) error {
	coordinate := snapshot.Conversation
	if len(messages) != int(coordinate.MessageCount) {
		return fmt.Errorf("frozen report context message count mismatch")
	}
	if len(messages) < 3 {
		return fmt.Errorf("frozen report context has no reportable candidate answer")
	}
	for index, message := range messages {
		expectedSeq := index + 1
		if message.SeqNo != expectedSeq {
			return fmt.Errorf("frozen report context message ordering mismatch")
		}
		expectedRole := "assistant"
		if expectedSeq%2 == 0 {
			expectedRole = "user"
		}
		if message.Role != expectedRole || strings.TrimSpace(message.Content) == "" {
			return fmt.Errorf("frozen report context message role/content mismatch")
		}
	}
	if messages[len(messages)-1].SeqNo != int(coordinate.LastMessageSeqNo) || messages[len(messages)-1].Role != "assistant" {
		return fmt.Errorf("frozen report context last message mismatch")
	}
	return nil
}
