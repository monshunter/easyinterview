package review

import (
	"context"
	"strings"
)

func (s *Service) ListTargetJobReports(ctx context.Context, in ListTargetJobReportsRequest) (TargetJobReportsOverviewRecord, error) {
	userID := strings.TrimSpace(in.UserID)
	targetJobID := strings.TrimSpace(in.TargetJobID)
	if userID == "" || targetJobID == "" {
		return TargetJobReportsOverviewRecord{}, ErrReportNotFound
	}
	reader, err := s.reportReader()
	if err != nil {
		return TargetJobReportsOverviewRecord{}, err
	}
	res, err := reader.ListTargetJobReports(ctx, userID, targetJobID)
	if err != nil {
		return TargetJobReportsOverviewRecord{}, err
	}
	return res, nil
}
