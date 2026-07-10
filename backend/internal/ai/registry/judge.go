package registry

import "context"

// FailClosedJudge is the fail-closed Judge default for callers that have not
// injected a concrete judge dependency. Every call returns ErrJudgeUnavailable.
//
// Wire it into business code as `var judge Judge = FailClosedJudge{}`
// so the dependency graph is explicit.
type FailClosedJudge struct{}

// Judge implements the Judge interface; see types.go for the contract.
func (FailClosedJudge) Judge(
	_ context.Context,
	_ string,
	_ string,
	_ []byte,
	_ string,
) ([]Score, Reasoning, error) {
	return nil, Reasoning{}, ErrJudgeUnavailable
}
