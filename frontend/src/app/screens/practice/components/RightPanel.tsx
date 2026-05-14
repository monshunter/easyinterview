import { type FC, type ReactNode } from "react";

import { ExpCard, type ExperienceItem } from "./ExpCard";

export interface AiTransparencyMeta {
  promptVersion: string;
  rubricVersion: string;
  modelId: string;
  language: string;
  personaLabel?: string;
}

export interface RightPanelProps {
  jdLinkLabel: string;
  jdProbesLabel: string;
  jdProbesText: string;
  experienceLabel: string;
  aiTransparencyLabel: string;
  aiTransparencyMeta: AiTransparencyMeta;
  strict: boolean;
  strictBannerText: string;
  experiences: ExperienceItem[];
  finishCta: ReactNode;
}

/**
 * Source-level mirror of `ui-design/src/screen-practice.jsx` lines 260-322
 * (right context panel + pinned bottom CTA). Strict mode replaces the hint /
 * experience block with a strict banner; pinned bottom CTA keeps the flex
 * shrink rule.
 */
export const RightPanel: FC<RightPanelProps> = ({
  jdLinkLabel,
  jdProbesLabel,
  jdProbesText,
  experienceLabel,
  aiTransparencyLabel,
  aiTransparencyMeta,
  strict,
  strictBannerText,
  experiences,
  finishCta,
}) => {
  return (
    <div
      data-testid="practice-rightpanel"
      style={{
        borderLeft: "1px solid var(--ei-color-rule-strong)",
        display: "flex",
        flexDirection: "column",
        background: "var(--ei-color-bg-soft)",
      }}
    >
      <div
        style={{
          flex: 1,
          overflowY: "auto",
          padding: "20px 18px",
        }}
      >
        <div
          data-testid="practice-rightpanel-jd"
          style={{
            padding: 12,
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 2,
            marginBottom: 14,
          }}
        >
          <div
            className="ei-label"
            style={{
              color: "var(--ei-color-fg-tertiary)",
              marginBottom: 4,
            }}
          >
            {jdLinkLabel}
          </div>
          <div
            style={{
              fontSize: 11.5,
              color: "var(--ei-color-fg-tertiary)",
              fontFamily: "var(--ei-font-mono)",
              marginBottom: 4,
            }}
          >
            {jdProbesLabel}
          </div>
          <div
            style={{
              fontSize: 13,
              color: "var(--ei-color-fg-primary)",
              lineHeight: 1.55,
            }}
          >
            {jdProbesText}
          </div>
        </div>

        {strict ? (
          <div
            data-testid="practice-rightpanel-strict-banner"
            style={{
              padding: "10px 12px",
              background: "var(--ei-color-bg-soft)",
              border: "1px solid var(--ei-color-rule-strong)",
              borderRadius: 2,
              marginBottom: 14,
            }}
          >
            <div
              style={{
                fontSize: 11,
                color: "var(--ei-color-fg-tertiary)",
                fontFamily: "var(--ei-font-mono)",
                lineHeight: 1.65,
              }}
            >
              {strictBannerText}
            </div>
          </div>
        ) : (
          <>
            <div
              data-testid="practice-rightpanel-experience-label"
              className="ei-label"
              style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 10 }}
            >
              {experienceLabel}
            </div>
            <div
              style={{ display: "flex", flexDirection: "column", gap: 8 }}
            >
              {experiences.map((exp, idx) => (
                <ExpCard
                  key={exp.id}
                  index={idx}
                  title={exp.title}
                  meta={exp.meta}
                  hot={exp.hot}
                />
              ))}
            </div>
          </>
        )}

        <div
          data-testid="practice-rightpanel-ai-transparency"
          style={{
            borderTop: "1px dotted var(--ei-color-rule-strong)",
            marginTop: 16,
            paddingTop: 14,
          }}
        >
          <div
            className="ei-label"
            style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 8 }}
          >
            {aiTransparencyLabel}
          </div>
          <div
            style={{
              fontSize: 11.5,
              color: "var(--ei-color-fg-tertiary)",
              lineHeight: 1.55,
              fontFamily: "var(--ei-font-mono)",
            }}
          >
            prompt {aiTransparencyMeta.promptVersion}
            <br />
            rubric {aiTransparencyMeta.rubricVersion}
            <br />
            model · {aiTransparencyMeta.modelId}
            <br />
            lang · {aiTransparencyMeta.language}
            {aiTransparencyMeta.personaLabel ? (
              <>
                <br />
                role · {aiTransparencyMeta.personaLabel}
              </>
            ) : null}
          </div>
        </div>
      </div>

      {finishCta}
    </div>
  );
};
