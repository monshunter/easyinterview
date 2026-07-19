import type { FC } from "react";

import { useI18n } from "../i18n/messages";
import type { LooseRoute } from "../normalizeRoute";
import type { Route } from "../routes";
import { AuthShell } from "./AuthShell";

export interface AuthLogoutScreenProps {
  route: Route;
  onNavigate: (next: LooseRoute) => void;
  /** Wires `POST /auth/logout`. The session cookie is cleared by the server. */
  onLogout: () => Promise<void>;
}

export const AuthLogoutScreen: FC<AuthLogoutScreenProps> = ({
  onNavigate,
  onLogout,
}) => {
  const { t } = useI18n();
  const confirm = async () => {
    await onLogout();
    onNavigate({ name: "home", params: {} });
  };
  const cancel = () => onNavigate({ name: "home", params: {} });
  return (
    <AuthShell
      routeName="auth_logout"
      eyebrowKey="auth.logout.eyebrow"
      titleKey="auth.logout.title"
      subKey="auth.logout.sub"
    >
      <p
        data-testid="auth-logout-data-hint"
        className="ei-auth-status ei-auth-status--warn"
      >
        {t("auth.logoutHint")}
      </p>
      <div className="ei-auth-row ei-auth-row--stacked">
        <button
          type="button"
          data-testid="auth-logout-confirm"
          className="ei-auth-cta"
          onClick={confirm}
        >
          {t("auth.confirmLogout")}
        </button>
        <button
          type="button"
          data-testid="auth-logout-cancel"
          className="ei-auth-secondary-link"
          onClick={cancel}
        >
          {t("common.back")}
        </button>
      </div>
    </AuthShell>
  );
};
