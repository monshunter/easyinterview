import type {
  ApiErrorCode,
  Confidence,
  DimensionStatus,
  ReadinessTier,
} from "../../../api/generated/types";
import type { MessageKey } from "../../i18n/messages";

/** 4-tier readiness label map (D-10). Forbidden 5-tier numeric labels are gone. */
const READINESS_TIER_LABEL: Record<ReadinessTier, MessageKey> = {
  not_ready: "report.readiness.tier.notReady",
  needs_practice: "report.readiness.tier.needsPractice",
  basically_ready: "report.readiness.tier.basicallyReady",
  well_prepared: "report.readiness.tier.wellPrepared",
} as const;

export function readinessTierLabel(tier: ReadinessTier): MessageKey {
  return READINESS_TIER_LABEL[tier];
}

const DIMENSION_STATUS_LABEL: Record<DimensionStatus, MessageKey> = {
  strong: "report.dimension.status.strong",
  meets_bar: "report.dimension.status.meetsBar",
  needs_work: "report.dimension.status.needsWork",
} as const;

export function dimensionStatusLabel(status: DimensionStatus): MessageKey {
  return DIMENSION_STATUS_LABEL[status];
}

const CONFIDENCE_LABEL: Record<Confidence, MessageKey> = {
  high: "report.confidence.high",
  medium: "report.confidence.medium",
  low: "report.confidence.low",
};

export function confidenceLabel(confidence: Confidence): MessageKey {
  return CONFIDENCE_LABEL[confidence];
}

const FAILURE_LABEL_BY_CODE: Partial<Record<ApiErrorCode, MessageKey>> = {
  AI_PROVIDER_TIMEOUT: "report.failureState.errorCode.AI_PROVIDER_TIMEOUT",
  AI_PROVIDER_SECRET_MISSING:
    "report.failureState.errorCode.AI_PROVIDER_SECRET_MISSING",
  AI_PROVIDER_CONFIG_INVALID:
    "report.failureState.errorCode.AI_PROVIDER_CONFIG_INVALID",
  AI_OUTPUT_INVALID: "report.failureState.errorCode.AI_OUTPUT_INVALID",
  AI_FALLBACK_EXHAUSTED: "report.failureState.errorCode.AI_FALLBACK_EXHAUSTED",
  AI_UNSUPPORTED_CAPABILITY:
    "report.failureState.errorCode.AI_UNSUPPORTED_CAPABILITY",
  REPORT_CONTEXT_TOO_LARGE:
    "report.failureState.errorCode.REPORT_CONTEXT_TOO_LARGE",
};

/**
 * Maps a failure errorCode to a report.failureState.errorCode.* i18n key.
 * REPORT_NOT_FOUND is intentionally NOT routed through this map — callers
 * branch to the `failureState.notFound.*` keys instead so the cross-user
 * not-found UI stays visually and semantically distinct from AI_* failures.
 */
export function failureErrorCodeKey(
  code: ApiErrorCode | string | null,
): MessageKey {
  if (code && code in FAILURE_LABEL_BY_CODE) {
    return FAILURE_LABEL_BY_CODE[code as ApiErrorCode]!;
  }
  return "report.failureState.errorCode.UNKNOWN";
}
