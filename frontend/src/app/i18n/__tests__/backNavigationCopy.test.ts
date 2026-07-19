import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

import { en } from "../locales/en";
import { zh } from "../locales/zh";

const RETURN_CONTROL_SOURCES = [
  "../../auth/AuthLogoutScreen.tsx",
  "../../screens/parse/ParseScreen.tsx",
  "../../screens/practice/PracticeScreen.tsx",
  "../../screens/practice/components/PracticeSessionLostState.tsx",
  "../../screens/reports/ReportsScreen.tsx",
  "../../screens/generating/GeneratingScreen.tsx",
  "../../screens/generating/components/GeneratingErrorState.tsx",
  "../../screens/report/components/ReportDashboard.tsx",
  "../../screens/report/components/ReportFailureState.tsx",
  "../../screens/report/components/ReportMissingState.tsx",
  "../../screens/report-conversation/ReportConversationScreen.tsx",
  "../../screens/resume-workshop/create/ResumeCreateFlow.tsx",
  "../../screens/resume-workshop/components/ResumeDetailView.tsx",
  "../../screens/resume-workshop/components/NotFoundEmptyState.tsx",
] as const;

const RETIRED_RETURN_KEYS = [
  "auth.backHome",
  "reports.backToPlan",
  "reports.returnToPlan",
  "reports.backToWorkspace",
  "parse.failedHome",
  "parse.errorHome",
  "resumeWorkshop.create.back",
  "resumeWorkshop.detail.back",
  "resumeWorkshop.detail.notFoundCta",
  "practice.sessionLost.cta",
  "practice.terminal.backToPlan",
  "generating.errors.backToReports",
  "generating.errors.backToWorkspace",
  "report.back",
  "report.conversation.back",
  "report.conversation.loading.back",
  "report.conversation.unavailable.back",
  "report.failureState.backToReports",
  "report.failureState.backToWorkspace",
  "report.missingReport.cta",
] as const;

describe("shared secondary-page Back copy", () => {
  it("defines one locale-owned label for Chinese and English", () => {
    expect(zh).toHaveProperty("common.back", "返回");
    expect(en).toHaveProperty("common.back", "Back");
  });

  it.each(RETURN_CONTROL_SOURCES)("uses common.back in %s", (path) => {
    const source = readFileSync(resolve(__dirname, path), "utf8");
    expect(source).toContain('t("common.back")');
  });

  it("retires target-specific return action keys from production sources and catalogs", () => {
    const production = RETURN_CONTROL_SOURCES
      .map((path) => readFileSync(resolve(__dirname, path), "utf8"))
      .join("\n");
    const catalogs = `${JSON.stringify(zh)}\n${JSON.stringify(en)}`;

    for (const key of RETIRED_RETURN_KEYS) {
      expect(production).not.toContain(key);
      expect(catalogs).not.toContain(`"${key}"`);
    }
  });
});
