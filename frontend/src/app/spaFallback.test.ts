/**
 * Plan 004 Phase 4.1 — host fallback contract.
 *
 * `frontend/scripts/spaFallback.mjs` mirrors `frontend/src/app/routeUrl.ts`
 * canonical path table. This test guards three properties:
 *   1. Every canonical Route path returned by the App codec is served by
 *      the SPA fallback (otherwise a deep-link reload returns 404).
 *   2. Unsafe / non-frontend paths are rejected by the fallback so
 *      `/api/*`, `/openapi/*`, `/ui-design/*`, `/health`, scenario script
 *      paths and asset 404s remain transparent.
 *   3. Static asset requests (with a file extension) are never routed
 *      through the SPA fallback.
 */
import { describe, expect, it } from "vitest";

import {
  FALLBACK_DENY_PREFIXES,
  FRONTEND_CANONICAL_PATHS,
  FRONTEND_LEGACY_PATHS,
  isCanonicalFrontendPath,
  resolveSpaFallback,
} from "../../scripts/spaFallback.mjs";
import { LEGACY_PATH_TO_ROUTE, ROUTE_TO_PATH } from "./routeUrl";

describe("spaFallback canonical path drift gate", () => {
  it("covers every Route path declared by the App codec (ROUTE_TO_PATH)", () => {
    for (const path of Object.values(ROUTE_TO_PATH)) {
      expect(
        FRONTEND_CANONICAL_PATHS.includes(path),
        `routeUrl.ROUTE_TO_PATH includes ${path}; spaFallback.FRONTEND_CANONICAL_PATHS must include it too`,
      ).toBe(true);
    }
  });

  it("does not enumerate paths the App codec does not own", () => {
    for (const path of FRONTEND_CANONICAL_PATHS) {
      expect(
        Object.values(ROUTE_TO_PATH).includes(path),
        `spaFallback enumerates ${path} which is not in routeUrl.ROUTE_TO_PATH`,
      ).toBe(true);
    }
  });

  it("mirrors the App codec legacy path table (retired routes still load the SPA)", () => {
    const legacyPaths = [...LEGACY_PATH_TO_ROUTE.keys()];
    expect([...FRONTEND_LEGACY_PATHS].sort()).toEqual(legacyPaths.sort());
    for (const path of FRONTEND_LEGACY_PATHS) {
      expect(
        isCanonicalFrontendPath(path),
        `${path} must still be served by the SPA fallback for App-side normalization`,
      ).toBe(true);
      expect(
        Object.values(ROUTE_TO_PATH).includes(path),
        `${path} must not remain a canonical Route path`,
      ).toBe(false);
    }
  });
});

describe("isCanonicalFrontendPath", () => {
  it("returns true for every canonical frontend path", () => {
    for (const path of FRONTEND_CANONICAL_PATHS) {
      expect(isCanonicalFrontendPath(path), `${path} must be canonical`).toBe(
        true,
      );
    }
  });

  it("returns true with trailing slash or query string", () => {
    expect(isCanonicalFrontendPath("/workspace/")).toBe(true);
    expect(isCanonicalFrontendPath("/workspace?targetJobId=tj-1")).toBe(true);
    expect(isCanonicalFrontendPath("/auth/login/?next=/workspace")).toBe(true);
  });

  it("returns false for /api/*, /openapi/*, /ui-design/*, /health, /assets/*", () => {
    expect(isCanonicalFrontendPath("/api/health")).toBe(false);
    expect(isCanonicalFrontendPath("/openapi/openapi.yaml")).toBe(false);
    expect(isCanonicalFrontendPath("/ui-design/index.html")).toBe(false);
    expect(isCanonicalFrontendPath("/health")).toBe(false);
    expect(isCanonicalFrontendPath("/assets/index.js")).toBe(false);
  });

  it("returns false for file requests (with extension) under any path", () => {
    expect(isCanonicalFrontendPath("/index.html")).toBe(false);
    expect(isCanonicalFrontendPath("/workspace.json")).toBe(false);
    expect(isCanonicalFrontendPath("/auth/login.css")).toBe(false);
  });

  it("returns false for unknown / retired paths so the App handles fallback semantically", () => {
    expect(isCanonicalFrontendPath("/totally-unknown")).toBe(false);
    expect(isCanonicalFrontendPath("/voice")).toBe(false);
    expect(isCanonicalFrontendPath("/welcome")).toBe(false);
    expect(isCanonicalFrontendPath("/mistakes")).toBe(false);
    expect(isCanonicalFrontendPath("/drill")).toBe(false);
    expect(isCanonicalFrontendPath("/growth")).toBe(false);
  });

  it("returns false for empty / non-string input", () => {
    expect(isCanonicalFrontendPath("")).toBe(false);
    expect(isCanonicalFrontendPath(null)).toBe(false);
    expect(isCanonicalFrontendPath(undefined)).toBe(false);
  });
});

describe("resolveSpaFallback", () => {
  it("returns index.html absolute path for canonical routes", () => {
    expect(resolveSpaFallback("/workspace?targetJobId=tj-1", "/tmp/dist")).toEqual({
      kind: "file",
      absolute: "/tmp/dist/index.html",
    });
    expect(resolveSpaFallback("/auth/login", "/tmp/dist")).toEqual({
      kind: "file",
      absolute: "/tmp/dist/index.html",
    });
  });

  it("returns null for /api/*, /openapi/*, /health, file requests, unknown paths", () => {
    expect(resolveSpaFallback("/api/health", "/tmp/dist")).toBeNull();
    expect(resolveSpaFallback("/openapi/openapi.yaml", "/tmp/dist")).toBeNull();
    expect(resolveSpaFallback("/health", "/tmp/dist")).toBeNull();
    expect(resolveSpaFallback("/assets/index.js", "/tmp/dist")).toBeNull();
    expect(resolveSpaFallback("/totally-unknown", "/tmp/dist")).toBeNull();
  });
});

describe("FALLBACK_DENY_PREFIXES contract", () => {
  it("denies API / openapi / ui-design / health / assets / __vite", () => {
    expect(FALLBACK_DENY_PREFIXES).toContain("/api/");
    expect(FALLBACK_DENY_PREFIXES).toContain("/openapi/");
    expect(FALLBACK_DENY_PREFIXES).toContain("/ui-design/");
    expect(FALLBACK_DENY_PREFIXES).toContain("/health");
    expect(FALLBACK_DENY_PREFIXES).toContain("/assets/");
    expect(FALLBACK_DENY_PREFIXES).toContain("/__vite");
  });
});
