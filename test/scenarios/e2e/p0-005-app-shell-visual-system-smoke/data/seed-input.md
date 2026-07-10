# Seed Input

- 用户登录态：未登录（无保存 session、无保存 route）。
- API 数据源：OpenAPI fixture transport 直接消费仓库内 fixture，无外部依赖。
  - `getRuntimeConfig` 使用 `default` scenario。
  - `getMe` 使用 `unauthenticated` scenario（401 + B1 error envelope）。
- 浏览器状态：jsdom 默认状态；不预设 localStorage / hash route / language
  override / dark mode / customAccent，显示偏好从 `ocean/light` 开始。
- CSS 注入：`themes.css` / `typography.css` / `topbar.css` / `auth.css` /
  `screens.css` 在 `beforeEach` 注入到 document，让
  `getComputedStyle(documentElement)` 能解析 `:root[data-theme=X][data-mode=Y]`
  selector 上的 CSS variables。
