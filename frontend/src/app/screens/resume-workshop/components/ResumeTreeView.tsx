import { useState, type FC } from "react";

import { useI18n, type MessageKey } from "../../../i18n/messages";
import type {
  UiResumeSource,
  UiResumeVersion,
} from "../adapters/resume";
import { ResumeWorkshopIcon } from "./ResumeWorkshopIcon";
import { ResumeVersionRow } from "./ResumeVersionRow";

export interface ResumeTreeViewProps {
  sources: UiResumeSource[];
  versionsByAsset: Map<string, UiResumeVersion[]>;
  onOpenVersion: (version: UiResumeVersion) => void;
  selectedSourceId: string | null;
  onSelectSource: (sourceId: string | null) => void;
  onCreate: () => void;
}

export const ResumeTreeView: FC<ResumeTreeViewProps> = ({
  sources,
  versionsByAsset,
  onOpenVersion,
  selectedSourceId,
  onSelectSource,
  onCreate,
}) => {
  const { t } = useI18n();
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});

  const toggle = (id: string) =>
    setCollapsed((prev) => ({ ...prev, [id]: !prev[id] }));

  return (
    <div className="ei-resume-workshop-tree">
      {sources.map((source) => {
        const tree = versionsByAsset.get(source.id) ?? [];
        const hasVersions = tree.length > 0;
        const isCollapsed = !!collapsed[source.id];
        const isSelected = selectedSourceId === source.id;
        return (
          <section
            key={source.id}
            data-testid={`resume-tree-row-${source.id}`}
            data-source-status={source.status}
            data-collapsed={isCollapsed ? "true" : "false"}
            data-selected={isSelected ? "true" : "false"}
            className="ei-resume-workshop-tree-row"
          >
            <header className="ei-resume-workshop-tree-row-header">
              <button
                type="button"
                aria-expanded={!isCollapsed}
                aria-controls={`resume-tree-row-${source.id}-versions`}
                data-testid={`resume-tree-row-${source.id}-toggle`}
                onClick={() => toggle(source.id)}
                className="ei-resume-workshop-tree-toggle"
              >
                <ResumeWorkshopIcon
                  name={isCollapsed ? "chevronRight" : "chevronDown"}
                  size={12}
                />
                <span className="ei-resume-workshop-tree-source-copy">
                  <span className="ei-resume-workshop-tree-title-line">
                    <ResumeWorkshopIcon name="file" size={14} />
                    <span className="ei-text-title">{source.name}</span>
                    <span className="ei-text-label">{source.langTag}</span>
                    {source.status === "archived" ? (
                      <span
                        data-testid={`resume-tree-row-${source.id}-archived`}
                        className="ei-text-label"
                      >
                        {t("resumeWorkshop.tree.archived")}
                      </span>
                    ) : null}
                  </span>
                  <span className="ei-resume-workshop-tree-meta">
                    {sourceTypeLabel(source.type, t)} · {source.createdAt} ·{" "}
                    <span>{source.summary}</span> ·{" "}
                    {t("resumeWorkshop.tree.versionsCount").replace(
                      "{count}",
                      String(tree.length),
                    )}
                  </span>
                </span>
              </button>
              <div className="ei-resume-workshop-tree-row-actions">
                <button
                  type="button"
                  data-testid={`resume-tree-row-${source.id}-use-as-base`}
                  onClick={() => onSelectSource(isSelected ? null : source.id)}
                  data-selected={isSelected ? "true" : "false"}
                >
                  <ResumeWorkshopIcon
                    name={isSelected ? "check" : "plus"}
                    size={11}
                  />
                  {isSelected
                    ? t("resumeWorkshop.tree.selected")
                    : t("resumeWorkshop.tree.useAsBase")}
                </button>
              </div>
            </header>
            {isCollapsed ? null : hasVersions ? (
              <ul
                id={`resume-tree-row-${source.id}-versions`}
                className="ei-resume-workshop-tree-versions"
              >
                {tree.map((version) => (
                  <ResumeVersionRow
                    key={version.id}
                    version={version}
                    onOpen={onOpenVersion}
                    indent
                    variant="tree"
                  />
                ))}
              </ul>
            ) : (
              <p
                data-testid={`resume-tree-row-${source.id}-no-versions`}
                className="ei-text-body"
              >
                {t("resumeWorkshop.tree.noVersions")}
              </p>
            )}
          </section>
        );
      })}
      <button
        type="button"
        className="ei-resume-workshop-upload-another"
        data-testid="resume-workshop-upload-another"
        onClick={onCreate}
      >
        <ResumeWorkshopIcon name="upload" size={14} />
        {t("resumeWorkshop.tree.uploadAnother")}
      </button>
    </div>
  );
};

const sourceTypeLabel = (
  value: string,
  translate: (key: MessageKey) => string,
): string => {
  const normalized = value.toLowerCase();
  if (normalized.includes("upload")) {
    return translate("resumeWorkshop.sourceType.upload");
  }
  if (normalized.includes("paste")) {
    return translate("resumeWorkshop.sourceType.paste");
  }
  if (normalized.includes("guided")) {
    return translate("resumeWorkshop.sourceType.guided");
  }
  return translate("resumeWorkshop.sourceType.unknown");
};
