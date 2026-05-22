//go:build integration

package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/monshunter/easyinterview/backend/internal/profile"
	profilestore "github.com/monshunter/easyinterview/backend/internal/profile/store"
)

// openProfileDB returns a *sql.DB pointed at the local dev-stack Postgres or
// the DATABASE_URL configured for backend integration tests. Caller-owned
// cleanup: t.Cleanup closes the connection pool.
func openProfileDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Skipf("postgres ping failed (%v); skipping profile store integration test", err)
	}
	return db
}

// profileIntegrationFixture provisions a clean (users, user_settings,
// candidate_profiles, experience_cards) row set for a single test, then
// schedules cleanup via t.Cleanup.
type profileIntegrationFixture struct {
	UserA string
	UserB string
	UserC string
}

func newProfileFixture(t *testing.T, ctx context.Context, db *sql.DB) profileIntegrationFixture {
	t.Helper()
	fx := profileIntegrationFixture{
		UserA: "01918fa0-0000-7000-8000-0000000aa101",
		UserB: "01918fa0-0000-7000-8000-0000000bb201",
		UserC: "01918fa0-0000-7000-8000-0000000cc301",
	}
	for _, id := range []string{fx.UserA, fx.UserB, fx.UserC} {
		email := id + "@profile-integration.local"
		if _, err := db.ExecContext(ctx, `
insert into users (id, email, status)
values ($1, $2, 'active')
on conflict (id) do nothing`, id, email); err != nil {
			t.Fatalf("seed user %s: %v", id, err)
		}
		if _, err := db.ExecContext(ctx, `
insert into user_settings (user_id, ui_language, preferred_practice_language, region)
values ($1, 'zh-CN', 'en', 'CN-SH')
on conflict (user_id) do update set
  ui_language = excluded.ui_language,
  preferred_practice_language = excluded.preferred_practice_language,
  region = excluded.region`, id); err != nil {
			t.Fatalf("seed user_settings %s: %v", id, err)
		}
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = db.ExecContext(cleanupCtx, `delete from experience_cards where user_id = any($1)`, fixtureUserIDs(fx))
		_, _ = db.ExecContext(cleanupCtx, `delete from candidate_profiles where user_id = any($1)`, fixtureUserIDs(fx))
		_, _ = db.ExecContext(cleanupCtx, `delete from audit_events where user_id = any($1) and action = 'profile.privacy_delete'`, fixtureUserIDs(fx))
		_, _ = db.ExecContext(cleanupCtx, `delete from user_settings where user_id = any($1)`, fixtureUserIDs(fx))
		_, _ = db.ExecContext(cleanupCtx, `delete from users where id = any($1)`, fixtureUserIDs(fx))
	})
	return fx
}

func fixtureUserIDs(fx profileIntegrationFixture) interface{} {
	return pqArray([]string{fx.UserA, fx.UserB, fx.UserC})
}

// pqArray returns a value suitable for `id = any($1)` queries with a
// text/uuid array. We avoid importing pq just for tests by formatting the
// values into a Postgres array literal.
func pqArray(values []string) interface{} {
	buf := []byte("{")
	for i, v := range values {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, '"')
		buf = append(buf, v...)
		buf = append(buf, '"')
	}
	buf = append(buf, '}')
	return string(buf)
}

