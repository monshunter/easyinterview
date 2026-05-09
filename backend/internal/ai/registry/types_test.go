package registry

import (
	"context"
	"reflect"
	"testing"
)

// TestTypeShape freezes the struct field set and order for downstream
// consumers (RegistryAdapter, future plan 002 Judge implementations).
// Adding a field is fine; reordering or removing requires a spec revision.
func TestTypeShape(t *testing.T) {
	t.Parallel()

	checks := []struct {
		name  string
		typ   reflect.Type
		want  []string
	}{
		{
			name: "PromptResolution",
			typ:  reflect.TypeOf(PromptResolution{}),
			want: []string{
				"FeatureKey",
				"PromptVersion",
				"RubricVersion",
				"ModelProfileName",
				"DataSourceVersion",
				"FeatureFlag",
				"SystemMessage",
				"UserMessageTemplate",
				"Tools",
				"OutputSchema",
				"StreamWire",
			},
		},
		{
			name: "PromptMeta",
			typ:  reflect.TypeOf(PromptMeta{}),
			want: []string{"FeatureKey", "Version", "Language", "TemplateHash", "Status", "CreatedAt"},
		},
		{
			name: "RubricDimension",
			typ:  reflect.TypeOf(RubricDimension{}),
			want: []string{"Name", "Weight", "Description", "ScoreLevels"},
		},
		{
			name: "ScoreLevel",
			typ:  reflect.TypeOf(ScoreLevel{}),
			want: []string{"Label", "Threshold", "Description"},
		},
	}

	for _, c := range checks {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := make([]string, c.typ.NumField())
			for i := range got {
				got[i] = c.typ.Field(i).Name
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("%s field order drift: got %v, want %v", c.name, got, c.want)
			}
		})
	}
}

// stubJudge is a local test placeholder so TestJudgeSignature does not
// depend on judge.go (which lands at plan item 2.6). Once judge.go ships
// the production NotImplementedJudge, judge_test.go re-exercises the
// interface contract with the real implementation.
type stubJudge struct{}

func (stubJudge) Judge(
	_ context.Context,
	_ string,
	_ string,
	_ []byte,
	_ string,
) (Score, Reasoning, error) {
	return Score{}, Reasoning{}, nil
}

// TestJudgeSignature freezes the LLM Judge interface signature against
// spec §3.1 D-9. The exact Go idiom adds ctx and trailing error to the
// spec's logical signature; the spec parameter order
// (featureKey, prompt_version, output, rubric_version) and the (score,
// reasoning) result tuple must remain in the same order.
func TestJudgeSignature(t *testing.T) {
	t.Parallel()

	// Reflect on the Judge interface type so receiver is not counted as
	// an input (which would happen with a concrete type).
	interfaceType := reflect.TypeOf((*Judge)(nil)).Elem()
	method, ok := interfaceType.MethodByName("Judge")
	if !ok {
		t.Fatalf("Judge interface missing Judge method")
	}
	// Compile-time assertion that stubJudge satisfies Judge (and so does
	// the production NotImplementedJudge once judge.go lands at item 2.6).
	var _ Judge = stubJudge{}

	in := method.Type
	// reflect on an interface method uses the method type directly: no
	// receiver. Inputs: ctx, featureKey, promptVersion, output, rubricVersion.
	if in.NumIn() != 5 {
		t.Fatalf("Judge.Judge expected 5 inputs, got %d", in.NumIn())
	}
	wantInKinds := []reflect.Kind{reflect.Interface, reflect.String, reflect.String, reflect.Slice, reflect.String}
	for i, want := range wantInKinds {
		if got := in.In(i).Kind(); got != want {
			t.Fatalf("Judge.Judge input %d kind: got %v, want %v", i, got, want)
		}
	}
	// First input must be context.Context.
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !in.In(0).Implements(ctxType) {
		t.Fatalf("Judge.Judge input 0 must implement context.Context")
	}
	// Output []byte (kind Slice with elem Uint8).
	if elem := in.In(3).Elem(); elem.Kind() != reflect.Uint8 {
		t.Fatalf("Judge.Judge input 3 (output) must be []byte, got []%v", elem.Kind())
	}

	// Outputs: Score, Reasoning, error.
	if in.NumOut() != 3 {
		t.Fatalf("Judge.Judge expected 3 outputs, got %d", in.NumOut())
	}
	wantOutNames := []string{"Score", "Reasoning"}
	for i, want := range wantOutNames {
		if got := in.Out(i).Name(); got != want {
			t.Fatalf("Judge.Judge output %d name: got %q, want %q", i, got, want)
		}
	}
	if !in.Out(2).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		t.Fatalf("Judge.Judge output 2 must implement error")
	}
}
