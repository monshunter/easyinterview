import type { ReportConversation } from "../../../api/generated/types";
import { isValidReportContext } from "../report/reportContract";

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const REPORT_STATUSES = new Set(["queued", "generating", "ready", "failed"]);
const MESSAGE_ROLES = new Set(["user", "assistant"]);
const CONVERSATION_KEYS = ["context", "messages", "reportId", "reportStatus"];
const MESSAGE_KEYS = ["content", "createdAt", "role", "sequence"];

/**
 * Validates the report-owned transcript as one closed read projection.
 *
 * Empty message arrays are legal: a report can exist even when the durable
 * message projection has no readable turns. Any malformed identity, unknown
 * role, out-of-order sequence, or extra locator fails the whole projection.
 */
export function isValidReportConversation(
  value: unknown,
  expectedReportId: string,
): value is ReportConversation {
  if (!exactKeys(value, CONVERSATION_KEYS)) return false;
  const conversation = value as ReportConversation;
  if (
    !uuid(expectedReportId) ||
    conversation.reportId !== expectedReportId ||
    !uuid(conversation.reportId) ||
    !REPORT_STATUSES.has(conversation.reportStatus) ||
    !isValidReportContext(conversation.context) ||
    !Array.isArray(conversation.messages)
  ) return false;

  let previousSequence = 0;
  return conversation.messages.every((message) => {
    if (!exactKeys(message, MESSAGE_KEYS)) return false;
    if (
      !Number.isInteger(message.sequence) ||
      message.sequence < 1 ||
      message.sequence <= previousSequence ||
      !MESSAGE_ROLES.has(message.role) ||
      !text(message.content) ||
      !dateTime(message.createdAt)
    ) return false;
    previousSequence = message.sequence;
    return true;
  });
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

function text(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

function dateTime(value: unknown): value is string {
  return typeof value === "string" && !Number.isNaN(Date.parse(value));
}
