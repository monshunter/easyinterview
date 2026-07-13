# 001 Current Conversation Funnel Journey

> **版本**: 3.5
> **状态**: completed
> **更新日期**: 2026-07-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把既有 full-funnel owner 原地修订为当前连续对话漏斗的双层 gate：P0.098 自动化契约组合 + P0.099 共享真实环境浏览器验收。

## 2 范围

- 重写 P0.098/P0.099 scenario README、四段脚本和 expected outcome。
- 删除失去后端 server owner 的 `frontend/playwright.e2e.config.ts` 与专用 full-funnel spec。
- P0.099 真实流程覆盖 Mailpit 登录、resume/JD、连续聊天、完成、异步报告与截图。
- 修复真实 PostgreSQL 流程暴露的当前契约漂移。
- 历史 Phase 2 曾同步 P0.100 术语；当前所有 P0.100 Phase/BDD/可靠性工作由 `002-manual-uat-real-provider-full-funnel` 独占，本 plan 不再修改其场景合同。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract + tooling + real integration。
- **TDD 策略**: 运行时修复必须先由 focused Go test 复现，再实现并回归。
- **BDD 策略**: `E2E.P0.098` 与 `E2E.P0.099` 分别覆盖自动契约组合和真实用户路径；P0.100 由 plan 002 独立验收，不是本 plan 的 output-digest dependency，也不进入本 plan BDD。
- **替代验证 gate**: codegen, migration, full backend/frontend, prompt/eval, docs/index, negative search。

## 4 Coverage Matrix

| 行为 | 类型 | Gate | 负向断言 |
|------|------|------|----------|
| plan → continuous chat | primary | P0.022/P0.023/P0.044/P0.098/P0.099 | empty array NULL, append-event |
| chat failure/retry | recovery | P0.046/P0.098 | duplicate messages/provider calls |
| completion → report | primary | P0.047/P0.056/P0.099 | stale event column |
| report persistence/retry | recovery | P0.058/P0.098/P0.099 | JSON written to text[], stuck generating |
| voice disabled | boundary | P0.007/P0.099 | clickable control/provider call |
| desktop/mobile UI | responsive | P0.099 + pixel parity | sidebar/counter/current question |

## 5 实施步骤

### Phase 1: Real-path Red/Green

- Wire resume parse through shared observability.
- Normalize optional focus codes to non-null PostgreSQL arrays.
- Remove stale completion event columns.
- Persist report retry focus as PostgreSQL `text[]` and make generating retry idempotent.

### Phase 2: Scenario reconciliation

- Convert P0.098 to current contract composition.
- Convert P0.099 to hybrid shared-environment browser acceptance.
- Delete the orphaned dedicated Playwright server/config/spec.
- Record the historical P0.100 terminology handoff only; active P0.100 ownership is plan 002.

### Phase 3: Real browser acceptance

- Run shared environment reset/redeploy and verify services.
- Complete Mailpit login, resume/JD import, continuous chat, finish/report.
- Capture exact six-image report/generating desktop/mobile matrix plus redacted evidence.

### Phase 4: Closeout

- Run focused/full tests, scenarios, codegen/migration/prompt/eval/docs gates.
- Run active negative search and post-pass reconcile.
- Record bug/retrospective/work journal and restore documents to completed.

### Phase 5: P0.099 exact six-image report acceptance

- Create current-run en/zh ready rows for P0.099 and collect exactly six redacted full-page images at 1440x1200 and 390x844。For every ready row, bind DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` and report/session/context digest；P0.100 output digest is not a prerequisite.
- Both desktop and 390x844 real report images cover the action region and prove their actual English / zh-CN labels satisfy `<=24 whitespace words` / `<=64 Unicode code points` with complete visibility, no clipping/ellipsis/hidden content and mobile `scrollWidth=390`。Separate deterministic ui-design/OpenAPI boundary fixtures use exactly24 words /64 code points and pass prototype/formal pixel parity。200-code-point fuse and18/52 targeted-repair margin cannot pass either gate.

## 6 验收标准

- Current real funnel reaches a ready report and the DB evidence matches current schemas.
- P0.098/P0.099 scripts pass without no-op or stale infrastructure.
- Six full-page screenshots close each ready row's current-run canonical audit chain；desktop+390 action regions prove actual `<=24-whitespace-word` / `<=64-Unicode-code-point` copy is fully visible without clipping/ellipsis/overflow，while deterministic parity proves exact boundaries.
- Active artifacts have zero stale structured-question contract references.

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-13 | 3.5 | Bind final P0.099 run12381 exact-six evidence and restore the fully checked owner plan to completed. |
| 2026-07-13 | 3.4 | Finalize scheme A P0.099 UX：24-whitespace-word/64-code-point exact boundary，desktop+390 complete wrapping；200 fuse and18/52 repair margin are not UX evidence. |
| 2026-07-13 | 3.3 | Close P0.099 through per-ready-row current-run canonical audit + screenshot digests；prove exact 14/40 only with deterministic parity and keep P0.100 independent. |
| 2026-07-12 | 3.2 | Keep Phase 5 scoped to P0.099 exact six-image acceptance and hand all active P0.100 reliability work to plan 002. |
| 2026-07-12 | 3.1 | Reopen Phase 5 for exact six-image grounded report acceptance and content-level P0.100 reliability. |
| 2026-07-12 | 3.0 | Rebase the completed owner onto the continuous-conversation real acceptance flow. |
