package practice

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestListPracticeSessionsParsesFiltersAndPageInfo(t *testing.T) {
	cursor := "cursor-2"
	session := fixtureSessionRecord()
	session.Status = sharedtypes.SessionStatusCompleted
	service := &fakePlanService{
		listResult: domain.ListSessionsResult{
			Items:      []domain.SessionRecord{session},
			NextCursor: cursor,
			HasMore:    true,
			PageSize:   5,
		},
	}
	handler := newTestHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/practice/sessions?targetJobId=target-1&status=completed&pageSize=5&cursor=cursor-1", nil)
	req = req.WithContext(contextWithUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()
	handler.ListPracticeSessions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if service.listRequest.UserID != "user-1" ||
		service.listRequest.TargetJobID != "target-1" ||
		service.listRequest.Status != sharedtypes.SessionStatusCompleted ||
		service.listRequest.PageSize != 5 ||
		service.listRequest.Cursor != "cursor-1" {
		t.Fatalf("list request not mapped to service: %+v", service.listRequest)
	}
	var out api.PaginatedPracticeSession
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode PaginatedPracticeSession: %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].Id != session.ID {
		t.Fatalf("unexpected items: %+v", out.Items)
	}
	if out.PageInfo.NextCursor == nil || *out.PageInfo.NextCursor != cursor || !out.PageInfo.HasMore || out.PageInfo.PageSize != 5 {
		t.Fatalf("unexpected pageInfo: %+v", out.PageInfo)
	}
}

func TestListPracticeSessionsRejectsInvalidStatus(t *testing.T) {
	service := &fakePlanService{
		listErr: &domain.ServiceError{
			Code:    sharederrors.CodeValidationFailed,
			Message: "status is invalid",
			Details: map[string]any{"field": "status"},
		},
	}
	handler := newTestHandler(service)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/practice/sessions?status=not-a-status", nil)
	req = req.WithContext(contextWithUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	handler.ListPracticeSessions(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertAPIError(t, rec, sharederrors.CodeValidationFailed, false)
}

func TestListPracticeSessionsFixtureParity(t *testing.T) {
	fixture := loadListPracticeSessionsFixture(t)
	scenario := fixture.Scenarios["default"]
	service := &fakePlanService{
		listResult: domain.ListSessionsResult{
			Items:    sessionsFromFixture(scenario.Response.Body.Items),
			HasMore:  scenario.Response.Body.PageInfo.HasMore,
			PageSize: scenario.Response.Body.PageInfo.PageSize,
		},
	}
	if scenario.Response.Body.PageInfo.NextCursor != nil {
		service.listResult.NextCursor = *scenario.Response.Body.PageInfo.NextCursor
	}
	handler := newTestHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/practice/sessions?targetJobId=target-1&status=completed", nil)
	req = req.WithContext(contextWithUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()
	handler.ListPracticeSessions(rec, req)

	if rec.Code != scenario.Response.Status {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, scenario.Response.Status, rec.Body.String())
	}
	assertJSONEqual(t, mustJSON(t, scenario.Response.Body), rec.Body.Bytes())
}

type listPracticeSessionsFixture struct {
	Scenarios map[string]struct {
		Response struct {
			Status int                          `json:"status"`
			Body   api.PaginatedPracticeSession `json:"body"`
		} `json:"response"`
	} `json:"scenarios"`
}

func loadListPracticeSessionsFixture(t *testing.T) listPracticeSessionsFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "PracticeSessions", "listPracticeSessions.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture listPracticeSessionsFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return fixture
}

func sessionsFromFixture(items []api.PracticeSession) []domain.SessionRecord {
	out := make([]domain.SessionRecord, 0, len(items))
	for _, item := range items {
		out = append(out, sessionRecordFromFixture(item))
	}
	return out
}
