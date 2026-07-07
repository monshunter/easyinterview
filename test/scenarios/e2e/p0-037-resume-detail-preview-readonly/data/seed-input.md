# Seed input — E2E.P0.037

| Operation | Fixture | Scenario | 用途 |
|-----------|---------|----------|------|
| getRuntimeConfig | `openapi/fixtures/Auth/getRuntimeConfig.json` | default | App runtime bootstrap |
| getMe | `openapi/fixtures/Auth/getMe.json` | authenticated | App auth state |
| getResume | `openapi/fixtures/Resumes/getResume.json` | default / not-found | Flat resume detail + 404 |

User states:
- 已登录访问已有 resumeId 期待只读 detail 正文渲染。
- 已登录携带旧 `tab=rewrites` / `tailorRunId` 参数访问时，期待旧参数被忽略且不 materialize 二次操作。
- 已登录访问不存在 resumeId 期待 NotFoundEmptyState 渲染。
