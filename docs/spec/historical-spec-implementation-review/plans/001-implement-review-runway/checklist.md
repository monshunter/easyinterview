# Historical Spec Implement Review Runway Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-04

**关联计划**: [plan](./plan.md)

## Phase 1: Scope Inventory

- [x] 1.1 读取 `docs/work-journal/INDEX.md` 与最新工作日志，确认最近完成事项和未完成事项
  <!-- verified: 2026-05-04 evidence=docs/work-journal/INDEX.md + docs/work-journal/2026-05-04.md; latest completed work is docs(historical-spec): add implementation review runway; implementation not yet launched before this run -->
- [x] 1.2 读取强制校正参考资料：`docs/spec/product-scope/spec.md`、`docs/spec/engineering-roadmap/spec.md`、`docs/ui-design/INDEX.md`、目标相关 `docs/ui-design/*.md` 与 `ui-design/` 静态原型 / contract tests
  <!-- verified: 2026-05-04 evidence=docs/reports/2026-05-04-historical-spec-implementation-review-runway-verification.md#21-前置上下文; references include product-scope, engineering-roadmap, docs/ui-design active docs, ui-design/src/app.jsx, ui-design/src/data.jsx, ui-design/ui-design-contract.test.mjs -->
- [x] 1.3 读取 `docs/spec/INDEX.md` 与目标 subject 的 `plans/INDEX.md`，形成候选 historical spec / plan 清单
  <!-- verified: 2026-05-04 evidence=docs/reports/2026-05-04-historical-spec-implementation-review-runway-verification.md#22-候选-subject--plan-表; all real docs/spec subjects and plans/INDEX files inspected -->
- [x] 1.4 对候选 context.yaml 执行 validation，并记录 target、status、checklist 状态、依赖、写入范围和对应 product / roadmap / UI 参考章节
  <!-- verified: 2026-05-04 evidence=docs/reports/2026-05-04-historical-spec-implementation-review-runway-verification.md#23-context-validation-evidence; 15/15 context manifests passed validate_context.py for default targets -->
- [x] 1.5 标记每个候选 plan 的并行安全级别：read-only-parallel-safe / docs-write-serial / truth-source-serial / implementation-disjoint / implementation-serial
  <!-- verified: 2026-05-04 evidence=docs/reports/2026-05-04-historical-spec-implementation-review-runway-verification.md#24-parallelism-classification; truth-source serial zones recorded -->

## Phase 2: L1 Reconcile

- [x] 2.1 对候选 plan 执行 `/plan-review`，识别 spec / plan / checklist / context 与 product-scope、engineering-roadmap、docs/ui-design、ui-design 静态原型和工程 truth source 的漂移
  <!-- verified: 2026-05-04 evidence=docs/reports/2026-05-04-historical-spec-implementation-review-runway-verification.md#4-phase-2-l1-reconcile; findings: 9 plans missing quality gate classification, repo-scaffold lifecycle drift, engineering-roadmap future-child items, ai-gateway draft gate -->
- [x] 2.2 对需要修订的 subject 执行 `/plan-review --fix`，原地修复 spec / plan / checklist / context / INDEX
  <!-- verified: 2026-05-04 evidence=docs/reports/2026-05-04-historical-spec-implementation-review-runway-verification.md#43-applied-fixes; fixed 9 missing quality gate classifications and synced affected plans/INDEX via sync-doc-index --fix-index -->
- [x] 2.3 若 completed plan 需要修订，按规则切回 active 或追加 remediation 项，验证后再恢复 completed
  <!-- verified: 2026-05-04 evidence=repo-scaffold/001-bootstrap lifecycle closed to completed after 16/16 checklist items were already checked; engineering-roadmap future-child items intentionally left unchecked; ai-gateway 002 intentionally left draft -->
- [x] 2.4 执行 context validation、sync-doc-index、Markdown link check 与必要的 heading anchor audit，确认 L1 零漂移
  <!-- verified: 2026-05-04 evidence=docs/reports/2026-05-04-historical-spec-implementation-review-runway-verification.md#44-l1-verification-evidence; 15 context manifests PASS, sync-doc-index zero drift, check_md_links OK, heading anchor audit TOTAL 0, make docs-check PASS, git diff --check PASS -->

## Phase 3: Implementation Runway

