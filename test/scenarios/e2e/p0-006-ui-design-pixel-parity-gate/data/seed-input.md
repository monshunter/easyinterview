# Seed Input

- 用户登录态：未登录（无保存 session、无保存 route）。
- API 数据源：本场景不消费后端 API；frontend dist 由静态服务器托管，
  TopBar / auth_login 行为不依赖 `/me` 等运行时 fetch。
- 浏览器：chromium 二进制已通过 `pnpm --filter @easyinterview/frontend
  test:pixel-parity:install` 安装到本地 Playwright 缓存。
- 构建：`frontend/dist/index.html` 已通过 `pnpm --filter
  @easyinterview/frontend build` 构建；包含 fontsource bundle、ei-* 视觉
  className、theme/typography/topbar/auth/screens 样式。
- ui-design：`ui-design/index.html` 静态原型可加载，依赖外网 CDN
  （unpkg.com + Google Fonts）。
- 端口：`http://127.0.0.1:4173`（由 `frontend/scripts/serve-pixel-parity
  .mjs` 提供）。
