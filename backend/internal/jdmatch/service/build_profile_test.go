package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/service"
	"github.com/monshunter/easyinterview/backend/internal/profile"
)

// callCounter records how many times each cross-owner dependency was
// invoked, so the test can prove the 7-cross-owner spec D-17 / D-18
// contract: each dep is called exactly once per BuildJobMatchProfile.
type callCounter struct {
	mu sync.Mutex
	n  map[string]int
}

func (c *callCounter) inc(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.n == nil {
		c.n = map[string]int{}
	}
	c.n[name]++
}

func (c *callCounter) get(name string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n[name]
}

func happyPathDeps(t *testing.T, c *callCounter) service.ProfileDeps {
	t.Helper()
	headline := "Senior frontend engineer"
	years := int32(6)
	return service.ProfileDeps{
		GetUserIdentity: func(ctx context.Context, userID string) (auth.UserIdentity, error) {
			c.inc("identity")
			return auth.UserIdentity{DisplayName: "Alice Example", EmailMasked: "a***e@example.com"}, nil
		},
		GetCandidateProfile: func(ctx context.Context, userID string) (*api.CandidateProfile, error) {
			c.inc("candidate-profile")
			return &api.CandidateProfile{
				Headline:                  &headline,
				YearsOfExperience:         &years,
				PreferredPracticeLanguage: "en",
				UiLanguage:                "en",
			}, nil
		},
		CountExperienceCardsBySource: func(ctx context.Context, userID string) (profile.SourceCounts, error) {
			c.inc("experience-cards")
			return profile.SourceCounts{"manual": 2, "resume_parse": 1}, nil
		},
		CountResumes:          func(ctx context.Context, userID string) (int, error) { c.inc("resumes"); return 3, nil },
		CountTargetJobs:       func(ctx context.Context, userID string) (int, error) { c.inc("jds"); return 5, nil },
		CountPracticeSessions: func(ctx context.Context, userID string) (int, error) { c.inc("mocks"); return 8, nil },
		CountDebriefs:         func(ctx context.Context, userID string) (int, error) { c.inc("debriefs"); return 2, nil },
	}
}

func TestBuildJobMatchProfileAggregationHappyPath(t *testing.T) {
	cnt := &callCounter{}
	deps := happyPathDeps(t, cnt)
	out, err := service.BuildJobMatchProfile(context.Background(), "user-A", deps)
	if err != nil {
		t.Fatalf("BuildJobMatchProfile: %v", err)
	}
	// Each of the 7 cross-owner APIs invoked exactly once.
	for _, name := range []string{"identity", "candidate-profile", "experience-cards", "resumes", "jds", "mocks", "debriefs"} {
		if got := cnt.get(name); got != 1 {
			t.Fatalf("%s call count = %d, want 1", name, got)
		}
	}
	if out.Profile.DisplayName != "Alice Example" {
		t.Fatalf("displayName = %q, want Alice Example", out.Profile.DisplayName)
	}
	if out.Profile.AvatarUrl != nil {
		t.Fatalf("avatarUrl must be nil at P0 baseline, got %v", out.Profile.AvatarUrl)
	}
	if out.Profile.Headline == nil || *out.Profile.Headline != "Senior frontend engineer" {
		t.Fatalf("headline = %v", out.Profile.Headline)
	}
	if out.Profile.YearsOfExperience == nil || *out.Profile.YearsOfExperience != 6 {
		t.Fatalf("yearsOfExperience = %v", out.Profile.YearsOfExperience)
	}
	if out.Profile.LocationText != nil {
		t.Fatalf("locationText must be nil at P0 baseline, got %v", out.Profile.LocationText)
	}
	if out.Profile.CompensationText != nil {
		t.Fatalf("compensationText must be nil at P0 baseline, got %v", out.Profile.CompensationText)
	}
	if out.Profile.Skills == nil || len(out.Profile.Skills) != 0 {
		t.Fatalf("skills must be [] at P0 baseline, got %#v", out.Profile.Skills)
	}
	if out.Profile.Sources.Resumes != 3 || out.Profile.Sources.Jds != 5 || out.Profile.Sources.Mocks != 8 || out.Profile.Sources.Debriefs != 2 {
		t.Fatalf("sources mismatch: %#v", out.Profile.Sources)
	}
	if out.ExperienceCardCount != 3 {
		t.Fatalf("experienceCardCount = %d, want 3 (D-2 trace-only)", out.ExperienceCardCount)
	}
	if !out.CallTrace.IdentityOK || !out.CallTrace.CandidateProfileOK || !out.CallTrace.ExperienceCardsOK || !out.CallTrace.ResumesOK || !out.CallTrace.TargetJobsOK || !out.CallTrace.PracticeOK || !out.CallTrace.DebriefsOK {
		t.Fatalf("call trace must all be ok: %#v", out.CallTrace)
	}
}

