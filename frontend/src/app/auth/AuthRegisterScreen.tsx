import { useState, type FC, type FormEvent } from "react";

import type { AuthEmailStartRequest } from "../../api/generated/types";
import { useI18n } from "../i18n/messages";
import type { LooseRoute } from "../normalizeRoute";
import type { Route } from "../routes";
import { AuthShell } from "./AuthShell";
import { decodePendingActionRoute } from "./pendingAction";

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
  const hasPendingAction = decodePendingActionRoute(route.params) !== null;

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
    <AuthShell
      routeName="auth_register"
      eyebrowKey="auth.register.eyebrow"
      titleKey="auth.register.title"
      subKey="auth.register.sub"
      pendingAction={hasPendingAction}
    >
      <form
        data-testid="auth-register-form"
        className="ei-auth-form"
        onSubmit={submit}
      >
        <label className="ei-auth-field">
          <span className="ei-auth-field-label ei-text-label">
            {t("auth.displayName")}
          </span>
          <input
            data-testid="auth-register-name"
            className="ei-auth-field-input"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoComplete="name"
          />
        </label>
        <label className="ei-auth-field">
          <span className="ei-auth-field-label ei-text-label">
            {t("auth.email")}
          </span>
          <input
            data-testid="auth-register-email"
            className="ei-auth-field-input"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />
        </label>
        <fieldset
          data-testid="auth-register-password-stub"
          className="ei-auth-stub"
          disabled
        >
          <legend className="ei-text-label">
            {t("auth.setPasswordUnavailable")}
          </legend>
          <input
            aria-label="password"
            type="password"
            autoComplete="new-password"
            className="ei-auth-field-input"
          />
        </fieldset>
        <label className="ei-text-body">
          <input
            data-testid="auth-register-terms"
            type="checkbox"
            checked={agreed}
            onChange={(e) => setAgreed(e.target.checked)}
          />{" "}
          {t("auth.acceptTerms")}
        </label>
        <button
          type="submit"
          data-testid="auth-register-submit"
          className="ei-auth-cta"
          disabled={!email.trim() || !agreed}
        >
          {t("auth.createAndVerify")}
        </button>
      </form>
    </AuthShell>
  );
};
