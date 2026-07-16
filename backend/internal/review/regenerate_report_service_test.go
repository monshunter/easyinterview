package review

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestRegenerateReportServiceCreatesFreshSameReportAction(t *testing.T) {
	now := time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC)
	reportID := "019f6a6c-4ccb-768f-901f-54b1e126555b"
	repository := &regenerationRepositoryStub{}
	repository.regenerate = func(in RegenerateReportStoreInput) (RegenerateReportStoreResult, error) {
		return RegenerateReportStoreResult{
			ReportID: in.ReportID,
			Job: ReportJobRecord{
				ID: in.JobID, JobType: string(sharedjobs.JobTypeReportGenerate), ResourceType: "feedback_report",
				ResourceID: in.ReportID, Status: sharedtypes.JobStatusQueued, CreatedAt: in.Now, UpdatedAt: in.Now,
			},
		}, nil
	}
	ids := []string{"019f6a6c-4ccb-768f-901f-54b1e126555c", "019f6a6c-4ccb-768f-901f-54b1e126555d"}
	service := NewService(ServiceOptions{
		Repository: repository,
		Now:        func() time.Time { return now },
		NewID: func() string {
			id := ids[0]
			ids = ids[1:]
			return id
		},
	})

	result, err := service.RegenerateReport(context.Background(), RegenerateReportRequest{UserID: " user-1 ", ReportID: reportID})
	if err != nil {
		t.Fatalf("RegenerateReport: %v", err)
	}
	if repository.calls != 1 || repository.input.UserID != "user-1" || repository.input.ReportID != reportID || repository.input.JobID != "019f6a6c-4ccb-768f-901f-54b1e126555c" || repository.input.AuditEventID != "019f6a6c-4ccb-768f-901f-54b1e126555d" || !repository.input.Now.Equal(now) {
		t.Fatalf("store input=%+v calls=%d", repository.input, repository.calls)
	}
	if result.ReportID != reportID || result.Job.ResourceID != reportID || result.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("result=%+v", result)
	}
}

func TestRegenerateReportServicePropagatesTypedEligibilityFailures(t *testing.T) {
	reportID := "019f6a6c-4ccb-768f-901f-54b1e126555b"
	for _, want := range []error{ErrReportNotFound, ErrReportNotReady, ErrReportInvalidStateTransition, ErrReportContextTooLarge} {
		t.Run(want.Error(), func(t *testing.T) {
			repository := &regenerationRepositoryStub{regenerate: func(RegenerateReportStoreInput) (RegenerateReportStoreResult, error) {
				return RegenerateReportStoreResult{}, want
			}}
			service := NewService(ServiceOptions{Repository: repository})
			result, err := service.RegenerateReport(context.Background(), RegenerateReportRequest{UserID: "user-1", ReportID: reportID})
			if !errors.Is(err, want) || result.ReportID != "" || result.Job.ID != "" || repository.calls != 1 {
				t.Fatalf("result=%+v err=%v calls=%d want=%v", result, err, repository.calls, want)
			}
		})
	}
}

func TestRegenerateReportServiceHidesInvalidReportIdentityBeforeStore(t *testing.T) {
	repository := &regenerationRepositoryStub{}
	result, err := NewService(ServiceOptions{Repository: repository}).RegenerateReport(context.Background(), RegenerateReportRequest{UserID: "user-1", ReportID: "report-1"})
	if !errors.Is(err, ErrReportNotFound) || result.ReportID != "" || repository.calls != 0 {
		t.Fatalf("result=%+v err=%v calls=%d", result, err, repository.calls)
	}
}

type regenerationRepositoryStub struct {
	calls      int
	input      RegenerateReportStoreInput
	regenerate func(RegenerateReportStoreInput) (RegenerateReportStoreResult, error)
}

func (r *regenerationRepositoryStub) RegenerateFeedbackReport(_ context.Context, in RegenerateReportStoreInput) (RegenerateReportStoreResult, error) {
	r.calls++
	r.input = in
	if r.regenerate == nil {
		return RegenerateReportStoreResult{}, errors.New("unexpected regeneration call")
	}
	return r.regenerate(in)
}

func (*regenerationRepositoryStub) LoadReportContext(context.Context, string) (ReportContext, error) {
	return ReportContext{}, nil
}

func (*regenerationRepositoryStub) AssertCurrentReportJobLease(context.Context, string, int32) error {
	return nil
}

func (*regenerationRepositoryStub) PersistReportResult(context.Context, ReportResultPersistence) error {
	return nil
}

func (*regenerationRepositoryStub) PersistReportFailure(context.Context, ReportFailurePersistence) error {
	return nil
}

func TestReportRegenerationDomainContractExists(t *testing.T) {
	types, methods, source := parseReviewProductionContract(t)

	assertStructFields(t, types, "RegenerateReportRequest", "UserID", "ReportID")
	assertStructFields(t, types, "RegenerateReportResult", "ReportID", "Job")
	assertStructFields(t, types, "RegenerateReportStoreInput", "UserID", "ReportID", "JobID", "AuditEventID", "Now")
	assertStructFields(t, types, "RegenerateReportStoreResult", "ReportID", "Job")
	if _, ok := methods["Service.RegenerateReport"]; !ok {
		t.Fatal("review Service must expose RegenerateReport(context.Context, RegenerateReportRequest)")
	}
	for _, required := range []string{
		"RegenerateFeedbackReport",
		"RegenerateReportStoreInput",
		"ErrReportNotReady",
		"ErrReportInvalidStateTransition",
		"idx.RequireServerID",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("report regeneration domain contract missing %q", required)
		}
	}
}

func TestReportRegenerationServiceDoesNotCallProviderOrMutateTranscript(t *testing.T) {
	raw, err := os.ReadFile("regenerate_report.go")
	if err != nil {
		t.Fatalf("read regenerate_report.go: %v", err)
	}
	for _, forbidden := range []string{
		"aiclient", ".Complete(", "practice_messages", "generation_context =", "outbox_events",
	} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("report regeneration service must not contain %q", forbidden)
		}
	}
}

func parseReviewProductionContract(t *testing.T) (map[string]*ast.StructType, map[string]*ast.FuncDecl, string) {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	types := map[string]*ast.StructType{}
	methods := map[string]*ast.FuncDecl{}
	var source strings.Builder
	files := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatal(err)
		}
		source.Write(raw)
		source.WriteByte('\n')
		file, err := parser.ParseFile(files, name, raw, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range file.Decls {
			switch node := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range node.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						types[typeSpec.Name.Name] = structType
					}
				}
			case *ast.FuncDecl:
				if node.Recv == nil || len(node.Recv.List) != 1 {
					continue
				}
				receiver := receiverName(node.Recv.List[0].Type)
				methods[receiver+"."+node.Name.Name] = node
			}
		}
	}
	return types, methods, source.String()
}

func receiverName(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return receiverName(value.X)
	default:
		return ""
	}
}

func assertStructFields(t *testing.T, types map[string]*ast.StructType, typeName string, fields ...string) {
	t.Helper()
	structType, ok := types[typeName]
	if !ok {
		t.Fatalf("missing domain type %s", typeName)
	}
	present := map[string]bool{}
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			present[name.Name] = true
		}
	}
	for _, field := range fields {
		if !present[field] {
			t.Fatalf("%s missing field %s", typeName, field)
		}
	}
}
