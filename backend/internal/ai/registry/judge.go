package registry

import "context"

// NotImplementedJudge is the safe-default Judge implementation shipped by plan
// 001. Every call returns ErrJudgeNotImplemented; plan 004 adds the real
// LLMJudge while keeping NotImplementedJudge as the fail-closed default for
// callers that have not injected a judge dependency.
//
// Wire it into business code as `var judge Judge = NotImplementedJudge{}`
// so the dependency graph is explicit and swapping to LLMJudge is a one-line
// change.
type NotImplementedJudge struct{}

// Judge implements the Judge interface; see types.go for the contract.
func (NotImplementedJudge) Judge(
	_ context.Context,
	_ string,
	_ string,
	_ []byte,
	_ string,
) ([]Score, Reasoning, error) {
	return nil, Reasoning{}, ErrJudgeNotImplemented
}
