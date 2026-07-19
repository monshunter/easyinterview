/**
 * Auth contract gate (Phase 3.3) — keeps the D1 frontend within the C1 / B2
 * email-code contract. Any new auth surface (password login, OAuth,
 * Bearer-style headers, custom session storage) must first land an explicit
 * spec / OpenAPI revision; this test enforces the freeze.
 */
import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { extname, join, resolve } from "node:path";

import { describe, expect, it } from "vitest";

import { ALL_OPERATION_IDS } from "../../api/generated/client";

const FRONTEND_SRC = resolve(__dirname, "../..");
const AUTH_DIR = resolve(__dirname);

const ALLOWED_AUTH_OPERATIONS = new Set<string>([
  "startAuthEmailChallenge",
  "verifyAuthEmailChallenge",
  "updateMe",
  "getMe",
  "deleteMe",
  "logout",
]);

/** D1 also bootstraps the public runtime config, but that is not an auth op. */
const ALLOWED_NON_AUTH_OPERATIONS = new Set<string>([
  "getRuntimeConfig",
  // Phase 2-4 TargetJobs operations (frontend-home-job-picks-and-parse)
  "listTargetJobs",
  "importTargetJob",
  "archiveTargetJob",
  "createUploadPresign",
  "getTargetJob",
  "updateTargetJob",
  // Target-scoped current-round report overview (frontend-report-dashboard)
  "listTargetJobReports",
  // Phase 2-4 workspace-and-practice operations (frontend-workspace-and-practice)
  "getResume",
  "getPracticePlan",
  "createPracticePlan",
  "startPracticeSession",
  // D-20 flat resume workshop operations (frontend-resume-workshop)
  "listResumes",
  "updateResume",
  "duplicateResume",
  "archiveResume",
  "exportResume",
  "requestResumeTailor",
  "getResumeTailorRun",
  // Phase 1-6 resume workshop create flow (frontend-resume-workshop/002)
  "registerResume",
  "getPracticeSession",
  "getJob",
  // Report-owned read-only conversation record (frontend-report-dashboard)
  "getReportConversation",
]);

const ALL_CLIENT_CALL_RE = /\.client\.(\w+)\(/g;
const APP_AUTH_DISPATCH_RE = /runtime\.client\.(\w+)\(/g;

function* walk(dir: string): Generator<string> {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      yield* walk(full);
      continue;
    }
    const ext = extname(entry);
    if (ext === ".ts" || ext === ".tsx") yield full;
  }
}

function isTestFile(file: string): boolean {
  return /\.test\.(ts|tsx)$/.test(file);
}

function isGenerated(file: string): boolean {
  return file.includes(`${"/api/"}generated/`);
}

