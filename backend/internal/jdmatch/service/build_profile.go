// Package service hosts the JD-Match orchestration layer. Handlers
// consume Service-typed structs; the Service composes the cross-owner
// internal APIs (backend-auth identity / backend-profile candidate
// profile + experience cards counts / 4 counter packages) and the
// JD-Match store layer.
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/profile"
)

// ProfileDeps wires the 7 cross-owner internal APIs the JD-Match
// orchestrator consumes per spec §4.4 / D-17 / D-18. Each function is
// pluggable so handler tests can inject deterministic stubs.
type ProfileDeps struct {
	GetUserIdentity              func(ctx context.Context, userID string) (auth.UserIdentity, error)
	GetCandidateProfile          func(ctx context.Context, userID string) (*api.CandidateProfile, error)
	CountExperienceCardsBySource func(ctx context.Context, userID string) (profile.SourceCounts, error)
	CountResumes                 func(ctx context.Context, userID string) (int, error)
	CountTargetJobs              func(ctx context.Context, userID string) (int, error)
	CountPracticeSessions        func(ctx context.Context, userID string) (int, error)
	CountDebriefs                func(ctx context.Context, userID string) (int, error)
}

// JobMatchProfileResult bundles the orchestrator output and a trace of
// which cross-owner calls succeeded vs. fell back to a safe default. The
// trace lets the HTTP layer attach observability metadata without
// leaking PII (raw email, candidate_profile text, etc.) into the
// response envelope.
type JobMatchProfileResult struct {
	Profile     api.JobMatchProfile
	CallTrace   ProfileCallTrace
	ExperienceCardCount int64
}

// ProfileCallTrace records pass/fail state for each cross-owner call.
// CountExperienceCardsBySource is included for the P1 enrichment anchor
// but does NOT enter `sources` per spec D-2.
type ProfileCallTrace struct {
	IdentityOK         bool
	CandidateProfileOK bool
	ExperienceCardsOK  bool
	ResumesOK          bool
	TargetJobsOK       bool
	PracticeOK         bool
	DebriefsOK         bool
}

// BuildJobMatchProfile orchestrates the 7 cross-owner internal APIs and
// projects their results into the generated `JobMatchProfile` DTO.
// Spec D-17 / D-18 boundaries are enforced here:
//
//   - displayName is mandatory, non-null. Identity failure falls back
//     to backend-auth's AnonymousDisplayName (`Candidate`) — never raw
//     email or empty string.
//   - avatarUrl / locationText / compensationText are nil at P0
//     baseline (returned as null over JSON). skills returns an empty
//     slice ([]).
//   - headline / yearsOfExperience are projected from the candidate
//     profile when present; missing profile → both nil.
//   - sources is populated from the four counter packages. A counter
//     failure logs a redacted warning and falls back to 0 for that
//     facet, never aborting the whole call.
//   - The optional CountExperienceCardsBySource sum is exposed via
//     ExperienceCardCount for trace / P1 enrichment but never lands in
//     the sources object (D-2).
//
// userID must be the session-resolved current user; cross-user
// isolation is the caller's responsibility (handler wires session
// middleware then passes the resolved userID).
func BuildJobMatchProfile(ctx context.Context, userID string, deps ProfileDeps) (JobMatchProfileResult, error) {
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return JobMatchProfileResult{}, fmt.Errorf("jdmatch: BuildJobMatchProfile requires a non-empty userID")
	}
	out := JobMatchProfileResult{}

	// Identity (D-17). Failure → anonymous fallback.
	displayName := auth.AnonymousDisplayName
	if deps.GetUserIdentity != nil {
		ident, err := deps.GetUserIdentity(ctx, uid)
		if err != nil {
			log.Printf("jdmatch.buildProfile identity fallback for userID=%s reason=%s", redactID(uid), redactErr(err))
		} else {
			out.CallTrace.IdentityOK = true
			if ident.DisplayName != "" {
				displayName = ident.DisplayName
			}
		}
	}
	out.Profile.DisplayName = displayName
	// avatarUrl baseline null per D-18; pointer left nil.

	// Candidate profile (read-only, no seed side effect per backend-
	// profile D-13). Failure → null headline / yearsOfExperience.
	if deps.GetCandidateProfile != nil {
		cp, err := deps.GetCandidateProfile(ctx, uid)
		if err != nil {
			log.Printf("jdmatch.buildProfile candidate-profile fallback for userID=%s reason=%s", redactID(uid), redactErr(err))
		} else {
			out.CallTrace.CandidateProfileOK = true
			if cp != nil {
				if cp.Headline != nil && *cp.Headline != "" {
					h := *cp.Headline
					out.Profile.Headline = &h
				}
				if cp.YearsOfExperience != nil {
					y := *cp.YearsOfExperience
					out.Profile.YearsOfExperience = &y
				}
			}
		}
	}

	// Experience card count (D-2 trace-only; never enters sources).
	if deps.CountExperienceCardsBySource != nil {
		counts, err := deps.CountExperienceCardsBySource(ctx, uid)
		if err != nil {
			log.Printf("jdmatch.buildProfile experience-cards fallback for userID=%s reason=%s", redactID(uid), redactErr(err))
		} else {
			out.CallTrace.ExperienceCardsOK = true
			var total int64
			for _, n := range counts {
				total += n
			}
			out.ExperienceCardCount = total
		}
	}

	// Skills baseline (D-18: P0 returns []).
	out.Profile.Skills = []string{}

	// Sources counts (D-18 sources object; each missing counter logs
	// a redacted warn and falls back to 0).
	out.Profile.Sources = api.JobMatchProfileSourceCounts{}
	out.Profile.Sources.Resumes = int32(callCounter(ctx, uid, "resumes", deps.CountResumes, &out.CallTrace.ResumesOK))
	out.Profile.Sources.Jds = int32(callCounter(ctx, uid, "jds", deps.CountTargetJobs, &out.CallTrace.TargetJobsOK))
	out.Profile.Sources.Mocks = int32(callCounter(ctx, uid, "mocks", deps.CountPracticeSessions, &out.CallTrace.PracticeOK))
	out.Profile.Sources.Debriefs = int32(callCounter(ctx, uid, "debriefs", deps.CountDebriefs, &out.CallTrace.DebriefsOK))

	return out, nil
}

func callCounter(ctx context.Context, userID, label string, fn func(context.Context, string) (int, error), okOut *bool) int {
	if fn == nil {
		return 0
	}
	n, err := fn(ctx, userID)
	if err != nil {
		log.Printf("jdmatch.buildProfile counter fallback for userID=%s facet=%s reason=%s", redactID(userID), label, redactErr(err))
		return 0
	}
	*okOut = true
	if n < 0 {
		return 0
	}
	return n
}

// redactID truncates a userID to its first 6 chars + ... so warn logs
// can still correlate without spilling the full ID into long-term log
// retention. Empty / short IDs are passed through unchanged.
func redactID(id string) string {
	if len(id) <= 6 {
		return id
	}
	return id[:6] + "..."
}

// redactErr trims the error to a short shape that never carries
// candidate_profile text, raw email, or other PII surfaces. Underlying
// errors are mapped to error class names where possible.
func redactErr(err error) string {
	switch {
	case errors.Is(err, auth.ErrUserNotFound):
		return "auth.ErrUserNotFound"
	}
	return fmt.Sprintf("%T", err)
}
