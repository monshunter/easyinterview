import type { FC, ReactNode } from "react";

import { useI18n } from "../i18n/messages";
import type { MessageKey } from "../i18n/locales/zh";

export interface AuthShellProps {
  /** Maps to `route-{name}` testid contract for auth screens. */
  routeName:
    | "auth_login"
    | "auth_register"
    | "auth_verify"
    | "auth_reset"
    | "auth_logout";
  /** i18n key for the eyebrow (e.g., `auth.login.eyebrow`). */
  eyebrowKey: MessageKey;
  /** i18n key for the headline. */
  titleKey: MessageKey;
  /** i18n key for the supporting copy. */
  subKey: MessageKey;
  /**
   * Whether the side panel should render the pending-action callout (when a
   * pendingAction is encoded in route.params) instead of the default
   * authentication principle panel.
   */
  pendingAction?: boolean;
  /** Right column form content. */
  children: ReactNode;
}

/**
 * Two-column auth shell transcribed from `ui-design/src/screen-auth.jsx`.
 * The shell never owns form state or pendingAction wiring; it only handles
 * the visual rhythm (max-width 1160 / padding 54 48 96 / grid 0.88fr 1.12fr /
 * gap 44) and the optional side panel. Per-screen testid stays on the outer
 * `<section>` so D1 route-state tests keep working.
 */
export const AuthShell: FC<AuthShellProps> = ({
  routeName,
  eyebrowKey,
  titleKey,
  subKey,
  pendingAction,
  children,
}) => {
  const { t } = useI18n();
  return (
    <section
      data-testid={`route-${routeName}`}
      data-route-name={routeName}
      className="ei-auth-shell"
    >
      <div className="ei-auth-side">
        <span className="ei-auth-eyebrow ei-text-label">{t(eyebrowKey)}</span>
        <h1 className="ei-text-display">{t(titleKey)}</h1>
        <p className="ei-auth-sub ei-text-body">{t(subKey)}</p>
        {pendingAction ? (
          <div
            className="ei-auth-side-panel ei-auth-side-panel-pending"
            data-testid="auth-side-pending-action"
          >
            <span className="ei-auth-eyebrow ei-text-label">
              {t("auth.pendingAction.eyebrow")}
            </span>
            <p className="ei-text-body">{t("auth.pendingAction.body")}</p>
          </div>
        ) : (
          <div
            className="ei-auth-side-panel"
            data-testid="auth-side-principle"
          >
            <span className="ei-auth-eyebrow ei-text-label">
              {t("auth.principle.eyebrow")}
            </span>
            <p className="ei-text-body">{t("auth.principle.body")}</p>
          </div>
        )}
      </div>
      <div className="ei-auth-card">{children}</div>
    </section>
  );
};
