# E2E.P0.019 Workspace Context Loading + Empty States

> **场景 ID**: E2E.P0.019
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given

4 种子场景：(A) getTargetJob=with-rounds+getResume=default+getPracticePlan=ready;
(B) getTargetJob=not-found; (C) getTargetJob=with-rounds+getResume=not-found;
(D) getPracticePlan=archived/not-found

## 2 When

分别加载 workspace route 并验证各 variant 的 UI 状态。

## 3 Then
- A: workspace 完整渲染，InterviewContext hydrate
- B: WorkspaceEmptyState testid 命中，CTA → home
- C: WorkspaceMissingResumeState testid 命中，CTA → resume_versions?flow=create
- D: plan refresh 正确处理 archived/not-found

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```
