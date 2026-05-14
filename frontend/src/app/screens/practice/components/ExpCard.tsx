import { type FC } from "react";

export interface ExperienceItem {
  id: string;
  title: string;
  meta: string;
  hot?: boolean;
}

export interface ExpCardProps {
  index: number;
  title: string;
  meta: string;
  hot?: boolean;
}

/**
 * Source-level mirror of `ui-design/src/screen-practice.jsx::ExpCard` lines
 * 643-651. Phase 1 renders without data; Phase 3.4 wires real experience
 * candidates from InterviewContext / future resume tailoring.
 */
export const ExpCard: FC<ExpCardProps> = ({ index, title, meta, hot }) => {
  return (
    <div
      data-testid={`practice-rightpanel-exp-${index}`}
      style={{
        padding: 10,
        background: "var(--ei-color-bgCard)",
        border: `1px solid ${
          hot ? "var(--ei-color-accent)" : "var(--ei-color-rule)"
        }`,
        borderRadius: 2,
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          gap: 8,
          alignItems: "center",
        }}
      >
        <div
          style={{
            fontSize: 12.5,
            color: "var(--ei-color-ink)",
            fontWeight: 500,
          }}
        >
          {title}
        </div>
      </div>
      <div
        style={{
          fontSize: 11,
          color: "var(--ei-color-ink3)",
          marginTop: 4,
          fontFamily: "var(--ei-mono)",
        }}
      >
        {meta}
      </div>
    </div>
  );
};
