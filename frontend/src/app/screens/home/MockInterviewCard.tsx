import type { FC } from "react";

import type { TargetJob, TargetJobStatus } from "../../../api/generated/types";

/** Default interview rounds when no real round data is available. */
const DEFAULT_ROUNDS = [
  "R1 Phone Screen",
  "R2 Technical",
  "R3 Culture",
  "R4 Final",
];

function statusTone(
  status: TargetJobStatus,
): "muted" | "amber" | "neutral" {
  switch (status) {
    case "draft":
    case "preparing":
      return "muted";
    case "applied":
    case "interviewing":
      return "amber";
    case "offer":
    case "rejected":
    case "archived":
      return "neutral";
  }
}

function statusLabel(status: TargetJobStatus): string {
  switch (status) {
    case "draft":
      return "Draft";
    case "preparing":
      return "Preparing";
    case "applied":
      return "Applied";
    case "interviewing":
      return "Interviewing";
    case "offer":
      return "Offer";
    case "rejected":
      return "Rejected";
    case "archived":
      return "Archived";
  }
}

function roundIndexFromStatus(
  status: TargetJobStatus,
  roundCount: number,
): number {
  if (roundCount === 0) return 0;
  switch (status) {
    case "draft":
    case "preparing":
      return 0;
    case "applied":
    case "interviewing":
      return 1;
    case "offer":
    case "rejected":
    case "archived":
      return roundCount - 1;
  }
}

const statusColorMap: Record<
  string,
  { bg: string; fg: string }
> = {
  amber: {
    bg: "var(--ei-color-amber-soft)",
    fg: "var(--ei-color-amber)",
  },
  neutral: {
    bg: "var(--ei-color-bg-soft)",
    fg: "var(--ei-color-fg-secondary)",
  },
  muted: {
    bg: "transparent",
    fg: "var(--ei-color-fg-tertiary)",
  },
};

// ─── MiniRoundRail ────────────────────────────────────────────────

interface MiniRoundRailProps {
  rounds: string[];
  currentIndex: number;
}

const MiniRoundRail: FC<MiniRoundRailProps> = ({ rounds, currentIndex }) => (
  <div style={{ marginTop: 18 }}>
    <div style={{ position: "relative", height: 34 }}>
      <div
        style={{
          position: "absolute",
          top: 9,
          left: 8,
          right: 8,
          height: 1,
          background: "var(--ei-color-rule-strong)",
        }}
      />
      <div
        style={{
          display: "grid",
          gridTemplateColumns: `repeat(${rounds.length}, 1fr)`,
        }}
      >
        {rounds.map((round, i) => {
          const done = i < currentIndex;
          const current = i === currentIndex;
          const align =
            i === 0
              ? "flex-start"
              : i === rounds.length - 1
                ? "flex-end"
                : "center";
          const textAlign =
            i === 0
              ? "left"
              : i === rounds.length - 1
                ? "right"
                : "center";
          return (
            <div
              key={round}
              style={{
                position: "relative",
                display: "flex",
                flexDirection: "column",
                alignItems: align,
              }}
            >
              <div
                style={{
                  width: 18,
                  height: 18,
                  borderRadius: 9,
                  border: `1px solid ${
                    done
                      ? "var(--ei-color-ok)"
                      : current
                        ? "var(--ei-color-accent)"
                        : "var(--ei-color-rule-strong)"
                  }`,
                  background: done
                    ? "var(--ei-color-ok)"
                    : current
                      ? "var(--ei-color-accent)"
                      : "var(--ei-color-bg-card)",
                  color:
                    done || current
                      ? "#fff"
                      : "var(--ei-color-fg-tertiary)",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  zIndex: 1,
                }}
              >
                {done ? (
                  <svg
                    width={10}
                    height={10}
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="#fff"
                    strokeWidth={2.2}
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  >
                    <path d="M5 12l5 5L20 7" />
                  </svg>
                ) : (
                  <span
                    style={{
                      fontSize: 9,
                      fontFamily: "var(--ei-font-mono)",
                    }}
                  >
                    {i + 1}
                  </span>
                )}
              </div>
              <div
                style={{
                  marginTop: 6,
                  fontSize: 10.5,
                  color: current
                    ? "var(--ei-color-fg-primary)"
                    : "var(--ei-color-fg-tertiary)",
                  maxWidth: 68,
                  textAlign,
                  whiteSpace: "nowrap",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                }}
              >
                {round}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  </div>
);

// ─── MockInterviewCard ────────────────────────────────────────────

interface MockInterviewCardProps {
  job: TargetJob;
  onClick: () => void;
}

export const MockInterviewCard: FC<MockInterviewCardProps> = ({
  job,
  onClick,
}) => {
  const tone = statusTone(job.status);
  const colors = statusColorMap[tone] ?? { bg: "transparent", fg: "var(--ei-color-fg-tertiary)" };
  const ci = roundIndexFromStatus(job.status, DEFAULT_ROUNDS.length);

  return (
    <div
      data-testid={`home-recent-mock-card-${job.id}`}
      onClick={onClick}
      style={{
        background: "var(--ei-color-bg-card)",
        border: "1px solid var(--ei-color-rule-strong)",
        borderRadius: 3,
        padding: 20,
        cursor: "pointer",
        display: "flex",
        flexDirection: "column",
        gap: 14,
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          gap: 12,
        }}
      >
        <div>
          <div
            style={{
              fontSize: 11,
              fontFamily: "var(--ei-font-mono)",
              color: "var(--ei-color-fg-tertiary)",
              marginBottom: 4,
            }}
          >
            {job.companyName.toUpperCase()} ·{" "}
            {statusLabel(job.status)}
          </div>
          <div
            style={{
              fontSize: 19,
              color: "var(--ei-color-fg-primary)",
              fontFamily: "var(--ei-font-serif)",
              letterSpacing: "-0.01em",
            }}
          >
            {job.title}
          </div>
          <div
            style={{
              fontSize: 12.5,
              color: "var(--ei-color-fg-tertiary)",
              marginTop: 4,
            }}
          >
            {job.locationText || "Remote / TBD"}
          </div>
        </div>
        <div
          style={{
            padding: "3px 8px",
            height: "fit-content",
            background: colors.bg,
            color: colors.fg,
            fontSize: 11,
            fontFamily: "var(--ei-font-mono)",
            borderRadius: 2,
            whiteSpace: "nowrap",
          }}
        >
          {statusLabel(job.status)}
        </div>
      </div>
      <div data-testid={`home-recent-mock-rail-${job.id}`}>
        <MiniRoundRail
          rounds={DEFAULT_ROUNDS}
          currentIndex={ci}
        />
      </div>
    </div>
  );
};
