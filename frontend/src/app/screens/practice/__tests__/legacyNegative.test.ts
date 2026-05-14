import { describe, expect, it } from "vitest";
import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { join, resolve } from "node:path";

const FRONTEND_ROOT = process.cwd();
const REPO_ROOT = resolve(FRONTEND_ROOT, "..");
const PRACTICE_DIR = join(
  FRONTEND_ROOT,
  "src/app/screens/practice",
);

function runtimeFiles(dir: string): string[] {
  const out: string[] = [];
  for (const entry of readdirSync(dir)) {
    const absolute = join(dir, entry);
    const stat = statSync(absolute);
    if (stat.isDirectory()) {
      if (entry === "__tests__") continue;
      out.push(...runtimeFiles(absolute));
      continue;
    }
    if (!/\.(tsx?|css)$/.test(entry)) continue;
    if (/\.test\./.test(entry)) continue;
    out.push(absolute);
  }
  return out;
}

function runtimeText(): string {
  return runtimeFiles(PRACTICE_DIR)
    .map((file) => readFileSync(file, "utf8"))
    .join("\n");
}

describe("practice legacy-negative gates (Phase 5)", () => {
  it("has the practice Playwright pixel-parity spec checked in", () => {
    expect(existsSync(join(FRONTEND_ROOT, "tests/pixel-parity/practice.spec.ts"))).toBe(
      true,
    );
  });

  it("keeps retired prototype data, voice surface, and route/testid names out of runtime code", () => {
    const text = runtimeText();
    expect(text).not.toMatch(/ui-design\/src\/data|window\.EI_DATA/);
    expect(text).not.toMatch(
      /VoiceSessionSurface|PracticeWaveformBars|PracticeAnnotatedWaveform|VoiceExpressionPanel/,
    );
    expect(text).not.toMatch(
      /practice-mode-card-|growth-summary|drill-builder-|mistakes-queue-/,
    );
    expect(text).not.toMatch(
      /data-testid="practice-voice-(?!coming-soon)/,
    );
    expect(text).not.toMatch(/practiceMode\s*[=:]\s*['"]debrief['"]/);
    expect(text).not.toMatch(/切到语音|Switch to voice/);
    expect(text).not.toMatch(/\bgetFeedbackReport\b|\bcreatePracticeVoiceTurn\b/);
    expect(text).not.toMatch(
      /AI_PROVIDER_API_KEY|AI_PROVIDER_BASE_URL|prompt-registry|provider-registry|AIClient|LLM endpoint/,
    );
  });

  it("scenario triggers include the dedicated Phase 4/5 remediation tests", () => {
    const p46 = readFileSync(
      join(
        REPO_ROOT,
        "test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/trigger.sh",
      ),
      "utf8",
    );
    const p47 = readFileSync(
      join(
        REPO_ROOT,
        "test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/scripts/trigger.sh",
      ),
      "utf8",
    );
    expect(p46).toContain("practiceErrors.test.tsx");
    expect(p46).toContain("practiceClientEventConflict.test.tsx");
    expect(p46).toContain("practiceConflict.test.tsx");
    expect(p47).toContain("practicePrivacy.test.tsx");
  });
});
