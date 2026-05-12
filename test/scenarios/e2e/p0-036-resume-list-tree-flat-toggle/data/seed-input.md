# Seed input — E2E.P0.036

| Operation | Fixture | Scenario | 用途 |
|-----------|---------|----------|------|
| getRuntimeConfig | `openapi/fixtures/Auth/getRuntimeConfig.json` | default | App runtime bootstrap |
| getMe | `openapi/fixtures/Auth/getMe.json` | authenticated / unauthenticated | App auth state per case |
| listResumes | `openapi/fixtures/Resumes/listResumes.json` | default | 提供 2 个 asset，第 2 个无匹配版本 |
| listResumeVersions | `openapi/fixtures/Resumes/listResumeVersions.json` | default | scenario-scoped 版本聚合（与第 1 个 asset 关联） |

User states:
- 未登录访问 `/resume_versions` 期待 auth gate 渲染。
- 登录后访问 `/resume_versions` 期待 list 渲染。
