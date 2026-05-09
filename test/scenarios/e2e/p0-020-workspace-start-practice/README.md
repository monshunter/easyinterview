# E2E.P0.020 Start Practice + Auth Gate

> **场景 ID**: E2E.P0.020
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given
5 子场景：(A1) 已登录+无 plan→首次面试；(A2) createPracticePlan 422 missing-resume；
(A3) startPracticeSession 502+重试；(B1) 有 plan+ready→跳过 create；(C1) 未登录→requestAuth→登录恢复

## 2 When
点击 workspace-cta-start 触发立即面试流程

## 3 Then
- A1: createPracticePlan + startPracticeSession 双步调用，Idempotency-Key 双键稳定，nav practice
- A2: 422 error inline + focus 更换简历，不进入 startPracticeSession
- A3: 502 retry Idempotency-Key 复用，3 次失败 fallback CTA
- B1: 跳过 createPracticePlan
- C1: requestAuth 触发 auth_login 携带 autoStartPractice=1

## 4 执行
```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```
