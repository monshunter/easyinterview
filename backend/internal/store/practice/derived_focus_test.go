package practice

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestResolveDerivedSemanticFocusStripsAnchorsAndRequiresMeaning(t *testing.T) {
	dimensions := []byte(`[
  {"code":"system_design","label":"系统设计","status":"needs_work","confidence":"high"},
  {"code":"communication","label":"沟通表达","status":"meets_bar","confidence":"medium"}
]`)
	issues := []byte(`[
  {"dimensionCode":"system_design","evidence":"未说明容量估算","confidence":"high","sourceMessageSeqNos":[2]},
  {"dimensionCode":"system_design","evidence":"故障恢复取舍不完整","confidence":"medium","sourceMessageSeqNos":[4]}
]`)

	focus, err := resolveDerivedSemanticFocus([]string{"system_design"}, dimensions, issues)
	if err != nil {
		t.Fatalf("resolveDerivedSemanticFocus: %v", err)
	}
	if len(focus) != 1 || focus[0].Code != "system_design" || focus[0].Label != "系统设计" ||
		!reflect.DeepEqual(focus[0].Issues, []string{"未说明容量估算", "故障恢复取舍不完整"}) {
		t.Fatalf("semantic focus = %+v", focus)
	}
	raw, _ := json.Marshal(focus)
	for _, forbidden := range []string{"sourceMessageSeqNos", "rawTranscript", "2", "4"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("semantic focus leaked anchor/raw source %q: %s", forbidden, raw)
		}
	}
}

func TestResolveDerivedSemanticFocusEmptyDoesNotFabricateGuidance(t *testing.T) {
	focus, err := resolveDerivedSemanticFocus([]string{}, nil, nil)
	if err != nil {
		t.Fatalf("resolveDerivedSemanticFocus: %v", err)
	}
	if focus == nil || len(focus) != 0 {
		t.Fatalf("empty semantic focus must be an explicit empty array: %#v", focus)
	}
}

func TestResolveDerivedSemanticFocusFailsClosedOnCodeOnlyOrUnsupportedFocus(t *testing.T) {
	for _, tc := range []struct {
		name       string
		codes      []string
		dimensions string
		issues     string
	}{
		{name: "code only dimension", codes: []string{"system_design"}, dimensions: `[{"code":"system_design","label":"","status":"needs_work","confidence":"high"}]`, issues: `[{"dimensionCode":"system_design","evidence":"gap","confidence":"high","sourceMessageSeqNos":[2]}]`},
		{name: "missing issue", codes: []string{"system_design"}, dimensions: `[{"code":"system_design","label":"系统设计","status":"needs_work","confidence":"high"}]`, issues: `[]`},
		{name: "unsupported code", codes: []string{"delivery"}, dimensions: `[{"code":"system_design","label":"系统设计","status":"needs_work","confidence":"high"}]`, issues: `[{"dimensionCode":"system_design","evidence":"gap","confidence":"high","sourceMessageSeqNos":[2]}]`},
		{name: "raw transcript field", codes: []string{"system_design"}, dimensions: `[{"code":"system_design","label":"系统设计","status":"needs_work","confidence":"high"}]`, issues: `[{"dimensionCode":"system_design","evidence":"gap","confidence":"high","sourceMessageSeqNos":[2],"rawTranscript":"candidate answer"}]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if focus, err := resolveDerivedSemanticFocus(tc.codes, []byte(tc.dimensions), []byte(tc.issues)); err == nil {
				t.Fatalf("focus=%+v want fail closed", focus)
			}
		})
	}
}
