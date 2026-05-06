# Seed Input

- 用户登录态：未登录（`getMe` `unauthenticated` scenario，401 + B1 error envelope）。
- 起始 route：`workspace`，无业务参数。
- 待恢复 pending action：

| 字段 | 值 |
|------|----|
| type | `start_practice` |
| label | `立即面试` |
| route | `practice` |
| planId | `plan-tj-1` |
| targetJobId | `tj-1` |
| jdId | `jd-tj-1` |
| resumeVersionId | `frontend-v3` |
| roundId | `round-manager` |

- API fixture：
  - `getRuntimeConfig` `default`
  - `getMe` `unauthenticated`
  - `startAuthEmailChallenge` `default`
  - `verifyAuthEmailChallenge` `default`（响应包含 `Set-Cookie:
    ei_session=…` 模拟 first-party session 签发）
- 浏览器状态：jsdom 默认；不预设 localStorage / hash route / language override。