describe("auth contract gate (Phase 3.3)", () => {
  it("does not introduce Bearer-style Authorization headers in active code", () => {
    const pattern = /["'`]Bearer\s/i;
    const offenders: string[] = [];
    for (const file of walk(FRONTEND_SRC)) {
      if (isTestFile(file)) continue;
      if (isGenerated(file)) continue;
      const content = readFileSync(file, "utf8");
      if (pattern.test(content)) offenders.push(file);
    }
    expect(offenders).toEqual([]);
  });

  it("does not persist auth tokens / sessions in localStorage or sessionStorage", () => {
    const offenders: string[] = [];
    const callPattern = /(localStorage|sessionStorage)\.setItem\([^)]*\)/g;
    const sensitive = /(token|session|auth|jwt|bearer|cookie)/i;
    for (const file of walk(FRONTEND_SRC)) {
      if (isTestFile(file)) continue;
      const content = readFileSync(file, "utf8");
      const matches = content.match(callPattern) ?? [];
      for (const match of matches) {
        if (sensitive.test(match)) {
          offenders.push(`${file}: ${match}`);
        }
      }
    }
    expect(offenders).toEqual([]);
  });

  it("does not reference password / OAuth API paths in active code", () => {
    const forbidden = [
      "/auth/password",
      "/auth/login/password",
      "/auth/oauth",
      "/oauth/authorize",
      "/oauth/callback",
      "/auth/token",
    ];
    const offenders: Array<{ file: string; needle: string }> = [];
    for (const file of walk(FRONTEND_SRC)) {
      if (isTestFile(file)) continue;
      if (isGenerated(file)) continue;
      const content = readFileSync(file, "utf8");
      for (const needle of forbidden) {
        if (content.includes(needle)) offenders.push({ file, needle });
      }
    }
    expect(offenders).toEqual([]);
  });

  it("does not import the generated EasyInterviewClient inside auth/ screens (callbacks only)", () => {
    const offenders: string[] = [];
    for (const file of walk(AUTH_DIR)) {
      if (isTestFile(file)) continue;
      const content = readFileSync(file, "utf8");
      if (/from\s+["'][^"']*api\/generated\/client/.test(content)) {
        offenders.push(file);
      }
    }
    expect(offenders).toEqual([]);
  });

  it("App.tsx auth dispatch only references the allowed auth operations", () => {
    // App.tsx wires the public email challenge, verify + `/me` follow-up,
    // first-login profile completion, and logout. Other auth surfaces must
    // still land explicit spec/OpenAPI work before appearing here.
    const file = resolve(__dirname, "../App.tsx");
    const content = readFileSync(file, "utf8");
    const used = new Set<string>();
    let match: RegExpExecArray | null;
    while ((match = APP_AUTH_DISPATCH_RE.exec(content)) !== null) {
      const op = match[1];
      if (op) used.add(op);
    }
    const expected = new Set([
      "startAuthEmailChallenge",
      "verifyAuthEmailChallenge",
      "updateMe",
      "getMe",
      "logout",
    ]);
    expect([...used].sort()).toEqual([...expected].sort());
  });

  it("frontend/src only calls client operations covered by current owner matrices", () => {
    const allowed = new Set<string>([
      ...ALLOWED_AUTH_OPERATIONS,
      ...ALLOWED_NON_AUTH_OPERATIONS,
    ]);
    const offenders: Array<{ file: string; op: string }> = [];
    for (const file of walk(FRONTEND_SRC)) {
      if (isTestFile(file)) continue;
      if (isGenerated(file)) continue;
      const content = readFileSync(file, "utf8");
      let match: RegExpExecArray | null;
      const re = new RegExp(ALL_CLIENT_CALL_RE.source, "g");
      while ((match = re.exec(content)) !== null) {
        const op = match[1];
        if (op && !allowed.has(op)) offenders.push({ file, op });
      }
    }
    expect(offenders).toEqual([]);
  });

  it("owner operation allowlists only retain generated operations", () => {
    const allowed = new Set<string>([
      ...ALLOWED_AUTH_OPERATIONS,
      ...ALLOWED_NON_AUTH_OPERATIONS,
    ]);
    const generated = new Set<string>(ALL_OPERATION_IDS);

    expect([...allowed].filter((operationId) => !generated.has(operationId))).toEqual([]);
  });

  it("keeps zero auth_reset / forgot-password residue in the auth surface (D-16)", () => {
    // product-scope D-16: email-code is the only sign-in flow.
    // The reset screen file must be gone and no non-test auth source may
    // reference the out-of-scope route or forgot-password vocabulary.
    expect(existsSync(resolve(AUTH_DIR, "AuthResetScreen.tsx"))).toBe(false);

    const offenders: Array<{ file: string; needle: string }> = [];
    for (const file of walk(AUTH_DIR)) {
      if (/\.test\.tsx?$/.test(file)) continue;
      const content = readFileSync(file, "utf8");
      for (const needle of ["auth_reset", "AuthResetScreen", "forgotPassword"]) {
        if (content.includes(needle)) offenders.push({ file, needle });
      }
    }
    expect(offenders).toEqual([]);
  });
});
