# OpenAPI v1 Resume Summary BDD Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.034 Backend register and summary list

- [x] 更新既有场景断言：list item 精确为九字段 `ResumeSummary`，无详情/provenance。
- [x] 证明 upload/paste、parse/readability、cursor 与 cross-user 语义不变。
- [x] 显式 `getResume` 仍返回 full `Resume`，并记录当前 setup/trigger/verify/cleanup 证据。
  <!-- verified: 2026-07-14 scenario=E2E.P0.034 evidence="fresh four-stage PASS; exact nine-field list projection, full getResume, cursor/cross-user and real PostgreSQL gates" -->

## E2E.P0.036 Flat list auth and navigation

- [x] 保留未登录 0 `listResumes/getResume` 的 route-level gate。
- [x] 登录列表只使用 summary 字段并拒绝行级 N+1 `getResume`；每个 fixture item 仍对应一行。
- [x] 点击 open action 后才进入 `resumeId` detail route，并记录当前 setup/trigger/verify/cleanup 证据。
  <!-- verified: 2026-07-14 scenario=E2E.P0.036 evidence="fresh four-stage PASS; unauthenticated zero API, StrictMode list=1, retry=2 and navigation-only detail fetch" -->

## E2E.P0.037 Full read-only detail

- [x] 证明 detail route 通过 `getResume` 获取 full `Resume`，不依赖 list item 中已删除字段。
- [x] 保留 PDF/source、Markdown body、pending polling、failed-with-readable 单次请求与 404 privacy assertions。
- [x] 记录当前 setup/trigger/verify/cleanup 证据；历史 PASS 不得勾选本项。
  <!-- verified: 2026-07-14 scenario=E2E.P0.037 evidence="fresh four-stage PASS; full-detail ready/retry/polling/readable-failed/source/404 matrix with maxInFlight=1" -->

## Phase 7: 汇总 Gate

- [x] 三个复用场景均以当前 source/fixture/generated/backend/frontend 实现通过。
- [x] 主 checklist `7.6 BDD-Gate` 只在上述场景全部通过并记录证据后勾选。
  <!-- verified: 2026-07-14 evidence="P0.034/P0.036/P0.037 current source/fixture/generated/backend/frontend evidence is green and main 7.6 is closed" -->
