import { useState, type FC } from "react";

import { useI18n } from "../../../i18n/messages";

export type InterviewerPersona = "general" | "hr" | "manager";

export interface RoleDropdownProps {
  persona: InterviewerPersona;
  onChange: (next: InterviewerPersona) => void;
}

/**
 * Source-level mirror of `ui-design/src/screen-practice.jsx::RoleDropdown`
 * (lines 617-641). Phase 3.4 wires UI-only persona switching: clicking
 * the menu changes local persona state and the AI TRANSPARENCY label,
 * but no backend request is made.
 */
export const RoleDropdown: FC<RoleDropdownProps> = ({ persona, onChange }) => {
  const { t } = useI18n();
  const [open, setOpen] = useState(false);

  const labels: Record<InterviewerPersona, string> = {
    general: t("practice.toolbar.role.general"),
    hr: t("practice.toolbar.role.hr"),
    manager: t("practice.toolbar.role.manager"),
  };

  const choose = (next: InterviewerPersona) => {
    onChange(next);
    setOpen(false);
  };

  return (
    <div style={{ position: "relative" }}>
      <button
        data-testid="practice-topbar-role"
        type="button"
        onClick={() => setOpen((o) => !o)}
        style={{
          background: "transparent",
          border: "1px solid var(--ei-color-rule-strong)",
          padding: "6px 10px",
          borderRadius: 2,
          display: "flex",
          gap: 6,
          alignItems: "center",
          color: "var(--ei-color-fg-secondary)",
          fontSize: 12,
          cursor: "pointer",
        }}
      >
        {labels[persona]}
      </button>
      {open && (
        <div
          data-testid="practice-topbar-role-menu"
          role="menu"
          style={{
            position: "absolute",
            top: "100%",
            right: 0,
            marginTop: 4,
            background: "var(--ei-color-bg-card)",
            border: "1px solid var(--ei-color-rule-strong)",
            borderRadius: 2,
            minWidth: 200,
            zIndex: 20,
            boxShadow: "0 4px 16px rgba(0,0,0,0.08)",
          }}
        >
          {(["general", "hr", "manager"] as InterviewerPersona[]).map((k) => (
            <button
              key={k}
              data-testid={`practice-topbar-role-option-${k}`}
              type="button"
              role="menuitem"
              onClick={() => choose(k)}
              style={{
                display: "block",
                width: "100%",
                textAlign: "left",
                padding: "10px 12px",
                background:
                  persona === k
                    ? "var(--ei-color-bg-soft)"
                    : "transparent",
                border: "none",
                cursor: "pointer",
                fontSize: 13,
                color: "var(--ei-color-fg-primary)",
                fontWeight: persona === k ? 500 : 400,
              }}
            >
              {labels[k]}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};
