import { useState, type FC, type FormEvent } from "react";

import { useI18n } from "../i18n/messages";
import type { LooseRoute } from "../normalizeRoute";
import type { Route } from "../routes";
import { AuthShell } from "./AuthShell";

export interface AuthResetScreenProps {
  route: Route;
  onNavigate: (next: LooseRoute) => void;
  /**
   * Reserved for the future password-reset path. D1 keeps this prop only so
   * downstream callers can pass it without restructuring; the screen never
   * invokes the callback because B2 / C1 have not introduced a real reset
   * contract yet.
   */
  onStartChallenge?: () => Promise<void> | void;
}

export const AuthResetScreen: FC<AuthResetScreenProps> = ({ onNavigate }) => {
  const { t } = useI18n();
  const [email, setEmail] = useState("");
  const [submitted, setSubmitted] = useState(false);

  const submit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    // Reset is a UI shell only — no API is wired. The button updates local
    // state so the screen can show a "邮件已发送" placeholder hint without
    // pretending a request happened.
    if (!email.trim()) return;
    setSubmitted(true);
  };

  return (
    <AuthShell
      routeName="auth_reset"
      eyebrowKey="auth.reset.eyebrow"
      titleKey="auth.reset.title"
      subKey="auth.reset.sub"
    >
      <form
        data-testid="auth-reset-form"
        className="ei-auth-form"
        onSubmit={submit}
      >
        <label className="ei-auth-field">
          <span className="ei-auth-field-label ei-text-label">
            {t("auth.accountEmail")}
          </span>
          <input
            data-testid="auth-reset-email"
            className="ei-auth-field-input"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />
        </label>
        <button
          type="submit"
          data-testid="auth-reset-send-stub"
          className="ei-auth-cta"
        >
          {t("auth.sendResetUnavailable")}
        </button>
      </form>
      {submitted ? (
        <p
          data-testid="auth-reset-stub-hint"
          className="ei-auth-status ei-auth-status--neutral"
        >
          {t("auth.resetHint")}
        </p>
      ) : null}
      <button
        type="button"
        data-testid="auth-reset-link-login"
        className="ei-auth-secondary-link"
        onClick={() => onNavigate({ name: "auth_login", params: {} })}
      >
        {t("auth.backToLogin")}
      </button>
    </AuthShell>
  );
};
