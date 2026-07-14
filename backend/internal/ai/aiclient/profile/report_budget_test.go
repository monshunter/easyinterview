package profile_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/profile"
)

const (
	reportInputByteLimit       = 917_504
	reportProviderFrameReserve = 2_048
)

type reportBoundaryManifest struct {
	SchemaVersion                   string                       `json:"schemaVersion"`
	ReportSchemaVersion             string                       `json:"reportSchemaVersion"`
	SerializerCommand               string                       `json:"serializerCommand"`
	SerializerVersion               string                       `json:"serializerVersion"`
	Files                           []reportBoundaryManifestFile `json:"files"`
	SemanticReachableFocusMax       int                          `json:"semanticReachableFocusMax"`
	SemanticReachableFocusMaxReason string                       `json:"semanticReachableFocusMaxReason"`
}

type reportBoundaryManifestFile struct {
	Name      string `json:"name"`
	ByteCount int    `json:"byteCount"`
	SHA256    string `json:"sha256"`
	Locale    string `json:"locale"`
	Purpose   string `json:"purpose"`
}

func TestReportProfileOfflineCapacityConsumesReviewBoundaryGate(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", "..", "..", ".."))
	backendRoot := filepath.Join(repoRoot, "backend")
	marker := exec.Command("go", "test", "./internal/review", "-run", "^TestReportBoundaryFixtures$", "-count=1", "-v")
	marker.Dir = backendRoot
	markerOutput, err := marker.CombinedOutput()
	if err != nil {
		t.Fatalf("review boundary marker command failed: %v\n%s", err, markerOutput)
	}
	if !bytes.Contains(markerOutput, []byte("REPORT_BOUNDARY_FIXTURES_READY")) {
		t.Fatalf("review boundary marker missing from focused owner test:\n%s", markerOutput)
	}

	loader, err := profile.NewLoader(profile.Options{
		Path:         filepath.Join(repoRoot, "config", "ai-profiles.yaml"),
		PollInterval: -1,
	})
	if err != nil {
		t.Fatalf("load tracked profile catalog: %v", err)
	}
	defer loader.Close()
	reportProfile, err := loader.Resolve("report.generate.default")
	if err != nil {
		t.Fatalf("resolve report profile: %v", err)
	}

	fixtureDir := filepath.Join(backendRoot, "internal", "review", "testdata", "report-boundary")
	manifestRaw, err := os.ReadFile(filepath.Join(fixtureDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read report boundary manifest: %v", err)
	}
	var manifest reportBoundaryManifest
	decoder := json.NewDecoder(bytes.NewReader(manifestRaw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatalf("decode report boundary manifest: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		t.Fatalf("report boundary manifest contains trailing JSON: %v", err)
	}
	if manifest.SchemaVersion != "report-boundary-fixtures.v1" || manifest.ReportSchemaVersion != "v0.2.0" || manifest.SerializerVersion != "report-boundary-serializer.v1" || strings.TrimSpace(manifest.SerializerCommand) == "" {
		t.Fatalf("report boundary manifest coordinate drift: %+v", manifest)
	}
	if manifest.SemanticReachableFocusMax != 4 || strings.TrimSpace(manifest.SemanticReachableFocusMaxReason) == "" {
		t.Fatalf("report boundary semantic maximum drift: %+v", manifest)
	}

	expected := map[string]struct {
		bytes   int
		locale  string
		purpose string
	}{
		"output-worst-case-en.json":    {3_651, "en", "current_direct_schema_semantic_worst_case"},
		"output-worst-case-zh-CN.json": {7_811, "zh-CN", "current_direct_schema_semantic_worst_case"},
	}
	if len(manifest.Files) != len(expected) {
		t.Fatalf("manifest file count=%d, want %d", len(manifest.Files), len(expected))
	}
	for _, entry := range manifest.Files {
		want, ok := expected[entry.Name]
		if !ok {
			t.Fatalf("unexpected report boundary fixture %q", entry.Name)
		}
		if entry.ByteCount != want.bytes || entry.Locale != want.locale || entry.Purpose != want.purpose {
			t.Fatalf("fixture manifest drift for %s: %+v", entry.Name, entry)
		}
		body, err := os.ReadFile(filepath.Join(fixtureDir, entry.Name))
		if err != nil {
			t.Fatalf("read fixture %s: %v", entry.Name, err)
		}
		if len(body) != entry.ByteCount {
			t.Fatalf("fixture %s bytes=%d, manifest=%d", entry.Name, len(body), entry.ByteCount)
		}
		sum := sha256.Sum256(body)
		if got := hex.EncodeToString(sum[:]); got != entry.SHA256 {
			t.Fatalf("fixture %s sha256=%s, manifest=%s", entry.Name, got, entry.SHA256)
		}
		if strings.HasPrefix(entry.Name, "output-") {
			assertCurrentReportOutputShape(t, entry.Name, body)
		}
	}

	conservativeReservedTokens := reportInputByteLimit + reportProviderFrameReserve + reportProfile.MaxTokens
	if conservativeReservedTokens >= reportProfile.ContextWindowTokens {
		t.Fatalf("offline capacity exceeded: input=%d frame=%d output=%d total=%d context=%d", reportInputByteLimit, reportProviderFrameReserve, reportProfile.MaxTokens, conservativeReservedTokens, reportProfile.ContextWindowTokens)
	}
	if reportProfile.RateLimit.TPM != 60_000 {
		t.Fatalf("TPM throughput drift=%d, want 60000", reportProfile.RateLimit.TPM)
	}
	fmt.Printf("REPORT_PROFILE_6144_PASS input_bytes=%d framing_reserve=%d output_tokens=%d context_window=%d tpm_throughput_only=%d live_usage=required_by_P0.100\n", reportInputByteLimit, reportProviderFrameReserve, reportProfile.MaxTokens, reportProfile.ContextWindowTokens, reportProfile.RateLimit.TPM)
}

func assertCurrentReportOutputShape(t *testing.T, name string, body []byte) {
	t.Helper()
	var value map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	expected := []string{"summary", "preparednessLevel", "dimensionAssessments", "highlights", "issues", "nextActions", "retryFocusDimensionCodes"}
	if len(value) != len(expected) {
		t.Fatalf("%s top-level key count=%d, want %d", name, len(value), len(expected))
	}
	for _, key := range expected {
		if _, ok := value[key]; !ok {
			t.Fatalf("%s missing current report field %q", name, key)
		}
	}
}

func TestValidateReportProfileLiveProbeResult(t *testing.T) {
	reportProfile := &aiclient.ModelProfile{
		ContextWindowTokens: 1_000_000,
		MaxTokens:           6_144,
	}
	validMeta := aiclient.AICallMeta{InputTokens: 12_000, OutputTokens: 500}

	tests := map[string]struct {
		profile        *aiclient.ModelProfile
		meta           aiclient.AICallMeta
		finishReason   string
		maxInputTokens int
		wantError      string
	}{
		"valid-exact-input":          {reportProfile, validMeta, "stop", 0, ""},
		"valid-output-probe":         {reportProfile, aiclient.AICallMeta{InputTokens: 6_144, OutputTokens: 1}, "stop", 6_144, ""},
		"missing-input-usage":        {reportProfile, aiclient.AICallMeta{OutputTokens: 1}, "stop", 0, "input token usage"},
		"missing-output-usage":       {reportProfile, aiclient.AICallMeta{InputTokens: 1}, "stop", 0, "output token usage"},
		"missing-finish-reason":      {reportProfile, validMeta, "", 0, "finish reason"},
		"length-finish":              {reportProfile, validMeta, "length", 0, "finish_reason=length"},
		"output-fixture-over-budget": {reportProfile, aiclient.AICallMeta{InputTokens: 6_145, OutputTokens: 1}, "stop", 6_144, "input token budget"},
		"context-capacity-exceeded":  {&aiclient.ModelProfile{ContextWindowTokens: 9_000, MaxTokens: 6_144}, aiclient.AICallMeta{InputTokens: 1_000, OutputTokens: 1}, "stop", 0, "context window"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateReportProfileLiveProbeResult(tc.profile, tc.meta, tc.finishReason, tc.maxInputTokens)
			if tc.wantError == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("error=%v, want substring %q", err, tc.wantError)
			}
		})
	}
}

func validateReportProfileLiveProbeResult(reportProfile *aiclient.ModelProfile, meta aiclient.AICallMeta, finishReason string, maxInputTokens int) error {
	if reportProfile == nil || reportProfile.ContextWindowTokens <= 0 || reportProfile.MaxTokens <= 0 {
		return fmt.Errorf("report profile budget is unavailable")
	}
	if meta.InputTokens <= 0 {
		return fmt.Errorf("live probe missing input token usage")
	}
	if meta.OutputTokens <= 0 {
		return fmt.Errorf("live probe missing output token usage")
	}
	finishReason = strings.TrimSpace(finishReason)
	if finishReason == "" {
		return fmt.Errorf("live probe missing finish reason")
	}
	if strings.EqualFold(finishReason, "length") {
		return fmt.Errorf("live probe returned finish_reason=length")
	}
	if maxInputTokens > 0 && meta.InputTokens > maxInputTokens {
		return fmt.Errorf("live probe input token budget exceeded: got %d, max %d", meta.InputTokens, maxInputTokens)
	}
	reserved := meta.InputTokens + reportProviderFrameReserve + reportProfile.MaxTokens
	if reserved >= reportProfile.ContextWindowTokens {
		return fmt.Errorf("live probe context window exceeded: reserved %d, context %d", reserved, reportProfile.ContextWindowTokens)
	}
	return nil
}
