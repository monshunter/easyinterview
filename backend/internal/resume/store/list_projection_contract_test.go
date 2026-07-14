package store

import (
	"reflect"
	"strings"
	"testing"
)

func TestResumeSummaryProjectionKeepsOnlyScalarResultFields(t *testing.T) {
	typeOfRecord := reflect.TypeOf(ResumeSummaryRecord{})
	want := []string{
		"ID", "Title", "DisplayName", "Language", "SourceType", "ParseStatus",
		"SummaryHeadline", "HasReadableContent", "UpdatedAt",
	}
	if typeOfRecord.NumField() != len(want) {
		t.Fatalf("ResumeSummaryRecord field count = %d, want %d", typeOfRecord.NumField(), len(want))
	}
	for index, fieldName := range want {
		if got := typeOfRecord.Field(index).Name; got != fieldName {
			t.Fatalf("ResumeSummaryRecord field[%d] = %q, want %q", index, got, fieldName)
		}
	}
}

func TestResumeSummaryProjectionSQLLocksHeadlinePriorityAndReadablePredicates(t *testing.T) {
	headlinePaths := []string{
		"parsed_summary->'headline'",
		"parsed_summary->'basics'->'headline'",
		"structured_profile->'headline'",
		"structured_profile->'basics'->'headline'",
	}
	previous := -1
	for _, path := range headlinePaths {
		position := strings.Index(resumeSummarySelectColumns, "jsonb_typeof("+path+") = 'string'")
		if position < 0 {
			t.Fatalf("summary projection missing JSON string type gate for %s", path)
		}
		if position <= previous {
			t.Fatalf("summary headline priority is out of order at %s", path)
		}
		previous = position
	}

	readableStart := strings.Index(resumeSummarySelectColumns, "end as summary_headline")
	readableEnd := strings.Index(resumeSummarySelectColumns, "as has_readable_content")
	if readableStart < 0 || readableEnd <= readableStart {
		t.Fatal("summary projection missing readable-content expression")
	}
	readable := resumeSummarySelectColumns[readableStart:readableEnd]
	for _, required := range []string{
		"nullif(btrim(parsed_text_snapshot",
		"nullif(btrim(original_text",
		"jsonb_typeof(structured_profile) = 'object'",
		"structured_profile <> '{}'::jsonb",
	} {
		if !strings.Contains(readable, required) {
			t.Fatalf("readable-content expression missing %q", required)
		}
	}
	for _, forbidden := range []string{"file_object_id", "source_type", "parse_status"} {
		if strings.Contains(readable, forbidden) {
			t.Fatalf("readable-content expression infers from forbidden %q", forbidden)
		}
	}
}
