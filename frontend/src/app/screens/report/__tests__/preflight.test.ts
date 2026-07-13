// Phase 0 cross-owner preflight asserts the backend-review/001 Phase 0 contract
// deliverables landed in the repo before Phase 1 implementation starts. If any
// assert fails the implementation must stop and route the gap back through
// bug-report / retrospective to backend-review/001 owner — see
// docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/plan.md
// Phase 0 (B1 / B2 / fixtures / generated TS client).

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

import { ERROR_CODES } from "../../../../lib/conventions/errors";
import type { FeedbackReport } from "../../../../api/generated/types";

// preflight.test.ts lives at frontend/src/app/screens/report/__tests__/.
// new URL(".", import.meta.url) resolves to the __tests__ directory; six
// parent hops land on repo root (report → screens → app → src → frontend → repo).
const REPO_ROOT = resolve(
  fileURLToPath(new URL(".", import.meta.url)),
  "..",
  "..",
  "..",
  "..",
  "..",
  "..",
);

function readRepoFile(rel: string): string {
  return readFileSync(resolve(REPO_ROOT, rel), "utf8");
}

function readRepoJson<T = unknown>(rel: string): T {
  return JSON.parse(readRepoFile(rel)) as T;
}

interface FixtureScenario {
  response: {
    status: number;
    body: Record<string, unknown>;
  };
}

interface FixtureFile {
  scenarios: Record<string, FixtureScenario>;
}

