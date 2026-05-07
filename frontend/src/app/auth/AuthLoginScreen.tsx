import { useState, type FC, type FormEvent } from "react";

import type { AuthEmailStartRequest } from "../../api/generated/types";
import { useI18n } from "../i18n/messages";
import type { LooseRoute } from "../normalizeRoute";
import type { Route } from "../routes";

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
    <section data-testid="route-auth_login" data-route-name="auth_login">
      <h1>{t("auth.login")}</h1>
      <form data-testid="auth-login-email-form" onSubmit={submit}>
        <label>
          {t("auth.email")}
          <input
            data-testid="auth-login-email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>
        <button type="submit" data-testid="auth-login-submit-email">
          {t("auth.sendEmail")}
        </button>
      </form>
      <fieldset data-testid="auth-login-password-stub" disabled>
        <legend>{t("auth.passwordLoginUnavailable")}</legend>
        <input
          aria-label="password"
          type="password"
          autoComplete="current-password"
        />
        <button type="button">{t("auth.passwordLogin")}</button>
      </fieldset>
      <div data-testid="auth-login-oauth-stub" aria-disabled="true">
        <button type="button" disabled>
          {t("auth.oauthUnavailable")}
        </button>
      </div>
      <button
        type="button"
        data-testid="auth-login-link-register"
        onClick={() =>
          onNavigate({ name: "auth_register", params: { ...route.params } })
        }
      >
        {t("auth.registerNew")}
      </button>
      <button
        type="button"
        data-testid="auth-login-link-reset"
        onClick={() => onNavigate({ name: "auth_reset", params: {} })}
      >
        {t("auth.forgotPassword")}
      </button>
    </section>
  );
};
