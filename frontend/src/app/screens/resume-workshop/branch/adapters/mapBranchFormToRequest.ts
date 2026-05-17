import type { BranchResumeVersionRequest } from "../../../../../api/generated/types";
import type { BranchFocus, ResumeBranchFormDraft } from "../ResumeBranchFlow";

export interface BranchSubmitContext {
  parentVersionId: string;
  targetJobId: string;
}

/**
 * Project the local form draft (name, target text, focus enum, seed enum)
 * plus the resolved source context (master `parentVersionId` + target job
 * `targetJobId`) into the OpenAPI wire shape expected by
 * `branchResumeVersion`.
 *
 * The mapper is the only place that performs:
 *   - `displayName` derivation from the trimmed `name` field
 *   - `focusAngle` derivation from the chip key (we always carry the literal
 *     focus key so backend can persist either the slug or its localized label
 *     downstream; the form does not invent translation hints)
 *   - hard-stripping any extra keys that the form ever introduces by mistake
 *     (so contributors cannot accidentally smuggle `versionType`,
 *     `parentVersionId`-overrides, etc. into the wire payload).
 */
export function mapBranchFormToBranchResumeVersionRequest(
  draft: ResumeBranchFormDraft,
  context: BranchSubmitContext,
): BranchResumeVersionRequest {
  if (!context.parentVersionId) {
    throw new Error(
      "mapBranchFormToBranchResumeVersionRequest: parentVersionId is required",
    );
  }
  if (!context.targetJobId) {
    throw new Error(
      "mapBranchFormToBranchResumeVersionRequest: targetJobId is required",
    );
  }
  const displayName = draft.name.trim();
  if (!displayName) {
    throw new Error(
      "mapBranchFormToBranchResumeVersionRequest: displayName must be non-empty",
    );
  }
  const focusAngle = focusToWire(draft.focus, draft.target);
  return {
    parentVersionId: context.parentVersionId,
    targetJobId: context.targetJobId,
    seedStrategy: draft.seed,
    displayName,
    focusAngle,
  };
}

const FOCUS_WIRE_LITERAL: Readonly<Record<BranchFocus, string>> = {
  platform: "platform",
  collaboration: "collaboration",
  fullstack: "fullstack",
  leadership: "leadership",
  custom: "custom",
};

/**
 * `focus` chip enum → wire string. Custom focus carries the trimmed target
 * description so backend has a non-empty bias hint, mirroring the ui-design
 * "Custom — I'll write it" semantics where the user expresses the focus via
 * the target field itself.
 */
function focusToWire(focus: BranchFocus, target: string): string {
  if (focus === "custom") {
    const trimmed = target.trim();
    return trimmed ? `custom:${trimmed}` : "custom";
  }
  return FOCUS_WIRE_LITERAL[focus];
}
