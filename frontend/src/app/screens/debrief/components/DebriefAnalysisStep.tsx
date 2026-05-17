import { useState, type FC } from "react";

import { useI18n } from "../../../i18n/messages";
import type { Debrief, DebriefSelectedContext } from "../types";

interface DebriefAnalysisStepProps {
  debrief: Debrief;
  selectedContext: DebriefSelectedContext;
  onAdvance: () => void;
}

/**
 * Phase 6.1 — Step 1 analysis renderer. Surfaces risk items + 3 dimension
 * comparison cards (target JD / mock report / resume evidence) + the
 * "About this analysis" provenance disclosure. Deliberately skips
 * `nextRoundChecklist` / `thankYouDraft` — Phase 6.4 in the long-term
 * roadmap, intentionally out of P0 scope per spec §3.3.
 */
export const DebriefAnalysisStep: FC<DebriefAnalysisStepProps> = ({
  debrief,
  selectedContext,
  onAdvance,
}) => {
  const { t } = useI18n();
  const [showProvenance, setShowProvenance] = useState(false);
  const riskItems = debrief.riskItems ?? [];
  const provenance = debrief.provenance;

  const dimensions = [
    {
      key: "targetJob",
      eyebrow: t("debrief.contextStrip.targetJobLabel"),
      title: selectedContext.targetJob
        ? `${selectedContext.targetJob.companyName ?? ""} · ${selectedContext.targetJob.title ?? ""}`
        : t("debrief.contextStrip.unset"),
      body: debrief.questions?.[0]?.aiAnalysis ?? "—",
    },
    {
      key: "mockSession",
      eyebrow: t("debrief.contextStrip.mockSessionLabel"),
      title: selectedContext.mockSession?.id ?? t("debrief.contextStrip.unset"),
      body: debrief.questions?.[1]?.aiAnalysis ?? "—",
    },
    {
      key: "resume",
      eyebrow: t("debrief.contextStrip.resumeLabel"),
      title:
        selectedContext.resumeAsset?.title ??
        selectedContext.resumeVersion?.id ??
        t("debrief.contextStrip.unset"),
      body: debrief.questions?.[2]?.aiAnalysis ?? "—",
    },
  ];

  return (
    <section
      className="ei-debrief-analysis"
      data-testid="debrief-analysis-step"
    >
      <div className="ei-label">{t("debrief.analysis.eyebrow")}</div>
      <section
        className="ei-debrief-analysis__risks"
        data-testid="debrief-analysis-risks"
      >
        <div className="ei-label">{t("debrief.analysis.risksEyebrow")}</div>
        {riskItems.length === 0 ? (
          <p data-testid="debrief-analysis-risks-empty">
            {t("debrief.analysis.risksEmpty")}
          </p>
        ) : (
          <ul>
            {riskItems.map((risk, idx) => (
              <li
                key={`${risk.label}-${idx}`}
                data-testid="debrief-analysis-risk-item"
                data-severity={risk.severity}
              >
                <span className="ei-label">
                  {t(`debrief.severity.${risk.severity}` as never)}
                </span>
                <span>{risk.label}</span>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section
        className="ei-debrief-analysis__dimensions"
        data-testid="debrief-analysis-dimensions"
      >
        <div className="ei-label">{t("debrief.analysis.dimensionsEyebrow")}</div>
        <div>
          {dimensions.map((dim) => (
            <article
              key={dim.key}
              data-testid={`debrief-analysis-dimension-${dim.key}`}
            >
              <div className="ei-label">{dim.eyebrow}</div>
              <h3>{dim.title}</h3>
              <p>{dim.body}</p>
            </article>
          ))}
        </div>
      </section>

      <section
        className="ei-debrief-analysis__provenance"
        data-testid="debrief-analysis-provenance"
      >
        <button
          type="button"
          data-testid="debrief-analysis-provenance-toggle"
          aria-expanded={showProvenance}
          onClick={() => setShowProvenance((v) => !v)}
        >
          {t("debrief.analysis.provenance.title")}
        </button>
        {showProvenance && provenance && (
          <dl data-testid="debrief-analysis-provenance-fields">
            <div>
              <dt>{t("debrief.analysis.provenance.promptVersion")}</dt>
              <dd>{provenance.promptVersion}</dd>
            </div>
            <div>
              <dt>{t("debrief.analysis.provenance.rubricVersion")}</dt>
              <dd>{provenance.rubricVersion}</dd>
            </div>
            <div>
              <dt>{t("debrief.analysis.provenance.modelId")}</dt>
              <dd>{provenance.modelId}</dd>
            </div>
            <div>
              <dt>{t("debrief.analysis.provenance.language")}</dt>
              <dd>{provenance.language}</dd>
            </div>
            <div>
              <dt>{t("debrief.analysis.provenance.featureFlag")}</dt>
              <dd>{provenance.featureFlag}</dd>
            </div>
            <div>
              <dt>{t("debrief.analysis.provenance.dataSourceVersion")}</dt>
              <dd>{provenance.dataSourceVersion}</dd>
            </div>
          </dl>
        )}
      </section>

      <div className="ei-debrief-analysis__cta">
        <button
          type="button"
          className="ei-debrief-btn ei-debrief-btn--accent"
          data-testid="debrief-analysis-advance"
          onClick={onAdvance}
        >
          {t("debrief.analysis.cta")}
        </button>
      </div>
    </section>
  );
};
