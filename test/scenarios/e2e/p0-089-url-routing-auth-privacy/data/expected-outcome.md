# Expected Outcome

| 维度 | 期望 |
|------|------|
| `/auth/login` URL (重定向后) | pathname `/auth/login`；search 含 `pendingRoute=practice` / `pendingType=start_practice` / `planId=plan-1` / `targetJobId=tj-1` / `sessionId=01918fa0-...`；不含任何 raw marker |
| 登录 → verify 成功后 URL | pathname `/practice`；search 含 6 个 safe handoff key (`planId` / `targetJobId` / `jdId` / `resumeVersionId` / `roundId` / `sessionId`)；不含 raw marker |
| jd_match restore URL | `/jd-match?selectedJobMatchId=jm-restored&pendingJdMatchActionId=pjm-restored&tab=search`；不含 `query` / `label` 等 raw marker |
| Hostile `/auth/login` direct-open | search 保留 `pendingRoute=workspace` / `pendingType=start_practice` / `planId` / `targetJobId`；其余 raw marker 被 allowlist 拦截 |
| `window.history.state` | 始终为 `null`，不写入 React state 中的 raw 字段 |
| `localStorage` / `sessionStorage` | 完全空；测试在每个用例前 clear |
| console.log / warn / error capture | 不含任何 raw marker |

| 反向断言 | 含义 |
|----------|------|
| 所有 raw marker 在 URL / history / storage / console 出现 0 次 | Plan 004 §3.2 redline |
| `token=AUTH_SECRET_TOKEN_3745` / `password=AUTH_PASSWORD_4856` 全 surface 0 次 | auth secret 不进 query / pendingAction |
| `pendingRoute` / `pendingType` / `pendingLabel` 在 verify 完成后 URL 中不出现 | 还原后 reserved key 已剥离 |

证据：`.test-output/e2e/p0-089-url-routing-auth-privacy/trigger.log` 必须
出现 `Tests 3 passed (3)` 与 `Test Files 1 passed (1)` marker。
