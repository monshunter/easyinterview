import type { Route } from "../../routes";
import { isValidFeedbackReport } from "./reportContract";

/**
 * Resolves the only legal Back destination for Report/Generating.
 * Route params and display text are deliberately absent from the inputs:
 * target identity must come from a valid response for the current report ID.
 */
export function resolveReportBackRoute(
  report: unknown,
  expectedReportId: string,
): Route {
  if (!isValidFeedbackReport(report, expectedReportId)) {
    return { name: "workspace", params: {} };
  }
  return {
    name: "reports",
    params: { targetJobId: report.targetJobId },
  };
}
