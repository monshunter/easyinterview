/**
 * RuntimeConfig types mirror the Go-side response built by
 * `backend/internal/platform/config/runtime_config.go` (BuildRuntimeConfig).
 *
 * OpenAPI schema truth source is owned by B2 openapi-v1-contract — keep
 * this file in lockstep with the spec field allowlist (D-2). A4 only owns
 * the field set; B2 freezes the wire shape.
 */

export interface RuntimeFlag {
  enabled: boolean;
  variant?: string;
}

export interface RuntimeConfig {
  appVersion: string;
  defaultUiLanguage: string;
  analyticsEnabled: boolean;
  featureFlags: Record<string, RuntimeFlag>;
  postHogPublicKey?: string;
}
