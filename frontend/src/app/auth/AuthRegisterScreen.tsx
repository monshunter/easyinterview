import { useState, type FC, type FormEvent } from "react";

import type { AuthEmailStartRequest } from "../../api/generated/types";
import { useI18n } from "../i18n/messages";
import type { LooseRoute } from "../normalizeRoute";
import type { Route } from "../routes";

export interface AuthRegisterScreenProps {
  route: Route;
  onNavigate: (next: LooseRoute) => void;
  onStartChallenge: (req: AuthEmailStartRequest) => Promise<void>;
}

export const AuthRegisterScreen: FC<AuthRegisterScreenProps> = ({
  route,
  onNavigate,
  onStartChallenge,
}) => {
  const { t } = useI18n();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [agreed, setAgreed] = useState(false);
  const returnTo = route.params.returnTo;

  const submit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmedEmail = email.trim();
    if (!trimmedEmail || !agreed) return;
    await onStartChallenge(
      returnTo ? { email: trimmedEmail, returnTo } : { email: trimmedEmail },
    );
    // Forward the entire route.params so any encoded pendingAction reaches
    // verify; email + optional displayName overlay on top.
    onNavigate({
      name: "auth_verify",
      params: {
        ...route.params,
        email: trimmedEmail,
        ...(name.trim() ? { displayName: name.trim() } : {}),
      },
    });
  };

  return (
    <section data-testid="route-auth_register" data-route-name="auth_register">
      <h1>{t("auth.register")}</h1>
      <form data-testid="auth-register-form" onSubmit={submit}>
        <label>
          {t("auth.displayName")}
          <input
            data-testid="auth-register-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoComplete="name"
          />
        </label>
        <label>
          {t("auth.email")}
          <input
            data-testid="auth-register-email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />
        </label>
        <fieldset data-testid="auth-register-password-stub" disabled>
          <legend>{t("auth.setPasswordUnavailable")}</legend>
          <input
            aria-label="password"
            type="password"
            autoComplete="new-password"
          />
        </fieldset>
        <label>
          <input
            data-testid="auth-register-terms"
            type="checkbox"
            checked={agreed}
            onChange={(e) => setAgreed(e.target.checked)}
          />
          {t("auth.acceptTerms")}
        </label>
        <button
          type="submit"
          data-testid="auth-register-submit"
          disabled={!email.trim() || !agreed}
        >
          {t("auth.createAndVerify")}
        </button>
      </form>
    </section>
  );
};
