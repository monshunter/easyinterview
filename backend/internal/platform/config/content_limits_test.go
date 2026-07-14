package config_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

func TestContentLimitsUseCodeDefaultsWhenKeysAreMissing(t *testing.T) {
	loader, err := config.Load(config.Options{ConfigDir: t.TempDir()})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	got, err := loader.ContentLimits()
	if err != nil {
		t.Fatalf("ContentLimits: %v", err)
	}
	want := config.ContentLimits{
		HTTPMaxRequestBodyBytes:        10 * 1024 * 1024,
		ResumeUploadBytes:              10 * 1024 * 1024,
		PrivacyExportUploadBytes:       5 * 1024 * 1024,
		ResumeMaxActive:                10,
		ResumeMaxExtractedTextBytes:    384 * 1024,
		ResumeMaxPasteTextBytes:        384 * 1024,
		TargetJobMaxRawTextBytes:       96 * 1024,
		PracticeMaxMessageBytes:        32 * 1024,
		PracticeMaxSessionTextBytes:    256 * 1024,
		ReportMaxFramedInputBytes:      896 * 1024,
		AIProviderMaxResponseBodyBytes: 4 * 1024 * 1024,
	}
	if got != want {
		t.Fatalf("ContentLimits() = %+v, want %+v", got, want)
	}
	if defaults := config.DefaultContentLimits(); defaults != want {
		t.Fatalf("DefaultContentLimits() = %+v, want %+v", defaults, want)
	}
}

func TestContentLimitsHonorYAMLOverrides(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), `
http:
  maxRequestBodyBytes: 11000001
upload:
  maxBytes:
    resume: 11000002
    privacyExport: 6000003
resume:
  maxActive: 11
  maxExtractedTextBytes: 400004
  maxPasteTextBytes: 400003
targetJob:
  maxRawTextBytes: 100005
practice:
  maxMessageBytes: 33006
  maxSessionTextBytes: 260007
report:
  maxFramedInputBytes: 920008
ai:
  maxResponseBodyBytes: 4200009
`)

	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got, err := loader.ContentLimits()
	if err != nil {
		t.Fatalf("ContentLimits: %v", err)
	}
	want := config.ContentLimits{
		HTTPMaxRequestBodyBytes:        11000001,
		ResumeUploadBytes:              11000002,
		PrivacyExportUploadBytes:       6000003,
		ResumeMaxActive:                11,
		ResumeMaxExtractedTextBytes:    400004,
		ResumeMaxPasteTextBytes:        400003,
		TargetJobMaxRawTextBytes:       100005,
		PracticeMaxMessageBytes:        33006,
		PracticeMaxSessionTextBytes:    260007,
		ReportMaxFramedInputBytes:      920008,
		AIProviderMaxResponseBodyBytes: 4200009,
	}
	if got != want {
		t.Fatalf("ContentLimits() = %+v, want %+v", got, want)
	}
}

func TestContentLimitsRejectExplicitNonPositiveValues(t *testing.T) {
	tests := map[string]string{
		"http.maxRequestBodyBytes":      "http:\n  maxRequestBodyBytes: %d\n",
		"upload.maxBytes.resume":        "upload:\n  maxBytes:\n    resume: %d\n",
		"upload.maxBytes.privacyExport": "upload:\n  maxBytes:\n    privacyExport: %d\n",
		"resume.maxActive":              "resume:\n  maxActive: %d\n",
		"resume.maxExtractedTextBytes":  "resume:\n  maxExtractedTextBytes: %d\n",
		"resume.maxPasteTextBytes":      "resume:\n  maxPasteTextBytes: %d\n",
		"targetJob.maxRawTextBytes":     "targetJob:\n  maxRawTextBytes: %d\n",
		"practice.maxMessageBytes":      "practice:\n  maxMessageBytes: %d\n",
		"practice.maxSessionTextBytes":  "practice:\n  maxSessionTextBytes: %d\n",
		"report.maxFramedInputBytes":    "report:\n  maxFramedInputBytes: %d\n",
		"ai.maxResponseBodyBytes":       "ai:\n  maxResponseBodyBytes: %d\n",
	}
	for path, inputFormat := range tests {
		for _, value := range []int{0, -1} {
			name := fmt.Sprintf("%s=%d", path, value)
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				writeYAML(t, filepath.Join(dir, "config.yaml"), fmt.Sprintf(inputFormat, value))
				loader, err := config.Load(config.Options{ConfigDir: dir})
				if err != nil {
					t.Fatalf("Load: %v", err)
				}
				if _, err := loader.ContentLimits(); err == nil ||
					!strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "must be positive") {
					t.Fatalf("ContentLimits error = %v, want %s positive-value failure", err, path)
				}
				if err := loader.Validate(); err == nil ||
					!strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "must be positive") {
					t.Fatalf("Validate error = %v, want %s positive-value failure", err, path)
				}
			})
		}
	}
}

func TestContentLimitsRejectInvalidCrossFieldCombinations(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"paste exceeds extracted": {
			input: `resume:
  maxExtractedTextBytes: 100
  maxPasteTextBytes: 101
`,
			want: "resume.maxPasteTextBytes must not exceed resume.maxExtractedTextBytes",
		},
		"message exceeds session": {
			input: `practice:
  maxMessageBytes: 101
  maxSessionTextBytes: 100
`,
			want: "practice.maxMessageBytes must not exceed practice.maxSessionTextBytes",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeYAML(t, filepath.Join(dir, "config.yaml"), tc.input)
			loader, err := config.Load(config.Options{ConfigDir: dir})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if _, err := loader.ContentLimits(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ContentLimits error = %v, want %q", err, tc.want)
			}
			if err := loader.Validate(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Validate error = %v, want %q", err, tc.want)
			}
		})
	}
}
