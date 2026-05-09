# E2E.P0.021 Workspace Handoff + Privacy + Legacy Negative

> **场景 ID**: E2E.P0.021
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given
已登录，getTargetJob=with-rounds，sessionHistory 缺 typed contract

## 2 When
点击 workspace-companyintel-open；点击 workspace-history-empty；隐私反查；旧入口负向 grep

## 3 Then
- A: nav("company_intel", {targetJobId, jdId})，getCompanyIntel 调用 0
- B: history 区域 EmptyHistory / disabled placeholder，点击不触发 report nav
- D: JD原文/简历正文/questionText 不在 console/URL/localStorage/telemetry
- E: 旧 testid + 旧 route alias + prototype data import grep 0 命中
- F: D1+D2+D3 regression PASS

## 4 执行
```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```
