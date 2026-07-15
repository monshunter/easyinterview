package review

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func (r *Repository) GetReportConversation(ctx context.Context, userID, reportID string) (reviewdomain.ReportConversationRecord, error) {
	if err := r.checkDB(); err != nil {
		return reviewdomain.ReportConversationRecord{}, err
	}
	var (
		conversation            reviewdomain.ReportConversationRecord
		conversationSessionID   string
		conversationTargetJobID string
		generationRaw           []byte
		boundSessionID          string
		sessionUserID           string
		sessionTargetJob        string
		status                  string
	)
	err := r.db.QueryRowContext(ctx, `
select fr.id::text, fr.session_id::text, fr.target_job_id::text, fr.status, fr.generation_context,
       ps.id::text, ps.user_id::text, ps.target_job_id::text
from feedback_reports fr
join practice_sessions ps on ps.id = fr.session_id
where fr.id = $1 and fr.user_id = $2`, reportID, userID).Scan(
		&conversation.ReportID,
		&conversationSessionID,
		&conversationTargetJobID,
		&status,
		&generationRaw,
		&boundSessionID,
		&sessionUserID,
		&sessionTargetJob,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return reviewdomain.ReportConversationRecord{}, reviewdomain.ErrReportNotFound
	}
	if err != nil {
		return reviewdomain.ReportConversationRecord{}, fmt.Errorf("load report conversation: %w", err)
	}
	conversation.Status = sharedtypes.ReportStatus(status)
	if err := validateReportConversationIdentity(conversation.ReportID, reportID, conversationSessionID, conversationTargetJobID, boundSessionID, sessionUserID, sessionTargetJob, userID); err != nil {
		return reviewdomain.ReportConversationRecord{}, err
	}
	if !validReportConversationStatus(conversation.Status) {
		return reviewdomain.ReportConversationRecord{}, fmt.Errorf("%w: report status", reviewdomain.ErrReportConversationInvalid)
	}
	frozen, err := decodeFrozenReportContext(generationRaw)
	if err != nil {
		return reviewdomain.ReportConversationRecord{}, fmt.Errorf("%w: frozen context", reviewdomain.ErrReportConversationInvalid)
	}
	if frozen.Conversation.SessionID != conversationSessionID || frozen.TargetJob.ID != conversationTargetJobID {
		return reviewdomain.ReportConversationRecord{}, fmt.Errorf("%w: frozen context binding", reviewdomain.ErrReportConversationInvalid)
	}
	conversation.Context = reviewdomain.ProjectFrozenReportContext(frozen)
	conversation.Messages = []reviewdomain.ReportConversationMessageRecord{}

	rows, err := r.db.QueryContext(ctx, `
select seq_no, role, content, created_at
from practice_messages
where session_id = $1
order by seq_no asc`, conversationSessionID)
	if err != nil {
		return reviewdomain.ReportConversationRecord{}, fmt.Errorf("load report conversation messages: %w", err)
	}
	defer rows.Close()
	var previous int32
	for rows.Next() {
		var message reviewdomain.ReportConversationMessageRecord
		if err := rows.Scan(&message.Sequence, &message.Role, &message.Content, &message.CreatedAt); err != nil {
			return reviewdomain.ReportConversationRecord{}, fmt.Errorf("scan report conversation message: %w", err)
		}
		if message.Sequence <= previous || message.Sequence < 1 || !validReportConversationRole(message.Role) || strings.TrimSpace(message.Content) == "" || message.CreatedAt.IsZero() {
			return reviewdomain.ReportConversationRecord{}, fmt.Errorf("%w: message projection", reviewdomain.ErrReportConversationInvalid)
		}
		previous = message.Sequence
		conversation.Messages = append(conversation.Messages, message)
	}
	if err := rows.Err(); err != nil {
		return reviewdomain.ReportConversationRecord{}, fmt.Errorf("iterate report conversation messages: %w", err)
	}
	return conversation, nil
}

func validateReportConversationIdentity(reportID, requestedReportID, reportSessionID, reportTargetJobID, boundSessionID, sessionUserID, sessionTargetJobID, userID string) error {
	if strings.TrimSpace(reportID) == "" || reportID != requestedReportID ||
		strings.TrimSpace(reportSessionID) == "" || strings.TrimSpace(reportTargetJobID) == "" ||
		reportSessionID != boundSessionID || sessionUserID != userID || sessionTargetJobID != reportTargetJobID {
		return fmt.Errorf("%w: report/session identity", reviewdomain.ErrReportConversationInvalid)
	}
	return nil
}

func validReportConversationStatus(status sharedtypes.ReportStatus) bool {
	for _, allowed := range sharedtypes.AllReportStatuses {
		if status == allowed {
			return true
		}
	}
	return false
}

func validReportConversationRole(role string) bool {
	return role == "user" || role == "assistant"
}