describe("frontend-report-dashboard/001 Phase 0 preflight", () => {
  it("FeedbackReport schema in openapi.yaml declares optional nullable errorCode (TestB2FeedbackReportSchemaHasErrorCode)", () => {
    const spec = readRepoFile("openapi/openapi.yaml");
    const schemaIdx = spec.indexOf("    FeedbackReport:");
    expect(
      schemaIdx,
      "blocked on backend-review/001 Phase 0.2 + 0.4 (errorCode field in B2 schema): FeedbackReport schema missing",
    ).toBeGreaterThan(-1);

    // Slice the schema window up to the next sibling schema and require both
    // `errorCode` and a nullable ApiErrorCode oneOf.
    const tail = spec.slice(schemaIdx);
    const schemaWindow = tail.slice(0, tail.indexOf("\n    PaginatedFeedbackReport:"));
    expect(
      /\n        errorCode:/.test(schemaWindow),
      "blocked on backend-review/001 Phase 0.2 + 0.4 (errorCode field in B2 schema): errorCode property absent",
    ).toBe(true);
    expect(
      /errorCode:\s*\n\s+oneOf:\s*\n\s+- \$ref: '#\/components\/schemas\/ApiErrorCode'\s*\n\s+- type: 'null'/.test(
        schemaWindow,
      ),
      "blocked on backend-review/001 Phase 0.2 + 0.4 (errorCode field in B2 schema): errorCode must be oneOf [ApiErrorCode, null]",
    ).toBe(true);
  });

  it("generated TS FeedbackReport exposes the required nullable direct contract", () => {
    // Type-only assertion: compile-time + runtime mirror. The runtime check
    // grep'd from generated types defends against codegen drift.
    const generated = readRepoFile("frontend/src/api/generated/types.ts");
    expect(
      /errorCode\s*:\s*ApiErrorCode\s*\|\s*null;/.test(generated),
      "blocked on backend-review/001 Phase 0.2 (generated TS errorCode required nullable)",
    ).toBe(true);

    const sample: FeedbackReport = {
      id: "01918fa0-0000-7000-8000-000000000001",
      sessionId: "01918fa0-0000-7000-8000-000000000002",
      targetJobId: "01918fa0-0000-7000-8000-000000000003",
      status: "ready",
      errorCode: null,
      summary: "Grounded summary",
      context: {
        sourcePlanId: "01918fa0-0000-7000-8000-000000000004",
        targetJobTitle: "Platform Engineer",
        targetJobCompany: "Acme",
        resumeId: "01918fa0-0000-7000-8000-000000000005",
        resumeDisplayName: "Platform resume",
        roundId: "round-1-technical",
        roundSequence: 1,
        roundName: "Technical interview",
        roundType: "technical",
        language: "en",
        hasNextRound: true,
      },
      preparednessLevel: "needs_practice",
      highlights: [],
      issues: [{ dimensionCode: "technical_depth", evidence: "Needs a metric.", confidence: "medium" }],
      nextActions: [{ type: "retry_current_round", label: "Practice with a measurable result." }],
      dimensionAssessments: [{ code: "technical_depth", label: "Technical depth", status: "needs_work", confidence: "medium" }],
      retryFocusDimensionCodes: ["technical_depth"],
      provenance: {
        promptVersion: "v0.2.0",
        rubricVersion: "v0.2.0",
        modelId: "fixture",
        language: "en",
        featureFlag: "none",
        dataSourceVersion: "fixture.v1",
      },
      createdAt: "2026-05-15T00:00:00Z",
      updatedAt: "2026-05-15T00:00:00Z",
    };
    expect(sample.errorCode).toBeNull();
  });

  it("openapi report fixtures expose direct queued/generating/ready/failed variants", () => {
    const fixture = readRepoJson<FixtureFile>(
      "openapi/fixtures/Reports/getFeedbackReport.json",
    );
    const scenarios = fixture.scenarios;
    expect(
      scenarios.default,
      "blocked on backend-review/001 Phase 0.4: missing default fixture variant",
    ).toBeDefined();
    expect(
      scenarios.generating,
      "blocked on backend-review/001 Phase 0.4: missing generating fixture variant",
    ).toBeDefined();
    expect(
      scenarios.failed,
      "blocked on backend-review/001 Phase 0.4: missing failed fixture variant",
    ).toBeDefined();

    const failedBody = scenarios.failed!.response.body as {
      status: string;
      errorCode: string | null;
    };
    expect(failedBody.status).toBe("failed");
    expect(
      failedBody.errorCode,
      "blocked on backend-review/001 Phase 0.4: report-failed scenario must populate errorCode",
    ).not.toBeNull();
  });

  it("openapi/fixtures/Reports/listTargetJobReports.json has empty scenario shape (TestListTargetJobReportsEmptyFixtureVariantExists)", () => {
    const fixture = readRepoJson<FixtureFile>(
      "openapi/fixtures/Reports/listTargetJobReports.json",
    );
    const empty = fixture.scenarios.empty;
    expect(
      empty,
      "blocked on backend-review/001 Phase 0.4: listTargetJobReports.empty fixture missing",
    ).toBeDefined();
    const body = empty!.response.body as {
      items: unknown[];
      pageInfo: { hasMore: boolean; nextCursor: unknown };
    };
    expect(body.items).toEqual([]);
    expect(body.pageInfo.hasMore).toBe(false);
    expect(body.pageInfo.nextCursor).toBeNull();
  });

  it("shared/conventions.yaml registers REPORT_NOT_FOUND with 404 + retryable=false (TestReportNotFoundErrorCodeRegistered)", () => {
    const yaml = readRepoFile("shared/conventions.yaml");
    const idx = yaml.indexOf("  - code: REPORT_NOT_FOUND");
    expect(
      idx,
      "blocked on backend-review/001 Phase 0.1 (REPORT_NOT_FOUND in B1): code entry missing",
    ).toBeGreaterThan(-1);
    const block = yaml.slice(idx, idx + 400);
    expect(/retryable:\s*false/.test(block)).toBe(true);

    // The B1 conventions YAML doesn't carry httpStatus inline (HTTP mapping is
    // backend-owned), but the contract requires REPORT_NOT_FOUND to resolve to
    // 404 via the openapi.yaml response binding. Assert the response wiring.
    const spec = readRepoFile("openapi/openapi.yaml");
    const opIdx = spec.indexOf("operationId: getFeedbackReport");
    expect(opIdx, "getFeedbackReport operation missing in openapi.yaml").toBeGreaterThan(-1);
    const window = spec.slice(opIdx, opIdx + 4000);
    expect(/'404':/.test(window), "getFeedbackReport must declare a 404 response").toBe(true);
    expect(
      /REPORT_NOT_FOUND/.test(window),
      "getFeedbackReport 404 must surface REPORT_NOT_FOUND",
    ).toBe(true);
  });

  it("generated TS ApiErrorCode union exposes REPORT_NOT_FOUND + the AI_* family used by ReportFailureState (TestReportNotFoundErrorCodeRegistered TS half)", () => {
    const generated = readRepoFile("frontend/src/api/generated/types.ts");
    expect(/"REPORT_NOT_FOUND"/.test(generated)).toBe(true);
    expect(/"AI_PROVIDER_TIMEOUT"/.test(generated)).toBe(true);
    expect(/"AI_PROVIDER_SECRET_MISSING"/.test(generated)).toBe(true);
    expect(/"AI_PROVIDER_CONFIG_INVALID"/.test(generated)).toBe(true);
    expect(/"AI_OUTPUT_INVALID"/.test(generated)).toBe(true);

    // The conventions registry shim mirrors B1 in TS land — ensure plan 001's
    // failure-state mapping has a single source of truth to consume.
    expect(ERROR_CODES.REPORT_NOT_FOUND).toBe("REPORT_NOT_FOUND");
    expect(ERROR_CODES.AI_PROVIDER_TIMEOUT).toBe("AI_PROVIDER_TIMEOUT");
  });
});

