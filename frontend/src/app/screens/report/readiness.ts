import type {
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
