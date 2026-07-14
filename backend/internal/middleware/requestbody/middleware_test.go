package requestbody

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLimitAcceptsExactBytesAndRejectsPlusOneBeforeHandler(t *testing.T) {
	const maxBytes int64 = 6
	for _, tc := range []struct {
		name       string
		body       string
		wantStatus int
		wantCalls  int
	}{
		{name: "exact", body: "123456", wantStatus: http.StatusNoContent, wantCalls: 1},
		{name: "plus one", body: "1234567", wantStatus: http.StatusRequestEntityTooLarge, wantCalls: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(http.StatusNoContent)
			})
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/test", strings.NewReader(tc.body))

			Limit(maxBytes, next).ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus || calls != tc.wantCalls {
				t.Fatalf("status=%d calls=%d, want status=%d calls=%d body=%s", rec.Code, calls, tc.wantStatus, tc.wantCalls, rec.Body.String())
			}
			if tc.wantStatus == http.StatusRequestEntityTooLarge && !strings.Contains(rec.Body.String(), `"code":"VALIDATION_FAILED"`) {
				t.Fatalf("oversize response must use API error envelope: %s", rec.Body.String())
			}
		})
	}
}
