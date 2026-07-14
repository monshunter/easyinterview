package review

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/outputschema"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const (
	reportBoundaryManifestSchemaVersion = "report-boundary-fixtures.v1"
	reportBoundarySerializerVersion     = "report-boundary-serializer.v1"
	reportBoundarySerializerCommand     = "cd backend && UPDATE_REPORT_BOUNDARY_FIXTURES=1 go test ./internal/review -run '^TestReportBoundaryFixtures$' -count=1"
	reportBoundaryUpdateEnv             = "UPDATE_REPORT_BOUNDARY_FIXTURES"
)

type reportBoundaryManifest struct {
	SchemaVersion                   string                       `json:"schemaVersion"`
	ReportSchemaVersion             string                       `json:"reportSchemaVersion"`
	SerializerCommand               string                       `json:"serializerCommand"`
	SerializerVersion               string                       `json:"serializerVersion"`
	SemanticReachableFocusMax       int                          `json:"semanticReachableFocusMax"`
	SemanticReachableFocusMaxReason string                       `json:"semanticReachableFocusMaxReason"`
	Files                           []reportBoundaryManifestFile `json:"files"`
}

type reportBoundaryManifestFile struct {
	Name      string `json:"name"`
	ByteCount int    `json:"byteCount"`
	SHA256    string `json:"sha256"`
	Locale    string `json:"locale"`
	Purpose   string `json:"purpose"`
}

