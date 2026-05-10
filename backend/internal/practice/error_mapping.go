package practice

import (
	"context"
	stderrs "errors"
	"strings"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func serviceErrorFromAI(err error) error {
	if err == nil {
		return nil
	}
	code, ok := aiErrorCode(err)
	if !ok {
		return err
	}
	meta := sharederrors.CodeRegistry[code]
	return &ServiceError{Code: code, Message: meta.Message}
}

func aiErrorCode(err error) (string, bool) {
	var apiErr *sharederrors.APIError
	if stderrs.As(err, &apiErr) && isPracticeAIErrorCode(apiErr.Code) {
		return apiErr.Code, true
	}
	if stderrs.Is(err, context.DeadlineExceeded) {
		return sharederrors.CodeAiProviderTimeout, true
	}
	text := strings.ToUpper(err.Error())
	for _, code := range []string{
		sharederrors.CodeAiProviderTimeout,
		sharederrors.CodeAiOutputInvalid,
		sharederrors.CodeAiProviderConfigInvalid,
		sharederrors.CodeAiProviderSecretMissing,
		sharederrors.CodeAiFallbackExhausted,
		sharederrors.CodeAiUnsupportedCapability,
	} {
		if strings.Contains(text, code) {
			return code, true
		}
	}
	if strings.Contains(text, "TIMEOUT") || strings.Contains(text, "DEADLINE EXCEEDED") {
		return sharederrors.CodeAiProviderTimeout, true
	}
	return "", false
}

func isPracticeAIErrorCode(code string) bool {
	switch code {
	case sharederrors.CodeAiProviderTimeout,
		sharederrors.CodeAiOutputInvalid,
		sharederrors.CodeAiProviderConfigInvalid,
		sharederrors.CodeAiProviderSecretMissing,
		sharederrors.CodeAiFallbackExhausted,
		sharederrors.CodeAiUnsupportedCapability:
		return true
	default:
		return false
	}
}
