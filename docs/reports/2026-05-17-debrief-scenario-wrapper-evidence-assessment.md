# Debrief Scenario Wrapper Evidence 交付复盘报告

> **日期**: 2026-05-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`plan-code-review --fix` 后续审计 `backend-debrief/001-debrief-record-and-analysis` 与关联 `backend-practice/004-derived-plans-debrief` 的 BDD / scenario wrapper 证据闭环。
- 成功证据：
  - `E2E.P0.060-064` 四段脚本串行通过，verify 现在要求 `=== RUN` / `--- PASS` / package `ok` 并拒绝 `FAIL` / `no tests to run`。
  - `E2E.P0.070-073` 新增四段脚本并串行通过。
  - `python3 -m pytest scripts/lint/scenario_script_contract_test.py scripts/lint/backend_practice_legacy_test.py scripts/lint/backend_debrief_legacy_test.py -q` 通过。
  - `python3 scripts/lint/backend_practice_legacy.py --repo-root . --phase all`、`python3 scripts/lint/backend_debrief_legacy.py --phase all`、`make docs-check`、`git diff --check` 通过。

## 2 会话中的主要阻点/痛点

- `backend-practice/004` 的 `Ready` 场景没有脚本资产。
  - **证据**：`p0-070` 到 `p0-073` 初始只有 README/data，没有 `scripts/`。
  - **影响**：BDD checklist 已 completed，但场景框架无法按 setup -> trigger -> verify -> cleanup 自动执行。
- 既有 `backend-debrief/001` verify 过弱。
  - **证据**：`p0-060` 到 `p0-064` 只检查测试名和 `PASS`，缺少 no-op/failure 反查与 package `ok` 断言。
  - **影响**：focused Go test 若匹配不到测试或产生失败文本，wrapper 证据不足以独立证明 BDD gate。
- 文档证据与当前 artifact 不完全一致。
  - **证据**：completed checklist 保留 `pending` 注释，004 bdd-plan 的 P0.073 验证入口写了旧测试名。
  - **影响**：review 需要额外反向校对，不能直接相信 completed 状态。

## 3 根因归类

- 场景 wrapper 缺失属于 **spec-plan + test framework gate** 问题：004 的 BDD checklist 把 Go test 当成完整场景证据。
- verify 过弱属于 **skill/test gate** 问题：之前只有 backend-practice/003 的 wrapper  contract test，未覆盖 debrief 新场景。
- stale `pending` 注释属于 **spec-plan evidence drift**：completed 文档缺少 pending/next-pass 负向反查。

## 4 对流程资产的改进建议

- 扩大 `scenario_script_contract_test.py` 的覆盖面。
  - **落点**：`scripts/lint/scenario_script_contract_test.py`
  - **优先级**：high
  - **状态**：本次已把 backend-debrief 060-064 与 backend-practice 070-073 纳入检查。
- 在后续新增 Ready 场景时，先检查脚本矩阵再更新 INDEX。
  - **落点**：scenario-create / plan-code-review 执行习惯
  - **优先级**：high
- 对 completed checklist 增加 `pending|next pass|asset readiness` 反查。
  - **落点**：plan-code-review L2 checklist
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一轮 debrief/frontend-debrief L2 review 继续按 wrapper-first 方式检查场景资产，避免只读 Go/Vitest body。
- 可延后：把 completed checklist 的 stale evidence 搜索做成通用 lint，目前本次已通过人工反查和 docs-check 收口。
