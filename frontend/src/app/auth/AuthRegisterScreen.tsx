import { useState, type FC, type FormEvent } from "react";

import type { AuthEmailStartRequest } from "../../api/generated/types";
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
      <h1>注册</h1>
      <form data-testid="auth-register-form" onSubmit={submit}>
        <label>
          显示姓名
          <input
            data-testid="auth-register-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoComplete="name"
          />
        </label>
        <label>
          邮箱
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
          <legend>设置密码（暂未开放）</legend>
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
          同意服务条款与隐私政策
        </label>
        <button
          type="submit"
          data-testid="auth-register-submit"
          disabled={!email.trim() || !agreed}
        >
          创建账号并验证邮箱
        </button>
      </form>
    </section>
  );
};