func TestReportBoundaryFixtures(t *testing.T) {
	expectedFiles := reconstructReportBoundaryFixtures(t)
	expectedManifest := buildReportBoundaryManifest(expectedFiles)

	if os.Getenv(reportBoundaryUpdateEnv) == "1" {
		writeReportBoundaryFixtures(t, expectedFiles, expectedManifest)
	}

	fixtureDir := reportBoundaryFixtureDir(t)
	manifestRaw, err := os.ReadFile(filepath.Join(fixtureDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read committed report boundary manifest: %v", err)
	}
	var committedManifest reportBoundaryManifest
	if err := decodeClosedJSONValue(manifestRaw, &committedManifest); err != nil {
		t.Fatalf("decode committed report boundary manifest: %v", err)
	}
	if committedManifest.SchemaVersion != reportBoundaryManifestSchemaVersion ||
		committedManifest.ReportSchemaVersion != reportGeneratePromptVersion ||
		committedManifest.SerializerCommand != reportBoundarySerializerCommand ||
		committedManifest.SerializerVersion != reportBoundarySerializerVersion {
		t.Fatalf("manifest coordinate drift: %+v", committedManifest)
	}
	if committedManifest.SemanticReachableFocusMax != 4 || committedManifest.SemanticReachableFocusMaxReason == "" {
		t.Fatalf("manifest must record the issue-backed semantic focus maximum: %+v", committedManifest)
	}

	expectedManifestRaw := marshalReportBoundaryManifest(t, expectedManifest)
	if !bytes.Equal(manifestRaw, expectedManifestRaw) {
		t.Fatalf("manifest drift; regenerate with %q", reportBoundarySerializerCommand)
	}
	if len(committedManifest.Files) != len(expectedFiles) {
		t.Fatalf("manifest file count = %d, want %d", len(committedManifest.Files), len(expectedFiles))
	}
	directoryEntries, err := os.ReadDir(fixtureDir)
	if err != nil {
		t.Fatalf("read report boundary fixture directory: %v", err)
	}
	if len(directoryEntries) != len(expectedFiles)+1 {
		t.Fatalf("fixture directory entry count = %d, want %d committed fixtures plus manifest", len(directoryEntries), len(expectedFiles))
	}

	for _, entry := range committedManifest.Files {
		expected, ok := expectedFiles[entry.Name]
		if !ok {
			t.Fatalf("manifest contains unexpected fixture %q", entry.Name)
		}
		committed, err := os.ReadFile(filepath.Join(fixtureDir, entry.Name))
		if err != nil {
			t.Fatalf("read committed fixture %s: %v", entry.Name, err)
		}
		if !bytes.Equal(committed, expected) {
			t.Fatalf("fixture %s does not reconstruct byte-identically; regenerate with %q", entry.Name, reportBoundarySerializerCommand)
		}
		if got := len(committed); got != entry.ByteCount {
			t.Fatalf("fixture %s byte count = %d, manifest = %d", entry.Name, got, entry.ByteCount)
		}
		if got := reportBoundarySHA256(committed); got != entry.SHA256 {
			t.Fatalf("fixture %s SHA-256 = %s, manifest = %s", entry.Name, got, entry.SHA256)
		}
		if !utf8.Valid(committed) {
			t.Fatalf("fixture %s is not valid UTF-8", entry.Name)
		}
	}

	verifyReportBoundaryOutput(t, filepath.Join(fixtureDir, "output-worst-case-en.json"), "en")
	verifyReportBoundaryOutput(t, filepath.Join(fixtureDir, "output-worst-case-zh-CN.json"), "zh-CN")

}

func reconstructReportBoundaryFixtures(t *testing.T) map[string][]byte {
	t.Helper()
	return map[string][]byte{
		"output-worst-case-en.json":    marshalReportBoundaryOutput(t, worstCaseReportBoundaryOutput("en")),
		"output-worst-case-zh-CN.json": marshalReportBoundaryOutput(t, worstCaseReportBoundaryOutput("zh-CN")),
	}
}

func reportBoundaryResolution(t *testing.T) registry.PromptResolution {
	t.Helper()
	root := reportBoundaryRepoRoot(t)
	template, err := os.ReadFile(filepath.Join(root, "config", "prompts", "report.generate", "v0.2.0.md"))
	if err != nil {
		t.Fatalf("read current report prompt: %v", err)
	}
	schema := mustReadReportBoundaryFile(t, filepath.Join(root, "config", "prompts", "report.generate", "v0.2.0.schema.json"))
	rawSchema := json.RawMessage(append([]byte(nil), schema...))
	return registry.PromptResolution{
		FeatureKey:          reportGenerateFeatureKey,
		PromptVersion:       reportGeneratePromptVersion,
		RubricVersion:       reportGenerateRubricVersion,
		ModelProfileName:    "report.generate.default",
		FeatureFlag:         "none",
		DataSourceVersion:   practicedomain.ReportContextSchemaVersion,
		OutputSchema:        &rawSchema,
		UserMessageTemplate: string(template),
	}
}

func reportBoundaryContext(t *testing.T, language string) ReportContext {
	t.Helper()
	rounds := []practicedomain.ReportRoundSnapshot{
		{ID: "round-1-technical", Sequence: 1, Type: "technical", Name: "Technical", Focus: "system design", DurationMinutes: 45},
		{ID: "round-2-manager", Sequence: 2, Type: "manager", Name: "Manager", Focus: "leadership", DurationMinutes: 45},
	}
	targetSummary := fmt.Sprintf(`{"interviewRounds":[{"sequence":1,"type":"technical","name":"Technical","focus":"system design","durationMinutes":45},{"sequence":2,"type":"manager","name":"Manager","focus":"leadership","durationMinutes":45}],"provenance":{"promptVersion":"v0.2.0","rubricVersion":"v0.2.0","modelId":"synthetic-model","language":%q,"dataSourceVersion":"target-summary.v1"}}`, language)
	snapshot := practicedomain.ReportContextSnapshot{
		SchemaVersion: practicedomain.ReportContextSchemaVersion,
		TargetJob: practicedomain.ReportTargetJobSnapshot{
			ID: "00000000-0000-4000-8000-000000000005", Title: "Synthetic Platform Engineer", Company: "Synthetic Example", Language: language,
			RawJD:        "Synthetic non-sensitive role context.",
			Summary:      json.RawMessage(targetSummary),
			Requirements: []practicedomain.ReportRequirementSnapshot{{Kind: "must_have", Label: "Synthetic reliability requirement", EvidenceLevel: "explicit", DisplayOrder: 1}},
		},
		Resume: practicedomain.ReportResumeSnapshot{
			ID: "00000000-0000-4000-8000-000000000006", DisplayName: "Synthetic resume", Language: language,
			SourceSnapshot: "Synthetic candidate built reliable queue systems.", StructuredProfile: json.RawMessage(`{"skills":["Go","PostgreSQL"]}`),
		},
		Round: rounds[0], CanonicalRounds: rounds,
		Plan: practicedomain.ReportPlanSnapshot{
			ID: "00000000-0000-4000-8000-000000000004", Goal: "baseline", InterviewerPersona: "hiring_manager", Difficulty: "medium", Language: language,
			TimeBudgetMinutes: 45, ResumeID: "00000000-0000-4000-8000-000000000006", RoundID: rounds[0].ID, RoundSequence: rounds[0].Sequence,
		},
		Conversation: practicedomain.ReportConversationCoordinate{SessionID: "00000000-0000-4000-8000-000000000003", Language: language, MessageCount: 3, LastMessageSeqNo: 3},
		HasNextRound: true,
	}
	if err := practicedomain.ValidateReportContextSnapshot(snapshot); err != nil {
		t.Fatalf("synthetic report boundary context is invalid: %v", err)
	}
	return ReportContext{
		FrozenContext: snapshot,
		Session: SessionSnapshot{
			UserID: "00000000-0000-4000-8000-000000000001", ReportID: "00000000-0000-4000-8000-000000000002", SessionID: snapshot.Conversation.SessionID,
			TargetJobID: snapshot.TargetJob.ID, Language: language,
		},
		Messages: reportBoundaryMessages(language),
	}
}

func worstCaseReportBoundaryOutput(language string) ReportContentDraft {
	dimensions := make([]DimensionAssessmentDraft, 0, 6)
	highlights := make([]ReportEvidenceDraft, 0, 2)
	issues := make([]ReportEvidenceDraft, 0, 4)
	focus := make([]string, 0, 4)
	for index := 0; index < 6; index++ {
		code := fmt.Sprintf("d%d", index)
		status := sharedtypes.DimensionStatusStrong
		if index >= 2 {
			status = sharedtypes.DimensionStatusNeedsWork
			focus = append(focus, code)
		}
		labelRune := "l"
		evidenceRune := string(rune('a' + index))
		if language == "zh-CN" {
			labelRune = "维"
			if index < 2 {
				evidenceRune = "优"
			} else {
				evidenceRune = "缺"
			}
		}
		dimensions = append(dimensions, DimensionAssessmentDraft{Code: code, Label: strings.Repeat(labelRune, 48), Status: status, Confidence: sharedtypes.ConfidenceMedium})
		evidence := ReportEvidenceDraft{DimensionCode: code, Evidence: strings.Repeat(evidenceRune, 240), Confidence: sharedtypes.ConfidenceMedium, SourceMessageSeqNos: []int32{2}}
		if index < 2 {
			highlights = append(highlights, evidence)
		} else {
			issues = append(issues, evidence)
		}
	}

	summaryRune, retryRune, reviewRune := "s", "r", "v"
	actionLabelLength := reportActionLabelSchemaRuneLimit
	if language == "zh-CN" {
		summaryRune, retryRune, reviewRune = "总", "练", "看"
		actionLabelLength = reportActionLabelChineseRuneLimit
	}
	return ReportContentDraft{
		Summary:              strings.Repeat(summaryRune, 360),
		PreparednessLevel:    sharedtypes.ReadinessTierNeedsPractice,
		DimensionAssessments: dimensions,
		Highlights:           highlights,
		Issues:               issues,
		NextActions: []ReportNextActionDraft{
			{Type: string(NextActionRetryCurrentRound), Label: strings.Repeat(retryRune, actionLabelLength)},
			{Type: string(NextActionReviewEvidence), Label: strings.Repeat(reviewRune, actionLabelLength)},
		},
		RetryFocusDimensionCodes: focus,
	}
}

func marshalReportBoundaryOutput(t *testing.T, content ReportContentDraft) []byte {
	t.Helper()
	raw, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("marshal report boundary output: %v", err)
	}
	return raw
}

