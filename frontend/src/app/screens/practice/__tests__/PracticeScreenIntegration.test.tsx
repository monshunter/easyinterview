/**
 * @vitest-environment jsdom
 *
 * Item 1.6 — comprehensive integration test for PracticeScreen + the 14
 * components under `practice/components/`. Adds the assertions that the
 * colocated `PracticeScreen.test.tsx` cannot cover:
 *  - i18n zh ↔ en switching live re-renders the static shell
 *  - source files do not import voice surface DOM from `ui-design/`
 *  - source files do not import legacy prototype helpers
 */

import { readFileSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";
import { render, screen, act } from "@testing-library/react";
import type { ReactNode } from "react";

import {
  DisplayPreferencesProvider,
  useDisplayPreferences,
} from "../../../display/DisplayPreferencesProvider";
import { InterviewContextProvider } from "../../../interview-context/InterviewContext";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import type { Route } from "../../../routes";
import { PracticeScreen } from "../PracticeScreen";

const PRACTICE_ROUTE: Route = {
  name: "practice",
  params: {
    sessionId: "01918fa0-0000-7000-8000-000000005000",
    planId: "01918fa0-0000-7000-8000-000000004000",
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    jdId: "jd-1",
    resumeId: "01918fa0-0000-7000-8000-000000001000",
    roundId: "round-tech1",
    mode: "text",
    modality: "text",
    practiceMode: "assisted",
    practiceGoal: "baseline",
    hintUsed: "false",
    hintCount: "0",
  },
};

function withLangProbe(ui: ReactNode) {
  let setLangFn: ((next: "zh" | "en") => void) | null = null;
  function LangProbe() {
    const prefs = useDisplayPreferences();
    setLangFn = prefs.setLang;
    return null;
  }
  return {
    setLang: (next: "zh" | "en") => {
      if (!setLangFn) throw new Error("LangProbe not yet mounted");
      setLangFn(next);
    },
    ...render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <NavigationProvider value={{ navigate: () => undefined }}>
            <LangProbe />
            {ui}
          </NavigationProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

describe("PracticeScreen integration (item 1.6)", () => {
  it("renders en strings by default and switches to zh on lang change", () => {
    const harness = withLangProbe(<PracticeScreen route={PRACTICE_ROUTE} />);
    // en defaults (DEFAULT_LANG = "en")
    expect(screen.getByTestId("practice-input-textarea").getAttribute("placeholder"))
      .toContain("answer");
    expect(screen.getByTestId("practice-rightpanel-cta-finish").textContent)
      .toContain("Finish");
    // switch to zh
    act(() => harness.setLang("zh"));
    expect(screen.getByTestId("practice-input-textarea").getAttribute("placeholder"))
      .toContain("回答");
    expect(screen.getByTestId("practice-rightpanel-cta-finish").textContent)
      .toContain("结束并生成报告");
  });

  it("source files do not import voice surface DOM from ui-design", () => {
    const root = practiceSourceRoot();
    const files = collectSourceFiles(root);
    expect(files.length).toBeGreaterThan(0);
    const banned = [
      "VoiceSessionSurface",
      "PracticeWaveformBars",
      "PracticeAnnotatedWaveform",
      "VoiceExpressionPanel",
    ];
    const offenders: string[] = [];
    for (const file of files) {
      const content = readFileSync(file, "utf8");
      for (const symbol of banned) {
        // Only flag actual import statements / type references, not test
        // assertions that mention the symbol name as a negative gate string.
        const importPattern = new RegExp(
          `import[^;]*\\b${symbol}\\b|from\\s+["'][^"']*ui-design[^"']*["']`,
        );
        const fromUiDesign = new RegExp(
          `import[^;]*ui-design[^;]*${symbol}|from[^;]*ui-design[^;]*\\b${symbol}\\b`,
        );
        if (importPattern.test(content) && fromUiDesign.test(content)) {
          offenders.push(`${file}: imports ${symbol} from ui-design`);
        }
      }
    }
    expect(offenders).toEqual([]);
  });

  it("source files do not import prototype data helpers (data.jsx, EI_DATA, sample helpers)", () => {
    const root = practiceSourceRoot();
    const files = collectSourceFiles(root);
    const offenders: string[] = [];
    const bannedImports = [
      "ui-design/src/data",
      "window.EI_DATA",
      "EI_DATA",
      "getPracticeSampleQuestions",
      "getPracticeSampleTranscript",
      "getPracticeWaveformSamples",
    ];
    for (const file of files) {
      // Only inspect non-test source files for banned imports.
      if (file.includes(".test.") || file.includes("__tests__")) continue;
      const content = readFileSync(file, "utf8");
      for (const symbol of bannedImports) {
        if (content.includes(symbol)) {
          offenders.push(`${file}: references ${symbol}`);
        }
      }
    }
    expect(offenders).toEqual([]);
  });

  it("source files do not reference legacy prototype testids or routes", () => {
    const root = practiceSourceRoot();
    const files = collectSourceFiles(root);
    const offenders: string[] = [];
    const bannedTokens = [
      "practice-mode-card-",
      "growth-summary",
      "drill-builder-",
      "mistakes-queue-",
      "切到语音",
      "Switch to voice",
    ];
    for (const file of files) {
      if (file.includes(".test.") || file.includes("__tests__")) continue;
      const content = readFileSync(file, "utf8");
      for (const tok of bannedTokens) {
        if (content.includes(tok)) {
          offenders.push(`${file}: contains banned token ${tok}`);
        }
      }
    }
    expect(offenders).toEqual([]);
  });
});

function practiceSourceRoot(): string {
  // Walk from the test file location to the practice source root.
  return join(
    process.cwd(),
    "src",
    "app",
    "screens",
    "practice",
  );
}

function collectSourceFiles(dir: string, accumulator: string[] = []): string[] {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    const st = statSync(full);
    if (st.isDirectory()) {
      collectSourceFiles(full, accumulator);
    } else if (
      entry.endsWith(".ts") ||
      entry.endsWith(".tsx") ||
      entry.endsWith(".js") ||
      entry.endsWith(".jsx")
    ) {
      accumulator.push(full);
    }
  }
  return accumulator;
}
