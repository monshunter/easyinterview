# Harness 工程框架渐进收敛 Checklist

> **版本**: 3.1
> **状态**: active
> **更新日期**: 2026-07-18

**关联计划**: [plan](./plan.md)

## Phase 1: 恢复方案 A 基线与真实回放

- [x] 1.1 对齐 Spec 1.5、plan/checklist/context、history 与 INDEX，负向确认当前 owner 不再授权方案 B；验证：Header/index check + `rg` 决策搜索
<!-- verified: 2026-07-17 context=PASS header-index=PASS links=PASS decision-negative=PASS -->
- [x] 1.2 先补失败断言，再把 generated index 调整为迁移期 additive、dirty-worktree-aware；验证：`python3 -m pytest scripts/harness_index_test.py -q`
<!-- verified: 2026-07-17 red=import-error green=11-passed inventory=retained-input exact-route=high generic-route=low -->
- [x] 1.3 建立从实际仓库操作采集的 R0-R3、歧义、误阻塞、失败恢复和回退基线；验证：`python3 -m pytest scripts/harness_replay_test.py -q`
<!-- verified: 2026-07-17 red=missing-module-and-semantic-false-green green=19-passed frozen-source-match=true ambiguity=true false-block=true high-confidence-misroute=true recovery-success-and-fail-closed=true tracked-and-untracked-rollback=true -->

## Phase 2: 建立唯一编排 owner 的等价影子

- [x] 2.1 先写 workflow contract 失败断言，再创建 `docs/agent-workflow.md` 的 R0-R3、必读 owner、handoff、失败恢复和退出合同；验证：`python3 -m pytest scripts/harness_workflow_test.py -q`
<!-- verified: 2026-07-17 red=7-missing-owner-failures green=7-passed links=PASS header-index=PASS independent-review=PASS -->
- [x] 2.2 定义统一 handoff schema，并执行旧入口与影子 workflow 的代表性等价回放；验证：workflow contract + replay fixtures
<!-- verified: 2026-07-17 red=5-contract-failures+self-proof-review-blockers green=37-passed oracle-sha-pinned workflow-tamper=confirmation+mutation+recovery+exit rejected recovery-order=checkpoint-before-authorization+isolated-restore live-worktree-unchanged independent-review=APPROVED -->

## Phase 3: 收敛 AGENTS 公共政策

- [x] 3.1 先写公共政策唯一 owner 与重复编排失败断言，再收敛 AGENTS 的稳定 policy IDs 和唯一 workflow 引用；验证：governance contract tests
<!-- verified: 2026-07-17 red=5-structure-consumption-failures+semantic-tamper-blocker green=7-passed policies=8 boundaries=8 canonical-contract-sha=d7886aa95b4d322f16fcc12600d69438a554e70722c4aba845c72b9ce88e8d8 tamper=AUTH+TEST+EVIDENCE+GIT-fail-closed independent-review=APPROVED -->
- [x] 3.2 删除 AGENTS 中完整 Skill 表、自动触发矩阵和逐步 runbook，验证 R0-R3 触发/确认/恢复语义不变；验证：workflow replay + negative search
<!-- verified: 2026-07-17 agents=26546-to-2686-bytes/285-to-17-lines workflow=active-single-owner harness=44-passed r0-r3=4/4-equivalent read-bytes-all-reduced links+fragments+diff=PASS independent-review=APPROVED -->

## Phase 4: 试点风险证据经济

- [x] 4.1 让 design、TDD、review 消费公共风险政策且不复制全文；验证：普通/重要/关键 owner-evidence fixtures
<!-- verified: 2026-07-17 consumers=design+tdd+plan-review+plan-code-review handoff-slice=POL-TEST-only workflow-schema-owner=table-driven-validator hidden-gate-NLP=absent -->
- [x] 4.2 证明普通风险允许零新增专用测试、重要/关键风险保留唯一主证据 owner；验证：duplicate-evidence negative tests + review replay
<!-- verified: 2026-07-17 harness=132-passed ordinary-zero-new-test=PASS important-critical-focused-owner=PASS duplicate+stale+gap+basis+risk-drift=fail-closed r0-r3=4/4-equivalent read-bytes-delta=-58256,-95060,-114790,-134068 independent-attacks=35/35 independent-reviews=2-APPROVED -->

## Phase 5: 试点能力寻源

- [x] 5.1 将寻源顺序和依赖准入写入 `docs/development.md` 唯一 owner；验证：owner-link contract
<!-- verified: 2026-07-17 owner=docs/development.md#3 matrix=6-rows spec+workflow+skills=reference-only local-owner=real-path+fragment+plan+Decision+ADR -->
- [x] 5.2 让入口、design、review 使用显式 build-vs-adopt handoff，覆盖成熟依赖、自研边界、许可证/安全/退出成本与简单逻辑 N/A；验证：capability fixtures
<!-- verified: 2026-07-17 handoff=4-required+2-conditional statuses=recorded+not_applicable+needs_sourcing focused=46-passed harness=178-passed attacks=15/15 r0-r3=4/4-equivalent independent-reviews=3-APPROVED -->

