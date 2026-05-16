package review

import (
	"context"
	"strings"
)

func (s *Service) ListTargetJobReports(ctx context.Context, in ListTargetJobReportsRequest) (PaginatedFeedbackReportRecord, error) {
	userID := strings.TrimSpace(in.UserID)
	targetJobID := strings.TrimSpace(in.TargetJobID)
	if userID == "" || targetJobID == "" {
		return PaginatedFeedbackReportRecord{}, ErrReportNotFound
	}
	pageSize := EffectiveReportPageSize(in.PageSize)
	listInput := ListTargetJobReportsInput{
		UserID:      userID,
		TargetJobID: targetJobID,
		Cursor:      strings.TrimSpace(in.Cursor),
		PageSize:    pageSize,
	}
	if listInput.Cursor != "" {
		createdAt, id, err := DecodeCursor(listInput.Cursor)
		if err != nil {
			return PaginatedFeedbackReportRecord{}, ErrInvalidCursor
		}
		listInput.CursorCreatedAt = createdAt
		listInput.CursorID = id
	}
	reader, err := s.reportReader()
	if err != nil {
		return PaginatedFeedbackReportRecord{}, err
	}
	res, err := reader.ListTargetJobReports(ctx, listInput)
	if err != nil {
		return PaginatedFeedbackReportRecord{}, err
	}
	return PaginatedFeedbackReportRecord{
		Items: res.Items,
		PageInfo: PageInfo{
			NextCursor: res.NextCursor,
			PageSize:   pageSize,
			HasMore:    res.HasMore,
		},
	}, nil
}
