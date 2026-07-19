import { useState, type FC, type FormEvent } from "react";

import type { AuthEmailStartRequest } from "../../api/generated/types";
import { useI18n } from "../i18n/messages";
import type { LooseRoute } from "../normalizeRoute";
import type { Route } from "../routes";
import { AuthShell } from "./AuthShell";
import { decodePendingActionRoute } from "./pendingAction";

export interface AuthLoginScreenProps {
  route: Route;
  onNavigate: (next: LooseRoute) => void;
  /**
   * Wires the email-code challenge. Implementations delegate to
   * the generated `startAuthEmailChallenge` operation.
   */
  onStartChallenge: (req: AuthEmailStartRequest) => Promise<void>;
}

export const AuthLoginScreen: FC<AuthLoginScreenProps> = ({
  route,
  onNavigate,
  onStartChallenge,
}) => {
  const { t } = useI18n();
  const [email, setEmail] = useState("");
  const hasPendingAction = decodePendingActionRoute(route.params) !== null;

  const submit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmed = email.trim();
    if (!trimmed) return;
    await onStartChallenge({ email: trimmed });
    // Forward the entire route.params so any encoded pendingAction (pendingRoute
    // / pendingType / pendingLabel + interview-context keys) reaches verify.
    const { returnTo: _returnTo, ...safeRouteParams } = route.params;
    onNavigate({
      name: "auth_verify",
      params: { ...safeRouteParams, email: trimmed },
    });
  };

  return (
    <AuthShell
      routeName="auth_login"
      eyebrowKey="auth.login.eyebrow"
      titleKey="auth.login.title"
      subKey="auth.login.sub"
      pendingAction={hasPendingAction}
    >
      <form
        data-testid="auth-login-email-form"
        className="ei-auth-form"
        onSubmit={submit}
      >
        <label className="ei-auth-field">
          <span className="ei-auth-field-label ei-text-label">
            {t("auth.email")}
          </span>
          <input
            data-testid="auth-login-email"
            className="ei-auth-field-input"
            type="email"
            placeholder={t("auth.emailPlaceholder")}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>
        <button
          type="submit"
          data-testid="auth-login-submit-email"
          className="ei-auth-cta"
        >
          {t("auth.sendEmail")}
        </button>
      </form>
      <div data-testid="auth-login-help" className="ei-auth-help">
        <div>{t("auth.login.helpAccount")}</div>
        <div className="ei-auth-help-line">{t("auth.login.helpCode")}</div>
      </div>
    </AuthShell>
  );
};
