# Seed input — E2E.P0.037

| Operation | Fixture | Scenario | 用途 |
|-----------|---------|----------|------|
| getRuntimeConfig | `openapi/fixtures/Auth/getRuntimeConfig.json` | default | App runtime bootstrap |
| getMe | `openapi/fixtures/Auth/getMe.json` | authenticated | App auth state |
| getResume | `openapi/fixtures/Resumes/getResume.json` | default / not-found | Flat resume detail + 404 |
| exportResume | `openapi/fixtures/Resumes/exportResume.json` | p0-501-not-available | 触发 P0 not-available toast |

User states:
- 已登录访问已有 resumeId 期待 detail 渲染。
- 已登录访问不存在 resumeId 期待 NotFoundEmptyState 渲染。
