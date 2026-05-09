package registry

import "context"

// NotImplementedJudge is the default Judge implementation shipped by plan
// 001. Every call returns ErrJudgeNotImplemented; plan 002 replaces this
// with a real LLM-judge capability profile binding.
//
// Wire it into business code as `var judge Judge = NotImplementedJudge{}`
// so the dependency graph is explicit and the swap to a real implementation
// in plan 002 is a one-line change.
type NotImplementedJudge struct{}

// Judge implements the Judge interface; see types.go for the contract.
func (NotImplementedJudge) Judge(
	_ context.Context,
	_ string,
	_ string,
	_ []byte,
	_ string,
) (Score, Reasoning, error) {
	return Score{}, Reasoning{}, ErrJudgeNotImplemented
}