- [x] 3.1 按依赖顺序选择下一个可执行 plan，并记录任何顺序调整原因
  <!-- verified: 2026-05-04 evidence=docs/reports/2026-05-04-historical-spec-implementation-review-runway-verification.md#51-selection-decision; no implementation target selected after L1 because remaining candidates are completed, draft-gated, future-child rules, or no-plan references -->
- [x] 3.2 通过 `/implement` 解析目标 context，并由原 plan 进入 `/tdd`
  <!-- skipped: 2026-05-04 no eligible original implementation plan after L1; current runway target was already entered through /implement docs and no additional subject was launched -->
- [x] 3.3 按 checklist phase 顺序实施，不跳项；每完成一项立即同步 checklist
  <!-- skipped: 2026-05-04 no additional implementation checklist executed; current runway checklist was updated in order with evidence comments -->
- [x] 3.4 为每个 phase 记录 Red / Green / Refactor、focused tests、contract gates、generated artifacts 和 BDD-Gate 证据（如适用）
  <!-- skipped: 2026-05-04 no code/generator/BDD phase was launched; L1 docs verification evidence recorded instead -->

## Phase 4: L2 Plan Code Review

- [x] 4.1 对已完成实现或 remediation 的 plan 执行 `/plan-code-review --fix`
  <!-- skipped: 2026-05-04 evidence=docs/reports/2026-05-04-historical-spec-implementation-review-runway-verification.md#6-phase-4-l2-plan-code-review; no new code implementation or generated artifact delta in this run -->
- [x] 4.2 将每个 finding 映射到原 checklist section；无法映射的 finding 降级 preview-only 并请用户确认 owner section
  <!-- skipped: 2026-05-04 no L2 findings produced in this run -->
- [x] 4.3 对已接受 finding 通过 `/tdd --section` 或原 plan owner 修复
  <!-- skipped: 2026-05-04 no accepted L2 finding to remediate -->
- [x] 4.4 运行 focused tests、相邻 regression tests 与受影响 contract gates，确认 L2 修复通过
  <!-- skipped: 2026-05-04 no L2 code remediation; final global gates run in Phase 6 -->

## Phase 5: Parallelism Gate

- [x] 5.1 并行只读发现任务完成后，由当前 owner 汇总结果并去重 finding
  <!-- verified: 2026-05-04 evidence=docs/reports/2026-05-04-historical-spec-implementation-review-runway-verification.md#7-phase-5-parallelism-gate; current owner integrated all inventory and findings -->
- [x] 5.2 写入并行前确认 path ownership、disjoint scope、INDEX / generated output / truth source 不冲突
  <!-- verified: 2026-05-04 all writes were docs-only and performed serially by current owner; no disjoint parallel write launched -->
- [x] 5.3 对 shared / OpenAPI / events / migrations / config / generated artifacts / 全局 INDEX 执行串行写入
  <!-- verified: 2026-05-04 truth-source serial zones had no code/generated artifact writes; global report/spec/plans INDEX updates were serialized -->
- [x] 5.4 多 agent 或多任务合并后，当前 owner 执行集成检查和相关验证
  <!-- verified: 2026-05-04 no multi-agent write merge occurred; current owner executed integration checks -->

## Phase 6: Final Reconcile

- [x] 6.1 执行 sync-doc-index、Markdown link check、docs/spec heading anchor audit、docs-check、codegen-check、目标 contract gates、focused tests、make test、make build、git diff --check
  <!-- verified: 2026-05-04 evidence=docs/reports/2026-05-04-historical-spec-implementation-review-runway-verification.md#8-phase-6-final-reconcile; docs-check PASS, heading anchor audit TOTAL 0, codegen-check PASS, make test PASS, make build PASS, git diff --check PASS -->
- [x] 6.2 执行 `/retrospective` 判断并记录报告需求；涉及 bugfix 时执行 `/bug-report` 判断
  <!-- verified: 2026-05-04 retrospective report created at docs/reports/2026-05-04-historical-spec-implementation-review-runway-assessment.md; bug-report not applicable because this was docs/governance remediation, not a bugfix -->
- [x] 6.3 执行 `/work-journal`，按逻辑边界拆分 commit，并验证 git status / git log
  <!-- verified: 2026-05-04 work journal entry prepared for docs(historical-spec): execute implementation review runway; git status/log verification performed after commit -->
- [x] 6.4 输出新对话交接摘要：已完成 subject / plan / target、剩余 checklist、验证证据、风险和下一步入口
  <!-- verified: 2026-05-04 final response will summarize completed plan, zero remaining checklist items, validation evidence, residual risks, and next entry points -->
