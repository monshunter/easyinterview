import type { FC } from "react";

import { useI18n } from "../../../i18n/messages";

interface NotImplementedPlaceholderProps {
  flow: "branch";
}

export const NotImplementedPlaceholder: FC<NotImplementedPlaceholderProps> = ({
  flow,
}) => {
  const { t } = useI18n();
  return (
    <div
      data-testid="resume-workshop-not-implemented"
      data-flow={flow}
      className="ei-screen-card ei-screen-card--placeholder"
    >
      <h2 className="ei-text-title">{t("resumeWorkshop.notImplemented.title")}</h2>
      <p className="ei-text-body">{t("resumeWorkshop.notImplemented.body")}</p>
    </div>
  );
};