describe("frontend-report-dashboard/001 Phase 8 browser evidence", () => {
  it("keeps owner claims and P0.059 bound to deterministic semantic and pixel-difference gates", () => {
    const planRoot =
      "docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff";
    const ownerPaths = [
      "docs/spec/frontend-report-dashboard/spec.md",
      `${planRoot}/plan.md`,
      `${planRoot}/checklist.md`,
      `${planRoot}/test-plan.md`,
      `${planRoot}/test-checklist.md`,
      `${planRoot}/bdd-plan.md`,
      `${planRoot}/bdd-checklist.md`,
    ];
    const claimPaths = [
      ...ownerPaths,
      "frontend/tests/pixel-parity/generating.spec.ts",
      "frontend/tests/pixel-parity/report.spec.ts",
      "test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative/README.md",
      "test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative/data/seed-input.md",
      "test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative/data/expected-outcome.md",
    ];
    const claimText = claimPaths.map(readRepoFile).join("\n");
    expect(ownerPaths).toHaveLength(7);

    expect(claimText).toMatch(/DOM\/style\/bbox/);
    expect(claimText).toMatch(/pixelmatch/);
    expect(claimText).toMatch(/changed-pixel ratio.*0\.5%/i);
    expect(claimText).toMatch(/prototype\/formal\/diff/);
    expect(claimText).toMatch(/fullPage:\s*true/);

    const helper = readRepoFile(
      "frontend/tests/pixel-parity/report-parity-helpers.ts",
    );
    expect(helper).toMatch(/from "pixelmatch"/);
    expect(helper).toMatch(/document\.fonts\.ready/);
    expect(helper).toMatch(/clock\.setFixedTime/);
    expect(helper).toMatch(/animation:\s*none !important/);
    expect(helper).toMatch(/transition:\s*none !important/);
    expect(helper).toMatch(/threshold:\s*0\.1/);
    expect(helper).toMatch(/maxChangedRatio = 0\.005/);
    expect(helper).toMatch(/changed \/ \(width \* height\)/);
    expect(helper).toMatch(/formal\.png/);
    expect(helper).toMatch(/prototype\.png/);
    expect(helper).toMatch(/diff\.png/);

    for (const path of [
      "frontend/tests/pixel-parity/generating.spec.ts",
      "frontend/tests/pixel-parity/report.spec.ts",
    ] as const) {
      const source = readRepoFile(path);
      expect(source).toMatch(/configureDeterministicPage/);
      expect(source).toMatch(/injectPrototypeReportFixture/);
      expect(source).toMatch(/normalizedText/);
      expect(source).toMatch(/surfaceSnapshot/);
      expect(source).toMatch(/expectSurfaceParity/);
      expect(source).toMatch(/expectPixelParity/);
      expect(source).toMatch(/deviceScaleFactor:\s*1/);
      expect(source).toMatch(/timezoneId:\s*"UTC"/);
    }

    const reportParity = readRepoFile(
      "frontend/tests/pixel-parity/report.spec.ts",
    );
    expect(reportParity).toMatch(/report root absolute viewport geometry/);
    expect(reportParity).toMatch(/prototype-baseline/);
    expect(reportParity).toMatch(/ready-well-prepared/);
    expect(reportParity).toMatch(/failed-context-too-large/);

    const generatingParity = readRepoFile(
      "frontend/tests/pixel-parity/generating.spec.ts",
    );
    expect(generatingParity).toMatch(/data-report-status='generating'/);
    expect(generatingParity).toMatch(/generating-progress/);
    expect(generatingParity).toMatch(/failed-context-too-large/);

    const trigger = readRepoFile(
      "test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative/scripts/trigger.sh",
    );
    expect(trigger).toContain("src/app/screens/report/__tests__/preflight.test.ts");
  });
});

