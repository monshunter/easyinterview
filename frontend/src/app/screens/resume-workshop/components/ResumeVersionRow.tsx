import type { FC } from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";
import type { UiResumeVersion } from "../adapters/resume";

interface ResumeVersionRowProps {
  version: UiResumeVersion;
  onOpen: (version: UiResumeVersion) => void;
  indent?: boolean;
  variant?: "tree" | "flat";
}

export const ResumeVersionRow: FC<ResumeVersionRowProps> = ({
  version,
  onOpen,
  indent,
  variant = "tree",
}) => {
  const { t } = useI18n();
  return (
    <li
      data-testid={`resume-version-row-${version.id}`}
      data-tag={version.tag}
      data-indent={indent ? "1" : "0"}
      data-variant={variant}
      className={`ei-resume-workshop-version-row ei-resume-workshop-version-row--${variant}`}
    >
      <span className="ei-text-body">{version.name}</span>
      <span className="ei-text-label">{version.tag}</span>
      <span className="ei-text-body">{version.date}</span>
      {version.match !== null ? (
        <span
          data-testid={`resume-version-row-${version.id}-match`}
          className="ei-text-label"
        >
          {t("resumeWorkshop.flat.headerMatch" as MessageKey)} {version.match}%
        </span>
      ) : null}
      <button
        type="button"
        onClick={() => onOpen(version)}
        data-testid={`resume-version-row-${version.id}-open`}
      >
        {t("resumeWorkshop.openVersion")}
      </button>
    </li>
  );
};
