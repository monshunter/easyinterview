import type { FC } from "react";

import type { LooseRoute } from "../normalizeRoute";
import type { Route } from "../routes";

export interface AuthLogoutScreenProps {
  route: Route;
  onNavigate: (next: LooseRoute) => void;
  /** Wires `POST /auth/logout`. The session cookie is cleared by the server. */
  onLogout: () => Promise<void>;
}

export const AuthLogoutScreen: FC<AuthLogoutScreenProps> = ({
  onNavigate,
  onLogout,
}) => {
  const confirm = async () => {
    await onLogout();
    onNavigate({ name: "home", params: {} });
  };
  const cancel = () => onNavigate({ name: "home", params: {} });
  return (
    <section data-testid="route-auth_logout" data-route-name="auth_logout">
      <h1>退出登录</h1>
      <p data-testid="auth-logout-data-hint">
        退出后会清除本机登录态，账号数据保留。
      </p>
      <button type="button" data-testid="auth-logout-confirm" onClick={confirm}>
        确认退出
      </button>
      <button type="button" data-testid="auth-logout-cancel" onClick={cancel}>
        返回首页
      </button>
    </section>
  );
};
