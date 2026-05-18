import type { FC } from "react";

import { useRequestAuth } from "../../../auth/useRequestAuth";
import { useI18n } from "../../../i18n/messages";
import type { ResumeWorkshopParams } from "../params";

interface ResumeWorkshopAuthGateProps {
  params: ResumeWorkshopParams;
}

const buildPendingActionParams = (
  params: ResumeWorkshopParams,
): Record<string, string> => {
  const restored: Record<string, string> = {};
  if (params.flow !== "list") restored.flow = params.flow;
  if (params.versionId) restored.versionId = params.versionId;
  if (params.tab) restored.tab = params.tab;
  if (params.branchOriginalId)
    restored.branchOriginalId = params.branchOriginalId;
  if (params.targetJobId) restored.targetJobId = params.targetJobId;
  if (params.createMode) restored.createMode = params.createMode;
  return restored;
};

export const ResumeWorkshopAuthGate: FC<ResumeWorkshopAuthGateProps> = ({
  params,
}) => {
  const { t } = useI18n();
  const requestAuth = useRequestAuth();

  const onSignIn = () => {
    requestAuth({
      type: "open_resume_workshop",
      label: t("resumeWorkshop.auth.pendingLabel"),
      route: "resume_versions",
      params: buildPendingActionParams(params),
    });
  };

  return (
    <div
      data-testid="resume-workshop-auth-gate"
      className="ei-screen-card ei-screen-card--auth-gate"
    >
      <span className="ei-text-label">{t("resumeWorkshop.auth.eyebrow")}</span>
      <h2 className="ei-text-title">{t("resumeWorkshop.auth.title")}</h2>
      <p className="ei-text-body">{t("resumeWorkshop.auth.body")}</p>
      <button
        type="button"
        data-testid="resume-workshop-auth-cta"
        className="ei-auth-cta"
        onClick={onSignIn}
      >
        {t("resumeWorkshop.auth.cta")}
      </button>
    </div>
  );
};
