import type { FC } from "react";

import { useI18n } from "../i18n/messages";
import type { LooseRoute } from "../normalizeRoute";
import type { Route } from "../routes";

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
    <section data-testid="route-auth_logout" data-route-name="auth_logout">
      <h1>{t("user.logout")}</h1>
      <p data-testid="auth-logout-data-hint">
        {t("auth.logoutHint")}
      </p>
      <button type="button" data-testid="auth-logout-confirm" onClick={confirm}>
        {t("auth.confirmLogout")}
      </button>
      <button type="button" data-testid="auth-logout-cancel" onClick={cancel}>
        {t("auth.backHome")}
      </button>
    </section>
  );
};
