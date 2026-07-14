# E2E.P0.089 URL Routing — Auth pendingAction + Privacy Redline

> **场景 ID**: E2E.P0.089
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

未登录用户在带有 Plan 004 routeStore + pendingAction safe-param 过滤的正式
前端中直接打开 `/reports?targetJobId=<uuid>`；测试 harness 还在 URL 与 pendingAction params
里塞入代表性的 raw 标记（raw JD 文本、source URL、out-of-scope job-search 查询 /
标签、简历正文、out-of-scope guided answers、解析摘要、structured profile、suggestion、
question/answer 正文、debrief 备注、AI prompt/response token、auth secret
token / password 等）。

测试在 vitest + jsdom 中运行，使用 fixture-backed
`startAuthEmailChallenge` / `verifyAuthEmailChallenge` / `getMe` 切换 mock
session。

## 2 When

- 未登录用户直接打开带 `section=reports` / `reportId` / `status` / `roundId`
  与 raw markers 的 Reports deep-link；App 将其转换为 `pendingRoute=reports` 的
  `/auth/login`，只保留 `targetJobId`。
- 在 `/auth/login` 输入邮箱 → `/auth/verify` 输入验证码 → 完成
  email-code mock 登录。
- App 自动接续目标 route (`/reports?targetJobId=<uuid>`)。
- 继续验证 `立即面试` 的 practice pendingAction safe handoff keys 回归。
- 用指向 reports 且携带 route-incompatible params / raw markers 的 hostile
  `/auth/login?...` 直接打开输入，验证 parseUrlToRoute /
  decodePendingActionRoute 的 allowlist 拦截。
- 将 hostile `/workspace?...rawText=...#prompt` entry 写入浏览器 history 并触发
  `popstate`，验证 restore 会立即改写为 canonical URL 并清空 raw
  `history.state`。

## 3 Then

- Reports 重定向到 `/auth/login` 后 URL 仅包含 pendingAction reserved metadata
  与 `targetJobId`；不包含 `section` / `reportId` / `status` / `roundId`。
- verify 完成后 URL 重写为 `/reports?targetJobId=<uuid>`，不恢复其他报告 authority。
- practice 相邻回归仍保留 `planId` / `targetJobId` / `jdId` / `resumeId` /
  `roundId` / `sessionId` 六个 safe handoff key。
- hostile `/auth/login` direct-open 只保留 Reports 的 `targetJobId` 与 pendingAction reserved metadata。
- 任意 hostile 输入下，URL、`window.history.state`、`localStorage`、
  `sessionStorage`、console capture 都 ZERO 命中 raw 标记。
- hostile popstate 后地址栏规范化为 query-free `/workspace`，hash 与 raw
  `history.state` 均被 scrub。
- `token` / `password` / `prompt` / `response` 等敏感字段在所有 surface
  都缺失。

## 4 执行

```bash
./test/scenarios/e2e/p0-089-url-routing-auth-privacy/scripts/setup.sh
./test/scenarios/e2e/p0-089-url-routing-auth-privacy/scripts/trigger.sh
./test/scenarios/e2e/p0-089-url-routing-auth-privacy/scripts/verify.sh
./test/scenarios/e2e/p0-089-url-routing-auth-privacy/scripts/cleanup.sh
```

## 5 污染控制

场景在 Python source contract + vitest + jsdom 中运行，不写共享数据库；trigger.sh
仅产生 `.test-output/e2e/p0-089-url-routing-auth-privacy/trigger.log` 作为
验证证据，cleanup.sh 删除 setup marker，保留日志。
