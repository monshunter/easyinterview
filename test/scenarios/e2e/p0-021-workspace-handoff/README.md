# E2E.P0.021 Workspace Embedded-only + Privacy + Non-current Negative

> **场景 ID**: E2E.P0.021
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given
已登录，getTargetJob=with-rounds，当前规划记录 consumer 不从 TargetJob extension 取数

## 2 When
点击 workspace-insight-open；点击 records placeholder；隐私反查；非当前入口负向 grep

## 3 Then
- A: 点击 workspace insight 卡片动作会留在 `workspace` route，并保留
  `{targetJobId, jdId}`；不调用独立 insight API
- B: 当前规划记录区域 Empty placeholder，点击不触发 report nav
- D: JD原文/简历正文/questionText 不在 console/URL/localStorage/telemetry
- E: 非当前 testid + route alias + prototype data import grep 0 命中
- F: D1+D2+D3 regression PASS

## 4 执行
```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```
