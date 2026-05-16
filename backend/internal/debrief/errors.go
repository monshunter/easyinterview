package debrief

import stderrs "errors"

var (
	ErrDebriefNotFound     = stderrs.New("debrief not found")
	ErrDebriefConflict     = stderrs.New("debrief conflict")
	ErrDebriefValidation   = stderrs.New("debrief validation failed")
	ErrDebriefIllegalState = stderrs.New("debrief illegal state")
	ErrDebriefPrerequisite = stderrs.New("debrief prerequisite not found")
)
