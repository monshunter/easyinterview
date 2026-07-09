package practice

import (
	"fmt"
	"strings"
)

type TurnStatus string

const (
	TurnStatusAsked             TurnStatus = "asked"
	TurnStatusAnswered          TurnStatus = "answered"
	TurnStatusFollowUpRequested TurnStatus = "follow_up_requested"
	TurnStatusAssessed          TurnStatus = "assessed"
)

func ParseTurnStatus(raw string) (TurnStatus, error) {
	status := TurnStatus(strings.TrimSpace(raw))
	if !status.valid() {
		return "", fmt.Errorf("unknown practice turn status %q", raw)
	}
	return status, nil
}

func ParseWireTurnStatus(raw string) (TurnStatus, error) {
	return ParseTurnStatus(raw)
}

func (s TurnStatus) WireValue() (string, error) {
	if !s.valid() {
		return "", fmt.Errorf("unknown practice turn status %q", string(s))
	}
	return string(s), nil
}

func (s TurnStatus) valid() bool {
	switch s {
	case TurnStatusAsked,
		TurnStatusAnswered,
		TurnStatusFollowUpRequested,
		TurnStatusAssessed:
		return true
	default:
		return false
	}
}
