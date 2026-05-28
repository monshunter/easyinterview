/**
 * Host fallback resolver for the formal frontend App shell.
 *
 * Truth source: docs/spec/frontend-shell/plans/004-url-addressable-routing/
 * plan.md §6 Phase 4.1.
 *
 * Static hosts (`vite preview`, the pixel-parity Playwright server, any
 * production CDN configured by deployment) must return `index.html` when
 * the request path is a canonical frontend route — otherwise reload /
 * direct-open of `/workspace?targetJobId=...` returns 404 and the
 * URL-addressable routing contract breaks.
 *
 * The canonical path list mirrors `frontend/src/app/routeUrl.ts` which is
 * the TypeScript truth source. Both files are kept in sync by the focused
 * `spaFallback.test.ts` unit test in `frontend/src/app/`.
 */

/** Canonical frontend paths that the SPA fallback must serve. */
export const FRONTEND_CANONICAL_PATHS = Object.freeze([
  "/",
  "/jd-match",
  "/workspace",
  "/resume-versions",
  "/debrief",
  "/parse",
  "/practice",
  "/generating",
  "/report",
  "/company-intel",
  "/profile",
  "/settings",
  "/auth/login",
  "/auth/verify",
  "/auth/profile",
  "/auth/reset",
  "/auth/logout",
]);

/** Path prefixes that must NEVER be swallowed by the SPA fallback. */
export const FALLBACK_DENY_PREFIXES = Object.freeze([
  "/api/",
  "/openapi/",
  "/ui-design/",
  "/health",
  "/assets/",
  "/__vite",
]);

/**
 * Returns `true` when `path` is a canonical frontend path that should be
 * served by `index.html`.
 *
 * Behaviour:
 *  - `/workspace`, `/workspace/`, `/workspace?x=1` → true
 *  - `/`                                            → true (root)
 *  - `/totally-unknown`                             → false (legacy fallback
 *    is handled in the App, not in the host)
 *  - `/workspace.json` (any file extension)         → false
 *  - `/api/...`, `/health`, `/ui-design/...`        → false
 */
export function isCanonicalFrontendPath(path) {
  if (typeof path !== "string" || path === "") return false;
  const cleaned = path.split("?")[0].split("#")[0];
  if (cleaned === "") return false;
  for (const deny of FALLBACK_DENY_PREFIXES) {
    if (cleaned === deny.replace(/\/$/, "") || cleaned.startsWith(deny)) {
      return false;
    }
  }
  // Reject any path that looks like a file request (has a `.`).
  const lastSegment = cleaned.split("/").pop() ?? "";
  if (lastSegment.includes(".")) return false;
  // Normalize trailing slash for matching: `/workspace/` and `/workspace`
  // both resolve to the same SPA route.
  const normalized =
    cleaned !== "/" && cleaned.endsWith("/")
      ? cleaned.slice(0, -1)
      : cleaned;
  return FRONTEND_CANONICAL_PATHS.includes(normalized);
}

/**
 * Returns `index.html` resolution for SPA fallback, or `null` when the
 * request should fall through to the regular static file resolver.
 *
 * `frontendDistDir` is the absolute path to `frontend/dist`. Callers use
 * the returned `{absolute, contentType}` to write a 200 + index.html.
 */
export function resolveSpaFallback(path, frontendDistDir) {
  if (!isCanonicalFrontendPath(path)) return null;
  return {
    kind: "file",
    absolute: `${frontendDistDir}/index.html`,
  };
}
