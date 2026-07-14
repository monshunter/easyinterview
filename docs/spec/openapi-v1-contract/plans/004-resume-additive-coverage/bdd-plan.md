# OpenAPI v1 Resume Summary BDD Plan

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-07-14

## Phase 7: Resume list summary / detail split

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.034 | Backend register and summary list | 两个隔离用户拥有 upload/paste、不同 parse/readability 状态的简历 | 用户注册、分页调用 `listResumes` 并显式调用一条 `getResume` | list items 只返回九字段 closed `ResumeSummary` 且 readable fact 正确；跨用户隐藏、分页不变；显式 detail 返回 full `Resume` | `test/scenarios/e2e/p0-034-resume-register-and-list/` |
| E2E.P0.036 | Flat list auth and navigation | 未登录或已登录用户进入 Resume Workshop 列表 | route gate 或已登录列表加载并点击一行 open action | 未登录 0 API；已登录列表只消费 summary、每 item 一行且不 N+1；点击后进入 `resumeId` detail route | `test/scenarios/e2e/p0-036-resume-flat-list-auth-boundary/` |
| E2E.P0.037 | Full read-only detail | 用户从列表进入 owned resume detail，或使用缺失/不同 parse 状态的 resumeId | detail route 调用 `getResume` | full Resume 正文/source/read-only/failure/polling 合同保持；detail 不依赖 list 内详情字段，404 隐私边界不变 | `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/` |

### 场景复用约束

- 本 plan 复用既有三个 scenario 目录，不创建平行场景编号或复制 runner。
- scenario owner 在实施 Phase 更新现有断言以覆盖 summary-only list、full detail 与请求边界；历史 PASS 不作为本次 evidence。
- 主 checklist 只记录 `BDD-Gate` 汇总状态，逐场景证据记录在 `bdd-checklist.md`。
