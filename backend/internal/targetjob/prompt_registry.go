package targetjob

import (
	"context"
	"strings"
)

const (
	defaultTargetImportPromptVersion     = "targetjob.import.prompt.v0"
	defaultTargetImportRubricVersion     = "targetjob.import.rubric.v0"
	defaultTargetImportModelProfileName  = "target.import.default"
	defaultTargetImportDataSourceVersion = "targetjob.import.v1"
)

// StaticPromptRegistry is the cmd/api bootstrap bridge for the F3-owned
// prompt registry. It exposes only the spec-locked target.import.parse tuple
// until the F3 runtime package lands as a shared dependency.
type StaticPromptRegistry struct {
	Resolution PromptResolution
}

// NewStaticPromptRegistry returns the target.import.parse prompt resolution
// currently required by backend-targetjob plan 001.
func NewStaticPromptRegistry() *StaticPromptRegistry {
	return &StaticPromptRegistry{
		Resolution: PromptResolution{
			PromptVersion:     defaultTargetImportPromptVersion,
			RubricVersion:     defaultTargetImportRubricVersion,
			ModelProfileName:  defaultTargetImportModelProfileName,
			DataSourceVersion: defaultTargetImportDataSourceVersion,
			FeatureFlag:       "none",
		},
	}
}

// Resolve implements PromptRegistryClient.
func (r *StaticPromptRegistry) Resolve(_ context.Context, featureKey string, language string) (PromptResolution, error) {
	if strings.TrimSpace(featureKey) != FeatureKeyTargetImportParse || strings.TrimSpace(language) == "" {
		return PromptResolution{}, ErrPromptUnsupported
	}
	if r == nil || r.Resolution.ModelProfileName == "" {
		return NewStaticPromptRegistry().Resolution, nil
	}
	return r.Resolution, nil
}
