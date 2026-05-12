import type { FC } from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";

export interface ComingSoonTabProps {
  /** "rewrites" or "edit" — controls the i18n message and testid suffix. */
  variant: "rewrites" | "edit";
}

const MESSAGE_KEY: Record<ComingSoonTabProps["variant"], MessageKey> = {
  rewrites: "resumeWorkshop.detail.comingSoonRewrites",
  edit: "resumeWorkshop.detail.comingSoonEdit",
};

export const ComingSoonTab: FC<ComingSoonTabProps> = ({ variant }) => {
  const { t } = useI18n();
  return (
    <div
      data-testid={`resume-detail-tab-content-coming-soon-${variant}`}
      className="ei-screen-card ei-resume-detail-coming-soon"
      role="tabpanel"
    >
      <p className="ei-text-body">{t(MESSAGE_KEY[variant])}</p>
    </div>
  );
};
