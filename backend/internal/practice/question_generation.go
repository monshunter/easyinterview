package practice

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type questionGenerationKind string

const (
	questionGenerationFollowUp     questionGenerationKind = "follow_up"
	questionGenerationNextQuestion questionGenerationKind = "next_question"
)

type questionAttemptMode string

const (
	questionAttemptInitial questionAttemptMode = "initial"
	questionAttemptRepair  questionAttemptMode = "repair"
)

type questionTemplateData struct {
	Language            string
	PracticeGoal        string
	PracticeMode        string
	TurnStatus          string
	TargetJobID         string
	GenerationKind      questionGenerationKind
	AttemptMode         questionAttemptMode
	LastQuestion        string
	QuestionIntent      string
	LastAnswer          string
	FollowUpCount       int
	CoveredDimensions   []string
	RemainingDimensions []string
	CommittedContext    string
}

type questionGenerationRequest struct {
	Resolution   registry.PromptResolution
	TemplateData questionTemplateData
	TaskRun      aiclient.AITaskRunContext
}

var unresolvedQuestionMarker = regexp.MustCompile(`\{\{[a-zA-Z0-9_.-]+\}\}`)

func renderQuestionTemplate(template string, data questionTemplateData) (string, error) {
	replacements := map[string]string{
		"{{language}}":             strings.TrimSpace(data.Language),
		"{{practice_goal}}":        strings.TrimSpace(data.PracticeGoal),
		"{{practice_mode}}":        strings.TrimSpace(data.PracticeMode),
		"{{turn_status}}":          strings.TrimSpace(data.TurnStatus),
		"{{target_job_id}}":        strings.TrimSpace(data.TargetJobID),
		"{{generation_kind}}":      strings.TrimSpace(string(data.GenerationKind)),
		"{{attempt_mode}}":         strings.TrimSpace(string(data.AttemptMode)),
		"{{last_question}}":        strings.TrimSpace(data.LastQuestion),
		"{{question_intent}}":      strings.TrimSpace(data.QuestionIntent),
		"{{last_answer}}":          strings.TrimSpace(data.LastAnswer),
		"{{follow_up_count}}":      strconv.Itoa(data.FollowUpCount),
		"{{covered_dimensions}}":   joinQuestionDimensions(data.CoveredDimensions),
		"{{remaining_dimensions}}": joinQuestionDimensions(data.RemainingDimensions),
		"{{committed_context}}":    strings.TrimSpace(data.CommittedContext),
	}
	unresolved := ""
	rendered := unresolvedQuestionMarker.ReplaceAllStringFunc(template, func(marker string) string {
		value, ok := replacements[marker]
		if !ok {
			unresolved = marker
			return marker
		}
		return value
	})
	if unresolved != "" {
		return "", sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "question prompt contains an unresolved marker", false)
	}
	return strings.TrimSpace(rendered), nil
}

func joinQuestionDimensions(values []string) string {
	clean := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			clean = append(clean, value)
		}
	}
	return strings.Join(clean, ", ")
}

func validateGeneratedQuestionLanguage(text, language string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "generated question is empty", false)
	}
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(language), "_", "-"))
	hanCount := 0
	latinCount := 0
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			hanCount++
		}
		if unicode.Is(unicode.Latin, r) {
			latinCount++
		}
	}
	switch {
	case normalized == "zh" || strings.HasPrefix(normalized, "zh-"):
		// Keep ordinary technical names such as RAG, API, K8s, and React inside
		// Chinese copy, while rejecting sentences padded with nontechnical English.
		if hanCount > 0 && hanCount*5 >= (hanCount+latinCount)*3 {
			return nil
		}
	case normalized == "en" || strings.HasPrefix(normalized, "en-"):
		if latinCount > 0 && hanCount == 0 {
			return nil
		}
	default:
		return sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "practice session language is unsupported", false)
	}
	return sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "generated question language does not match the session", false)
}

func (s *Service) generateQuestion(ctx context.Context, request questionGenerationRequest) (firstQuestion, aiclient.AICallMeta, error) {
	if s == nil || s.ai == nil {
		return firstQuestion{}, aiclient.AICallMeta{}, fmt.Errorf("practice question generator is not initialised")
	}
	for attempt := 0; attempt < 2; attempt++ {
		data := request.TemplateData
		if attempt == 0 {
			data.AttemptMode = questionAttemptInitial
		} else {
			data.AttemptMode = questionAttemptRepair
		}
		userContent, err := renderQuestionTemplate(request.Resolution.UserMessageTemplate, data)
		if err != nil {
			return firstQuestion{}, aiclient.AICallMeta{}, err
		}
		messages := make([]aiclient.Message, 0, 2)
		if system := strings.TrimSpace(request.Resolution.SystemMessage); system != "" {
			messages = append(messages, aiclient.Message{Role: "system", Content: system})
		}
		messages = append(messages, aiclient.Message{Role: "user", Content: userContent})
		featureKey := strings.TrimSpace(request.Resolution.FeatureKey)
		if featureKey == "" {
			featureKey = followUpFeatureKey
		}
		payload := aiclient.CompletePayload{
			Messages: messages,
			Metadata: attachOutputSchema(aiclient.CallMetadata{
				FeatureKey:        featureKey,
				PromptVersion:     request.Resolution.PromptVersion,
				RubricVersion:     request.Resolution.RubricVersion,
				Language:          data.Language,
				FeatureFlag:       request.Resolution.FeatureFlag,
				DataSourceVersion: request.Resolution.DataSourceVersion,
				TaskRun:           request.TaskRun,
			}, request.Resolution),
		}
		response, meta, err := s.ai.Complete(ctx, request.Resolution.ModelProfileName, payload)
		if err == nil {
			var question firstQuestion
			question, err = parseFirstQuestion(response.Content)
			if err == nil {
				err = validateGeneratedQuestionLanguage(question.Text, data.Language)
			}
			if err == nil {
				return question, meta, nil
			}
		}
		if attempt == 0 && isRepairableQuestionError(err) {
			continue
		}
		return firstQuestion{}, meta, err
	}
	return firstQuestion{}, aiclient.AICallMeta{}, sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "question generation failed", false)
}

func isRepairableQuestionError(err error) bool {
	code, ok := aiErrorCode(err)
	return ok && code == sharederrors.CodeAiOutputInvalid
}
