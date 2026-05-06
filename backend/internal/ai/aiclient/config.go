package aiclient

import "errors"

// Config is the AIClient boot config injected by A4 secrets-and-config wire
// (spec §5 / D-4). Phase 4 enforces fail-fast when the provider registry or
// model profile path is empty in any non-test AppEnv.
type Config struct {
	AppEnv               string
	ProviderRegistryPath string
	ModelProfilePath     string
}

// AppEnvTest is the only AppEnv value that allows the stub provider without
// an explicit WithStubAllowed(true).
const AppEnvTest = "test"

// ErrMissingProviderConfig is returned by New when AppEnv is not "test" and
// either ProviderRegistryPath or ModelProfilePath is empty. Runtime callers
// convert it into a non-zero exit; A3 never silently degrades
// to stub.
var ErrMissingProviderConfig = errors.New("aiclient: AI_PROVIDER_REGISTRY_PATH and AI_MODEL_PROFILE_PATH are required outside APP_ENV=test")

// ErrCapabilityNotImplemented is returned when a profile resolves to a
// capability the current build cannot service (e.g. stt while plan 001 is the
// shipping baseline). Plan 002 lifts this error for the relevant capabilities.
var ErrCapabilityNotImplemented = errors.New("aiclient: capability not implemented in this build")
