import { type FC } from "react";

export interface HintBannerProps {
  show: boolean;
  prefix: string;
  text: string;
}

/**
 * Source-level mirror of `ui-design/src/screen-practice.jsx` lines 219-223.
 * Visible only when assisted mode + show=true. Hidden in strict mode (caller
 * decides). Phase 1 renders nothing when show=false.
 */
export const HintBanner: FC<HintBannerProps> = ({ show, prefix, text }) => {
  if (!show) return null;
  return (
    <div
      data-testid="practice-hint-banner"
      style={{
        marginBottom: 10,
        padding: "10px 12px",
        background: "var(--ei-color-amberSoft)",
        borderRadius: 2,
        fontSize: 13,
        color: "var(--ei-color-warn)",
      }}
    >
      <b>{prefix}</b> {text}
    </div>
  );
};
