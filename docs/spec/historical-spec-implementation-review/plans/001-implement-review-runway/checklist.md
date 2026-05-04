# Historical Spec Implement Review Runway Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-04

**关联计划**: [plan](./plan.md)

## Phase 1: Scope Inventory

- [ ] 1.1 读取 `docs/work-journal/INDEX.md` 与最新工作日志，确认最近完成事项和未完成事项
- [ ] 1.2 读取强制校正参考资料：`docs/spec/product-scope/spec.md`、`docs/spec/engineering-roadmap/spec.md`、`docs/ui-design/INDEX.md`、目标相关 `docs/ui-design/*.md` 与 `ui-design/` 静态原型 / contract tests
- [ ] 1.3 读取 `docs/spec/INDEX.md` 与目标 subject 的 `plans/INDEX.md`，形成候选 historical spec / plan 清单
- [ ] 1.4 对候选 context.yaml 执行 validation，并记录 target、status、checklist 状态、依赖、写入范围和对应 product / roadmap / UI 参考章节
- [ ] 1.5 标记每个候选 plan 的并行安全级别：read-only-parallel-safe / docs-write-serial / truth-source-serial / implementation-disjoint / implementation-serial

## Phase 2: L1 Reconcile

- [ ] 2.1 对候选 plan 执行 `/plan-review`，识别 spec / plan / checklist / context 与 product-scope、engineering-roadmap、docs/ui-design、ui-design 静态原型和工程 truth source 的漂移
- [ ] 2.2 对需要修订的 subject 执行 `/plan-review --fix`，原地修复 spec / plan / checklist / context / INDEX
- [ ] 2.3 若 completed plan 需要修订，按规则切回 active 或追加 remediation 项，验证后再恢复 completed
- [ ] 2.4 执行 context validation、sync-doc-index、Markdown link check 与必要的 heading anchor audit，确认 L1 零漂移

## Phase 3: Implementation Runway

- [ ] 3.1 按依赖顺序选择下一个可执行 plan，并记录任何顺序调整原因
- [ ] 3.2 通过 `/implement` 解析目标 context，并由原 plan 进入 `/tdd`
- [ ] 3.3 按 checklist phase 顺序实施，不跳项；每完成一项立即同步 checklist
- [ ] 3.4 为每个 phase 记录 Red / Green / Refactor、focused tests、contract gates、generated artifacts 和 BDD-Gate 证据（如适用）

## Phase 4: L2 Plan Code Review

- [ ] 4.1 对已完成实现或 remediation 的 plan 执行 `/plan-code-review --fix`
- [ ] 4.2 将每个 finding 映射到原 checklist section；无法映射的 finding 降级 preview-only 并请用户确认 owner section
- [ ] 4.3 对已接受 finding 通过 `/tdd --section` 或原 plan owner 修复
- [ ] 4.4 运行 focused tests、相邻 regression tests 与受影响 contract gates，确认 L2 修复通过

## Phase 5: Parallelism Gate

- [ ] 5.1 并行只读发现任务完成后，由当前 owner 汇总结果并去重 finding
- [ ] 5.2 写入并行前确认 path ownership、disjoint scope、INDEX / generated output / truth source 不冲突
- [ ] 5.3 对 shared / OpenAPI / events / migrations / config / generated artifacts / 全局 INDEX 执行串行写入
- [ ] 5.4 多 agent 或多任务合并后，当前 owner 执行集成检查和相关验证

## Phase 6: Final Reconcile

- [ ] 6.1 执行 sync-doc-index、Markdown link check、docs/spec heading anchor audit、docs-check、codegen-check、目标 contract gates、focused tests、make test、make build、git diff --check
- [ ] 6.2 执行 `/retrospective` 判断并记录报告需求；涉及 bugfix 时执行 `/bug-report` 判断
- [ ] 6.3 执行 `/work-journal`，按逻辑边界拆分 commit，并验证 git status / git log
- [ ] 6.4 输出新对话交接摘要：已完成 subject / plan / target、剩余 checklist、验证证据、风险和下一步入口
