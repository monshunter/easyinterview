# Seed Input

- 用户登录态：未登录（无保存 session、无保存 route）。
- API 数据源：OpenAPI fixture transport 直接消费仓库内 fixture，无外部依赖。
  - `getRuntimeConfig` 使用 `default` scenario。
  - `getMe` 使用 `unauthenticated` scenario（401 + B1 error envelope）。
- 浏览器状态：jsdom 默认状态；不预设 localStorage / hash route / language
  override。