## Phase 6: 建立 Project Arch 与 `init-arch` 合同

- [ ] 6.1 定义 Project Arch v1 的 Docs Arch、test、scenario、env 固定角色、共同 SOP、扩展点和项目血肉边界；验证：Spec semantic review + owner matrix
- [ ] 6.2 定义 `init-docs` → `init-arch` 的 bundled Blueprint、`init/check/upgrade/repair`、版本、幂等、冲突和回滚合同；验证：fresh/upgrade coverage matrix
- [ ] 6.3 定义 Skill 的指导思想、Arch compatibility、可执行 SOP、证据结果和失败恢复结构门禁；验证：Skill contract matrix + L1 independent review

## Phase 7: 实现 `init-arch` 与四子系统 Blueprint

- [ ] 7.1 先写 fresh init、same-version no-op、legacy upgrade、custom/conflict、rollback 和业务血肉保留失败断言；验证：`python3 -m pytest scripts/harness_arch_test.py -q`
- [ ] 7.2 将模板整理到 `.agent-skills/init-arch/blueprint/`，由 `.agent-skills/init-arch/scripts/init_arch.py` 实现四种模式，暂保 `/init-docs` alias；验证：focused init-arch tests
- [ ] 7.3 建立 `docs/README.md` 的 `<!-- project-arch: v1 -->` 标记和目标 `scripts/harness_arch.py` check，在 fresh fixture 与当前仓库验证四子系统完整；验证：arch check + second-run zero diff
- [ ] 7.4 建立 `/environment-build` 与 `/environment-operate`，使用至少两个 Harness 自有异构环境 fixture 证明不同环境 Spec 可生成不同资产和 lifecycle adapter，且 `local-dev-stack` 只作 EasyInterview upgrade/regression 输入；验证：environment fixture matrix + lifecycle replay + golden-fixture negative search

## Phase 8: 收敛 14 个独立 Skill 与退出项

- [ ] 8.1 按 Spec 第 7 节把当前 20 个 Skill 全量映射为 keep/rename/merge/tool/remove，并为 14 个目标 Skill 声明 Arch 兼容版本、输入/输出和依赖的 canonical interfaces；验证：target/removed matrix parity + Skill contract matrix
- [ ] 8.2 保留或精炼每个 Skill 的指导思想、可执行 SOP、证据与恢复，移除 Skill-to-Skill 调用及 EasyInterview 业务实例；验证：semantic Skill tests + static coupling lint
- [ ] 8.3 统一语义化入口和唯一编排接口；除 `/init-docs` 限时 upgrade alias 外旧名称 hard cut，并把 `/work-journal` name-only 迁移为 `/delivery-commit`；验证：discovery replay + hard-cut zero-reference + normalized manual/auto/commit/journal/INDEX/ASCII parity
- [ ] 8.4 删除 `/frontend-design`、`/skill-creator`、`/agent-browser` 的实体、清单、自动触发、调用方和当前治理入口，不创建同义替代 Skill；验证：owner-aware zero-reference + replacement-wrapper negative search

## Phase 9: 文档工具化与索引归位

- [ ] 9.1 将 Header/INDEX 检查、修复和投影能力并入 Project Arch tooling；验证：docs transaction/index tests + double-run zero diff
- [ ] 9.2 将 `/create-doc` 与 `/sync-doc-index` 的文档事务、Header/INDEX check/fix/projection 能力并入 Project Arch tooling，迁移调用方后删除两个顶层 Skill wrapper；验证：docs transaction/index tests + double-run zero diff + zero-reference search

## Phase 10: fresh/upgrade 系统回放与旧入口退出

- [ ] 10.1 在 fresh 与 upgraded EasyInterview 上执行 R0-R3 单 Skill 和组合链路回放；验证：`make harness-replay`
- [ ] 10.2 执行 ambiguity、规则级联、上下文放大、`delivery-execute`→`delivery-commit`、scenario/env 污染、恢复和 rollback replay；验证：failure/recovery fixtures
- [ ] 10.3 满足退出条件后删除 aliases、重复编排、旧 contract tests 和无独立价值资产，保留历史 owner；验证：owner-aware negative search + Markdown links

## Phase 11: 全量验证与生命周期收口

- [ ] 11.1 完成 Spec A1-A25、Arch 正向接口、业务实例负向搜索和范围复查；验证：completion audit
- [ ] 11.2 运行 focused tests、`make test`、`make build`、`make lint`、`make docs-check`、链接/索引/fresh-upgrade replay 和 `git diff --check`；验证：全部 PASS
- [ ] 11.3 完成 retrospective、plan/checklist/index 生命周期和交付收口；验证：无未解释 checklist/gate/dirty scope
