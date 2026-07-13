import type {
  Confidence,
  DimensionStatus,
  FeedbackReport,
  GenerationProvenance,
  ReadinessTier,
} from "../../../api/generated/types";

const READINESS = new Set<ReadinessTier>([
  "not_ready",
  "needs_practice",
  "basically_ready",
  "well_prepared",
]);
const STATUSES = new Set<DimensionStatus>(["strong", "meets_bar", "needs_work"]);
const CONFIDENCES = new Set<Confidence>(["high", "medium", "low"]);
const ACTION_TYPES = new Set(["retry_current_round", "next_round", "review_evidence"]);
export const ACTION_LABEL_WIRE_MAX_CODE_POINTS = 200;
const REPORT_LANGUAGES = new Set(["en", "zh-CN"]);
const ROUND_TYPES = new Set([
  "hr",
  "technical",
  "manager",
  "cross_functional",
  "culture",
  "final",
  "other",
]);
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export type ReadyFeedbackReport = FeedbackReport & {
  status: "ready";
  errorCode: null;
  summary: string;
  preparednessLevel: ReadinessTier;
  provenance: GenerationProvenance;
};

export function isValidReadyReport(report: FeedbackReport): report is ReadyFeedbackReport {
  if (
    !exactKeys(report, [
      "context", "createdAt", "dimensionAssessments", "errorCode", "highlights",
      "id", "issues", "nextActions", "preparednessLevel", "provenance",
      "retryFocusDimensionCodes", "sessionId", "status", "summary",
      "targetJobId", "updatedAt",
    ]) ||
    report.status !== "ready" ||
    report.errorCode !== null ||
    !uuid(report.id) ||
    !uuid(report.sessionId) ||
    !uuid(report.targetJobId) ||
    !text(report.summary) ||
    report.summary.length > 360 ||
    !report.preparednessLevel ||
    !READINESS.has(report.preparednessLevel) ||
    !validProvenance(report.provenance) ||
    report.provenance.language !== report.context?.language ||
    !dateTime(report.createdAt) ||
    !dateTime(report.updatedAt) ||
    !validContext(report)
  ) return false;

  const dimensions = report.dimensionAssessments;
  const codes = new Set<string>();
  if (!Array.isArray(dimensions) || dimensions.length === 0 || dimensions.length > 6) return false;
  for (const item of dimensions) {
    if (
      !exactKeys(item, ["code", "confidence", "label", "status"]) ||
      !text(item.code) ||
      !/^[a-z][a-z0-9_]{1,63}$/.test(item.code) ||
      !text(item.label) ||
      item.label.length > 48 ||
      codes.has(item.code) ||
      !STATUSES.has(item.status) ||
      !CONFIDENCES.has(item.confidence)
    ) return false;
    codes.add(item.code);
  }
  if (
    (report.preparednessLevel === "not_ready" ||
      report.preparednessLevel === "needs_practice") &&
    !dimensions.some((item) => item.status === "needs_work")
  ) return false;

  const highlights = report.highlights;
  const issues = report.issues;
  if (
    !Array.isArray(highlights) ||
    !Array.isArray(issues) ||
    highlights.length > 4 ||
    issues.length > 4 ||
    highlights.length + issues.length > 6 ||
    highlights.length + issues.length === 0
  ) return false;
  if (![...highlights, ...issues].every((item) =>
    exactKeys(item, ["confidence", "dimensionCode", "evidence"]) &&
    codes.has(item.dimensionCode) &&
    text(item.evidence) &&
    item.evidence.length <= 240 &&
    CONFIDENCES.has(item.confidence)
  )) return false;

  const highlightCodes = new Set(highlights.map((item) => item.dimensionCode));
  const issueCodes = new Set(issues.map((item) => item.dimensionCode));
  for (const item of dimensions) {
    if (!highlightCodes.has(item.code) && !issueCodes.has(item.code)) return false;
    if (item.status === "strong" && !highlightCodes.has(item.code)) return false;
    if (item.status === "needs_work" && !issueCodes.has(item.code)) return false;
  }

  const actions = report.nextActions;
  if (!Array.isArray(actions) || actions.length === 0 || actions.length > 2) return false;
  const actionTypes = new Set<string>();
  for (const action of actions) {
    if (
      !exactKeys(action, ["label", "type"]) ||
      !ACTION_TYPES.has(action.type) ||
      !validActionLabel(action.label, report.context.language) ||
      actionTypes.has(action.type)
    ) return false;
    if (action.type === "next_round" && !report.context.hasNextRound) return false;
    actionTypes.add(action.type);
  }

  const focus = report.retryFocusDimensionCodes;
  if (
    !Array.isArray(focus) ||
    focus.length > 6 ||
    new Set(focus).size !== focus.length
  ) return false;
  if (!actionTypes.has("retry_current_round") && focus.length > 0) return false;
  return focus.every((code) =>
    dimensions.some((item) => item.code === code && item.status === "needs_work") &&
    issueCodes.has(code)
  );
}

function validContext(report: FeedbackReport): boolean {
  const context = report.context;
  return Boolean(
    context &&
    exactKeys(context, [
      "hasNextRound", "language", "resumeDisplayName", "resumeId", "roundId",
      "roundName", "roundSequence", "roundType", "sourcePlanId",
      "targetJobCompany", "targetJobTitle",
    ]) &&
    uuid(context.sourcePlanId) &&
    text(context.targetJobTitle) &&
    text(context.targetJobCompany) &&
    uuid(context.resumeId) &&
    text(context.resumeDisplayName) &&
    text(context.roundId) &&
    ROUND_TYPES.has(context.roundType) &&
    new RegExp(`^round-${context.roundSequence}-${context.roundType}$`).test(context.roundId) &&
    text(context.roundName) &&
    REPORT_LANGUAGES.has(context.language) &&
    Number.isInteger(context.roundSequence) &&
    context.roundSequence > 0 &&
    typeof context.hasNextRound === "boolean"
  );
}

function validProvenance(value: GenerationProvenance | null): value is GenerationProvenance {
  return Boolean(
    value &&
    exactKeys(value, [
      "dataSourceVersion", "featureFlag", "language", "modelId",
      "promptVersion", "rubricVersion",
    ]) &&
    text(value.promptVersion) &&
    text(value.rubricVersion) &&
    text(value.modelId) &&
    text(value.language) &&
    text(value.featureFlag) &&
    text(value.dataSourceVersion)
  );
}

function exactKeys(value: unknown, expected: readonly string[]): boolean {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    return false;
  }
  const actual = Object.keys(value).sort();
  const wanted = [...expected].sort();
  return actual.length === wanted.length && actual.every((key, index) => key === wanted[index]);
}

function uuid(value: unknown): value is string {
  return typeof value === "string" && UUID.test(value);
}

function dateTime(value: unknown): value is string {
  return typeof value === "string" && !Number.isNaN(Date.parse(value));
}

function text(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

function validActionLabel(value: unknown, language: string): value is string {
  if (!text(value)) return false;
  const codePoints = [...value].length;
  if (codePoints > ACTION_LABEL_WIRE_MAX_CODE_POINTS) return false;
  if (language === "en") return value.trim().split(/\s+/u).length <= 24;
  if (language === "zh-CN") return codePoints <= 64;
  return false;
}
