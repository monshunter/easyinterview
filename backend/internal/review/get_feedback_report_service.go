package review

import (
	"context"
	"fmt"
	"strings"
)

type reportReadRepository interface {
	GetFeedbackReport(ctx context.Context, userID, reportID string) (FeedbackReportRecord, error)
	ListTargetJobReports(ctx context.Context, userID, targetJobID string) (TargetJobReportsOverviewRecord, error)
}

func (s *Service) GetFeedbackReport(ctx context.Context, userID, reportID string) (FeedbackReportRecord, error) {
	userID = strings.TrimSpace(userID)
	reportID = strings.TrimSpace(reportID)
	if userID == "" || reportID == "" {
		return FeedbackReportRecord{}, ErrReportNotFound
	}
	reader, err := s.reportReader()
	if err != nil {
		return FeedbackReportRecord{}, err
	}
	return reader.GetFeedbackReport(ctx, userID, reportID)
}

func (s *Service) reportReader() (reportReadRepository, error) {
	if s == nil || s.repository == nil {
		return nil, fmt.Errorf("review repository is not configured")
	}
	reader, ok := s.repository.(reportReadRepository)
	if !ok {
		return nil, fmt.Errorf("review repository does not implement report reads")
	}
	return reader, nil
}