func TestProfileStoreSeedPatchVersionAndDelete(t *testing.T) {
	db := openProfileDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	fx := newProfileFixture(t, ctx, db)

	repo := profilestore.NewRepositoryWith(db, profilestore.NewRepositoryOptions{
		NewID: deterministicIDFactory(),
	})

	region := "CN-SH"
	defaults := profile.UserSettings{
		PreferredPracticeLanguage: "en",
		UILanguage:                "zh-CN",
		Region:                    &region,
	}

	// Seed candidate_profile for user A.
	rec, err := repo.SeedCandidateProfile(ctx, fx.UserA, defaults)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if rec.ProfileVersion != 1 {
		t.Fatalf("seed profile_version = %d, want 1", rec.ProfileVersion)
	}
	if rec.Headline != nil || rec.YearsOfExperience != nil || rec.CurrentRole != nil {
		t.Fatalf("seed must produce null nullable columns, got %+v", rec)
	}

	// Second seed should fail with ErrValidationFailed (unique violation).
	if _, err := repo.SeedCandidateProfile(ctx, fx.UserA, defaults); err == nil {
		t.Fatal("second seed expected to fail; got nil")
	}

	// Patch: bump headline + years.
	headline := "Senior frontend"
	years := int32(5)
	out, err := repo.UpsertLite(ctx, fx.UserA, profile.ProfilePatch{
		Headline:          &headline,
		YearsOfExperience: &years,
	}, defaults)
	if err != nil {
		t.Fatalf("upsert lite: %v", err)
	}
	if out.ProfileVersion != 2 {
		t.Fatalf("post-patch profile_version = %d, want 2", out.ProfileVersion)
	}
	if out.Headline == nil || *out.Headline != headline {
		t.Fatalf("headline not persisted: %+v", out.Headline)
	}

	// Subsequent patch bumps version again (monotonic).
	role := "Tech Lead"
	out, err = repo.UpsertLite(ctx, fx.UserA, profile.ProfilePatch{CurrentRole: &role}, defaults)
	if err != nil {
		t.Fatalf("upsert lite second: %v", err)
	}
	if out.ProfileVersion != 3 {
		t.Fatalf("monotonic version = %d, want 3", out.ProfileVersion)
	}
	if out.Headline == nil || *out.Headline != headline {
		t.Fatalf("headline overwritten on subsequent patch: %+v", out.Headline)
	}

	// Read-only GetByUser returns the same row without seeding (D-13).
	got, err := repo.GetCandidateProfileByUser(ctx, fx.UserA)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.ProfileVersion != 3 {
		t.Fatalf("get returned %+v", got)
	}

	// User without seed returns ErrNotFound (D-13 nil semantics).
	if _, err := repo.GetCandidateProfileByUser(ctx, fx.UserC); err == nil {
		t.Fatal("expected ErrNotFound for unseeded user")
	}

	// Delete removes the row.
	removed, err := repo.DeleteCandidateProfileForUser(ctx, fx.UserA)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if removed != 1 {
		t.Fatalf("delete rows = %d, want 1", removed)
	}
	if _, err := repo.GetCandidateProfileByUser(ctx, fx.UserA); err == nil {
		t.Fatal("expected ErrNotFound after delete")
	}
}

func TestExperienceCardsCursorPaginationAndCrossUser(t *testing.T) {
	db := openProfileDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	fx := newProfileFixture(t, ctx, db)

	id := deterministicIDFactory()
	repo := profilestore.NewRepositoryWith(db, profilestore.NewRepositoryOptions{NewID: id})
	region := "CN-SH"
	defaults := profile.UserSettings{
		PreferredPracticeLanguage: "en",
		UILanguage:                "zh-CN",
		Region:                    &region,
	}
	if _, err := repo.SeedCandidateProfile(ctx, fx.UserA, defaults); err != nil {
		t.Fatalf("seed A: %v", err)
	}
	if _, err := repo.SeedCandidateProfile(ctx, fx.UserB, defaults); err != nil {
		t.Fatalf("seed B: %v", err)
	}

	cardIDs := makeCardIDs(26)
	for i := 0; i < 25; i++ {
		source := profile.SourceTypeManual
		switch {
		case i < 12:
			source = profile.SourceTypeManual
		case i < 20:
			source = profile.SourceTypeResumeParse
		case i < 23:
			source = profile.SourceTypePracticeReport
		default:
			source = profile.SourceTypeDebrief
		}
		_, err := repo.CreateExperienceCard(ctx, cardIDs[i], fx.UserA, profile.ExperienceCardAttrs{
			Title:       "card-" + cardIDs[i],
			CompanyName: "Acme",
			Situation:   "s",
			Task:        "t",
			Action:      "a",
			Result:      "r",
			Skills:      []string{"go"},
			Language:    "en",
		}, profile.ExperienceCardSource{SourceType: source, Confidence: profile.ConfidenceDefaultMedium})
		if err != nil {
			t.Fatalf("create card %d: %v", i, err)
		}
	}
	// Cross-user card.
	if _, err := repo.CreateExperienceCard(ctx, cardIDs[25], fx.UserB, profile.ExperienceCardAttrs{
		Title: "B-card", CompanyName: "BCo", Situation: "s", Task: "t", Action: "a", Result: "r", Skills: []string{}, Language: "en",
	}, profile.ExperienceCardSource{SourceType: profile.SourceTypeManual, Confidence: profile.ConfidenceDefaultMedium}); err != nil {
		t.Fatalf("create B card: %v", err)
	}

	first, err := repo.ListExperienceCardsByUser(ctx, fx.UserA, nil, 20)
	if err != nil {
		t.Fatalf("list first: %v", err)
	}
	if len(first.Items) != 20 || !first.HasMore || first.NextCursor == "" {
		t.Fatalf("first page len=%d hasMore=%v cursor=%q", len(first.Items), first.HasMore, first.NextCursor)
	}

	// Page 2 via decoded cursor (we read from store: nextCursor is opaque
	// base64(json{u,i})). For the integration test we re-pass it verbatim by
	// decoding via the handler-side decoder; but here we only need to confirm
	// behavior with the same encoded value handed back to the store layer
	// via a small helper that mirrors handler.decodeCursor logic.
	cursor := decodeCursorForTest(t, first.NextCursor)
	second, err := repo.ListExperienceCardsByUser(ctx, fx.UserA, &cursor, 20)
	if err != nil {
		t.Fatalf("list second: %v", err)
	}
	if len(second.Items) != 5 || second.HasMore {
		t.Fatalf("second page len=%d hasMore=%v", len(second.Items), second.HasMore)
	}

	// Cross-user isolation: user B sees only B card.
	listB, err := repo.ListExperienceCardsByUser(ctx, fx.UserB, nil, 20)
	if err != nil {
		t.Fatalf("list B: %v", err)
	}
	if len(listB.Items) != 1 || listB.Items[0].UserID != fx.UserB {
		t.Fatalf("B list = %+v", listB.Items)
	}

	// Counts.
	counts, err := repo.CountExperienceCardsBySource(ctx, fx.UserA)
	if err != nil {
		t.Fatalf("counts: %v", err)
	}
	want := profile.SourceCounts{
		profile.SourceTypeManual:         12,
		profile.SourceTypeResumeParse:    8,
		profile.SourceTypePracticeReport: 3,
		profile.SourceTypeDebrief:        2,
	}
	for k, v := range want {
		if counts[k] != v {
			t.Fatalf("counts[%s] = %d, want %d", k, counts[k], v)
		}
	}

	// Cross-user update returns ErrNotFound (D-8 isolation).
	other := "stranger"
	if _, err := repo.UpdateExperienceCard(ctx, cardIDs[0], fx.UserB, profile.ExperienceCardPatch{Title: &other}); err == nil {
		t.Fatal("cross-user update must fail with ErrNotFound")
	}

	// Delete experience cards for user A.
	removed, err := repo.DeleteExperienceCardsForUser(ctx, fx.UserA)
	if err != nil {
		t.Fatalf("delete A cards: %v", err)
	}
	if removed != 25 {
		t.Fatalf("delete A removed = %d, want 25", removed)
	}
}