func buildReportBoundaryManifest(files map[string][]byte) reportBoundaryManifest {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	entries := make([]reportBoundaryManifestFile, 0, len(names))
	for _, name := range names {
		locale := "en"
		purpose := "current_direct_schema_semantic_worst_case"
		switch name {
		case "output-worst-case-zh-CN.json":
			locale = "zh-CN"
		}
		entries = append(entries, reportBoundaryManifestFile{Name: name, ByteCount: len(files[name]), SHA256: reportBoundarySHA256(files[name]), Locale: locale, Purpose: purpose})
	}
	return reportBoundaryManifest{
		SchemaVersion: reportBoundaryManifestSchemaVersion, ReportSchemaVersion: reportGeneratePromptVersion,
		SerializerCommand: reportBoundarySerializerCommand, SerializerVersion: reportBoundarySerializerVersion,
		SemanticReachableFocusMax:       4,
		SemanticReachableFocusMaxReason: "retry focus must be unique and issue-backed while issues are capped at four",
		Files:                           entries,
	}
}

func verifyReportBoundaryOutput(t *testing.T, path, language string) {
	t.Helper()
	raw := mustReadReportBoundaryFile(t, path)
	schema := mustReadReportBoundaryFile(t, filepath.Join(reportBoundaryRepoRoot(t), "config", "prompts", "report.generate", "v0.2.0.schema.json"))
	if err := outputschema.Validate(json.RawMessage(schema), string(raw)); err != nil {
		t.Fatalf("boundary output %s violates current report schema: %v", filepath.Base(path), err)
	}
	content, decodeIssues := decodeReportContent(raw)
	if len(decodeIssues) != 0 {
		t.Fatalf("boundary output %s direct decode issues: %#v", filepath.Base(path), decodeIssues)
	}
	reportCtx := reportBoundaryContext(t, language)
	if issues := validateReportContent(content, reportCtx.FrozenContext, reportCtx.Messages); len(issues) != 0 {
		t.Fatalf("boundary output %s business validation issues: %#v", filepath.Base(path), issues)
	}
	if len([]rune(content.Summary)) != 360 || len(content.DimensionAssessments) != 6 || len(content.Highlights) != 2 || len(content.Issues) != 4 || len(content.Highlights)+len(content.Issues) != 6 || len(content.NextActions) != 2 || len(content.RetryFocusDimensionCodes) != 4 {
		t.Fatalf("boundary output %s collection/rune bounds drifted", filepath.Base(path))
	}
	for index, dimension := range content.DimensionAssessments {
		if len([]rune(dimension.Label)) != 48 {
			t.Fatalf("boundary output %s dimension %d label = %d runes", filepath.Base(path), index, len([]rune(dimension.Label)))
		}
	}
	for _, evidenceItems := range [][]ReportEvidenceDraft{content.Highlights, content.Issues} {
		for index, evidence := range evidenceItems {
			if len([]rune(evidence.Evidence)) != 240 {
				t.Fatalf("boundary output %s evidence %d = %d runes", filepath.Base(path), index, len([]rune(evidence.Evidence)))
			}
		}
	}
	actionLabelLength := reportActionLabelSchemaRuneLimit
	if language == "zh-CN" {
		actionLabelLength = reportActionLabelChineseRuneLimit
	}
	for index, action := range content.NextActions {
		if len([]rune(action.Label)) != actionLabelLength {
			t.Fatalf("boundary output %s action %d label = %d runes", filepath.Base(path), index, len([]rune(action.Label)))
		}
	}
}

