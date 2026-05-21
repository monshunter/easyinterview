package service

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/profile"
)

// fakeStore is a minimal in-memory profile.Store for service-layer tests.
type fakeStore struct {
	mu       sync.Mutex
	profiles map[string]*profile.CandidateProfileRecord
	cards    map[string]*profile.ExperienceCardRecord
	order    []string
	// behavior knobs
	failCount   error
	failDelete  error
	failProfile error
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		profiles: map[string]*profile.CandidateProfileRecord{},
		cards:    map[string]*profile.ExperienceCardRecord{},
	}
}

func (f *fakeStore) GetCandidateProfileByUser(_ context.Context, userID string) (*profile.CandidateProfileRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec, ok := f.profiles[userID]
	if !ok {
		return nil, profile.ErrNotFound
	}
	clone := *rec
	return &clone, nil
}

func (f *fakeStore) UpsertLite(_ context.Context, userID string, _ profile.ProfilePatch, _ profile.UserSettings) (*profile.CandidateProfileRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec, ok := f.profiles[userID]
	if !ok {
		rec = &profile.CandidateProfileRecord{UserID: userID, ProfileVersion: 1}
		f.profiles[userID] = rec
	}
	rec.ProfileVersion++
	clone := *rec
	return &clone, nil
}

func (f *fakeStore) SeedCandidateProfile(_ context.Context, userID string, _ profile.UserSettings) (*profile.CandidateProfileRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.profiles[userID]; ok {
		return nil, profile.ErrValidationFailed
	}
	rec := &profile.CandidateProfileRecord{UserID: userID, ProfileVersion: 1}
	f.profiles[userID] = rec
	clone := *rec
	return &clone, nil
}

