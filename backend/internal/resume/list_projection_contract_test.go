package resume_test

import (
	"os"
	"strings"
	"testing"
)

func TestListResumesDoesNotReuseFullDetailMappers(t *testing.T) {
	storeSource := readSourceFile(t, "store/resumes.go")
	listStore := functionSource(t, storeSource, "func (r *Repository) List(")
	for _, required := range []string{"resumeSummarySelectColumns", "scanResumeSummary"} {
		if !strings.Contains(listStore, required) {
			t.Fatalf("store List missing summary-only anchor %q", required)
		}
	}
	for _, forbidden := range []string{"resumeSelectColumns", "scanResume(", "select *"} {
		if strings.Contains(strings.ToLower(listStore), strings.ToLower(forbidden)) {
			t.Fatalf("store List reuses full-detail query/scanner %q", forbidden)
		}
	}

	serviceSource := readSourceFile(t, "service.go")
	listService := functionSource(t, serviceSource, "func (s *Service) ListResumes(")
	if !strings.Contains(listService, "resumeSummaryRecordToAPI") {
		t.Fatal("service ListResumes does not use the closed summary mapper")
	}
	if strings.Contains(listService, "resumeRecordToAPI") {
		t.Fatal("service ListResumes reuses the full-detail mapper")
	}
}

func readSourceFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func functionSource(t *testing.T, source string, signature string) string {
	t.Helper()
	start := strings.Index(source, signature)
	if start < 0 {
		t.Fatalf("function signature %q not found", signature)
	}
	rest := source[start+len(signature):]
	next := strings.Index(rest, "\nfunc ")
	if next < 0 {
		return source[start:]
	}
	return source[start : start+len(signature)+next]
}