func TestPrivacyDeleteWithAuditRollsBackAndWritesFailureAudit(t *testing.T) {
	db := openProfileDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	fx := newProfileFixture(t, ctx, db)

	repo := profilestore.NewRepositoryWith(db, profilestore.NewRepositoryOptions{
		NewID: deterministicIDFactory(),
	})
	region := "CN-SH"
	defaults := profile.UserSettings{
		PreferredPracticeLanguage: "en",
		UILanguage:                "zh-CN",
		Region:                    &region,
	}
	if _, err := repo.SeedCandidateProfile(ctx, fx.UserA, defaults); err != nil {
		t.Fatalf("seed A: %v", err)
	}
	for i, id := range makeCardIDs(2) {
		if _, err := repo.CreateExperienceCard(ctx, id, fx.UserA, profile.ExperienceCardAttrs{
			Title:       "privacy rollback card",
			CompanyName: "Acme",
			Situation:   "sensitive situation",
			Task:        "sensitive task",
			Action:      "sensitive action",
			Result:      "sensitive result",
			Skills:      []string{"go"},
			Language:    "en",
		}, profile.ExperienceCardSource{
			SourceType: profile.SourceTypeManual,
			Confidence: profile.ConfidenceDefaultMedium,
		}); err != nil {
			t.Fatalf("create card %d: %v", i, err)
		}
	}

	installCandidateProfileDeleteFailureTrigger(t, ctx, db, fx.UserA)

	err := repo.DeleteCandidateProfileForUserWithAudit(
		ctx,
		fx.UserA,
		"job-rollback",
		time.Date(2026, 5, 21, 11, 0, 0, 0, time.UTC),
	)
	if err == nil || !strings.Contains(err.Error(), "delete_candidate_profile") {
		t.Fatalf("privacy delete error = %v, want delete_candidate_profile", err)
	}

	var remainingProfiles, remainingCards int
	if err := db.QueryRowContext(ctx, `select count(*) from candidate_profiles where user_id = $1`, fx.UserA).Scan(&remainingProfiles); err != nil {
		t.Fatalf("count profiles: %v", err)
	}
	if err := db.QueryRowContext(ctx, `select count(*) from experience_cards where user_id = $1`, fx.UserA).Scan(&remainingCards); err != nil {
		t.Fatalf("count cards: %v", err)
	}
	if remainingProfiles != 1 || remainingCards != 2 {
		t.Fatalf("rollback left profiles=%d cards=%d, want 1/2", remainingProfiles, remainingCards)
	}

	var result string
	var metadataRaw []byte
	if err := db.QueryRowContext(ctx, `
select result, metadata::text
  from audit_events
 where user_id = $1 and action = 'profile.privacy_delete'
 order by created_at desc limit 1`, fx.UserA).Scan(&result, &metadataRaw); err != nil {
		t.Fatalf("failure audit lookup: %v", err)
	}
	if result != "failure" {
		t.Fatalf("audit result = %q, want failure", result)
	}
	var metadata map[string]any
	if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	if metadata["errorStage"] != "delete_candidate_profile" || metadata["jobId"] != "job-rollback" {
		t.Fatalf("failure audit metadata = %#v", metadata)
	}
	for _, forbidden := range []string{"privacy rollback card", "sensitive situation", "sensitive task", "sensitive action", "sensitive result", "Acme"} {
		if strings.Contains(string(metadataRaw), forbidden) {
			t.Fatalf("failure audit leaked raw content %q in %s", forbidden, metadataRaw)
		}
	}
}

