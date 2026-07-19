import { useCallback, useState, type FC, type FormEvent } from "react";

import type { LooseRoute } from "../normalizeRoute";
import { useI18n } from "../i18n/messages";
import type { Route } from "../routes";
import { AuthShell } from "./AuthShell";
import { decodePendingActionRoute } from "./pendingAction";
import { buildResumeRoute } from "./resumeRoute";

export interface AuthVerifyRequest {
  token: string;
}

export interface AuthVerifyResult {
  profileCompletionRequired: boolean;
}

export interface AuthVerifyScreenProps {
  route: Route;
  onNavigate: (next: LooseRoute) => void;
  /**
   * Wires `verifyAuthEmailChallenge`. The generated client mints the session
   * cookie automatically; on success we restore the pending action route from
   * the verify route params.
   */
  onVerify: (req: AuthVerifyRequest) => Promise<AuthVerifyResult>;
}

export const AuthVerifyScreen: FC<AuthVerifyScreenProps> = ({
  route,
  onNavigate,
  onVerify,
}) => {
  const { t } = useI18n();
  const [code, setCode] = useState("");
  const [verifyFailed, setVerifyFailed] = useState(false);
  const hasPendingAction = decodePendingActionRoute(route.params) !== null;
  const completeVerify = useCallback(
    () => {
      const next = buildResumeRoute(route.params);
      onNavigate(next);
    },
    [onNavigate, route.params],
  );

  const submit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmed = code.trim();
    if (trimmed.length !== 6) return;
    setVerifyFailed(false);
    try {
      const result = await onVerify({ token: trimmed });
      if (result.profileCompletionRequired) {
        onNavigate({
          name: "auth_profile_setup",
          params: { ...route.params },
        });
        return;
      }
      completeVerify();
    } catch {
      setVerifyFailed(true);
    }
  };

  const updateCode = (value: string) => {
    setCode(value.replace(/\D/g, "").slice(0, 6));
  };

  return (
    <AuthShell
      routeName="auth_verify"
      eyebrowKey="auth.verify.eyebrow"
      titleKey="auth.verify.title"
      subKey="auth.verify.sub"
      pendingAction={hasPendingAction}
    >
      {route.params.email ? (
        <p
          data-testid="auth-verify-email-hint"
          className="ei-auth-status ei-auth-status--neutral"
        >
          {t("auth.verifySentPrefix")} {route.params.email}
        </p>
      ) : null}
      {verifyFailed ? (
        <p
          data-testid="auth-verify-code-status"
          className="ei-auth-status ei-auth-status--warn"
        >
          {t("auth.verifyLinkFailed")}
        </p>
      ) : null}
      <form
        data-testid="auth-verify-form"
        className="ei-auth-form"
        onSubmit={submit}
      >
        <label className="ei-auth-field">
          <span className="ei-auth-field-label ei-text-label">
            {t("auth.code")}
          </span>
          <input
            data-testid="auth-verify-code"
            className="ei-auth-field-input"
            type="text"
            placeholder={t("auth.codePlaceholder")}
            inputMode="numeric"
            autoComplete="one-time-code"
            pattern="[0-9]*"
            maxLength={6}
            value={code}
            onChange={(e) => updateCode(e.target.value)}
            required
          />
        </label>
        <button
          type="submit"
          data-testid="auth-verify-submit"
          className="ei-auth-cta"
          disabled={code.length !== 6}
        >
          {t("auth.verifyContinue")}
        </button>
      </form>
      <button
        type="button"
        data-testid="auth-verify-resend"
        className="ei-auth-secondary-link"
      >
        {t("auth.resend")}
      </button>
    </AuthShell>
  );
};
