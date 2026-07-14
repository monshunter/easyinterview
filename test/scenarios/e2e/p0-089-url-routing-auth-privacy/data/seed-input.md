# Seed Input

| 触发器 | Payload |
|--------|---------|
| unauthenticated Reports direct-link | `/reports?targetJobId=01918fa0-0000-7000-8000-000000002000&section=reports&reportId=rpt-hostile&status=ready&roundId=round-hostile` + 19 个 raw markers；期望转为 `pendingRoute=reports` |
| `useRequestAuth` → `立即面试` | `PendingAction.route=practice`，params 含 `sessionId / planId / targetJobId / jdId / resumeId / roundId / mode / modality` + 19 个 raw markers |
| hostile `/auth/login` 直接打开 | `pendingRoute=reports` + `targetJobId` + route-incompatible `section / reportId / status / roundId` + 19 个 raw markers |
| hostile browser history popstate | `window.history.pushState` 写入 `/workspace?targetJobId=tj-popstate&resumeId=rv-private&planId=plan-private&autoStartPractice=1&unknownKey=private&rawText=...&token=...#prompt` 与 raw `history.state` 后触发 `popstate`；只允许保留只读详情定位符 `targetJobId` |

19 个 raw marker 覆盖 `rawText` / `rawDescription` / `sourceUrl` /
`query` / `label` / `guidedAnswers` / `parsedSummary` /
`structuredProfile` / `suggestion` / `originalBullet` / `suggestedBullet` /
`questionText` / `answerText` / `notes` / `prompt` / `response` / `file` /
`token` / `password`。每个 marker 含唯一 hex 后缀（如 `RAW_JD_TEXT_2c1a`）
便于 trigger.log 反向 grep。

Auth fixture：`Auth/getMe.json`、`Auth/getRuntimeConfig.json`、
`Auth/startAuthEmailChallenge.json`、`Auth/verifyAuthEmailChallenge.json`，
默认未登录态，verify 完成后切换为 authenticated。