func (f *fakeStore) DeleteCandidateProfileForUser(_ context.Context, userID string) (int64, error) {
	if f.failDelete != nil {
		return 0, f.failDelete
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.profiles[userID]; ok {
		delete(f.profiles, userID)
		return 1, nil
	}
	return 0, nil
}

func (f *fakeStore) ListExperienceCardsByUser(_ context.Context, userID string, _ *profile.ListCardsCursor, pageSize int32) (profile.ListCardsResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := profile.ListCardsResult{PageSize: pageSize}
	ids := make([]string, 0, len(f.cards))
	for id := range f.cards {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if c, ok := f.cards[id]; ok && c.UserID == userID {
			out.Items = append(out.Items, *c)
		}
	}
	return out, nil
}

func (f *fakeStore) CreateExperienceCard(_ context.Context, id string, userID string, attrs profile.ExperienceCardAttrs, source profile.ExperienceCardSource) (*profile.ExperienceCardRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec := &profile.ExperienceCardRecord{
		ID: id, UserID: userID, ProfileID: "p-" + userID,
		Title: attrs.Title, CompanyName: attrs.CompanyName, Situation: attrs.Situation,
		Task: attrs.Task, Action: attrs.Action, Result: attrs.Result,
		Skills: append([]string{}, attrs.Skills...), Language: attrs.Language,
		SourceType: source.SourceType, Confidence: source.Confidence,
	}
	f.cards[id] = rec
	f.order = append(f.order, id)
	return rec, nil
}

func (f *fakeStore) UpdateExperienceCard(_ context.Context, _, _ string, _ profile.ExperienceCardPatch) (*profile.ExperienceCardRecord, error) {
	return nil, nil
}

func (f *fakeStore) DeleteExperienceCardsForUser(_ context.Context, userID string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var removed int64
	for id, c := range f.cards {
		if c.UserID == userID {
			delete(f.cards, id)
			removed++
		}
	}
	return removed, nil
}

func (f *fakeStore) CountExperienceCardsBySource(_ context.Context, userID string) (profile.SourceCounts, error) {
	if f.failCount != nil {
		return nil, f.failCount
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	out := profile.SourceCounts{}
	for _, t := range profile.SourceTypes {
		out[t] = 0
	}
	for _, c := range f.cards {
		if c.UserID == userID {
			out[c.SourceType]++
		}
	}
	return out, nil
}

type recordedAudit struct {
	mu          sync.Mutex
	tombstones  []profile.CandidateProfileDeleteTombstone
	failTombstone error
}

func (r *recordedAudit) WriteCandidateProfileDeleteTombstone(_ context.Context, in profile.CandidateProfileDeleteTombstone) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failTombstone != nil {
		return r.failTombstone
	}
	r.tombstones = append(r.tombstones, in)
	return nil
}

func TestPrivacyDeleteOrderAndAudit(t *testing.T) {
	store := newFakeStore()
	store.profiles["user-a"] = &profile.CandidateProfileRecord{UserID: "user-a", ProfileVersion: 7}
	for i := 0; i < 5; i++ {
		id := "card-" + string(rune('a'+i))
		store.cards[id] = &profile.ExperienceCardRecord{ID: id, UserID: "user-a", SourceType: profile.SourceTypeManual}
	}
	// Plus a foreign card to ensure cross-user isolation.
	store.cards["other"] = &profile.ExperienceCardRecord{ID: "other", UserID: "user-b", SourceType: profile.SourceTypeManual}

	audit := &recordedAudit{}
	svc := New(Options{Store: store, Audit: audit, Now: func() time.Time {
		return time.Date(2026, 5, 21, 8, 0, 0, 0, time.UTC)
	}})

	if err := svc.DeleteCandidateProfileForUser(context.Background(), "user-a", "job-001"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, ok := store.profiles["user-a"]; ok {
		t.Fatal("user-a profile must be removed")
	}
	if _, ok := store.cards["card-a"]; ok {
		t.Fatal("user-a card-a must be removed")
	}
	if _, ok := store.cards["other"]; !ok {
		t.Fatal("user-b card must be preserved")
	}
	if len(audit.tombstones) != 1 {
		t.Fatalf("audit tombstones = %d, want 1", len(audit.tombstones))
	}
	ts := audit.tombstones[0]
	if ts.UserID != "user-a" {
		t.Fatalf("tombstone userId = %q", ts.UserID)
	}
	if ts.ExperienceCardCount != 5 {
		t.Fatalf("tombstone experienceCardCount = %d, want 5", ts.ExperienceCardCount)
	}
	if ts.JobID != "job-001" {
		t.Fatalf("tombstone jobId = %q", ts.JobID)
	}
	if ts.DeletedAt.IsZero() {
		t.Fatal("tombstone deletedAt must be set")
	}
}

func TestCountExperienceCardsBySource(t *testing.T) {
	store := newFakeStore()
	for i := 0; i < 3; i++ {
		store.cards[string(rune('a'+i))] = &profile.ExperienceCardRecord{
			ID: string(rune('a' + i)), UserID: "user-a", SourceType: profile.SourceTypeManual,
		}
	}
	for i := 0; i < 2; i++ {
		store.cards["r"+string(rune('a'+i))] = &profile.ExperienceCardRecord{
			ID: "r" + string(rune('a'+i)), UserID: "user-a", SourceType: profile.SourceTypeResumeParse,
		}
	}
	svc := New(Options{Store: store})
	counts, err := svc.CountExperienceCardsBySource(context.Background(), "user-a")
	if err != nil {
		t.Fatalf("counts: %v", err)
	}
	if counts[profile.SourceTypeManual] != 3 {
		t.Fatalf("manual = %d, want 3", counts[profile.SourceTypeManual])
	}
	if counts[profile.SourceTypeResumeParse] != 2 {
		t.Fatalf("resume_parse = %d, want 2", counts[profile.SourceTypeResumeParse])
	}
	if counts[profile.SourceTypePracticeReport] != 0 {
		t.Fatalf("practice_report = %d, want 0", counts[profile.SourceTypePracticeReport])
	}
	if counts[profile.SourceTypeDebrief] != 0 {
		t.Fatalf("debrief = %d, want 0", counts[profile.SourceTypeDebrief])
	}
}

func TestGetCandidateProfileForUserSeededAndNil(t *testing.T) {
	store := newFakeStore()
	headline := "Senior frontend"
	role := "Tech Lead"
	yoe := int32(5)
	region := "CN-SH"
	store.profiles["user-a"] = &profile.CandidateProfileRecord{
		UserID:                    "user-a",
		Headline:                  &headline,
		YearsOfExperience:         &yoe,
		CurrentRole:               &role,
		PreferredPracticeLanguage: "en",
		UiLanguage:                "zh-CN",
		Region:                    &region,
		ProfileVersion:            3,
	}
	svc := New(Options{Store: store})

	// Seeded user returns full profile.
	got, err := svc.GetCandidateProfileForUser(context.Background(), "user-a")
	if err != nil {
		t.Fatalf("seeded read: %v", err)
	}
	if got == nil {
		t.Fatal("seeded user returned nil; want CandidateProfile")
	}
	if got.Headline == nil || *got.Headline != headline {
		t.Fatalf("headline = %v", got.Headline)
	}
	if got.YearsOfExperience == nil || *got.YearsOfExperience != 5 {
		t.Fatalf("yearsOfExperience = %v", got.YearsOfExperience)
	}

	// Unseeded user returns (nil, nil) — no side effects.
	got, err = svc.GetCandidateProfileForUser(context.Background(), "user-c")
	if err != nil {
		t.Fatalf("unseeded read: %v", err)
	}
	if got != nil {
		t.Fatalf("unseeded user returned %+v; want nil", got)
	}
	// Store remains pristine: user-c not added.
	store.mu.Lock()
	_, exists := store.profiles["user-c"]
	store.mu.Unlock()
	if exists {
		t.Fatal("unseeded read created a profile row; D-13 requires read-only semantics")
	}
}
