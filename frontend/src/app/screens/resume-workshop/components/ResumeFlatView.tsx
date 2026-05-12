import { useMemo, type FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type {
  UiResumeSource,
  UiResumeVersion,
} from "../adapters/resume";

export interface ResumeFlatViewProps {
  sources: UiResumeSource[];
  versions: UiResumeVersion[];
  onOpenVersion: (version: UiResumeVersion) => void;
}

export const ResumeFlatView: FC<ResumeFlatViewProps> = ({
  sources,
  versions,
  onOpenVersion,
}) => {
  const { t } = useI18n();
  const sourceById = useMemo(() => {
    const map = new Map<string, UiResumeSource>();
    for (const source of sources) map.set(source.id, source);
    return map;
  }, [sources]);

  const sorted = useMemo(() => {
    return [...versions].sort((a, b) => {
      const am = a.match;
      const bm = b.match;
      if (am === null && bm !== null) return 1;
      if (bm === null && am !== null) return -1;
      if (am !== null && bm !== null && am !== bm) return bm - am;
      return b.date.localeCompare(a.date);
    });
  }, [versions]);

  if (sorted.length === 0) {
    return (
      <p data-testid="resume-workshop-flat-empty" className="ei-text-body">
        {t("resumeWorkshop.flat.empty")}
      </p>
    );
  }

  return (
    <div role="table" className="ei-resume-workshop-flat">
      <div role="row" className="ei-resume-workshop-flat-header">
        <span role="columnheader">
          {t("resumeWorkshop.flat.headerVersion")}
        </span>
        <span role="columnheader">
          {t("resumeWorkshop.flat.headerOriginal")}
        </span>
        <span role="columnheader">
          {t("resumeWorkshop.flat.headerTarget")}
        </span>
        <span role="columnheader">
          {t("resumeWorkshop.flat.headerMatch")}
        </span>
        <span role="columnheader">
          {t("resumeWorkshop.flat.headerUpdated")}
        </span>
        <span role="columnheader" aria-hidden="true" />
      </div>
      {sorted.map((version) => {
        const source = sourceById.get(version.originalId);
        return (
          <div
            key={version.id}
            role="row"
            data-testid={`resume-flat-row-${version.id}`}
            data-tag={version.tag}
            className="ei-resume-workshop-flat-row"
          >
            <span role="cell">{version.name}</span>
            <span role="cell">{source?.name ?? "—"}</span>
            <span role="cell">{version.target ?? "—"}</span>
            <span role="cell">
              {version.match !== null ? `${version.match}%` : "—"}
            </span>
            <span role="cell">{version.date}</span>
            <span role="cell">
              <button
                type="button"
                onClick={() => onOpenVersion(version)}
                data-testid={`resume-flat-row-${version.id}-open`}
              >
                {t("resumeWorkshop.openVersion")}
              </button>
            </span>
          </div>
        );
      })}
    </div>
  );
};
