import { type FC } from "react";

export interface LiveNotesProps {
  label: string;
  okText: string;
  warnText: string;
  note: string;
}

/**
 * Source-level mirror of `ui-design/src/screen-practice.jsx` lines 170-179.
 * Hidden in strict mode (caller decides). Phase 1 emits skeleton text only.
 */
export const LiveNotes: FC<LiveNotesProps> = ({ label, okText, warnText, note }) => {
  return (
    <div
      data-testid="practice-sessionmap-live-notes"
      style={{
        borderTop: "1px dotted var(--ei-color-rule)",
        marginTop: 14,
        paddingTop: 14,
      }}
    >
      <div
        className="ei-label"
        style={{ color: "var(--ei-color-ink3)", marginBottom: 6 }}
      >
        {label}
      </div>
      <div
        style={{
          fontSize: 12,
          color: "var(--ei-color-ink2)",
          lineHeight: 1.5,
          padding: "8px 10px",
          background: "var(--ei-color-bgCard)",
          borderRadius: 2,
          border: "1px solid var(--ei-color-rule)",
        }}
      >
        <div
          data-testid="practice-sessionmap-live-notes-ok"
          style={{ color: "var(--ei-color-ok)" }}
        >
          ● {okText}
        </div>
        <div
          data-testid="practice-sessionmap-live-notes-warn"
          style={{ color: "var(--ei-color-warn)", marginTop: 4 }}
        >
          ● {warnText}
        </div>
        <div
          style={{
            color: "var(--ei-color-ink3)",
            marginTop: 4,
            fontSize: 11,
          }}
        >
          {note}
        </div>
      </div>
    </div>
  );
};
