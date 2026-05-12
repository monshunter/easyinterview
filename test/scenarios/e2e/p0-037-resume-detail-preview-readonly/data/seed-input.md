# Seed input — E2E.P0.037

| Operation | Fixture | Scenario | 用途 |
|-----------|---------|----------|------|
| getRuntimeConfig | `openapi/fixtures/Auth/getRuntimeConfig.json` | default | App runtime bootstrap |
| getMe | `openapi/fixtures/Auth/getMe.json` | authenticated | App auth state |
| getResumeVersion | `openapi/fixtures/Resumes/getResumeVersion.json` | default / master-default / not-found-404 | TARGETED + MASTER + 404 三态 |
| exportResumeVersion | `openapi/fixtures/Resumes/exportResumeVersion.json` | (隐式 default = 501) | 触发 P0 not-available toast |

User states:
- 已登录访问已有 versionId 期待 detail 渲染。
- 已登录访问不存在 versionId 期待 NotFoundEmptyState 渲染。
