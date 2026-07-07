# Frontend Resume Workshop Create Flow BDD Plan

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 覆盖范围 | 关联 Checklist |
|---------|------|----------|----------------|
| E2E.P0.081 | primary | CreateFlow upload / paste happy path, presign, register, parse polling, privacy and UI parity | 5.1 |
| E2E.P0.082 | failure / recovery | Parse failure, timeout, cancel-and-return and retry-from-input | 5.2 |
| E2E.P0.083 | primary + handoff | Preview confirm, Home CTA, Workspace CTA and auth pending action | 5.3 |

## 2 场景说明

### E2E.P0.081

Given an authenticated user opens the Resume Workshop create route.
When the user completes upload or paste creation and parsing reaches ready.
Then CreateFlow reaches preview, side-effect requests use IK, polling omits IK, UI parity anchors render, and raw resume content is not stored in route state or browser storage.

### E2E.P0.082

Given a user is in parse progress.
When parsing fails, times out, or the user cancels.
Then the failed state, retry state and return-to-input state are visible, and the original input remains local to the component.

### E2E.P0.083

Given a ready parsed resume is shown in preview confirm.
When the user confirms or enters from Home / Workspace CTA.
Then the flat resume save path completes or shows inline validation, and auth pending action keeps only safe route params.
