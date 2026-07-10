package targetjob

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
)

// RegistryAdapter satisfies targetjob.PromptRegistryClient by translating
// registry.PromptResolution into the local targetjob.PromptResolution. It
// is the only path through which the parse executor reaches the F3
// registry; the executor must not import the registry package directly
// because that would couple targetjob to F3's import path.
//
// F3 owns prompt and rubric version selection; targetjob only consumes the
// resolved current metadata through this adapter.
type RegistryAdapter struct {
	client *registry.Client
}

// NewRegistryAdapter wires a *registry.Client into the targetjob domain.
// A nil client returns nil so callers can detect a misconfigured wiring
// during construction rather than at first request.
func NewRegistryAdapter(client *registry.Client) *RegistryAdapter {
	if client == nil {
		return nil
	}
	return &RegistryAdapter{client: client}
}

// Resolve implements PromptRegistryClient. The adapter forwards to the
// registry client, asserts that the returned FeatureKey matches the
// caller's input (so a registry-side bug cannot silently substitute a
// different baseline), and projects all 7 targetjob fields plus the
// caller's feature_key.
func (a *RegistryAdapter) Resolve(ctx context.Context, featureKey string, language string) (PromptResolution, error) {
	if a == nil || a.client == nil {
		return PromptResolution{}, ErrPromptUnsupported
	}
	if strings.TrimSpace(featureKey) == "" || strings.TrimSpace(language) == "" {
		return PromptResolution{}, ErrPromptUnsupported
	}

	resolved, err := a.client.ResolveActive(ctx, featureKey, language)
	if err != nil {
		if errors.Is(err, registry.ErrPromptUnsupported) ||
			errors.Is(err, registry.ErrLanguageUnsupported) {
			return PromptResolution{}, ErrPromptUnsupported
		}
		return PromptResolution{}, fmt.Errorf("targetjob: registry resolve: %w", err)
	}
	if resolved.FeatureKey != featureKey {
		return PromptResolution{}, fmt.Errorf(
			"targetjob: registry returned feature_key %q, expected %q",
			resolved.FeatureKey, featureKey,
		)
	}

	return PromptResolution{
		PromptVersion:       resolved.PromptVersion,
		RubricVersion:       resolved.RubricVersion,
		ModelProfileName:    resolved.ModelProfileName,
		DataSourceVersion:   resolved.DataSourceVersion,
		FeatureFlag:         resolved.FeatureFlag,
		SystemMessage:       resolved.SystemMessage,
		UserMessageTemplate: resolved.UserMessageTemplate,
		OutputSchema:        resolved.OutputSchema,
	}, nil
}
