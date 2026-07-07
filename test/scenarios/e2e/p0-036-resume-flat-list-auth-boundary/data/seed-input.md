# Seed input — E2E.P0.036

| Operation | Fixture | Scenario | 用途 |
|-----------|---------|----------|------|
| getRuntimeConfig | `openapi/fixtures/Auth/getRuntimeConfig.json` | default | App runtime bootstrap |
| getMe | `openapi/fixtures/Auth/getMe.json` | authenticated / unauthenticated | App auth state per case |
| listResumes | `openapi/fixtures/Resumes/listResumes.json` | default | 提供 flat resume 列表，两条 item 对应两行 |
| getResume | `openapi/fixtures/Resumes/getResume.json` | default | open action 后 detail preview 读取当前 resume |

User states:
- 未登录访问 `/resume_versions` 期待 auth gate 渲染。
- 登录后访问 `/resume_versions` 期待 list 渲染。
