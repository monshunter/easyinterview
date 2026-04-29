package aiclient

import "errors"

// Config is the AIClient boot config injected by A4 secrets-and-config wire
// (spec §5 / D-4). Phase 4 enforces fail-fast when GatewayBaseURL or
// GatewayAPIKey is empty in any non-test AppEnv.
type Config struct {
	AppEnv           string
	GatewayBaseURL   string
	GatewayAPIKey    string
	ModelProfilePath string
}

// AppEnvTest is the only AppEnv value that allows the stub provider without
// an explicit WithStubAllowed(true).
const AppEnvTest = "test"

// ErrMissingGatewayConfig is returned by New when AppEnv is not "test" and
// either GatewayBaseURL or GatewayAPIKey is empty. Callers in cmd/api and
// cmd/worker convert it into a non-zero exit; A3 never silently degrades to
// stub.
var ErrMissingGatewayConfig = errors.New("aiclient: AI_GATEWAY_BASE_URL and AI_GATEWAY_API_KEY are required outside APP_ENV=test")

// ErrTaskTypeNotImplemented is returned when a profile resolves to a task
// type the current build cannot service (e.g. stt while plan 001 is the
// shipping baseline). Plan 002 lifts this error for the relevant task types.
var ErrTaskTypeNotImplemented = errors.New("aiclient: task type not implemented in this build")
