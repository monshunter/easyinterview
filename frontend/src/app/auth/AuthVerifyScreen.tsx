import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type FC,
  type FormEvent,
} from "react";

import { normalizeRoute, type LooseRoute } from "../normalizeRoute";
import { useI18n } from "../i18n/messages";
import type { Route } from "../routes";
import { AuthShell } from "./AuthShell";
import {
  decodePendingActionRoute,
  PENDING_ACTION_INTERVIEW_KEYS,
} from "./pendingAction";

export interface AuthVerifyRequest {
  token: string;
}

export interface AuthVerifyScreenProps {
  route: Route;
  onNavigate: (next: LooseRoute) => void;
  onReplace?: (next: LooseRoute) => void;
  /**
   * Wires `verifyAuthEmailChallenge`. The generated client mints the session
   * cookie automatically; on success we restore the pending action route from
   * the verify route params.
   */
  onVerify: (req: AuthVerifyRequest) => Promise<void>;
}

function buildResumeRoute(params: Record<string, string>): LooseRoute {
  // Phase 3.2 source of truth: pendingAction encoded under reserved keys.
  const decoded = decodePendingActionRoute(params);
  if (decoded) return decoded;

  // Backwards-compat path: B2 / external email links may still surface a
  // raw `returnTo` path. Treat it as a route name candidate and pass any
  // interview-context keys through.
  const resumeParams: Record<string, string> = {};
  for (const key of PENDING_ACTION_INTERVIEW_KEYS) {
    const value = params[key];
    if (value !== undefined) resumeParams[key] = value;
  }
  const returnTo = params.returnTo;
  if (returnTo) {
    const parsed = new URL(returnTo, "http://easyinterview.local");
    for (const key of PENDING_ACTION_INTERVIEW_KEYS) {
      const value = parsed.searchParams.get(key);
      if (value !== null) resumeParams[key] = value;
    }
    const candidate = parsed.pathname.replace(/^\/+/, "") || "home";
    return normalizeRoute({ name: candidate, params: resumeParams });
  }
  return { name: "home", params: resumeParams };
}

export const AuthVerifyScreen: FC<AuthVerifyScreenProps> = ({
  route,
  onNavigate,
  onReplace,
  onVerify,
}) => {
  const { t } = useI18n();
  const [code, setCode] = useState("");
  const [linkStatus, setLinkStatus] = useState<"idle" | "pending" | "failed">(
    route.params.token ? "pending" : "idle",
  );
  const autoVerifyTokenRef = useRef("");
  const hasPendingAction = decodePendingActionRoute(route.params) !== null;
  const magicLinkToken = route.params.token?.trim() ?? "";
  const completeVerify = useCallback(
    (mode: "push" | "replace") => {
      const next = buildResumeRoute(route.params);
      if (mode === "replace" && onReplace) {
        onReplace(next);
        return;
      }
      onNavigate(next);
    },
    [onNavigate, onReplace, route.params],
  );

  useEffect(() => {
    if (!magicLinkToken || autoVerifyTokenRef.current === magicLinkToken) return;
    autoVerifyTokenRef.current = magicLinkToken;
    setLinkStatus("pending");
    void onVerify({ token: magicLinkToken })
      .then(() => completeVerify("replace"))
      .catch(() => setLinkStatus("failed"));
  }, [completeVerify, magicLinkToken, onVerify]);

  const submit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmed = code.trim();
    if (!trimmed) return;
    await onVerify({ token: trimmed });
    completeVerify("push");
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
      {magicLinkToken ? (
        <p
          data-testid="auth-verify-link-status"
          className={`ei-auth-status ${
            linkStatus === "failed"
              ? "ei-auth-status--warn"
              : "ei-auth-status--neutral"
          }`}
        >
          {linkStatus === "failed"
            ? t("auth.verifyLinkFailed")
            : t("auth.verifyLinkPending")}
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
            inputMode="text"
            autoComplete="one-time-code"
            maxLength={256}
            value={code}
            onChange={(e) => setCode(e.target.value)}
            required
          />
        </label>
        <button
          type="submit"
          data-testid="auth-verify-submit"
          className="ei-auth-cta"
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
