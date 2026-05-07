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
   * Wires the passwordless email magic-link challenge. Implementations
   * delegate to the generated `startAuthEmailChallenge` operation.
   */
  onStartChallenge: (req: AuthEmailStartRequest) => Promise<void>;
}

/**
 * D1 only wires the passwordless email path. Password / OAuth are rendered as
 * stubs to anchor the visual design but never call any API; Phase 3.3 enforces
 * this with a negative search and 4.x will only relax it after C1 / B2 ship
 * the matching backend contracts.
 */
export const AuthLoginScreen: FC<AuthLoginScreenProps> = ({
  route,
  onNavigate,
  onStartChallenge,
}) => {
  const { t } = useI18n();
  const [email, setEmail] = useState("");
  const returnTo = route.params.returnTo;
  const hasPendingAction = decodePendingActionRoute(route.params) !== null;

  const submit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmed = email.trim();
    if (!trimmed) return;
    await onStartChallenge(
      returnTo ? { email: trimmed, returnTo } : { email: trimmed },
    );
    // Forward the entire route.params so any encoded pendingAction (pendingRoute
    // / pendingType / pendingLabel + interview-context keys) reaches verify.
    onNavigate({
      name: "auth_verify",
      params: { ...route.params, email: trimmed },
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
      <fieldset
        data-testid="auth-login-password-stub"
        className="ei-auth-stub"
        disabled
      >
        <legend className="ei-text-label">
          {t("auth.passwordLoginUnavailable")}
        </legend>
        <input
          aria-label="password"
          type="password"
          autoComplete="current-password"
          className="ei-auth-field-input"
        />
        <button type="button" className="ei-auth-secondary-link">
          {t("auth.passwordLogin")}
        </button>
      </fieldset>
      <div
        data-testid="auth-login-oauth-stub"
        aria-disabled="true"
        className="ei-auth-stub"
      >
        <span className="ei-text-label">{t("auth.oauthUnavailable")}</span>
      </div>
      <div className="ei-auth-link-row">
        <button
          type="button"
          data-testid="auth-login-link-register"
          className="ei-auth-secondary-link"
          onClick={() =>
            onNavigate({ name: "auth_register", params: { ...route.params } })
          }
        >
          {t("auth.registerNew")}
        </button>
        <button
          type="button"
          data-testid="auth-login-link-reset"
          className="ei-auth-secondary-link"
          onClick={() => onNavigate({ name: "auth_reset", params: {} })}
        >
          {t("auth.forgotPassword")}
        </button>
      </div>
    </AuthShell>
  );
};
