package handler

import (
	"errors"
	"net/http"

	"github.com/monshunter/easyinterview/backend/internal/resume"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func writeResumeServiceError(w http.ResponseWriter, err error, fallbackMessage string) {
	if errors.Is(err, resume.ErrValidationFailed) || errors.Is(err, resumestore.ErrInvalidCursor) {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "resume validation failed", nil)
		return
	}
	writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, fallbackMessage, nil)
}
