package review

import (
	"math/rand"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestComputeReadinessTier(t *testing.T) {
	rubric := registry.RubricSchema{Dimensions: []registry.RubricDimension{{Name: "depth", Weight: 1}}}
	for _, tc := range []struct {
		score float64
		want  sharedtypes.ReadinessTier
	}{
		{score: 0.29, want: sharedtypes.ReadinessTierNotReady},
		{score: 0.30, want: sharedtypes.ReadinessTierNeedsPractice},
		{score: 0.54, want: sharedtypes.ReadinessTierNeedsPractice},
		{score: 0.55, want: sharedtypes.ReadinessTierBasicallyReady},
		{score: 0.74, want: sharedtypes.ReadinessTierBasicallyReady},
		{score: 0.75, want: sharedtypes.ReadinessTierWellPrepared},
		{score: 0.99, want: sharedtypes.ReadinessTierWellPrepared},
	} {
		got := computeReadinessTier([]QuestionAssessmentDraft{{
			DimensionResults: map[string]DimensionResultDraft{"depth": {Score: tc.score}},
		}}, rubric)
		if got != tc.want {
			t.Fatalf("score %.2f tier = %s, want %s", tc.score, got, tc.want)
		}
	}
	if got := computeReadinessTier(nil, rubric); got != sharedtypes.ReadinessTierNotReady {
		t.Fatalf("empty tier = %s", got)
	}
}

func TestComputeReadinessTierScoreLevelsAndDimensionStatusMapping(t *testing.T) {
	rubric := registry.RubricSchema{Dimensions: []registry.RubricDimension{
		{Name: "depth", Weight: 0.6},
		{Name: "clarity", Weight: 0.4},
	}}
	got := computeReadinessTier([]QuestionAssessmentDraft{{
		DimensionResults: map[string]DimensionResultDraft{
			"depth":   {ScoreLevel: "developing"},
			"clarity": {Status: sharedtypes.DimensionStatusStrong},
		},
	}}, rubric)
	if got != sharedtypes.ReadinessTierBasicallyReady {
		t.Fatalf("tier = %s, want basically_ready", got)
	}
	if status := dimensionStatusFromScoreLevel("weak"); status != sharedtypes.DimensionStatusNeedsWork {
		t.Fatalf("weak maps to %s", status)
	}
	if status := dimensionStatusFromScoreLevel("proficient"); status != sharedtypes.DimensionStatusMeetsBar {
		t.Fatalf("proficient maps to %s", status)
	}
	if status := dimensionStatusFromScoreLevel("strong"); status != sharedtypes.DimensionStatusStrong {
		t.Fatalf("strong maps to %s", status)
	}
}

func TestComputeReadinessTierPropertyRandomDimensions(t *testing.T) {
	rng := rand.New(rand.NewSource(15))
	validTiers := map[sharedtypes.ReadinessTier]struct{}{
		sharedtypes.ReadinessTierNotReady:       {},
		sharedtypes.ReadinessTierNeedsPractice:  {},
		sharedtypes.ReadinessTierBasicallyReady: {},
		sharedtypes.ReadinessTierWellPrepared:   {},
	}
	for i := 0; i < 100; i++ {
		dims := make([]registry.RubricDimension, 0, 5)
		results := map[string]DimensionResultDraft{}
		for j := 0; j < 5; j++ {
			name := string(rune('a' + j))
			dims = append(dims, registry.RubricDimension{Name: name, Weight: rng.Float64() + 0.1})
			results[name] = DimensionResultDraft{Score: rng.Float64()}
		}
		tier := computeReadinessTier([]QuestionAssessmentDraft{{DimensionResults: results}}, registry.RubricSchema{Dimensions: dims})
		if _, ok := validTiers[tier]; !ok {
			t.Fatalf("invalid tier %q", tier)
		}
	}
}
