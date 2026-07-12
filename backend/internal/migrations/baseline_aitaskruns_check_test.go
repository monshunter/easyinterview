package migrations

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBaselineAITaskRunsTaskTypeCheckUsesConversationTasks(t *testing.T) {
	root := repoRoot(t)
	up := strings.ToLower(readFile(t, filepath.Join(root, "migrations", "000001_create_baseline.up.sql")))

	for _, required := range []string{
		"'practice_chat'",
		"'report_generate'",
	} {
		if !strings.Contains(up, required) {
			t.Fatalf("ai_task_runs.task_type check must include %s", required)
		}
	}
	if strings.Contains(up, "'unknown_task'") {
		t.Fatalf("baseline migration must not allow unknown_task")
	}
}
