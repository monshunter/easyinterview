# 001 — Honest Grounded Report Screen Checklist

> **版本**: 3.5
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1-5: Conversation-level baseline（历史已完成）

- [x] Conversation report、generating、replay、routing and parity baseline completed.

## Phase 6: Honest GeneratingScreen

- [x] Remove fake progress/observations/notify and render only backend queued/generating/ready/failed truth.
- [x] Polling preserves attempt/delay across hidden/blur, resumes at n+1, rejects duplicate concurrency and keeps one run `<=49` calls.
- [x] Typed timeout/network/context-too-large/failure actions expose no provider/async attempt details.

## Phase 7: Direct ReportDashboard contract

- [x] Consume generated direct-report shape and fail closed on unknown/malformed context/focus.
- [x] Replay/next requests send no client focus/settings and use server-owned projection.
- [x] English 24/25 and zh-CN 64/65 code tests, delimiter parity and no raw/truncation/rewrite pass.

## Phase 8: Visual and real-environment separation

- [x] Deterministic formal DOM/style/viewport component assertions run as a frontend code gate, not E2E.
- [x] BDD-Gate: `BDD.REPORT.UI.001` 由 [BDD checklist](./bdd-checklist.md) 关联 report/generating owner behavior tests。
- [x] E2E-HANDOFF: P0.099 是唯一 real report/generating owner，要求 exactly six `fullPage: true` images 绑定 current API/DB/report/session/context/screenshot digests；本轮未运行，状态仍为 `Ready`。
- [x] P0.099 contract 要求 real mobile ready images 完整显示 action region 且无 clipping/ellipsis/hiding/overflow；exact 24/64 保持 code test。

## Phase 9: Context Strip privacy

- [x] Target/round/resume stay visible while report/session UUIDs are absent from text、tooltip、ARIA and accessible names.
- [x] Formal real-backend acceptance screenshots and manifest use bounded redacted state/hash/viewport evidence only.

## Phase 10: ReportsScreen

- [x] Current target joins canonical round display and renders current/latest only; cross-target/stale/mismatch data fail closed.
- [x] ReportsScreen is the sole list consumer; Parse/Report/Generating have zero list calls and no global/history center exists.
- [x] Report/Generating Back uses trusted target or Workspace fallback while routes stay reportId-only.

## Phase 11: Command/read navigation

- [x] Reports Back reaches targetJobId-only Workspace detail directly with no Parse detour、animation、import or polling.
- [x] Focused component/route/source tests and deterministic parity pass.

## Closeout

- [x] Root `make test` is the independent complete backend/frontend unit regression gate; focused test PASS is development feedback, not full regression.
- [x] P0.099、typecheck/build/lint/docs/index/diff are reported as separate gates; code gates are never wrapped as E2E.
