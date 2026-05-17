import { type FC, type ReactNode } from "react";

import type { Lang } from "../../../i18n/messages";
import type { AiTransparencyMeta } from "./RightPanel";

export interface PracticeVoiceRightPanelProps {
  lang: Lang;
  strict: boolean;
  aiTransparencyLabel: string;
  aiTransparencyMeta: AiTransparencyMeta;
  finishCta: ReactNode;
}

interface VoiceMetric {
  id: string;
  key: string;
  value: string;
  hint: string;
  tone: "ok" | "warn" | "danger";
  bar: number;
}

/**
 * Voice-mode right panel for PracticeScreen. Mirrors the ui-design voice branch
 * for expression metrics, assisted-mode nudge, AI transparency, retention note,
 * and the shared pinned finish CTA.
 */
export const PracticeVoiceRightPanel: FC<PracticeVoiceRightPanelProps> = ({
  lang,
  strict,
  aiTransparencyLabel,
  aiTransparencyMeta,
  finishCta,
}) => {
  const copy = voicePanelCopy(lang);
  const metrics = voiceMetrics(lang);
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
        <div data-testid="practice-voice-expression-panel">
          <div
            className="ei-label"
            style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 14 }}
          >
            {copy.expressionMetrics}
          </div>
          <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
            {metrics.map((metric) => {
              const toneColor = metricToneColor(metric.tone);
              return (
                <div
                  key={metric.id}
                  data-testid={`practice-voice-expression-metric-${metric.id}`}
                >
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "baseline",
                      marginBottom: 5,
                      gap: 10,
                    }}
                  >
                    <div style={{ fontSize: 13, color: "var(--ei-color-fg-secondary)" }}>
                      {metric.key}
                    </div>
                    <div
                      style={{
                        fontFamily: "var(--ei-font-mono)",
                        fontSize: 15,
                        color: toneColor,
                        fontWeight: 500,
                      }}
                    >
                      {metric.value}
                    </div>
                  </div>
                  <div
                    style={{
                      height: 3,
                      background: "var(--ei-color-rule-strong)",
                      borderRadius: 2,
                      overflow: "hidden",
                    }}
                  >
                    <div
                      style={{
                        width: `${metric.bar * 100}%`,
                        height: "100%",
                        background: toneColor,
                      }}
                    />
                  </div>
                  <div
                    style={{
                      fontSize: 11.5,
                      color: "var(--ei-color-fg-tertiary)",
                      marginTop: 4,
                      fontFamily: "var(--ei-font-mono)",
                    }}
                  >
                    {metric.hint}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {!strict && (
          <div
            data-testid="practice-voice-nudge"
            style={{
              marginTop: 26,
              padding: 14,
              background: "var(--ei-color-bg-card)",
              border: "1px dotted var(--ei-color-rule-strong)",
              borderRadius: 3,
            }}
          >
            <div
              className="ei-label"
              style={{ color: "var(--ei-color-fg-tertiary)", marginBottom: 8 }}
            >
              {copy.nudgeLabel}
            </div>
            <div
              style={{
                fontSize: 13,
                color: "var(--ei-color-fg-primary)",
                lineHeight: 1.55,
              }}
            >
              {copy.nudgeText}
            </div>
          </div>
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

        <div
          data-testid="practice-voice-audio-retention-note"
          style={{
            marginTop: 16,
            fontSize: 11,
            color: "var(--ei-color-fg-muted)",
            lineHeight: 1.5,
            fontFamily: "var(--ei-font-mono)",
          }}
        >
          {copy.retentionNote}
        </div>
      </div>

      {finishCta}
    </div>
  );
};

function metricToneColor(tone: VoiceMetric["tone"]): string {
  switch (tone) {
    case "ok":
      return "var(--ei-color-ok)";
    case "warn":
      return "var(--ei-color-warn)";
    case "danger":
      return "var(--ei-color-danger)";
    default:
      return "var(--ei-color-fg-secondary)";
  }
}

function voiceMetrics(lang: Lang): VoiceMetric[] {
  return lang === "en"
    ? [
      { id: "wpm", key: "Words / min", value: "186", hint: "steady 160-200 wpm", tone: "ok", bar: 0.7 },
      { id: "pauses", key: "Long pauses", value: "2", hint: "2 over 1.5s", tone: "warn", bar: 0.5 },
      { id: "fillers", key: "Fillers", value: "4", hint: "um x2 · basically x2", tone: "danger", bar: 0.6 },
      { id: "volume", key: "Volume", value: "stable", hint: "no drop-offs", tone: "ok", bar: 0.78 },
    ]
    : [
      { id: "wpm", key: "语速", value: "186", hint: "稳定在 160-200 字/分", tone: "ok", bar: 0.7 },
      { id: "pauses", key: "长停顿", value: "2", hint: "本题 2 次超过 1.5 秒", tone: "warn", bar: 0.5 },
      { id: "fillers", key: "口头禅", value: "4", hint: "嗯 x2 · 就是 x2", tone: "danger", bar: 0.6 },
      { id: "volume", key: "音量", value: "稳定", hint: "没有明显衰减", tone: "ok", bar: 0.78 },
    ];
}

function voicePanelCopy(lang: Lang): {
  expressionMetrics: string;
  nudgeLabel: string;
  nudgeText: string;
  retentionNote: string;
} {
  return lang === "en"
    ? {
      expressionMetrics: "Expression metrics",
      nudgeLabel: "GENTLE NUDGE",
      nudgeText:
        "You're midway through the situation. Name the concrete action you took before moving to the response.",
      retentionNote: "audio stays on-device during the session · deleted after report",
    }
    : {
      expressionMetrics: "表达层指标",
      nudgeLabel: "现场提示",
      nudgeText:
        "当前在讲情境。试着先说一句你采取的具体行动，再切到对方的反应。",
      retentionNote: "音频仅在本次会话缓存 · 报告生成后自动删除",
    };
}
