# Harness 文档与技能体系全量迁移 Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-07-17

## Phase 1: 冻结迁移合同

- [x] 1.1 建立 owner Spec、基线指标和迁移 fixture；验证：`python3 -m pytest scripts/harness_index_test.py -q`（2 passed）
- [x] 1.2 固化旧结构与必须保留质量边界的失败断言；验证：`python3 -m pytest scripts/harness_index_test.py -q`（4 passed，可识别六类旧包装层并锁定八项质量边界）

## Phase 2: 生成式索引与精确路由

- [ ] 2.1 实现可重建、commit-aware 索引；验证：index unit/contract tests
- [ ] 2.2 实现 commit 失效、精确路由和 low-confidence 结果；验证：routing fixture tests

## Phase 3: 文档结构全量迁移

- [ ] 3.1 将本计划转换为 `changes/2026-07-17-full-migration.md` 单一 Change；验证：change schema test
- [ ] 3.2 删除所有旧包装层；验证：repo structure zero-residual test
- [ ] 3.3 修复有效引用且保留必要历史证据；验证：Markdown link check + protected-history assertions

## Phase 4: 治理与 Skills 收敛

- [ ] 4.1 集中公共政策并实现 R0-R3 渐进加载；验证：governance contract tests
- [ ] 4.2 收敛 Skill 能力并移除旧级联/预读；验证：Skill contract tests
- [ ] 4.3 更新初始化、文档、索引、场景和收尾合同；验证：init/docs/scenario/closeout tests

## Phase 5: Replay 与成本验收

- [ ] 5.1 R1 replay 通过且不产生流程文档；验证：R1 replay fixture
- [ ] 5.2 R2 replay 只使用一个 Change 且保留跨层 gate；验证：R2 replay fixture
- [ ] 5.3 R2/R3 review 不信历史 PASS；验证：review replay fixture
- [ ] 5.4 预读量下降至少 50%，流程文件触碰中位数下降至少 50%，路由/缺陷指标不退化；验证：baseline comparison report

## Phase 6: 旧入口删除与全量收口

- [ ] 6.1 删除旧 validator、matcher、template 和 Skill 兼容入口；验证：zero-reference search
- [ ] 6.2 全量 test/build/lint/docs/link gate 通过；验证：根命令当前输出
- [ ] 6.3 冻结 Change 并完成 A1-A12 审计；验证：completion audit
