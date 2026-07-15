# Roadmap Rebaseline and Subspec Governance Checklist

> **版本**: 3.5
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 1: 当前 truth source 投影

- [x] 1.1 ADR-Q1..Q6 作为当前认证、异步、分析、部署、隐私和 AI 路由架构约束投影到 roadmap spec。
  <!-- verified: 2026-07-06 method=roadmap-plan-current-governance-reconcile evidence="engineering-roadmap spec v3.29 keeps ADR-Q1..Q6 as current constraints; plan v3.4 describes them as architecture constraints rather than spawn drivers." -->
- [x] 1.2 当前基础、契约、质量、产品和 UI owner spec 作为后续实施依赖。
  <!-- verified: 2026-07-06 method=roadmap-plan-current-governance-reconcile evidence="plan v3.4 Phase 1.2 and roadmap spec current workstream table point implementation to active owner specs and coded truth sources." -->
- [x] 1.3 `docs/spec/INDEX.md` 只投影真实存在且 Header 合规的 spec。
  <!-- verified: 2026-07-06 method=sync-doc-index evidence="sync-doc-index --fix-index updated engineering-roadmap plan index and docs/spec index during product-scope 1.38; follow-up sync-doc-index --check PASS." -->

## Phase 2: Roadmap 当前实施地图

- [x] 2.1 对齐 `product-scope`、`docs/ui-design/` 与 `frontend/src` 当前模块、route、上下文和范围外边界。
  <!-- verified: 2026-07-06 method=roadmap-plan-current-governance-reconcile evidence="plan v3.4 Phase 2.1 names product-scope, docs/ui-design, and frontend/src as the current truth sources." -->
- [x] 2.2 `engineering-roadmap/spec.md` 只描述当前 truth source 关系、active owner、P0 workstream、future candidates、实施顺序和验收标准。
  <!-- verified: 2026-07-06 method=engineering-roadmap-current-map-wording-reconcile evidence="engineering-roadmap spec v3.29 updated background, scope, decisions, constraints, workstreams, implementation sequence, and acceptance criteria to current execution-map wording." -->
- [x] 2.3 修订本 plan、checklist、context 和 plans/INDEX 为 roadmap rebaseline、按需 child 创建和 no-pending INDEX 合同。
  <!-- verified: 2026-07-06 method=roadmap-plan-current-governance-reconcile evidence="plan/checklist bumped to v3.4; context specVersion.to bumped to 3.29 and discovery keywords now use current-map / no-pending / out-of-scope-boundary wording." -->
- [x] 2.4 交叉引用和索引投影指向当前 roadmap 版本和当前 plan 版本。
  <!-- verified: 2026-07-06 method=sync-doc-index evidence="sync-doc-index --fix-index updated docs/spec/INDEX.md engineering-roadmap version to 3.29 and product-scope plans INDEX to 1.38; engineering-roadmap plans INDEX will project plan v3.4 after this checklist update." -->
- [x] 2.5 文档一致性验证通过。
  <!-- verified: 2026-07-06 method=docs-gates evidence="validate_context.py engineering-roadmap/001 docs PASS; sync-doc-index --check PASS; make docs-check PASS; git diff --check PASS." -->

## Phase 3: 后续 child 创建规则

- [x] 3.1 创建 child spec / plan 前必须确认 `product-scope` 与 UI 设计文档已明确保留对应用户行为或工程能力。
  <!-- verified: 2026-07-06 method=roadmap-plan-current-governance-reconcile evidence="plan v3.4 Phase 3.1 records current behavior / engineering capability as the first creation precondition." -->
- [x] 3.2 创建 child plan 时必须同步 `context.yaml`、`plan.md`、`checklist.md`，涉及用户行为时同步 BDD plan / checklist。
  <!-- verified: 2026-07-06 method=roadmap-plan-current-governance-reconcile evidence="plan v3.4 Phase 3.1 records paired context, plan, checklist, TDD, and BDD requirements." -->
- [x] 3.3 Future candidates 只有在产品 / UI / 合规设计确认后才创建 owner 文档。
  <!-- verified: 2026-07-06 method=roadmap-plan-current-governance-reconcile evidence="plan v3.4 Phase 3.3 and roadmap spec future candidate section keep future items out of INDEX and owner packages until design confirmation." -->

## Phase 4: 技术草稿引用边界

- [x] 4.1 `product-scope` §1.5 持有当前技术契约 owner matrix，roadmap 只消费当前 owner spec / coded truth source。
  <!-- verified: 2026-07-06 method=roadmap-plan-current-governance-reconcile evidence="plan v3.4 Phase 4.1 points technical contract ownership to product-scope §1.5 owner matrix." -->
- [x] 4.2 当前项目文档、代码注释、生成源、生成物、日志与报告不得把范围外技术草稿目录名或文件名作为 truth source。
  <!-- verified: 2026-07-06 method=roadmap-plan-current-governance-reconcile evidence="plan v3.4 Phase 4.2 records technical-draft zero-reference as the current boundary." -->
- [x] 4.3 `shared/conventions.yaml`、lint 脚本、codegen source 与 generated artifacts 的字段和 gate 归属到当前 owner。
  <!-- verified: 2026-07-06 method=roadmap-plan-current-governance-reconcile evidence="plan v3.4 Phase 4.3 assigns shared conventions, OpenAPI, DB, event, and observability gates to current owners." -->
- [x] 4.4 Technical-draft zero-reference gate 固化到本 plan。
  <!-- verified: 2026-07-06 method=roadmap-plan-current-governance-reconcile evidence="plan v3.4 Phase 4.4 requires zero-reference checks for directory names, filenames, shorthand, Markdown links, implementation prerequisites, and external truth-source wording." -->

## Phase 5: Context discovery deduplication

- [x] 5.1 删除 context discovery 中重复的 `AI provider` keyword；结构化 YAML 扫描确认全仓 context 同列表标量值零重复，owner context 与 docs gates 通过。<!-- verified: 2026-07-10 method=yaml-duplicate-scan evidence="Baseline scan found only engineering-roadmap spec.discovery.keywords AI provider x2; GREEN scan reports zero duplicate scalar entries across all plan contexts." -->
