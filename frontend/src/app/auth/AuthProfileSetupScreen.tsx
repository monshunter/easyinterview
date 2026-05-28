import { useState, type FC, type FormEvent } from "react";

import type { CompleteProfileRequest } from "../../api/generated/types";
import { useI18n } from "../i18n/messages";
import type { LooseRoute } from "../normalizeRoute";
import type { Route } from "../routes";
import { AuthShell } from "./AuthShell";
import { decodePendingActionRoute } from "./pendingAction";
import { buildResumeRoute } from "./resumeRoute";

export interface AuthProfileSetupScreenProps {
  route: Route;
  onNavigate: (next: LooseRoute) => void;
  onCompleteProfile: (req: CompleteProfileRequest) => Promise<void>;
}

export const AuthProfileSetupScreen: FC<AuthProfileSetupScreenProps> = ({
  route,
  onNavigate,
  onCompleteProfile,
}) => {
  const { t } = useI18n();
  const [displayName, setDisplayName] = useState("");
  const [agreed, setAgreed] = useState(false);
  const [submitFailed, setSubmitFailed] = useState(false);
  const hasPendingAction = decodePendingActionRoute(route.params) !== null;

  const submit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmed = displayName.trim();
    if (!trimmed || !agreed) return;
    setSubmitFailed(false);
    try {
      await onCompleteProfile({
        displayName: trimmed,
        acceptedTerms: true,
      });
    } catch {
      setSubmitFailed(true);
      return;
    }
    onNavigate(buildResumeRoute(route.params));
  };

  return (
    <AuthShell
      routeName="auth_profile_setup"
      eyebrowKey="auth.profileSetup.eyebrow"
      titleKey="auth.profileSetup.title"
      subKey="auth.profileSetup.sub"
      pendingAction={hasPendingAction}
    >
      {submitFailed ? (
        <p
          data-testid="auth-profile-status"
          className="ei-auth-status ei-auth-status--warn"
        >
          {t("auth.profileSetupFailed")}
        </p>
      ) : null}
      <form
        data-testid="auth-profile-form"
        className="ei-auth-form"
        onSubmit={submit}
      >
        <label className="ei-auth-field">
          <span className="ei-auth-field-label ei-text-label">
            {t("auth.displayName")}
          </span>
          <input
            data-testid="auth-profile-name"
            className="ei-auth-field-input"
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            autoComplete="name"
            required
          />
        </label>
        <label className="ei-text-body">
          <input
            data-testid="auth-profile-terms"
            type="checkbox"
            checked={agreed}
            onChange={(e) => setAgreed(e.target.checked)}
          />{" "}
          {t("auth.acceptTerms")}
        </label>
        <button
          type="submit"
          data-testid="auth-profile-submit"
          className="ei-auth-cta"
          disabled={!displayName.trim() || !agreed}
        >
          {t("auth.completeProfile")}
        </button>
      </form>
    </AuthShell>
  );
};
