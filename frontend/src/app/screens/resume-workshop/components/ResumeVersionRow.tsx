import type { FC } from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";
import type { UiResumeVersion } from "../adapters/resume";
import { ResumeWorkshopIcon } from "./ResumeWorkshopIcon";

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
  const { lang, t } = useI18n();
  const isMaster = version.tag === "MASTER";
  const meta = [
    version.date,
    `${version.bullets} ${lang === "en" ? "bullets" : "条 bullet"}`,
    version.accepted > 0
      ? `${version.accepted} ${lang === "en" ? "accepted" : "已采纳"}`
      : null,
  ].filter((item): item is string => item !== null);
  return (
    <li
      data-testid={`resume-version-row-${version.id}`}
      data-tag={version.tag}
      data-indent={indent ? "1" : "0"}
      data-variant={variant}
      className={`ei-resume-workshop-version-row ei-resume-workshop-version-row--${variant}`}
    >
      {indent ? (
        <span
          className="ei-resume-workshop-version-indent"
          data-testid={`resume-version-row-${version.id}-indent`}
        >
          └
        </span>
      ) : null}
      <span className="ei-resume-workshop-version-main">
        <ResumeWorkshopIcon
          name={isMaster ? "resume" : "briefcase"}
          size={13}
          data-testid={`resume-version-row-${version.id}-icon`}
        />
        <span className="ei-resume-workshop-version-name">{version.name}</span>
        <span className="ei-resume-workshop-version-tag">{version.tag}</span>
        <span className="ei-resume-workshop-version-meta">
          {meta.join(" · ")}
        </span>
      </span>
      <span className="ei-resume-workshop-version-actions">
        {version.match !== null ? (
          <span
            data-testid={`resume-version-row-${version.id}-match`}
            className="ei-resume-workshop-version-match"
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
          <ResumeWorkshopIcon name="arrowRight" size={10} />
        </button>
      </span>
    </li>
  );
};
