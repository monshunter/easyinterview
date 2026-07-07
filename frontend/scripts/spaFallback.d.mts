/**
 * Type declarations for the host fallback resolver consumed by both
 * `serve-pixel-parity.mjs` (Node) and the focused unit/scenario tests
 * (`frontend/src/app/spaFallback.test.ts` + `nonCurrentRouteNegative.test.ts`
 * + `p0-090-url-routing-hash-non-current-negative.test.tsx`). The .mjs source
 * stays Node-only; this .d.mts gives TypeScript-aware tooling type info
 * without forcing a build step.
 */

export const FRONTEND_CANONICAL_PATHS: readonly string[];
export const FRONTEND_NON_CURRENT_PATHS: readonly string[];
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
