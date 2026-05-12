import { useState, type FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type {
  UiResumeSource,
  UiResumeVersion,
} from "../adapters/resume";
import { ResumeVersionRow } from "./ResumeVersionRow";
import { fireResumeWorkshopToast } from "./toast";

export interface ResumeTreeViewProps {
  sources: UiResumeSource[];
  versionsByAsset: Map<string, UiResumeVersion[]>;
  onOpenVersion: (version: UiResumeVersion) => void;
}

export const ResumeTreeView: FC<ResumeTreeViewProps> = ({
  sources,
  versionsByAsset,
  onOpenVersion,
}) => {
  const { t } = useI18n();
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});

  const toggle = (id: string) =>
    setCollapsed((prev) => ({ ...prev, [id]: !prev[id] }));

  const fireToast = (message: string) =>
    fireResumeWorkshopToast(message, "neutral");

  return (
    <div className="ei-resume-workshop-tree">
      {sources.map((source) => {
        const tree = versionsByAsset.get(source.id) ?? [];
        const hasVersions = tree.length > 0;
        const isCollapsed = !!collapsed[source.id];
        return (
          <section
            key={source.id}
            data-testid={`resume-tree-row-${source.id}`}
            data-source-status={source.status}
            data-collapsed={isCollapsed ? "true" : "false"}
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
                <span className="ei-text-body">
                  {t("resumeWorkshop.tree.versionsCount").replace(
                    "{count}",
                    String(tree.length),
                  )}
                </span>
              </button>
              <div className="ei-resume-workshop-tree-row-actions">
                <button
                  type="button"
                  data-testid={`resume-tree-row-${source.id}-use-as-base`}
                  onClick={() =>
                    fireToast(t("resumeWorkshop.tree.toastSelect"))
                  }
                >
                  {t("resumeWorkshop.tree.useAsBase")}
                </button>
                <button
                  type="button"
                  data-testid={`resume-tree-row-${source.id}-new-version`}
                  onClick={() =>
                    fireToast(t("resumeWorkshop.tree.toastBranch"))
                  }
                >
                  {t("resumeWorkshop.tree.newVersion")}
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
    </div>
  );
};