func TestBuildJobMatchProfileIdentityFailureFallsBack(t *testing.T) {
	cnt := &callCounter{}
	deps := happyPathDeps(t, cnt)
	deps.GetUserIdentity = func(ctx context.Context, userID string) (auth.UserIdentity, error) {
		cnt.inc("identity")
		return auth.UserIdentity{}, auth.ErrUserNotFound
	}
	out, err := service.BuildJobMatchProfile(context.Background(), "user-A", deps)
	if err != nil {
		t.Fatalf("BuildJobMatchProfile: %v", err)
	}
	if out.Profile.DisplayName != auth.AnonymousDisplayName {
		t.Fatalf("displayName fallback = %q, want %q", out.Profile.DisplayName, auth.AnonymousDisplayName)
	}
	if out.CallTrace.IdentityOK {
		t.Fatalf("identity trace must be false after fallback")
	}
	// Other cross-owner calls still complete.
	if !out.CallTrace.CandidateProfileOK || !out.CallTrace.ResumesOK {
		t.Fatalf("non-identity calls must still complete: %#v", out.CallTrace)
	}
}

func TestBuildJobMatchProfileNilCandidateProfileKeepsNullFields(t *testing.T) {
	cnt := &callCounter{}
	deps := happyPathDeps(t, cnt)
	deps.GetCandidateProfile = func(ctx context.Context, userID string) (*api.CandidateProfile, error) {
		cnt.inc("candidate-profile")
		return nil, nil
	}
	out, err := service.BuildJobMatchProfile(context.Background(), "user-A", deps)
	if err != nil {
		t.Fatalf("BuildJobMatchProfile: %v", err)
	}
	if out.Profile.Headline != nil || out.Profile.YearsOfExperience != nil {
		t.Fatalf("missing candidate_profile should leave headline / years nil, got %v / %v", out.Profile.Headline, out.Profile.YearsOfExperience)
	}
	if !out.CallTrace.CandidateProfileOK {
		t.Fatalf("call trace must record OK when fn returned nil err")
	}
}

func TestBuildJobMatchProfileCounterFailureFallsBackToZero(t *testing.T) {
	cnt := &callCounter{}
	deps := happyPathDeps(t, cnt)
	deps.CountResumes = func(ctx context.Context, userID string) (int, error) {
		cnt.inc("resumes")
		return 0, errors.New("transient db error")
	}
	out, err := service.BuildJobMatchProfile(context.Background(), "user-A", deps)
	if err != nil {
		t.Fatalf("BuildJobMatchProfile: %v", err)
	}
	if out.Profile.Sources.Resumes != 0 {
		t.Fatalf("resumes fallback should be 0, got %d", out.Profile.Sources.Resumes)
	}
	if out.CallTrace.ResumesOK {
		t.Fatalf("resumes trace must be false after fallback")
	}
	// Other counters unaffected.
	if out.Profile.Sources.Jds != 5 || !out.CallTrace.TargetJobsOK {
		t.Fatalf("other counters must remain consistent: %#v / %#v", out.Profile.Sources, out.CallTrace)
	}
}

func TestBuildJobMatchProfileExperienceCardsNeverEnterSources(t *testing.T) {
	cnt := &callCounter{}
	deps := happyPathDeps(t, cnt)
	deps.CountExperienceCardsBySource = func(ctx context.Context, userID string) (profile.SourceCounts, error) {
		cnt.inc("experience-cards")
		return profile.SourceCounts{"manual": 10, "resume_parse": 7, "practice_report": 3, "debrief": 5}, nil
	}
	out, err := service.BuildJobMatchProfile(context.Background(), "user-A", deps)
	if err != nil {
		t.Fatalf("BuildJobMatchProfile: %v", err)
	}
	if out.ExperienceCardCount != 25 {
		t.Fatalf("experienceCardCount = %d, want 25", out.ExperienceCardCount)
	}
	// Sources must remain the 4-counter projection (D-2).
	if out.Profile.Sources.Resumes != 3 || out.Profile.Sources.Jds != 5 || out.Profile.Sources.Mocks != 8 || out.Profile.Sources.Debriefs != 2 {
		t.Fatalf("experience cards leaked into sources: %#v", out.Profile.Sources)
	}
}

func TestBuildJobMatchProfileRejectsEmptyUserID(t *testing.T) {
	_, err := service.BuildJobMatchProfile(context.Background(), "  ", service.ProfileDeps{})
	if err == nil {
		t.Fatalf("expected error for empty userID")
	}
}

func TestBuildJobMatchProfileDisplayNameNeverEmpty(t *testing.T) {
	cnt := &callCounter{}
	deps := happyPathDeps(t, cnt)
	deps.GetUserIdentity = func(ctx context.Context, userID string) (auth.UserIdentity, error) {
		cnt.inc("identity")
		// Identity OK but display name is empty — spec D-17 says fall
		// back to anonymous display name rather than return "".
		return auth.UserIdentity{DisplayName: "", EmailMasked: "x***x@example.com"}, nil
	}
	out, err := service.BuildJobMatchProfile(context.Background(), "user-A", deps)
	if err != nil {
		t.Fatalf("BuildJobMatchProfile: %v", err)
	}
	if out.Profile.DisplayName == "" {
		t.Fatalf("displayName must never be empty")
	}
	if out.Profile.DisplayName != auth.AnonymousDisplayName {
		t.Fatalf("displayName fallback = %q, want %q", out.Profile.DisplayName, auth.AnonymousDisplayName)
	}
}
