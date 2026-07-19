import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const APP_ROOT = resolve(__dirname, "..");
const readAppSource = (path: string) => readFileSync(resolve(APP_ROOT, path), "utf8");

const topbarCss = readAppSource("topbar/topbar.css");
const authCss = readAppSource("auth/auth.css");
const screensCss = readAppSource("screens/screens.css");

const cssInventory = [
  ["TopBar navigation", topbarCss, ".ei-topbar-nav-button"],
  ["TopBar language option", topbarCss, ".ei-topbar-lang-option"],
  ["TopBar unauthenticated login", topbarCss, ".ei-topbar-auth-login"],
  ["Auth primary action", authCss, ".ei-auth-cta"],
  ["Home submit", screensCss, ".ei-home-submit"],
  ["Home recent interview action", screensCss, ".ei-home-recent-action"],
  ["Workspace create", screensCss, ".ei-workspace-plan-create"],
  ["Workspace primary action", screensCss, ".ei-workspace-card-primary"],
  ["Workspace delete icon action", screensCss, ".ei-workspace-card-delete"],
  ["Report pending action", screensCss, ".ei-report-pending-cta"],
  ["Report header action", screensCss, ".ei-report-header-cta"],
  ["Report records action", screensCss, ".ei-reports-action"],
  ["Resume create", screensCss, ".ei-resume-workshop-create"],
  ["Resume card open", screensCss, ".ei-resume-workshop-card-open"],
  ["Resume card delete", screensCss, ".ei-resume-workshop-card-delete"],
  ["Mock interview compact delete", screensCss, ".ei-mock-interview-card-delete"],
  ["Practice session action", screensCss, ".ei-practice-session-button"],
  ["Practice finish", screensCss, ".ei-practice-finish-button"],
  ["Practice message retry", screensCss, ".ei-practice-message-retry"],
  ["Practice send", screensCss, ".ei-practice-input-send"],
  ["Resume intake submit", screensCss, ".ei-resume-create-cta-accent"],
  ["Interview plan primary action", screensCss, ".ei-plan-detail-primary-action"],
  ["Interview plan secondary action", screensCss, ".ei-plan-detail-secondary-action"],
  ["Settings theme option", screensCss, ".ei-settings-theme-option"],
  ["Settings save", screensCss, ".ei-settings-primary-action"],
  ["Settings secondary action", screensCss, ".ei-settings-secondary-action"],
  ["Settings danger action", screensCss, ".ei-settings-danger-action"],
  ["Shared transition recovery action", screensCss, ".ei-transition-scene__action"],
] as const;

function radiusForSelector(source: string, selector: string): string | null {
  const blocks = source.matchAll(/([^{}]+)\{([^{}]*)\}/g);
  for (const match of blocks) {
    const selectors = (match[1] ?? "").split(",").map((entry) => entry.trim());
    const body = match[2] ?? "";
    if (!selectors.includes(selector)) continue;
    const radius = body.match(/border-radius:\s*([^;]+);/);
    if (radius?.[1]) return radius[1].trim();
  }
  return null;
}

describe("framed action radius contract (Phase 23)", () => {
  it("routes every framed CSS action through the semantic control radius", () => {
    const drift = cssInventory.flatMap(([label, source, selector]) => {
      const radius = radiusForSelector(source, selector);
      return radius === "var(--ei-radius-control)"
        ? []
        : [`${label} (${selector}) uses ${radius ?? "no radius"}`];
    });

    expect(drift).toEqual([]);
  });

  it("routes inline recovery actions through the same control radius", () => {
    const inlineInventory = [
      [
        "Parse failed reparse",
        readAppSource("screens/parse/ParseScreen.tsx"),
        /data-testid="parse-failed-reparse"[\s\S]{0,500}?borderRadius:\s*"var\(--ei-radius-control\)"/,
      ],
      [
        "Parse failed home",
        readAppSource("screens/parse/ParseScreen.tsx"),
        /data-testid="parse-failed-home"[\s\S]{0,500}?borderRadius:\s*"var\(--ei-radius-control\)"/,
      ],
      [
        "Parse error home",
        readAppSource("screens/parse/ParseScreen.tsx"),
        /onClick=\{handleCancel\}[\s\S]{0,500}?borderRadius:\s*"var\(--ei-radius-control\)"/,
      ],
      [
        "Practice session lost",
        readAppSource("screens/practice/components/PracticeSessionLostState.tsx"),
        /data-testid="practice-session-lost-cta"[\s\S]{0,500}?borderRadius:\s*"var\(--ei-radius-control\)"/,
      ],
      [
        "Practice inline retry",
        readAppSource("screens/practice/components/ErrorState.tsx"),
        /data-testid="practice-error-state-retry"[\s\S]{0,500}?borderRadius:\s*"var\(--ei-radius-control\)"/,
      ],
      [
        "Practice terminal recovery",
        readAppSource("screens/practice/components/TerminalRecovery.tsx"),
        /data-testid="practice-terminal-recovery-cta"[\s\S]{0,700}?borderRadius:\s*"var\(--ei-radius-control\)"/,
      ],
      [
        "Report missing recovery",
        readAppSource("screens/report/components/ReportMissingState.tsx"),
        /data-testid="report-missing-report-cta"[\s\S]{0,500}?borderRadius:\s*"var\(--ei-radius-control\)"/,
      ],
      [
        "Report failure retry",
        readAppSource("screens/report/components/ReportFailureState.tsx"),
        /data-testid="report-failure-retry-cta"[\s\S]{0,500}?borderRadius:\s*"var\(--ei-radius-control\)"/,
      ],
      [
        "Report failure back",
        readAppSource("screens/report/components/ReportFailureState.tsx"),
        /data-testid="report-failure-back-to-workspace"[\s\S]{0,500}?borderRadius:\s*"var\(--ei-radius-control\)"/,
      ],
      [
        "Generating recovery actions",
        readAppSource("screens/generating/components/GeneratingErrorState.tsx"),
        /function buttonStyle[\s\S]{0,900}?borderRadius:\s*"var\(--ei-radius-control\)"/,
      ],
    ] as const;

    const drift = inlineInventory.flatMap(([label, source, pattern]) =>
      pattern.test(source) ? [] : [label],
    );
    expect(drift).toEqual([]);
  });

  it("keeps circular, pill, borderless and non-button surfaces outside the action token", () => {
    expect(radiusForSelector(topbarCss, ".ei-topbar-settings")).toBe(
      "var(--ei-radius-pill)",
    );
    expect(radiusForSelector(topbarCss, ".ei-topbar-control")).toBe("18px");
    expect(radiusForSelector(screensCss, ".ei-screen-card")).toBe(
      "var(--ei-radius-md)",
    );
    expect(screensCss).toMatch(
      /\.ei-resume-detail-back\s*\{[^}]*background:\s*transparent;[^}]*border:\s*0;/,
    );
    expect([topbarCss, authCss, screensCss].join("\n")).not.toMatch(
      /(^|})\s*button\s*\{/m,
    );
  });
});
