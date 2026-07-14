/**
 * Type declarations for the host fallback resolver consumed by both
 * `serve-pixel-parity.mjs` (Node) and the focused code-level tests
 * (`frontend/src/app/spaFallback.test.ts` + `outOfScopeRouteNegative.test.ts`
 * + `url-routing-negative.test.tsx`). The .mjs source
 * stays Node-only; this .d.mts gives TypeScript-aware tooling type info
 * without forcing a build step.
 */

export const FRONTEND_CANONICAL_PATHS: readonly string[];
export const FRONTEND_OUT_OF_SCOPE_PATHS: readonly string[];
export const FALLBACK_DENY_PREFIXES: readonly string[];

export function isCanonicalFrontendPath(path: unknown): boolean;

export interface SpaFallbackResolution {
  kind: "file";
  absolute: string;
}

export function resolveSpaFallback(
  path: string,
  frontendDistDir: string,
): SpaFallbackResolution | null;
