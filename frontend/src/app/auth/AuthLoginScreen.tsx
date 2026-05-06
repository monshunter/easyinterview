import { useState, type FC, type FormEvent } from "react";

import type { AuthEmailStartRequest } from "../../api/generated/types";
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
      <h1>登录</h1>
      <form data-testid="auth-login-email-form" onSubmit={submit}>
        <label>
          邮箱
          <input
            data-testid="auth-login-email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>
        <button type="submit" data-testid="auth-login-submit-email">
          发送登录邮件
        </button>
      </form>
      <fieldset data-testid="auth-login-password-stub" disabled>
        <legend>密码登录（暂未开放）</legend>
        <input
          aria-label="password"
          type="password"
          autoComplete="current-password"
        />
        <button type="button">密码登录</button>
      </fieldset>
      <div data-testid="auth-login-oauth-stub" aria-disabled="true">
        <button type="button" disabled>
          第三方登录（暂未开放）
        </button>
      </div>
      <button
        type="button"
        data-testid="auth-login-link-register"
        onClick={() => onNavigate({ name: "auth_register", params: {} })}
      >
        注册新账号
      </button>
      <button
        type="button"
        data-testid="auth-login-link-reset"
        onClick={() => onNavigate({ name: "auth_reset", params: {} })}
      >
        忘记密码
      </button>
    </section>
  );
};
