package practice

type TurnStatus string

const (
	TurnStatusAsked             TurnStatus = "asked"
	TurnStatusAnswered          TurnStatus = "answered"
	TurnStatusFollowUpRequested TurnStatus = "follow_up_requested"
	TurnStatusAssessed          TurnStatus = "assessed"
)