describe("frontend-report-dashboard/001 Phase 9 direct-start evidence", () => {
  it("keeps replay owners and P0.057 on the generated-client direct-start flow", () => {
    const currentOwnerFiles = [
      "docs/spec/frontend-report-dashboard/spec.md",
      "docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/plan.md",
      "docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/checklist.md",
      "docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/test-plan.md",
      "docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/test-checklist.md",
      "docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/bdd-plan.md",
      "docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/bdd-checklist.md",
      "test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/README.md",
      "test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/data/expected-outcome.md",
      "test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/data/seed-input.md",
      "test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/verify.sh",
    ];
    const ownerText = currentOwnerFiles.map(readRepoFile).join("\n");
    expect(ownerText).not.toMatch(/autoStartPractice/);
    expect(ownerText).not.toMatch(/workspace auto-start/i);
    expect(ownerText).not.toMatch(/route:\s*["'`]workspace["'`]/);

    const handlers = readRepoFile(
      "frontend/src/app/screens/report/useReplayCtaHandlers.ts",
    );
    expect(handlers).toContain(
      "startPracticeFromParams(runtime.client, params, lang)",
    );
    expect(handlers).toContain(
      'navigate({ name: "practice", params: started.params })',
    );
    expect(handlers.match(/route: "report"/g)).toHaveLength(2);

    const startPractice = readRepoFile(
      "frontend/src/app/interview-context/startPractice.ts",
    );
    expect(startPractice).toContain("client.createPracticePlan(");
    expect(startPractice).toContain("client.startPracticeSession(");

    const trigger = readRepoFile(
      "test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/scripts/trigger.sh",
    );
    expect(trigger).toContain("src/app/screens/report/__tests__/preflight.test.ts");
  });
});

describe("frontend-report-dashboard/001 Phase 10 P0.056 evidence", () => {
  it("keeps P0.056 claims within its four focused owner test files", () => {
    const scenarioRoot =
      "test/scenarios/e2e/p0-056-generating-to-report-happy-path";
    const readScenario = (relativePath: string) =>
      readRepoFile(`${scenarioRoot}/${relativePath}`);
    const trigger = readScenario("scripts/trigger.sh");
    const verify = readScenario("scripts/verify.sh");
    const readme = readScenario("README.md");
    const currentClaims = [
      readme,
      readScenario("data/seed-input.md"),
      readScenario("data/expected-outcome.md"),
      verify,
      readRepoFile(
        "docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/bdd-plan.md",
      ),
      readRepoFile(
        "docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/bdd-checklist.md",
      ),
    ].join("\n");

    for (const unsupportedClaim of [
      /end-to-end mount/i,
      /transcripts/i,
      /resumeVersionId/,
      /getResumeVersion/,
      /fake timer 推进/,
      /切 dark \+ customAccent/,
      /3 次（轮询）\+ 1/,
    ]) {
      expect(currentClaims).not.toMatch(unsupportedClaim);
    }
    expect(readme).toContain("four focused owner test files");
    expect(readme).toContain("composed backend/frontend gate");
    expect(readme).toContain("not a single browser journey");

    const focusedFiles = [
      "src/app/screens/report/__tests__/preflight.test.ts",
      "src/app/screens/generating/__tests__/useReportGenerationPoll.test.tsx",
      "src/app/screens/generating/__tests__/GeneratingScreen.test.tsx",
      "src/app/screens/report/__tests__/ConversationReport.test.tsx",
    ];
    for (const path of focusedFiles) {
      expect(trigger).toContain(path);
      expect(verify).toContain(path.split("/").at(-1));
    }
  });
});

describe("frontend-report-dashboard/001 Phase 11 P0.058 evidence", () => {
  it("keeps P0.058 claims within its focused failure contracts", () => {
    const scenarioRoot =
      "test/scenarios/e2e/p0-058-report-failure-and-missing-session";
    const readScenario = (relativePath: string) =>
      readRepoFile(`${scenarioRoot}/${relativePath}`);
    const trigger = readScenario("scripts/trigger.sh");
    const verify = readScenario("scripts/verify.sh");
    const readme = readScenario("README.md");
    const currentClaims = [
      readme,
      readScenario("data/seed-input.md"),
      readScenario("data/expected-outcome.md"),
      verify,
      readRepoFile(
        "docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/bdd-plan.md",
      ),
      readRepoFile(
        "docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/bdd-checklist.md",
      ),
    ].join("\n");

    for (const unsupportedClaim of [
      /GeneratingScreen `timeout` state surfaces GeneratingErrorState/,
      /3 次 timeout/,
      /分六子场景/,
      /console\.log \/ URL search params \/ localStorage \/ sessionStorage \/ telemetry/,
      /with backend 404/i,
      /errorCode.*URL params.*raw provider/i,
    ]) {
      expect(currentClaims).not.toMatch(unsupportedClaim);
    }
    expect(readme).toContain("seven focused frontend Vitest files");
    expect(readme).toContain("REPORT_CONTEXT_TOO_LARGE");

    const focusedFiles = [
      "src/app/screens/report/__tests__/preflight.test.ts",
      "src/app/screens/report/__tests__/ReportFailureState.test.tsx",
      "src/app/screens/report/__tests__/ReportMissingSessionState.test.tsx",
      "src/app/screens/report/__tests__/useFeedbackReport.test.tsx",
      "src/app/screens/report/__tests__/ConversationReport.test.tsx",
      "src/app/screens/generating/__tests__/useReportGenerationPoll.test.tsx",
      "src/app/screens/generating/__tests__/GeneratingScreen.test.tsx",
    ];
    for (const path of focusedFiles) {
      expect(trigger).toContain(path);
      expect(verify).toContain(path.split("/").at(-1));
    }
  });
});
