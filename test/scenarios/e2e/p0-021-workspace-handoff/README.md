# E2E.P0.021 Workspace Boundary + Privacy + Out-of-scope Negative

> **场景 ID**: E2E.P0.021
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given
已登录，workspace plan-list 是当前 runtime；insight / records typed consumers 不由 workspace runtime 拼接。

## 2 When
运行 workspace source negative、report replay handoff regression、隐私反查和 out-of-scope 入口负向 grep。

## 3 Then
- A: workspace runtime 不调用独立 insight API / report API，不从 prototype helper 或 TargetJob extension 拼接记录行
- B: report replay handoff regression 由 report owner 覆盖，workspace 只保留 source-level 边界
- D: JD原文/简历正文/questionText 不在 console/URL/localStorage/telemetry
- E: out-of-scope testid + prototype data import grep 0 命中
- F: D1+D2+D3 regression PASS

## 4 执行
```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```
