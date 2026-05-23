package handler

import (
	"context"
	"sort"
	"sync"

	"github.com/monshunter/easyinterview/backend/internal/profile"
)

// fakeStore is an in-memory profile.Store used by the handler unit tests.
// It does not aim to replicate SQL transaction semantics — it implements just
// enough behavior for the handler-level invariants (seed-once, patch
// semantics, cross-user 404, cursor pagination ordering, source_type force).
type fakeStore struct {
	mu       sync.Mutex
	profiles map[string]*profile.CandidateProfileRecord
	cards    map[string]*profile.ExperienceCardRecord
	order    []string // insertion order to drive stable iteration in tests
	newID    func() string
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

func (f *fakeStore) UpsertLite(_ context.Context, userID string, patch profile.ProfilePatch, defaults profile.UserSettings) (*profile.CandidateProfileRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec, ok := f.profiles[userID]
	if !ok {
		rec = &profile.CandidateProfileRecord{
			UserID:                    userID,
			PreferredPracticeLanguage: defaults.PreferredPracticeLanguage,
			UILanguage:                defaults.UILanguage,
			Region:                    defaults.Region,
			ProfileVersion:            1,
		}
		f.profiles[userID] = rec
	}
	if patch.Headline != nil {
		v := *patch.Headline
		rec.Headline = &v
	}
	if patch.YearsOfExperience != nil {
		v := *patch.YearsOfExperience
		rec.YearsOfExperience = &v
	}
	if patch.CurrentRole != nil {
		v := *patch.CurrentRole
		rec.CurrentRole = &v
	}
	if patch.PreferredPracticeLanguage != nil {
		rec.PreferredPracticeLanguage = *patch.PreferredPracticeLanguage
	}
	if patch.UILanguage != nil {
		rec.UILanguage = *patch.UILanguage
	}
	if patch.Region != nil {
		v := *patch.Region
		rec.Region = &v
	}
	rec.ProfileVersion++
	clone := *rec
	return &clone, nil
}

func (f *fakeStore) SeedCandidateProfile(_ context.Context, userID string, defaults profile.UserSettings) (*profile.CandidateProfileRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.profiles[userID]; ok {
		return nil, profile.ErrValidationFailed
	}
	rec := &profile.CandidateProfileRecord{
		UserID:                    userID,
		PreferredPracticeLanguage: defaults.PreferredPracticeLanguage,
		UILanguage:                defaults.UILanguage,
		Region:                    defaults.Region,
		ProfileVersion:            1,
	}
	f.profiles[userID] = rec
	clone := *rec
	return &clone, nil
}

func (f *fakeStore) DeleteCandidateProfileForUser(_ context.Context, userID string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.profiles[userID]; ok {
		delete(f.profiles, userID)
		return 1, nil
	}
	return 0, nil
}

func (f *fakeStore) ListExperienceCardsByUser(_ context.Context, userID string, cursor *profile.ListCardsCursor, pageSize int32) (profile.ListCardsResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if pageSize <= 0 {
		pageSize = profile.DefaultExperienceCardSize
	}
	cards := make([]profile.ExperienceCardRecord, 0)
	for _, id := range f.order {
		card, ok := f.cards[id]
		if !ok || card.UserID != userID {
			continue
		}
		cards = append(cards, *card)
	}
	sort.Slice(cards, func(i, j int) bool {
		if !cards[i].UpdatedAt.Equal(cards[j].UpdatedAt) {
			return cards[i].UpdatedAt.After(cards[j].UpdatedAt)
		}
		return cards[i].ID > cards[j].ID
	})
	if cursor != nil {
		idx := 0
		for idx < len(cards) {
			c := cards[idx]
			if c.UpdatedAt.Before(cursor.UpdatedAt) || (c.UpdatedAt.Equal(cursor.UpdatedAt) && c.ID < cursor.ID) {
				break
			}
			idx++
		}
		cards = cards[idx:]
	}
	res := profile.ListCardsResult{PageSize: pageSize}
	if int32(len(cards)) > pageSize {
		res.Items = append(res.Items, cards[:pageSize]...)
		res.HasMore = true
		last := res.Items[len(res.Items)-1]
		res.NextCursor = encodeCursor(last.UpdatedAt, last.ID)
	} else {
		res.Items = append(res.Items, cards...)
	}
	return res, nil
}

func (f *fakeStore) CreateExperienceCard(_ context.Context, id string, userID string, attrs profile.ExperienceCardAttrs, source profile.ExperienceCardSource) (*profile.ExperienceCardRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.cards[id]; ok {
		return nil, profile.ErrValidationFailed
	}
	rec := &profile.ExperienceCardRecord{
		ID:          id,
		UserID:      userID,
		ProfileID:   "fake-profile-" + userID,
		Title:       attrs.Title,
		CompanyName: attrs.CompanyName,
		Situation:   attrs.Situation,
		Task:        attrs.Task,
		Action:      attrs.Action,
		Result:      attrs.Result,
		Skills:      append([]string{}, attrs.Skills...),
		Language:    attrs.Language,
		SourceType:  source.SourceType,
		Confidence:  source.Confidence,
	}
	if f.newID != nil {
		_ = f.newID() // sink to track call count if needed
	}
	f.cards[id] = rec
	f.order = append(f.order, id)
	clone := *rec
	return &clone, nil
}

func (f *fakeStore) UpdateExperienceCard(_ context.Context, cardID string, userID string, patch profile.ExperienceCardPatch) (*profile.ExperienceCardRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec, ok := f.cards[cardID]
	if !ok || rec.UserID != userID {
		return nil, profile.ErrNotFound
	}
	if patch.Title != nil {
		rec.Title = *patch.Title
	}
	if patch.CompanyName != nil {
		rec.CompanyName = *patch.CompanyName
	}
	if patch.Situation != nil {
		rec.Situation = *patch.Situation
	}
	if patch.Task != nil {
		rec.Task = *patch.Task
	}
	if patch.Action != nil {
		rec.Action = *patch.Action
	}
	if patch.Result != nil {
		rec.Result = *patch.Result
	}
	if patch.Skills != nil {
		rec.Skills = append([]string{}, (*patch.Skills)...)
	}
	if patch.Language != nil {
		rec.Language = *patch.Language
	}
	clone := *rec
	return &clone, nil
}

func (f *fakeStore) DeleteExperienceCardsForUser(_ context.Context, userID string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var removed int64
	keep := make([]string, 0, len(f.order))
	for _, id := range f.order {
		card, ok := f.cards[id]
		if ok && card.UserID == userID {
			delete(f.cards, id)
			removed++
			continue
		}
		keep = append(keep, id)
	}
	f.order = keep
	return removed, nil
}

func (f *fakeStore) CountExperienceCardsBySource(_ context.Context, userID string) (profile.SourceCounts, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make(profile.SourceCounts, len(profile.SourceTypes))
	for _, t := range profile.SourceTypes {
		out[t] = 0
	}
	for _, card := range f.cards {
		if card.UserID != userID {
			continue
		}
		out[card.SourceType]++
	}
	return out, nil
}

// fakeSettings returns a static UserSettings payload for every user.
type fakeSettings struct {
	defaults profile.UserSettings
}

func (f fakeSettings) GetUserSettings(_ context.Context, _ string) (profile.UserSettings, error) {
	return f.defaults, nil
}
