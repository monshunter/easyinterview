package migrations

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBaselineAITaskRunsTaskTypeCheckIncludesHintGenerate(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))

	required := "task_type text not null check (task_type in ('jd_parse', 'resume_parse', 'question_generate', 'followup_generate', 'report_generate', 'resume_tailor', 'debrief_generate', 'hint_generate'))"
	if !strings.Contains(up, required) {
		t.Fatalf("ai_task_runs.task_type check must include hint_generate")
	}
	if strings.Contains(up, "'unknown_task'") {
		t.Fatalf("baseline migration must not allow unknown_task")
	}
}
