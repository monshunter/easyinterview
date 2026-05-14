import { type FC } from "react";

export interface SessionMapItem {
  id: string;
  topic: string;
  duration: string;
  status: "active" | "done" | "pending" | "skipped" | "follow_up_requested";
}

export interface SessionMapProps {
  label: string;
  items: SessionMapItem[];
  activeIndex: number;
}

/**
 * Source-level mirror of `ui-design/src/screen-practice.jsx` lines 138-180
 * (left rail SESSION MAP). turn status rendering rules expand in Phase 3.5.
 */
export const SessionMap: FC<SessionMapProps> = ({ label, items, activeIndex }) => {
  return (
    <>
      <div
        data-testid="practice-sessionmap-label"
        className="ei-label"
        style={{ color: "var(--ei-color-ink3)", marginBottom: 12 }}
      >
        {label}
      </div>
      {items.map((item, idx) => {
        const isActive = idx === activeIndex;
        const isDone = idx < activeIndex || item.status === "done";
        const isSkipped = item.status === "skipped";
        return (
          <div
            key={item.id}
            data-testid={`practice-sessionmap-item-${idx}`}
            data-status={
              isActive
                ? "active"
                : isDone
                  ? "done"
                  : isSkipped
                    ? "skipped"
                    : "pending"
            }
            style={{
              padding: "10px 12px",
              marginBottom: 6,
              borderRadius: 2,
              background: isActive
                ? "var(--ei-color-bgCard)"
                : "transparent",
              border: `1px solid ${
                isActive ? "var(--ei-color-rule)" : "transparent"
              }`,
              display: "flex",
              gap: 10,
              alignItems: "flex-start",
            }}
          >
            <div
              style={{
                width: 22,
                height: 22,
                borderRadius: 11,
                flexShrink: 0,
                border: `1px solid ${
                  isActive ? "var(--ei-color-accent)" : "var(--ei-color-rule)"
                }`,
                background: isDone
                  ? "var(--ei-color-ok)"
                  : isActive
                    ? "var(--ei-color-accentSoft)"
                    : "transparent",
                color: isDone
                  ? "#fff"
                  : isActive
                    ? "var(--ei-color-accent)"
                    : "var(--ei-color-ink3)",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                fontSize: 11,
                fontFamily: "var(--ei-mono)",
              }}
            >
              {isDone ? "✓" : isSkipped ? "↷" : idx + 1}
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div
                style={{
                  fontSize: 12.5,
                  color: isActive
                    ? "var(--ei-color-ink)"
                    : "var(--ei-color-ink2)",
                  fontWeight: isActive ? 500 : 400,
                }}
              >
                {item.topic}
              </div>
              <div
                style={{
                  fontSize: 11,
                  color: "var(--ei-color-ink3)",
                  marginTop: 2,
                  fontFamily: "var(--ei-mono)",
                }}
              >
                {item.duration}
              </div>
            </div>
          </div>
        );
      })}
    </>
  );
};
