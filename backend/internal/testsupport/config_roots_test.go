package testsupport

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigRootsPointAtCurrentTruthSources(t *testing.T) {
	prompts, rubrics := ConfigRoots(t)
	for name, root := range map[string]string{
		"prompts": prompts,
		"rubrics": rubrics,
	} {
		info, err := os.Stat(root)
		if err != nil {
			t.Fatalf("stat %s root %q: %v", name, root, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s root is not a directory: %s", name, root)
		}
		if filepath.Base(root) != name {
			t.Fatalf("%s root = %q", name, root)
		}
	}
}