func writeReportBoundaryFixtures(t *testing.T, files map[string][]byte, manifest reportBoundaryManifest) {
	t.Helper()
	dir := reportBoundaryFixtureDir(t)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create report boundary fixture directory: %v", err)
	}
	for name, raw := range files {
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0o644); err != nil {
			t.Fatalf("write report boundary fixture %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), marshalReportBoundaryManifest(t, manifest), 0o644); err != nil {
		t.Fatalf("write report boundary manifest: %v", err)
	}
}

func marshalReportBoundaryManifest(t *testing.T, manifest reportBoundaryManifest) []byte {
	t.Helper()
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		t.Fatalf("marshal report boundary manifest: %v", err)
	}
	return output.Bytes()
}

func reportBoundarySHA256(raw []byte) string {
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:])
}

func reportBoundaryFixtureDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve report boundary test source path")
	}
	return filepath.Join(filepath.Dir(file), "testdata", "report-boundary")
}

func reportBoundaryRepoRoot(t *testing.T) string {
	t.Helper()
	return filepath.Clean(filepath.Join(reportBoundaryFixtureDir(t), "..", "..", "..", "..", ".."))
}

func mustReadReportBoundaryFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return raw
}

func reportBoundaryMessages(language string) []MessageSnapshot {
	if language == "zh-CN" {
		return []MessageSnapshot{
			{Role: "assistant", Content: "请说明一个合成的系统设计方案。", SeqNo: 1},
			{Role: "user", Content: "我会先灰度发布，监控错误率，并在超过阈值时回滚。", SeqNo: 2},
			{Role: "assistant", Content: "你会怎样确定阈值？", SeqNo: 3},
		}
	}
	return []MessageSnapshot{
		{Role: "assistant", Content: "Describe a synthetic system design approach.", SeqNo: 1},
		{Role: "user", Content: "I would stage the rollout, monitor errors, and roll back above a threshold.", SeqNo: 2},
		{Role: "assistant", Content: "How would you define that threshold?", SeqNo: 3},
	}
}
