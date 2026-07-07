package registry

import "context"

// NotImplementedJudge is the fail-closed Judge default for callers that have
// not injected a concrete judge dependency. Every call returns
// ErrJudgeNotImplemented.
//
// Wire it into business code as `var judge Judge = NotImplementedJudge{}`
// so the dependency graph is explicit.
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
