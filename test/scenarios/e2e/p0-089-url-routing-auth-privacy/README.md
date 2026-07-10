# E2E.P0.089 URL Routing — Auth pendingAction + Privacy Redline

> **场景 ID**: E2E.P0.089
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

未登录用户在带有 Plan 004 routeStore + pendingAction safe-param 过滤的正式
前端中触发 auth-gated workflow URL；测试 harness 还在 pendingAction params
里塞入代表性的 raw 标记（raw JD 文本、source URL、out-of-scope job-search 查询 /
标签、简历正文、out-of-scope guided answers、解析摘要、structured profile、suggestion、
question/answer 正文、debrief 备注、AI prompt/response token、auth secret
token / password 等）。

测试在 vitest + jsdom 中运行，使用 fixture-backed
`startAuthEmailChallenge` / `verifyAuthEmailChallenge` / `getMe` 切换 mock
session。

## 2 When

- 未登录用户点击 `立即面试`，触发 `useRequestAuth` → navigate
  `/auth/login`；URL params 含 encoded pendingAction + safe handoff keys。
- 在 `/auth/login` 输入邮箱 → `/auth/verify` 输入验证码 → 完成
  email-code mock 登录。
- App 自动接续目标 route (`/practice?...`)。
- 用 hostile URL 直接打开 `/auth/login?...rawText=...&token=...`，验证
  parseUrlToRoute / decodePendingActionRoute 的 allowlist 拦截。
- 从浏览器 back stack 回到 hostile `/workspace?...rawText=...#prompt` entry，验证
  popstate restore 会立即改写为 canonical URL 并清空 raw `history.state`。

## 3 Then

- 重定向到 `/auth/login` 后 URL 仅包含 `pendingRoute` / `pendingType` /
  `pendingLabel` + practice safe params（含 `planId` / `targetJobId` /
  `jdId` / `resumeId` / `roundId` / `sessionId`）。
- verify 完成后 URL 重写为 `/practice?...`，保留全部 6 个 safe handoff key。
- 任意 hostile 输入下，URL、`window.history.state`、`localStorage`、
  `sessionStorage`、console capture 都 ZERO 命中 raw 标记。
- hostile popstate 后地址栏只保留 `/workspace?targetJobId=...`，hash 与
  raw `history.state` 均被 scrub。
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

场景在 vitest + jsdom 中运行，不写共享数据库，不启动 Kind cluster；trigger.sh
仅产生 `.test-output/e2e/p0-089-url-routing-auth-privacy/trigger.log` 作为
验证证据，cleanup.sh 删除 setup marker，保留日志。