func installCandidateProfileDeleteFailureTrigger(t *testing.T, ctx context.Context, db *sql.DB, userID string) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `drop trigger if exists profile_delete_fail_rollback_test on candidate_profiles`); err != nil {
		t.Fatalf("drop old trigger: %v", err)
	}
	if _, err := db.ExecContext(ctx, `drop function if exists profile_delete_fail_rollback_test()`); err != nil {
		t.Fatalf("drop old function: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
create function profile_delete_fail_rollback_test()
returns trigger
language plpgsql
as $$
begin
  if old.user_id = '`+userID+`'::uuid then
    raise exception 'profile delete failure injected';
  end if;
  return old;
end;
$$`); err != nil {
		t.Fatalf("create failure function: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
create trigger profile_delete_fail_rollback_test
before delete on candidate_profiles
for each row
execute function profile_delete_fail_rollback_test()`); err != nil {
		t.Fatalf("create failure trigger: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, _ = db.ExecContext(cleanupCtx, `drop trigger if exists profile_delete_fail_rollback_test on candidate_profiles`)
		_, _ = db.ExecContext(cleanupCtx, `drop function if exists profile_delete_fail_rollback_test()`)
	})
}

func makeCardIDs(n int) []string {
	out := make([]string, n)
	for i := range out {
		hex := []byte("0123456789abcdef")
		buf := []byte("01918fa1-0000-7000-8000-000000000000")
		v := i + 1
		for j := len(buf) - 1; j >= 0 && v > 0; j-- {
			if buf[j] == '-' {
				continue
			}
			buf[j] = hex[v%16]
			v /= 16
		}
		out[i] = string(buf)
	}
	return out
}

func deterministicIDFactory() func() string {
	counter := 0
	return func() string {
		counter++
		hex := []byte("0123456789abcdef")
		buf := []byte("01918fa2-0000-7000-8000-000000000000")
		v := counter
		for j := len(buf) - 1; j >= 0 && v > 0; j-- {
			if buf[j] == '-' {
				continue
			}
			buf[j] = hex[v%16]
			v /= 16
		}
		return string(buf)
	}
}

// decodeCursorForTest mirrors handler.decodeCursor so the store-only
// integration test can re-issue paged queries without depending on the
// handler package.
func decodeCursorForTest(t *testing.T, raw string) profile.ListCardsCursor {
	t.Helper()
	if raw == "" {
		t.Fatal("decodeCursorForTest: empty cursor")
	}
	// We reuse the handler-side encoder format (base64-url JSON {u,i}); the
	// concrete decoder is private to handler, so reimplement here.
	payload, err := decodeBase64URL(raw)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	type wire struct {
		U string `json:"u"`
		I string `json:"i"`
	}
	var w wire
	if err := jsonUnmarshal(payload, &w); err != nil {
		t.Fatalf("unmarshal cursor: %v", err)
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, w.U)
	if err != nil {
		t.Fatalf("parse updatedAt: %v", err)
	}
	return profile.ListCardsCursor{UpdatedAt: updatedAt.UTC(), ID: w.I}
}
