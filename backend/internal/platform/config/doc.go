// Package config is the truth source for runtime configuration loading,
// validation and redaction in easyinterview backend processes.
//
// Layer priority (secrets-and-config spec D-1, highest wins):
//
//  1. Runtime secret (injected via SecretSource)
//  2. Process environment variable
//  3. config/{APP_ENV}.yaml (per-environment overrides)
//  4. config/config.yaml (defaults, repository-versioned)
//
// Public read API uses the Get* family (GetString / GetInt / GetBool /
// GetDuration / GetSecret). Keys use dot paths that match the canonical
// config schema in spec §3.1.2 (for example app.listenAddr, log.level,
// auth.sessionCookieSecret). Business packages must consume configuration
// through these accessors and never call os.Getenv directly; see spec §4.1
// for the boundary lint enforced by Phase 4.
//
// Sensitive values are wrapped in RedactedString so that fmt, JSON marshal
// and error wrapping always emit *** and only Reveal() returns plaintext.
//
// This package is owned by docs/spec/secrets-and-config (A4).
package config
