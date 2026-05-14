import { type FC } from "react";

/**
 * Phase 1: skeleton dropdown trigger button. The full menu + UI-only persona
 * switch lands in Phase 3.4. RoleDropdown is rendered as a button (not a
 * <select>) to mirror `ui-design/src/screen-practice.jsx::RoleDropdown`.
 */
export const RoleDropdown: FC = () => {
  return (
    <button
      data-testid="practice-topbar-role"
      type="button"
      style={{
        background: "transparent",
        border: "1px solid var(--ei-color-rule)",
        padding: "6px 10px",
        borderRadius: 2,
        display: "flex",
        gap: 6,
        alignItems: "center",
        color: "var(--ei-color-ink2)",
        fontSize: 12,
        cursor: "pointer",
      }}
    >
      role
    </button>
  );
};
