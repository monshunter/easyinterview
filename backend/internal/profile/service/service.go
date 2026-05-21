// Package service implements the profile internal API surface consumed by
// privacy delete runners and cross-owner aggregation (backend-jobs-recommendations).
// All three services are deliberately HTTP-free: they accept domain types and
// return domain types, with auditing surfaces wired by the caller.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/profile"
)

// Options bundles dependencies common to all profile internal services.
type Options struct {
	Store profile.Store
	Audit profile.AuditTombstoneWriter
	Now   func() time.Time
}

// Service hosts privacy delete + source counts + cross-owner profile read.
type Service struct {
	store profile.Store
	audit profile.AuditTombstoneWriter
	now   func() time.Time
}

type transactionalPrivacyDeleteStore interface {
	DeleteCandidateProfileForUserWithAudit(ctx context.Context, userID string, jobID string, deletedAt time.Time) error
}

// New constructs a Service. Required: Store. Audit defaults to a no-op when
// nil. Now defaults to time.Now().UTC().
func New(opts Options) *Service {
	s := &Service{store: opts.Store, audit: opts.Audit}
	if opts.Now != nil {
		s.now = opts.Now
	} else {
		s.now = func() time.Time { return time.Now().UTC() }
	}
	if s.audit == nil {
		s.audit = noopTombstoneWriter{}
	}
	return s
}

// DeleteCandidateProfileForUser hard-deletes the user's candidate_profiles
// row and all related experience_cards, in that order (spec D-9). Writes an
// audit tombstone with only userId + experienceCardCount + deletedAt (PII
// redline) when the delete chain succeeds.
func (s *Service) DeleteCandidateProfileForUser(ctx context.Context, userID, jobID string) error {
	if s == nil || s.store == nil {
		return fmt.Errorf("profile service is not configured")
	}
	if txStore, ok := s.store.(transactionalPrivacyDeleteStore); ok {
		return txStore.DeleteCandidateProfileForUserWithAudit(ctx, userID, jobID, s.now().UTC())
	}
	counts, err := s.store.CountExperienceCardsBySource(ctx, userID)
	if err != nil {
		return fmt.Errorf("count experience cards: %w", err)
	}
	var cardCount int64
	for _, n := range counts {
		cardCount += n
	}
	if _, err := s.store.DeleteExperienceCardsForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete experience cards: %w", err)
	}
	if _, err := s.store.DeleteCandidateProfileForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete candidate profile: %w", err)
	}
	tombstone := profile.CandidateProfileDeleteTombstone{
		UserID:              userID,
		ExperienceCardCount: cardCount,
		DeletedAt:           s.now().UTC(),
		JobID:               jobID,
	}
	if err := s.audit.WriteCandidateProfileDeleteTombstone(ctx, tombstone); err != nil {
		return fmt.Errorf("write privacy tombstone: %w", err)
	}
	return nil
}

// CountExperienceCardsBySource proxies to the store (spec D-11). Returned map
// always includes every value in profile.SourceTypes with default 0.
func (s *Service) CountExperienceCardsBySource(ctx context.Context, userID string) (profile.SourceCounts, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("profile service is not configured")
	}
	return s.store.CountExperienceCardsBySource(ctx, userID)
}

// GetCandidateProfileForUser returns the stored CandidateProfile shape for
// cross-owner aggregation (spec D-13). Read-only: missing rows return
// (nil, nil) without seeding; no audit_events / profile_version side effects.
func (s *Service) GetCandidateProfileForUser(ctx context.Context, userID string) (*api.CandidateProfile, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("profile service is not configured")
	}
	rec, err := s.store.GetCandidateProfileByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, profile.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("read candidate profile: %w", err)
	}
	if rec == nil {
		return nil, nil
	}
	dto := mapCandidateProfile(rec)
	return &dto, nil
}

func mapCandidateProfile(rec *profile.CandidateProfileRecord) api.CandidateProfile {
	if rec == nil {
		return api.CandidateProfile{}
	}
	return api.CandidateProfile{
		Headline:                  copyStringPtr(rec.Headline),
		YearsOfExperience:         copyInt32Ptr(rec.YearsOfExperience),
		CurrentRole:               copyStringPtr(rec.CurrentRole),
		PreferredPracticeLanguage: rec.PreferredPracticeLanguage,
		UiLanguage:                rec.UiLanguage,
		Region:                    copyStringPtr(rec.Region),
	}
}

func copyStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := *s
	return &v
}

func copyInt32Ptr(v *int32) *int32 {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}

type noopTombstoneWriter struct{}

func (noopTombstoneWriter) WriteCandidateProfileDeleteTombstone(context.Context, profile.CandidateProfileDeleteTombstone) error {
	return nil
}
