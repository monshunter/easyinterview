package review

import (
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func computeReadinessTier(assessments []QuestionAssessmentDraft, rubric registry.RubricSchema) sharedtypes.ReadinessTier {
	if len(assessments) == 0 || len(rubric.Dimensions) == 0 {
		return sharedtypes.ReadinessTierNotReady
	}
	total := 0.0
	count := 0
	for _, assessment := range assessments {
		score, ok := weightedAssessmentScore(assessment, rubric)
		if !ok {
			continue
		}
		total += score
		count++
	}
	if count == 0 {
		return sharedtypes.ReadinessTierNotReady
	}
	return readinessTierFromScore(total / float64(count))
}

func weightedAssessmentScore(assessment QuestionAssessmentDraft, rubric registry.RubricSchema) (float64, bool) {
	total := 0.0
	weightTotal := 0.0
	for _, dim := range rubric.Dimensions {
		result, ok := assessment.DimensionResults[dim.Name]
		if !ok {
			continue
		}
		weight := dim.Weight
		if weight <= 0 {
			weight = 1
		}
		total += dimensionScoreValue(result) * weight
		weightTotal += weight
	}
	if weightTotal == 0 {
		return 0, false
	}
	return total / weightTotal, true
}

func dimensionScoreValue(result DimensionResultDraft) float64 {
	if result.Score > 0 {
		return clamp01(result.Score)
	}
	switch strings.ToLower(strings.TrimSpace(result.ScoreLevel)) {
	case "weak":
		return 0.2
	case "developing":
		return 0.5
	case "proficient":
		return 0.8
	case "strong":
		return 1.0
	}
	switch result.Status {
	case sharedtypes.DimensionStatusNeedsWork:
		return 0.2
	case sharedtypes.DimensionStatusMeetsBar:
		return 0.8
	case sharedtypes.DimensionStatusStrong:
		return 1.0
	default:
		return 0.0
	}
}

func dimensionStatusFromScoreLevel(level string) sharedtypes.DimensionStatus {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "weak", "developing", "needs_work":
		return sharedtypes.DimensionStatusNeedsWork
	case "proficient", "meets_bar":
		return sharedtypes.DimensionStatusMeetsBar
	case "strong":
		return sharedtypes.DimensionStatusStrong
	default:
		return sharedtypes.DimensionStatusNeedsWork
	}
}

func readinessTierFromScore(score float64) sharedtypes.ReadinessTier {
	switch {
	case score < 0.30:
		return sharedtypes.ReadinessTierNotReady
	case score < 0.55:
		return sharedtypes.ReadinessTierNeedsPractice
	case score < 0.75:
		return sharedtypes.ReadinessTierBasicallyReady
	default:
		return sharedtypes.ReadinessTierWellPrepared
	}
}

func validReadinessTier(tier sharedtypes.ReadinessTier) bool {
	switch tier {
	case sharedtypes.ReadinessTierNotReady,
		sharedtypes.ReadinessTierNeedsPractice,
		sharedtypes.ReadinessTierBasicallyReady,
		sharedtypes.ReadinessTierWellPrepared:
		return true
	default:
		return false
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
