package events

import (
	"errors"
	"fmt"
)

// ErrManualFormNotEventSource is returned when business code attempts to
// surface the B2 API source variant `manual_form` as a B3 event sourceType.
// `manual_form` is the synchronous ready fallback path (backend-targetjob 001
// plan D-13 / B3 spec D-13) and MUST NOT emit `target.import.requested`.
var ErrManualFormNotEventSource = errors.New("manual_form must not appear in target.import.requested.sourceType")

// MapAPISourceTypeToEvent maps a B2 OpenAPI `ImportTargetJobRequest.source.type`
// literal to the event-local TargetImportSourceType used in
// `target.import.requested.sourceType`. The mapping is fixed by B3 spec D-13:
//
//	url         -> url
//	manual_text -> text
//	file        -> file
//	manual_form -> error (synchronous path; no event)
//
// Any other value is rejected so business code cannot smuggle unknown source
// variants past the runner contract.
func MapAPISourceTypeToEvent(apiSourceType string) (TargetImportSourceType, error) {
	switch apiSourceType {
	case "url":
		return TargetImportSourceTypeUrl, nil
	case "manual_text":
		return TargetImportSourceTypeText, nil
	case "file":
		return TargetImportSourceTypeFile, nil
	case "manual_form":
		return "", ErrManualFormNotEventSource
	default:
		return "", fmt.Errorf("unknown api source type %q (allowed: url, manual_text, file)", apiSourceType)
	}
}
