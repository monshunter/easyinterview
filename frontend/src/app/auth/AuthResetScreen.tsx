import { useState, type FC, type FormEvent } from "react";

import type { LooseRoute } from "../normalizeRoute";
import type { Route } from "../routes";

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
    <section data-testid="route-auth_reset" data-route-name="auth_reset">
      <h1>重置登录</h1>
      <form data-testid="auth-reset-form" onSubmit={submit}>
        <label>
          账号邮箱
          <input
            data-testid="auth-reset-email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />
        </label>
        <button type="submit" data-testid="auth-reset-send-stub">
          发送重置说明（暂未开放）
        </button>
      </form>
      {submitted ? (
        <p data-testid="auth-reset-stub-hint">
          重置流程暂未开放。请使用邮箱验证码登录或联系支持。
        </p>
      ) : null}
      <button
        type="button"
        data-testid="auth-reset-link-login"
        onClick={() => onNavigate({ name: "auth_login", params: {} })}
      >
        返回登录
      </button>
    </section>
  );
};
