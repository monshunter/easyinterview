import type { FC } from "react";

import type { Route } from "../routes";

/**
 * D1 placeholder for the Settings & Privacy route. Mirrors
 * docs/ui-design/user-profile-and-settings.md §5.
 *
 * Settings holds account basics, login security, font preset, and the
 * privacy / data control pane. It does NOT restore Growth / Experiences /
 * Mistakes / Drill modules or any job-target / skill-tag metadata.
 *
 * Notifications + Subscription are P1 placeholders.
 */
export const SettingsScreen: FC<{ route: Route }> = ({ route }) => (
  <section
    data-testid={`route-${route.name}`}
    data-route-name={route.name}
    data-route-params={JSON.stringify(route.params)}
  >
    <header>
      <h1>设置与隐私</h1>
    </header>
    <div data-testid="settings-account">
      <h2>账号基础信息</h2>
      <ul>
        <li>显示姓名</li>
        <li>登录邮箱</li>
        <li>手机号</li>
        <li>界面语言</li>
        <li>时区</li>
      </ul>
    </div>
    <div data-testid="settings-login-security">
      <h2>登录与安全</h2>
      <ul>
        <li>密码（C1 / B2 接入后可用）</li>
        <li>登录方式</li>
        <li>两步验证</li>
      </ul>
    </div>
    <div data-testid="settings-font-preset">
      <h2>字体预设</h2>
      <ul>
        <li>编辑级: Noto Serif SC + Inter</li>
        <li>现代: Source Serif Pro + Geist</li>
        <li>杂志: Cormorant Garamond + IBM Plex Sans</li>
      </ul>
    </div>
    <div data-testid="settings-privacy">
      <h2>隐私与数据</h2>
      <ul>
        <li>数据留存开关</li>
        <li>数据概览</li>
        <li>导出数据</li>
        <li>删除单次会话</li>
        <li>删除所有练习数据</li>
        <li>注销账号</li>
      </ul>
    </div>
    <div data-testid="settings-notifications-placeholder">
      <h2>通知（P1 占位）</h2>
    </div>
    <div data-testid="settings-subscription-placeholder">
      <h2>订阅（P1 占位）</h2>
    </div>
  </section>
);
