import { type FC } from "react";

export interface QuestionCardProps {
  badgeText: string;
  topic: string;
  tags: string[];
  prompt: string;
}

/**
 * Source-level mirror of `ui-design/src/screen-practice.jsx` lines 196-205
 * (current question header). Phase 1 renders skeleton; Phase 2 hooks in
 * `currentTurn.questionText` from `getPracticeSession.currentTurn`.
 */
export const QuestionCard: FC<QuestionCardProps> = ({
  badgeText,
  topic,
  tags,
  prompt,
}) => {
  return (
    <div
      data-testid="practice-question"
      style={{
        padding: "28px 40px 20px",
        borderBottom: "1px solid var(--ei-color-rule)",
        background: "var(--ei-color-bgCard)",
      }}
    >
      <div
        style={{
          display: "flex",
          gap: 8,
          marginBottom: 10,
          flexWrap: "wrap",
        }}
      >
        <span
          data-testid="practice-question-badge"
          className="ei-mono"
          style={{
            display: "inline-flex",
            alignItems: "center",
            padding: "3px 8px",
            borderRadius: 3,
            fontSize: 11.5,
            background: "var(--ei-color-accentSoft)",
            color: "var(--ei-color-accent)",
          }}
        >
          {badgeText}
        </span>
        <span
          data-testid="practice-question-topic"
          className="ei-mono"
          style={{
            display: "inline-flex",
            alignItems: "center",
            padding: "3px 8px",
            borderRadius: 3,
            fontSize: 11.5,
            background: "var(--ei-color-bgSoft)",
            color: "var(--ei-color-ink3)",
          }}
        >
          {topic}
        </span>
        {tags.map((t, idx) => (
          <span
            key={t}
            data-testid={`practice-question-tag-${idx}`}
            className="ei-mono"
            style={{
              display: "inline-flex",
              alignItems: "center",
              padding: "3px 8px",
              borderRadius: 3,
              fontSize: 11.5,
              background: "var(--ei-color-bgSoft)",
              color: "var(--ei-color-ink3)",
            }}
          >
            {t}
          </span>
        ))}
      </div>
      <div
        data-testid="practice-question-prompt"
        className="ei-serif"
        style={{
          fontSize: 22,
          color: "var(--ei-color-ink)",
          lineHeight: 1.35,
        }}
      >
        {prompt}
      </div>
    </div>
  );
};
